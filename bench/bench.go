package bench

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/lhy1024/bench/utils"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"go.uber.org/zap"
)

type bench interface {
	Run() error
	Collect() error
}

func createScaleOutCase(cluster *cluster) *benchCase {
	return &benchCase{
		generator: newYCSB(cluster, "workload-scale-out"),
		bench:     newScaleOut(cluster),
	}
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
	PrevApplyLog           float64 `json:"prevApplyLog"`
	CurApplyLog            float64 `json:"curApplyLog"`
	PrevDbMutex            float64 `json:"prevDbMutex"`
	CurDbMutex             float64 `json:"curDbMutex"`
}

const (
	typeInt = iota
	typeFloat64
)

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

func (s *scaleOut) Run() error {
	preStoreNum := s.c.getStoreNum()
	for i := 0; i < s.num; i++ {
		if err := s.c.AddStore(); err != nil {
			return err
		}
	}
	s.waitScaleOut(preStoreNum)
	s.t.addTime = time.Now()
	for {
		bal, err := s.isBalance()
		if err != nil {
			return err
		}
		if bal {
			return nil
		}
		time.Sleep(time.Second)
	}
}

func (s *scaleOut) waitScaleOut(preStoreNum int) {
	for {
		time.Sleep(time.Second)
		if s.num+preStoreNum == s.c.getStoreNum() {
			return
		}
	}
}

func (s *scaleOut) isBalance() (bool, error) {
	r := v1.Range{
		Start: time.Now().Add(-9 * time.Minute),
		End:   time.Now(),
		Step:  time.Minute,
	}
	matrix, err := s.c.getMatrixMetric("pd_scheduler_store_status{type=\"region_score\"}", r)
	if err != nil {
		return false, err
	}
	// if low deviation in a series of scores applies to all stores, then it is balanced.
	for _, scores := range matrix {
		if len(scores) != 10 {
			return false, nil
		}
		mean := 0.0
		dev := 0.0
		for _, score := range scores {
			mean += score / 10
		}
		for _, score := range scores {
			dev += (score - mean) * (score - mean) / 10
		}
		if mean*mean*0.02 < dev {
			return false, nil
		}
	}
	log.Info("balanced")
	s.t.balanceTime = time.Now()
	return true, nil
}

func (s *scaleOut) Collect() error {
	// create report data
	data, err := s.createReport()
	if err != nil {
		return err
	}

	// try get last data
	lastReport, err := s.c.GetLastReport()
	if err != nil {
		return err
	}

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

	return s.c.SendReport(data, plainText)
}

func (s *scaleOut) queryPrevCur(query string, prevArg, curArg interface{}, typ int) error {
	prevValue, err := s.c.getMetric(query, s.t.addTime)
	if err != nil {
		return err
	}
	curValue, err := s.c.getMetric(query, s.t.balanceTime)
	if err != nil {
		return err
	}
	if typ == typeInt {
		*(prevArg.(*int)) = int(prevValue)
		*(curArg.(*int)) = int(curValue)
	} else if typ == typeFloat64 {
		*(prevArg.(*float64)) = prevValue
		*(curArg.(*float64)) = curValue
	} else {
		return errors.New("Type not supported")
	}
	return nil
}

func (s *scaleOut) createReport() (string, error) {
	rep := &stats{BalanceInterval: int(s.t.balanceTime.Sub(s.t.addTime).Seconds())}
	err := s.queryPrevCur("sum(tidb_server_handle_query_duration_seconds_sum{sql_type!=\"internal\"})"+
		" / (sum(tidb_server_handle_query_duration_seconds_count{sql_type!=\"internal\"}) + 1)",
		&rep.PrevLatency, &rep.CurLatency, typeFloat64)
	if err != nil {
		return "", err
	}

	err = s.queryPrevCur("pd_scheduler_event_count{type=\"balance-leader-scheduler\", name=\"schedule\"}",
		&rep.PrevBalanceLeaderCount, &rep.CurBalanceLeaderCount, typeInt)
	if err != nil {
		return "", err
	}

	err = s.queryPrevCur("pd_scheduler_event_count{type=\"balance-region-scheduler\", name=\"schedule\"}",
		&rep.PrevBalanceRegionCount, &rep.CurBalanceRegionCount, typeInt)
	if err != nil {
		return "", err
	}

	err = s.queryPrevCur("sum(tikv_engine_compaction_flow_bytes)", &rep.PrevCompactionRate, &rep.CurCompactionRate, typeFloat64)
	if err != nil {
		return "", err
	}

	err = s.queryPrevCur("sum(tikv_raftstore_apply_log_duration_seconds_sum) / (sum(tikv_raftstore_apply_log_duration_seconds_count) + 1)",
		&rep.PrevApplyLog, &rep.CurApplyLog, typeFloat64)
	if err != nil {
		return "", err
	}

	err = s.queryPrevCur("sum(tikv_raftstore_apply_perf_context_time_duration_secs_sum{type=\"db_mutex_lock_nanos\"}) / "+
		"(sum(tikv_raftstore_apply_perf_context_time_duration_secs_count{type=\"db_mutex_lock_nanos\"}) + 1)",
		&rep.PrevDbMutex, &rep.CurDbMutex, typeFloat64)
	if err != nil {
		return "", err
	}

	bytes, err := json.Marshal(rep)
	if err != nil {
		log.Error("marshal error", zap.Error(err))
	}

	return string(bytes), err
}

