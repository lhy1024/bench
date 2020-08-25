package main

import "net/http"

const (
	stores     = "/stores"       // post request
	testStatus = "/tests/status" // post status kind
	report     = "/report"       // post report
	err        = "/error"        // post err msg
)

type statusKind uint64

const (
	startLoad statusKind = iota
	startBench
	finishBench
)

type storageKind uint64

const (
	nvme storageKind = iota
	sharedNvme
)

type request struct {
	id      uint64
	cpu     uint64 // example 40c
	mem     uint64 // example 160G
	disk    uint64 // example 1024G
	storage storageKind
}

type cluster struct {
	id         uint64 // make convenient api server to distinguish different benches
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

func (c *cluster) addStore(req *request) error {
	// todo
	return nil
}

func (c *cluster) changeStatus(s statusKind) {

}

func (c *cluster) reportErr(err error) {

}

func (c *cluster) report() {

}
