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
	return &scaleOut{
		c: c,
	}
}

func (s *scaleOut) run() error {
	if err := s.c.addStores(1); err != nil {
		return err
	}
	for {
		time.Sleep(time.Minute)
		if s.isBalance() {
			return nil
		}
	}
}

func (s *scaleOut) isBalance() bool {
	// todo get data from prometheus
	// todo @zeyuan
	return false
}

func (s *scaleOut) collect() error {
	// create createReport
	r, err := s.createReport()
	if err != nil {
		return err
	}

	// send current createReport
	err = s.c.sendReport(r)
	if err != nil {
		return err
	}

	// merge two reports
	lastReport, err := s.c.getLastReport()
	if len(lastReport) == 0 {
		return nil
	}
	markdown := s.mergeReport(lastReport, r)
	return s.c.sendResult(markdown)
}

func (s *scaleOut) createReport() (string, error) {
	//todo @zeyuan
	return "", nil
}

func (s *scaleOut) mergeReport(lastReport, report string) (markdown string) {
	//todo @zeyuan
	return ""
}
