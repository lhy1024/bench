package bench

type benchCase struct {
	generator
	bench
}

type benchCases struct {
	cases map[string]*benchCase
}

// NewBenches return bench cases
func NewBenches(cluster *cluster) *benchCases {
	caseMap := make(map[string]*benchCase)
	caseMap["scale-out"] = createScaleOutCase(cluster)
	caseMap["sim-import"] = createSimulatorCase(cluster, "import")
	return &benchCases{
		cases: caseMap,
	}
}

// GetBench return bench with name
func (c *benchCases) GetBench(name string) *benchCase {
	if f, ok := c.cases[name]; ok {
		return f
	}
	return nil
}

// SupportList return all support bench cases
func (c *benchCases) SupportList() []string {
	var ret []string
	for name := range c.cases {
		ret = append(ret, name)
	}
	return ret
}
