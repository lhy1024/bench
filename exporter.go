package main

type Exporter interface {
	Export() error
}

type BR struct {
	c *Cluster
}

func NewBR(c *Cluster) Exporter {
	return &BR{
		c: c,
	}
}

func (l *BR) Export() error {
	// todo @zeyuan
	return nil
}
