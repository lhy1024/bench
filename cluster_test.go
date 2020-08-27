package main

import (
	"testing"

	. "github.com/pingcap/check"
)

func Test(t *testing.T) {
	TestingT(t)
}

type testClusterSuite struct{}

var _ = Suite(&testClusterSuite{})

func (s *testClusterSuite) TestScaleOut(c *C) {
	cluster := newCluster("test", "", "", "", "http://172.16.5.110:8000")
	err := cluster.addStore()
	c.Assert(err, IsNil)
}

func (s *testClusterSuite) TestReport(c *C) {
	cluster := newCluster("test", "", "", "", "http://172.16.5.110:8000")
	lastReport, err := cluster.getLastReport()
	c.Assert(err, IsNil)
	c.Assert(lastReport, IsNil)
	report := "report"
	err = cluster.sendReport(report, "")
	c.Assert(err, IsNil)
	lastReport, err = cluster.getLastReport()
	c.Assert(err, IsNil)
	c.Assert(lastReport, NotNil)
	c.Assert(lastReport.Data, Equals, report)
}
