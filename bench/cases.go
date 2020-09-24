package bench

type Case struct {
	Generator
	Bench
}

type Benches struct {
	cases map[string]*Case
}

func NewBenches(cluster *Cluster) *Benches {
	caseMap := make(map[string]*Case)
	caseMap["scale-out"] = CreateScaleOutCase(cluster)
	caseMap["sim-import"] = CreateSimulatorCase(cluster, "import")
	return &Benches{
		cases: caseMap,
	}
}

func (c *Benches) GetBench(name string) *Case {
	if f, ok := c.cases[name]; ok {
		return f
	}
	return nil
}

func (c *Benches) SupportList() []string {
	var ret []string
	for name := range c.cases {
		ret = append(ret, name)
	}
	return ret
}
