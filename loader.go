package main

import (
	"bytes"
	"errors"
	"go.uber.org/zap"
	"os/exec"
	"strings"

	"github.com/pingcap/log"
	"github.com/siddontang/go-mysql/client"
)

type loader interface {
	load() error
}

type ycsb struct {
	c *cluster
}

func newYcsb(c *cluster) loader {
	return &ycsb{
		c: c,
	}
}

func splitAddr(addr string) (string, string, error) {
	subs := strings.Split(addr, ":")
	if len(subs) != 2 {
		return "", "", errors.New("addr is wrong")
	}
	return subs[0], subs[1], nil
}

func (l *ycsb) load() error {
	host, port, err := splitAddr(l.c.tidb)
	if err != nil {
		return err
	}
	dbName := "test"

	// go-ycsb insert
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("./go-ycsb/go-ycsb", "load", "mysql", "-P", "./go-ycsb/workload", "-p", "mysql.user=root", "-p", "mysql.db="+dbName,
		"-p", "mysql.host="+host, "-p", "mysql.port="+port)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	log.Debug("go-ycsb", zap.Strings("cmd", cmd.Args), zap.String("stdout", stdout.String()), zap.String("stderr", stderr.String()))
	if err != nil {
		return err
	}

	// split table
	conn, err := client.Connect(l.c.tidb, "root", "", dbName)
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
	return err
}

type br struct {
	c *cluster
}

func newBr(c *cluster) *ycsb {
	return &ycsb{
		c: c,
	}
}

func (l *br) load() error {
	// todo @zeyuan
	return nil
}
