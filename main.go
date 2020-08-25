package main

import (
	"flag"
)

var testID = flag.Uint64("id", 0, "bench suit id")
var tidbServer = flag.String("tidb", "", "tidb server")
var pdServer = flag.String("pd", "", "pd server")
var prometheusServer = flag.String("prom", "", "prometheus server")
var apiServer = flag.String("api", "", "api server")

func main() {
	flag.Parse()
	cluster := newCluster(testID, tidbServer, pdServer, prometheusServer, apiServer)

	// load data
	loader := newBr(cluster)
	err := loader.load()
	if err != nil {
		cluster.reportErr(err)
		return
	}

	// bench
	bench := newScaleOut(cluster)
	err = bench.run()
	if err != nil {
		cluster.reportErr(err)
		return
	}

	// report
	err = bench.collect()
	if err != nil {
		cluster.reportErr(err)
		return
	}
}
