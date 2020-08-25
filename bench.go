package main

import "time"

type bench interface {
	run() error
	collect() error
}

type scaleOut struct {
	c *cluster
}

func newScaleOut(c *cluster) *scaleOut {
	c.changeStatus(startBench)
	return &scaleOut{
		c: c,
	}
}

func (s *scaleOut) run() error {
	for {
		if s.isBalance() {
			return nil
		}
		time.Sleep(time.Minute)
	}
}

func (s *scaleOut) isBalance() bool {
	// todo get data from prometheus
	return false
}

func (s *scaleOut) collect() error {
	// todo
	s.c.report()
	return nil
}
