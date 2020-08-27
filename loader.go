package main

import (
	"bytes"
	"errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"os/exec"
	"strings"
)

type loader interface {
	load() error
}

type ycsb struct {
	c *cluster
}

func newYcsb(c *cluster) *ycsb {
	return &ycsb{
		c: c,
	}
}

func (l *ycsb) load() error {
	subs := strings.Split(l.c.tidb, ":")
	if len(subs) != 2 {
		return errors.New("tidb addr is wrong")
	}
	host, port := subs[0], subs[1]
	cmd := exec.Command("./go-ycsb/go-ycsb", "load", "mysql", "-P", "./go-ycsb/workload", "-p", "mysql.user=root", "-p mysql.db=test",
		"-p mysql.host="+host, "-p", "mysql.port="+port)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Info("go-ycsb", zap.Strings("cmd", cmd.Args), zap.String("stdout", stdout.String()), zap.String("stderr", stderr.String()))
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
