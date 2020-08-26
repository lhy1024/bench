package main

import (
	"os"
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

	// bench
	bench := newScaleOut(cluster)
	err = bench.run()

	// sendReport
	err = bench.collect()
}
