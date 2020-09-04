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
	cluster := NewCluster("test", "", "", "", "http://172.16.5.110:8000")
	err := cluster.AddStore()
	c.Assert(err, IsNil)
}

func (s *testClusterSuite) TestReport(c *C) {
	cluster := NewCluster("test", "", "", "", "http://172.16.5.110:8000")
	lastReport, err := cluster.GetLastReport()
	c.Assert(err, IsNil)
	c.Assert(lastReport, IsNil)
	report := "report"
	err = cluster.SendReport(report, "")
	c.Assert(err, IsNil)
	lastReport, err = cluster.GetLastReport()
	c.Assert(err, IsNil)
	c.Assert(lastReport, NotNil)
	c.Assert(lastReport.Data, Equals, report)
}
