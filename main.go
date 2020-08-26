package main

import (
	"os"

	"github.com/pingcap/log"
	"go.uber.org/zap"
)

func main() {
	var clusterName = os.Getenv("CLUSTER_NAME")
	var tidbServer = os.Getenv("TIDB_ADDR")
	var pdServer = os.Getenv("PD_ADDR")
	var prometheusServer = os.Getenv("PROM_ADDR")
	var apiServer = os.Getenv("API_SERVER")
	cluster := newCluster(clusterName, tidbServer, pdServer, prometheusServer, apiServer)

	// load data
	loader := newBr(cluster)
	err := loader.load()
	if err != nil {
		log.Fatal("failed when load", zap.Error(err))
	}

	// bench
	bench := newScaleOut(cluster)
	err = bench.run()
	if err != nil {
		log.Fatal("failed when bench", zap.Error(err))
	}

	// sendReport
	err = bench.collect()
	if err != nil {
		log.Fatal("failed when collect report", zap.Error(err))
	}
}
