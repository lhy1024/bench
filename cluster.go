package main

import "net/http"

const (
	nodes   = "/nodes"   // Get: get all node status; Post: scale in or scale out
	reports = "/reports" // Get: get last sendReport; Post: add new sendReport
	results = "/results" // Post: markdown string
)

type storageKind uint64

const (
	nvme storageKind = iota
	sharedNvme
)

type node struct {
	cpu     uint64 // example 40c
	mem     uint64 // example 160G
	disk    uint64 // example 1024G
	storage storageKind
}

type cluster struct {
	id         uint64
	tidb       string
	pd         string
	prometheus string
	api        string
	client     *http.Client
}

func newCluster(id *uint64, tidb, pd, prometheus, api *string) *cluster {
	return &cluster{
		id:         *id,
		tidb:       *tidb,
		pd:         *pd,
		prometheus: *prometheus,
		api:        *api,
		client:     &http.Client{},
	}
}

func (c *cluster) getCurrentTiKV() node {
	// todo
	return node{}
}

func (c *cluster) getAvailableNodes(n node) []node {
	// todo
	return []node{}
}

func (c *cluster) addStore(n node) error {
	// todo
	return nil
}

func (c *cluster) addStores(num uint64) error {
	tikv := c.getCurrentTiKV()
	nodes := c.getAvailableNodes(tikv)
	if len(nodes) != 0 {
		err := c.addStore(tikv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *cluster) sendReport(report string) error {
	// todo
	return nil
}

func (c *cluster) sendResult(markdown string) error {
	// todo
	return nil
}

func (c *cluster) getLastReport() (string, error) {
	// todo
	return "", nil
}
