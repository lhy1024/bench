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
	if err := s.c.addStore(); err != nil {
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
	// create report
	report, err := s.createReport()
	if err != nil {
		return err
	}

	// try get last report
	lastReport, err := s.c.getLastReport()
	if err != nil {
		return err
	}

	// send report
	var plainText string
	if len(lastReport) == 0 { //first send
		plainText = ""
	} else { //second send
		plainText = s.mergeReport(lastReport, report)
	}

	return s.c.sendReport(report, plainText)
}

func (s *scaleOut) createReport() (string, error) {
	//todo @zeyuan
	return "", nil
}

// lastReport is
func (s *scaleOut) mergeReport(lastReport, report string) (plainText string) {
	//todo @zeyuan
	return ""
}
