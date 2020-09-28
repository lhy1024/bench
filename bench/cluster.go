package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

const (
	resourcePrefix = "api/cluster/resource/%v"
	scaleOutPrefix = "api/cluster/scale_out/%v/%v/%v"
	resultsPrefix  = "api/cluster/workload/%v/result"
)

// ResourceRequestItem ...
type ResourceRequestItem struct {
	gorm.Model
	ItemID       uint   `gorm:"column:item_id;unique;not null" json:"item_id"`
	InstanceType string `gorm:"column:instance_type;type:varchar(100);not null" json:"instance_type"`
	RRID         uint   `gorm:"column:rr_id;not null" json:"rr_id"`
	RID          uint   `gorm:"column:r_id" json:"r_id"`
	// Components records which *_servers are serving on this machine
	Components string `gorm:"column:components" json:"components"`
}

func (r *ResourceRequestItem) hasNum(style string) (num int) {
	components := strings.Split(r.Components, "|")
	for _, component := range components {
		if component == style {
			num++
		}
	}
	return num
}

// WorkloadReport ...
type WorkloadReport struct {
	gorm.Model
	CRID      uint    `gorm:"column:cr_id;not null" json:"cr_id"`
	Data      string  `gorm:"column:result;type:longtext;not null" json:"data"`
	PlainText *string `gorm:"column:plaintext" json:"plaintext,omitempty"`
}

type cluster struct {
	id             string
	name           string
	tidbAddr       string
	pdAddr         string
	prometheusAddr string
	apiAddr        string
	client         *http.Client
}

// NewCluster return cluster
func NewCluster() *cluster {
	return &cluster{
		id:             os.Getenv("CLUSTER_ID"),
		name:           os.Getenv("CLUSTER_NAME"),
		tidbAddr:       os.Getenv("TIDB_ADDR"),
		pdAddr:         os.Getenv("PD_ADDR"),
		prometheusAddr: os.Getenv("PROM_ADDR"),
		apiAddr:        os.Getenv("API_SERVER"),
		client:         &http.Client{},
	}
}

// SetAPIServer is used to set config.
func (c *cluster) SetAPIServer(apiAddr string) {
	c.apiAddr = apiAddr
}

// SetID is used to set config.
func (c *cluster) SetID(id string) {
	c.id = id
}

// SetName is used to set config.
func (c *cluster) SetName(name string) {
	c.name = name
}

func (c *cluster) joinURL(prefix string) string {
	return c.apiAddr + "/" + prefix
}

func (c *cluster) getAllResource() ([]ResourceRequestItem, error) {
	prefix := fmt.Sprintf(resourcePrefix, c.id)
	url := c.joinURL(prefix)
	resp, err := doRequest(url, http.MethodGet)
	if err != nil {
		return nil, err
	}
	resources := make([]ResourceRequestItem, 0)
	err = json.Unmarshal([]byte(resp), &resources)
	return resources, err
}

func (c *cluster) getAvailableResourceID(component string) (uint, error) {
	resources, err := c.getAllResource()
	if err != nil {
		return 0, errors.New("failed to get all resource")
	}
	// select available
	for _, resource := range resources {
		if resource.hasNum(component) == 0 {
			return resource.ID, nil
		}
	}
	return 0, errors.New("no available resources")
}

func (c *cluster) getStoreNum() (num int) {
	resources, err := c.getAllResource()
	if err != nil {
		return 0
	}
	for _, resource := range resources {
		num += resource.hasNum("tikv")
	}
	return num
}

func (c *cluster) scaleOut(component string, id uint) error {
	prefix := fmt.Sprintf(scaleOutPrefix, c.id, id, component)
	url := c.joinURL(prefix)
	_, err := doRequest(url, http.MethodPost)
	return err
}

// AddStore is used to add store.
func (c *cluster) AddStore() error {
	component := "tikv"
	id, err := c.getAvailableResourceID(component)
	if err != nil {
		return err
	}
	return c.scaleOut(component, id)
}

// SendReport is used to send report.
func (c *cluster) SendReport(data, plainText string) error {
	prefix := fmt.Sprintf(resultsPrefix, c.id)
	url := c.joinURL(prefix)
	return postJSON(url, map[string]interface{}{
		"data":      data,
		"plaintext": plainText,
	})
}

// GetLastReport is used to get the last report.
func (c *cluster) GetLastReport() (*WorkloadReport, error) {
	prefix := fmt.Sprintf(resultsPrefix, c.id)
	url := c.joinURL(prefix)
	resp, err := doRequest(url, http.MethodGet)
	if err != nil {
		return nil, err
	}

	reports := make([]WorkloadReport, 0)
	err = json.Unmarshal([]byte(resp), &reports)
	if err != nil || len(reports) == 0 {
		return nil, err
	}
	return &reports[0], nil
}

func (c *cluster) getMetric(query string, t time.Time) (float64, error) {
	client, err := api.NewClient(api.Config{
		Address: c.prometheusAddr,
	})
	if err != nil {
		log.Error("error creating client", zap.Error(err))
		return 0, err
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, query, t)
	if err != nil {
		log.Error("error querying Prometheus", zap.Error(err))
		return 0, err
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	vector := result.(model.Vector)
	if len(vector) >= 1 {
		return float64(vector[0].Value), nil
	}
	return 0, nil
}

func (c *cluster) getMatrixMetric(query string, r v1.Range) ([][]float64, error) {
	client, err := api.NewClient(api.Config{
		Address: c.prometheusAddr,
	})
	if err != nil {
		log.Error("error creating client", zap.Error(err))
		return nil, err
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.QueryRange(ctx, query, r)
	if err != nil {
		log.Error("error querying Prometheus", zap.Error(err))
		return nil, err
	}
	if len(warnings) > 0 {
		log.Warn("query has warnings")
	}
	matrix := result.(model.Matrix)
	var ret [][]float64
	for _, m := range matrix {
		var r []float64
		for _, v := range m.Values {
			r = append(r, float64(v.Value))
		}
		ret = append(ret, r)
	}
	return ret, nil
}
