package main

import (
	"flag"

	"github.com/lhy1024/bench/bench"
	"github.com/pingcap/log"
	"go.uber.org/zap"
)

var (
	withBench    = flag.Bool("bench", true, "bench mode, it will bench this workload-scale-out")
	withGenerate = flag.Bool("generate", false, "generate mode,it will allow bench in empty database or only generate data")
	caseName     = flag.String("case", "", "case name, support list:scale-out, tpcc")
)

func main() {
	flag.Parse()
	cluster := bench.NewCluster()
	benchCases := bench.NewBenches(cluster)
	benchCase := benchCases.GetBench(*caseName)
	if benchCase == nil {
		log.Fatal("error with case name", zap.String("name", *caseName), zap.Strings("support list", benchCases.SupportList()))
		return
	}

	if *withGenerate {
		err := benchCase.Generate()
		if err != nil {
			log.Fatal("failed when generate data", zap.Error(err))
		}
		log.Info("generate data finish")
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
