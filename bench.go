package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pingcap/log"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

type bench interface {
	run() error
	collect() error
}

type timePoint struct {
	addTime     time.Time
	balanceTime time.Time
}

type stats struct {
	BalanceInterval        int     `json:"balanceInterval"`
	PrevBalanceLeaderCount int     `json:"prevBalanceLeaderCount"`
	PrevBalanceRegionCount int     `json:"prevBalanceRegionCount"`
	CurBalanceLeaderCount  int     `json:"curBalanceLeaderCount"`
	CurBalanceRegionCount  int     `json:"curBalanceRegionCount"`
	PrevLatency            float64 `json:"prevLatency"`
	CurLatency             float64 `json:"curLatency"`
	PrevCompactionRate     float64 `json:"prevCompactionRate"`
	CurCompactionRate      float64 `json:"curCompactionRate"`
}

type scaleOut struct {
	c   *cluster
	t   timePoint
	num int //scale out num
}

func newScaleOut(c *cluster) bench {
	num, err := strconv.Atoi(os.Getenv("SCALE_NUM"))
	if err != nil {
		num = 1 // default
	}
	return &scaleOut{
		c:   c,
		num: num,
	}
}

func (s *scaleOut) run() error {
	for i := 0; i < s.num; i++ {
		if err := s.c.addStore(); err != nil {
			return err
		}
	}
	s.t.addTime = time.Now()
	for {
		time.Sleep(time.Minute)
		bal, err := s.isBalance()
		if err != nil {
			return err
		}
		if bal {
			return nil
		}
	}
}

