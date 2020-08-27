package main

import (
	"os"

	"github.com/pingcap/log"
	"go.uber.org/zap"
)

var benchCases = map[string]func(*cluster) bench{
	"scaleOut": newScaleOut,
}

func newBench(name string, c *cluster) bench {
	if f, ok := benchCases[name]; ok {
		return f(c)
	}
	return nil
}

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
	log.Info("finish load")

	// bench
	bench := newBench("scaleOut", cluster)
	if bench == nil {
		log.Fatal("error bench name", zap.Error(err))
		return
	}
	err = bench.run()
	if err != nil {
		log.Fatal("failed when bench", zap.Error(err))
	}
	log.Info("finish bench")

	// sendReport
	err = bench.collect()
	if err != nil {
		log.Fatal("failed when collect report", zap.Error(err))
	}
	log.Info("finish report")
}
