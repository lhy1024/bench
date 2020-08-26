package main

import (
	"encoding/json"
	"fmt"
	"github.com/pingcap/errors"
	"net/http"
	"strings"
)

const (
	resourcePrefix = "api/cluster/resource/%v"
	scaleOutPrefix = "api/cluster/scale_out/%v/%v/%v"
	resultsPrefix  = "api/cluster/workload/%v/result"
)

type Spec struct {
	CPU  string
	Mem  string
	Disk string
}

// todo update after @junpeng
type ResourceRequestItem struct {
	ItemID uint   `gorm:"column:item_id;unique;not null"`
	Spec   Spec   `gorm:"column:spec;type:longtext;not null"`
	Status string `gorm:"column:status;type:varchar(255);not null"`
	RRID   uint   `gorm:"column:rr_id;not null"`
	RID    uint   `gorm:"column:r_id;not null"`
	// Components records which *_servers are serving on this machine
	Components string `gorm:"column:components"`
}

func (r *ResourceRequestItem) isAvailable(toScale string) bool {
	components := strings.Split(r.Components, "|")
	for _, component := range components {
		if component == toScale {
			return false
		}
	}
	return true
}

type WorkloadReport struct {
	CRID      uint    `json:"cr_id"`
	Data      string  `json:"data"`
	PlainText *string `json:"plaintext,omitempty"`
}

type cluster struct {
	name       string
	tidb       string
	pd         string
	prometheus string
	api        string
	client     *http.Client
}

func newCluster(name, tidb, pd, prometheus, api string) *cluster {
	return &cluster{
		name:       name,
		tidb:       tidb,
		pd:         pd,
		prometheus: prometheus,
		api:        api,
		client:     &http.Client{},
	}
}

func (c *cluster) joinUrl(prefix string) string {
	return c.api + "/" + prefix
}

func (c *cluster) getAvailableResourceID(component string) (uint, error) {
	// get all nodes
	prefix := fmt.Sprintf(resourcePrefix, c.name)
	url := c.joinUrl(prefix)
	resp, err := doRequest(url, http.MethodGet)
	if err != nil {
		return 0, err
	}
	resources := make([]ResourceRequestItem, 0, 0)
	err = json.Unmarshal([]byte(resp), &resources)
	if err != nil {
		return 0, err
	}

	// select available
	for _, resource := range resources {
		if resource.isAvailable(component) {
			return resource.ItemID, nil
		}
	}

	return 0, errors.New("no available resources")
}

func (c *cluster) scaleOut(component string, id uint) error {
	prefix := fmt.Sprintf(scaleOutPrefix, c.name, id, component)
	url := c.joinUrl(prefix)
	_, err := doRequest(url, http.MethodPost)
	return err
}

func (c *cluster) addStore() error {
	component := "tikv"
	id, err := c.getAvailableResourceID(component)
	if err != nil {
		return err
	}
	return c.scaleOut(component, id)
}

func (c *cluster) sendReport(data, plainText string) error {
	prefix := fmt.Sprintf(resultsPrefix, c.name)
	url := c.joinUrl(prefix)
	return postJSON(url, map[string]interface{}{
		"data":      data,
		"plaintext": plainText,
	})
}

func (c *cluster) getLastReportData() (string, error) {
	prefix := fmt.Sprintf(resultsPrefix, c.name)
	url := c.joinUrl(prefix)
	resp, err := doRequest(url, http.MethodGet)
	if err != nil {
		return "", err
	}

	reports := make([]WorkloadReport, 0, 0)
	err = json.Unmarshal([]byte(resp), &reports)
	if err != nil {
		return "", err
	}
	return reports[0].Data, nil
}