func (s *scaleOut) isBalance() (bool, error) {
	client, err := api.NewClient(api.Config{
		Address: s.c.prometheus,
	})
	if err != nil {
		log.Error("error creating client", zap.Error(err))
		return false, nil
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r := v1.Range{
		Start: time.Now().Add(-9 * time.Minute),
		End:   time.Now(),
		Step:  time.Minute,
	}
	result, warnings, err := v1api.QueryRange(ctx, "pd_scheduler_store_status{type=\"region_score\"}", r)
	if err != nil {
		return false, err
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	matrix := result.(model.Matrix)
	for _, data := range matrix {
		if len(data.Values) != 10 {
			return false, nil
		}
		mean := 0.0
		dev := 0.0
		for _, v := range data.Values {
			mean += float64(v.Value) / 10
		}
		for _, v := range data.Values {
			dev += (float64(v.Value) - mean) * (float64(v.Value) - mean) / 10
		}
		if mean*mean*0.1 < dev {
			return false, nil
		}
	}
	log.Info("balanced")
	s.t.balanceTime = time.Now()
	return true, nil
}

func (s *scaleOut) collect() error {
	// create report data
	data, err := s.createReport()
	if err != nil {
		return err
	}

	// try get last data
	lastReport, err := s.c.getLastReport()
	if err != nil {
		return err
	}
	fmt.Println(s.mergeReport(data, data))
	// send data
	var plainText string
	if lastReport == nil { //first send
		plainText = ""
	} else { //second send
		plainText, err = s.mergeReport(lastReport.Data, data)
		log.Info("Merge report success", zap.String("merge result", plainText))
		if err != nil {
			return err
		}
	}

	return s.c.sendReport(data, plainText)
}

func (s *scaleOut) createReport() (string, error) {
	rep := &stats{BalanceInterval: int(s.t.balanceTime.Sub(s.t.addTime).Seconds())}
	client, err := api.NewClient(api.Config{
		Address: s.c.prometheus,
	})
	if err != nil {
		log.Error("error creating client", zap.Error(err))
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx,
		"sum(tidb_server_handle_query_duration_seconds_sum{sql_type!=\"internal\"})"+
			" / sum(tidb_server_handle_query_duration_seconds_count{sql_type!=\"internal\"})", s.t.addTime)
	if err != nil {
		log.Error("error querying Prometheus", zap.Error(err))
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	vector := result.(model.Vector)
	if len(vector) >= 1 {
		rep.PrevLatency = float64(vector[0].Value)
	}

	result, warnings, err = v1api.Query(ctx,
		"sum(tidb_server_handle_query_duration_seconds_sum{sql_type!=\"internal\"})"+
			" / sum(tidb_server_handle_query_duration_seconds_count{sql_type!=\"internal\"})", s.t.balanceTime)
	if err != nil {
		log.Error("error querying Prometheus", zap.Error(err))
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	vector = result.(model.Vector)
	if len(vector) >= 1 {
		rep.CurLatency = float64(vector[0].Value)
	}

	result, warnings, err = v1api.Query(ctx,
		"pd_scheduler_event_count{type=\"balance-leader-scheduler\", name=\"schedule\"}", s.t.balanceTime)
	if err != nil {
		log.Error("error querying Prometheus", zap.Error(err))
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	vector = result.(model.Vector)
	if len(vector) >= 1 {
		rep.CurBalanceLeaderCount = int(vector[0].Value)
	}

	result, warnings, err = v1api.Query(ctx,
		"pd_scheduler_event_count{type=\"balance-region-scheduler\", name=\"schedule\"}", s.t.balanceTime)
	if err != nil {
		log.Error("error querying Prometheus", zap.Error(err))
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	vector = result.(model.Vector)
	if len(vector) >= 1 {
		rep.CurBalanceRegionCount = int(vector[0].Value)
	}

	result, warnings, err = v1api.Query(ctx,
		"pd_scheduler_event_count{type=\"balance-leader-scheduler\", name=\"schedule\"}", s.t.addTime)
	if err != nil {
		log.Error("error querying Prometheus", zap.Error(err))
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	vector = result.(model.Vector)
	if len(vector) >= 1 {
		rep.PrevBalanceLeaderCount = int(vector[0].Value)
	}

	result, warnings, err = v1api.Query(ctx,
		"pd_scheduler_event_count{type=\"balance-region-scheduler\", name=\"schedule\"}", s.t.addTime)
	if err != nil {
		log.Error("error querying Prometheus", zap.Error(err))
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	vector = result.(model.Vector)
	if len(vector) >= 1 {
		rep.PrevBalanceRegionCount = int(vector[0].Value)
	}
	bytes, err := json.Marshal(rep)
	if err != nil {
		log.Error("marshal error", zap.Error(err))
	}

	result, warnings, err = v1api.Query(ctx,
		"sum(tikv_engine_compaction_flow_bytes)", s.t.balanceTime)
	if err != nil {
		log.Error("error querying Prometheus", zap.Error(err))
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	vector = result.(model.Vector)
	if len(vector) >= 1 {
		rep.PrevCompactionRate = float64(vector[0].Value)
	}

	result, warnings, err = v1api.Query(ctx,
		"sum(tikv_engine_compaction_flow_bytes)", time.Now())
	if err != nil {
		log.Error("error querying Prometheus", zap.Error(err))
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	vector = result.(model.Vector)
	if len(vector) >= 1 {
		rep.CurCompactionRate = float64(vector[0].Value)
	}

	return string(bytes), err
}

func reportLine(head string, last float64, cur float64) string {
	headPart := "\t* " + head + ": "
	curPart := fmt.Sprintf("%.2f ", cur)
	deltaPart := fmt.Sprintf("delta: %.2f%%  \n", (cur-last)*100/(last+1))
	return headPart + curPart + deltaPart
}

// lastReport is
func (s *scaleOut) mergeReport(lastReport, report string) (plainText string, err error) {
	last := &stats{}
	cur := &stats{}
	err = json.Unmarshal([]byte(lastReport), last)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(report), cur)
	if err != nil {
		return
	}
	title := "@@\t\t\tBenchmark diff\t\t\t@@\n"
	splitLine := ""
	for i := 0; i < 58; i++ {
		splitLine += "="
	}
	splitLine += "\n"
	plainText += title
	plainText += splitLine
	balanceTag := "balance:  \n"
	scheduleTag := "schedule:  \n"
	compactionTag := "compaction:  \n"
	latencyTag := "latency:  \n"
	plainText += balanceTag + reportLine("balance_time", float64(last.BalanceInterval), float64(cur.BalanceInterval))
	plainText += scheduleTag + reportLine("balance_leader",
		float64(last.CurBalanceLeaderCount-last.PrevBalanceLeaderCount), float64(cur.CurBalanceLeaderCount-cur.PrevBalanceLeaderCount))
	plainText += reportLine("balance_leader",
		float64(last.CurBalanceRegionCount-last.PrevBalanceRegionCount), float64(cur.CurBalanceRegionCount-cur.PrevBalanceRegionCount))
	plainText += compactionTag + reportLine("compaction_flow",
		last.CurCompactionRate-last.PrevCompactionRate, cur.CurCompactionRate-cur.PrevCompactionRate)
	plainText += latencyTag + reportLine("prev_query_latency", last.PrevLatency, cur.PrevLatency)
	plainText += reportLine("cur_query_latency", last.CurLatency, cur.CurLatency)
	return
}
