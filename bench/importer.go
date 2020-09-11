package bench

import (
	"strings"

	"github.com/lhy1024/bench/utils"
	"github.com/pingcap/errors"
	"github.com/siddontang/go-mysql/client"
)

type Importer interface {
	Import() error
}

type ycsb struct {
	c               *Cluster
	workload        string
	dbName          string
	withEmptyRegion bool
}

func NewYCSB(c *Cluster, workload string) Importer {
	return &ycsb{
		c:               c,
		workload:        workload,
		dbName:          "test",
		withEmptyRegion: false,
	}
}

func splitAddr(addr string) (string, string, error) {
	subs := strings.Split(addr, ":")
	if len(subs) != 2 {
		return "", "", errors.New("addr is wrong")
	}
	return subs[0], subs[1], nil
}

func (l *ycsb) Import() error {
	host, port, err := splitAddr(l.c.tidbAddr)
	if err != nil {
		return err
	}
	// go-ycsb insert
	cmd := utils.NewCommand("./go-ycsb/go-ycsb", "Import", "mysql", "-P", "./go-ycsb/"+l.workload, "-p", "mysql.user=root", "-p", "mysql.db="+l.dbName,
		"-p", "mysql.host="+host, "-p", "mysql.port="+port)
	err = cmd.Run()
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
	res, err := conn.Execute("split table test_go_ycsb BETWEEN (0) AND (9223372036854775807) REGIONS 1000;")
	if err != nil {
		return err
	}
	res, err = conn.Execute("show table test_go_ycsb regions;")
	if err != nil {
		return err
	}
	if res.RowNumber() < 1000 {
		return errors.New("split region failed")
	}
	return nil
}
