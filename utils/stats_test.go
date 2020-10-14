package utils

import (
	"encoding/json"
	"strings"
	"testing"

	. "github.com/pingcap/check"
)

func Test(t *testing.T) {
	TestingT(t)
}

type testStatsSuite struct{}

var _ = Suite(&testStatsSuite{})

func (s *testStatsSuite) TestScaleOutStats(c *C) {
	prev := ScaleOutOnce{10, 11, 12, 13,
		12, 11, 10, 9, 8, 7,
		6, 5, 4}
	cur := ScaleOutOnce{10, 9, 8, 7,
		6, 5, 6, 7, 8, 9,
		10, 11, 12}
	bytes1, _ := json.Marshal(prev)
	bytes2, _ := json.Marshal(cur)
	stats := &ScaleOutStats{}
	err := stats.Init(string(bytes1), string(bytes2))
	c.Assert(err, IsNil)
	err = stats.RenderTo("test_scale.html")
	c.Assert(err, IsNil)
	report, err := stats.Report()
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(report, "p0: BalanceInterval"), Equals, true)
	c.Assert(strings.Contains(report, "PR(last, red) is"), Equals, true)
}
