package test

import (
	"testing"
	"time"

	"github.com/lhy1024/bench/bench"
	. "github.com/pingcap/check"
)

func Test(t *testing.T) {
	TestingT(t)
}

type testClusterSuite struct{}

var _ = Suite(&testClusterSuite{})

func (s *testClusterSuite) SetUpSuite(c *C) {
	go mockServer()
	time.Sleep(1 * time.Second)
}

func (s *testClusterSuite) TestScaleOut(c *C) {
	cluster := bench.NewCluster()
	cluster.SetApiServer("http://" + mockServerAddr)
	cluster.SetID("1")
	cluster.SetName("test")

	err := cluster.AddStore()
	c.Assert(err, IsNil)
}

func (s *testClusterSuite) TestReport(c *C) {
	cluster := bench.NewCluster()
	cluster.SetApiServer("http://" + mockServerAddr)
	cluster.SetID("1")
	cluster.SetName("test")

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
