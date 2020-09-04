package main

import (
	"flag"
	"os"

	"github.com/pingcap/log"
	"go.uber.org/zap"
)

var (
	withExport = flag.Bool("export", false, "export mode, bench will creat data and export by br")
	withBench  = flag.Bool("bench", true, "bench mode, it will bench this workload-scale-out")
	withImport = flag.Bool("import", false, "import mode, it can be used with bench mode")
	caseName   = flag.String("case", "", "case name, support list:scale-out,tpcc")
)

func main() {
	flag.Parse()
	// todo http head and format valid check
	var clusterName = os.Getenv("CLUSTER_NAME")
	var tidbServer = os.Getenv("TIDB_ADDR")
	var pdServer = os.Getenv("PD_ADDR")
	var prometheusServer = os.Getenv("PROM_ADDR")
	var apiServer = os.Getenv("API_SERVER")
	// todo @zeyuan may additional parameters with export

	cluster := NewCluster(clusterName, tidbServer, pdServer, prometheusServer, apiServer)
	benchCases := NewBenches(cluster)
	benchCase := benchCases.GetBench(*caseName)
	if benchCase == nil {
		log.Fatal("error with case name", zap.String("name", *caseName), zap.Strings("support list", benchCases.SupportList()))
		return
	}

	if *withExport {
		err := benchCase.Import()
		if err != nil {
			log.Fatal("failed when load data", zap.Error(err))
		}
		log.Info("load finish")
		err = benchCase.Export()
		if err != nil {
			log.Fatal("failed when backup data", zap.Error(err))
		}
		log.Info("backup finish")
		return
	}

	log.Info("run in bench mode")
	if *withImport {
		err := benchCase.Import()
		if err != nil {
			log.Fatal("failed when import", zap.Error(err))
		}
		log.Info("import finish")
	}

	if *withBench {
		err := benchCase.Run()
		if err != nil {
			log.Fatal("failed when bench", zap.Error(err))
		}
		err = benchCase.Collect()
		if err != nil {
			log.Fatal("failed when collect report", zap.Error(err))
		}
		log.Info("bench finish")
	}
}