func reportLine(head string, last float64, cur float64) string {
	headPart := "\t* " + head + ": "
	curPart := fmt.Sprintf("%.8f ", cur)
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
	title := "```diff  \n@@\t\t\tBenchmark diff\t\t\t@@\n"
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
	plainText += scheduleTag + reportLine("balance_leader_operator_count",
		float64(last.CurBalanceLeaderCount-last.PrevBalanceLeaderCount), float64(cur.CurBalanceLeaderCount-cur.PrevBalanceLeaderCount))
	plainText += reportLine("balance_region_operator_count",
		float64(last.CurBalanceRegionCount-last.PrevBalanceRegionCount), float64(cur.CurBalanceRegionCount-cur.PrevBalanceRegionCount))
	plainText += compactionTag + reportLine("compaction_flow_bytes",
		last.CurCompactionRate-last.PrevCompactionRate, cur.CurCompactionRate-cur.PrevCompactionRate)
	plainText += latencyTag + reportLine("prev_query_latency", last.PrevLatency, cur.PrevLatency)
	plainText += reportLine("cur_query_latency", last.CurLatency, cur.CurLatency)
	plainText += reportLine("prev_apply_log_latency", last.PrevApplyLog, cur.PrevApplyLog)
	plainText += reportLine("cur_apply_log_latency", last.CurApplyLog, cur.CurApplyLog)
	plainText += reportLine("prev_db_mutex_latency", last.PrevDbMutex, cur.PrevDbMutex)
	plainText += reportLine("cur_db_mutex_latency", last.CurDbMutex, cur.CurDbMutex)
	plainText += "```  \n"
	return
}

type simulatorBench struct {
	simPath string
	c       *cluster
	report  string
}

func (s *simulatorBench) Run() error {
	cmd := utils.NewCommand(s.simPath, s.c.pdAddr)
	limit := os.Getenv("STORE_LIMIT")
	if limit == "" {
		limit = "2000"
	}
	ctl := utils.NewCommand("/bin/pd-ctl", "store", "limit", "all", limit)
	go func() {
		time.Sleep(3 * time.Second)
		_, err := ctl.Run()
		if err != nil {
			log.Error("pd-ctl", zap.Error(err))
		}
	}()
	out, err := cmd.Run()
	if err != nil {
		return err
	}
	s.report = out
	return nil
}

func (s *simulatorBench) Collect() error {
	lastReport, err := s.c.GetLastReport()
	if err != nil {
		return err
	}

	var plainText string
	var data string
	if lastReport == nil { //first send
		plainText = ""
		data = createSimReport("last", s.report)
	} else { //second send
		data = createSimReport("cur", s.report)
		plainText = "```diff  \n"
		plainText += lastReport.Data
		plainText += data
		plainText += "```  \n"
		log.Info("Concat report success", zap.String("concat result", plainText))
	}
	return s.c.SendReport(data, plainText)
}

func newSimulator(cluster *cluster, simCase string) bench {
	path := "/scripts/simulator/" + simCase
	return &simulatorBench{simPath: path, c: cluster}
}

func createSimulatorCase(cluster *cluster, simCase string) *benchCase {
	return &benchCase{
		generator: newEmptyGenerator(),
		bench:     newSimulator(cluster, simCase),
	}
}

func createSimReport(head, report string) string {
	plainText := head + ":  \n"
	plainText += "\t*artifacts link: " + os.Getenv("ARTIFACT_URL") + "/workload.tar.gz   \n"
	plainText += "\t*simulator report:  \n"
	plainText += report + "  \n"

	return plainText
}
