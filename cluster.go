package main

import "net/http"

const (
	resourcePrefix = "api/cluster/resource/%v"
	scaleOutPrefix = "api/cluster/scale_out/%v/%v/%v"
	resultsPrefix  = "api/cluster/workload/%v/result"
)

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

func (c *cluster) getAvailableResourceID(component string) (uint64, error) {
	// get all nodes

	// select available
	return 0, nil
}

func (c *cluster) scaleOut(component string, id uint64) error {
	// todo
	return nil
}

func (c *cluster) addStore() error {
	id, err := c.getAvailableResourceID("tikv")
	if err != nil {
		return err
	}
	return c.scaleOut("tikv", id)
}

func (c *cluster) sendReport(report, plainText string) error {
	// todo
	return nil
}

func (c *cluster) getLastReport() (string, error) {
	// todo
	return "", nil
}
