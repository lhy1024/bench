package main

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
	// todo
	return nil
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