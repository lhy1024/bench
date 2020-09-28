package bench

import (
	"strings"

	"github.com/lhy1024/bench/utils"
	"github.com/pingcap/errors"
	"github.com/siddontang/go-mysql/client"
)

type generator interface {
	Generate() error
}

type ycsb struct {
	c               *cluster
	workload        string
	dbName          string
	withEmptyRegion bool
}

func newYCSB(c *cluster, workload string) generator {
	return &ycsb{
		c:               c,
		workload:        workload,
		dbName:          "test",
		withEmptyRegion: false,
	}
}

func splitAddr(addr string) (host string, port string, err error) {
	subs := strings.Split(addr, ":")
	if len(subs) != 2 {
		return "", "", errors.New("addr is wrong")
	}
	host, port = subs[0], subs[1]
	return host, port, nil
}

// Generate is used to generate data.
func (l *ycsb) Generate() error {
	host, port, err := splitAddr(l.c.tidbAddr)
	if err != nil {
		return err
	}
	// go-ycsb insert
	cmd := utils.NewCommand("./go-ycsb/go-ycsb", "load", "mysql", "-P", "./go-ycsb/"+l.workload, "-p", "mysql.user=root", "-p", "mysql.db="+l.dbName,
		"-p", "mysql.host="+host, "-p", "mysql.port="+port)
	_, err = cmd.Run()
	if err != nil {
		return err
	}

	if l.withEmptyRegion {
		return l.split()
	}
	return nil
}

func (l *ycsb) split() error {
	// split table
	conn, err := client.Connect(l.c.tidbAddr, "root", "", l.dbName)
	if err != nil {
		return err
	}
	_, err = conn.Execute("split table test_go_ycsb BETWEEN (0) AND (9223372036854775807) REGIONS 1000;")
	if err != nil {
		return err
	}
	res, err := conn.Execute("show table test_go_ycsb regions;")
	if err != nil {
		return err
	}
	if res.RowNumber() < 1000 {
		return errors.New("split region failed")
	}
	return nil
}

func newEmptyGenerator() generator {
	return &emptyGenerator{}
}

type emptyGenerator struct {
}

// Generate is used to generate data.
func (l *emptyGenerator) Generate() error {
	return nil
}
