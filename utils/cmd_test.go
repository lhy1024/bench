package utils

import (
	"strings"
	"testing"

	. "github.com/pingcap/check"
)

func TestCmd(t *testing.T) {
	TestingT(t)
}

type testCmdSuite struct{}

var _ = Suite(&testCmdSuite{})

func (s *testCmdSuite) TestCmd(c *C) {
	cmd := NewCommand("echo", "hello test")
	out, err := cmd.Run()
	c.Assert(strings.Contains(out, "hello"), Equals, true)
	c.Assert(err, IsNil)
}
