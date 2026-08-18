package quota

import "sync"

type Usage struct {
	Keys  int   `json:"keys"`
	Bytes int64 `json:"bytes"`
}

func (u Usage) Add(keys int, bytes int64) Usage {
	u.Keys += keys
	u.Bytes += bytes
	if u.Keys < 0 {
		u.Keys = 0
	}
	if u.Bytes < 0 {
		u.Bytes = 0
	}
	return u
}

type Snapshot struct {
	NS    string `json:"ns"`
	Keys  int    `json:"keys"`
	Bytes int64  `json:"bytes"`
}

type Counter struct {
	mu  sync.Mutex
	use map[string]Usage
}

func NewCounter() *Counter {
	return &Counter{use: make(map[string]Usage)}
}

func (c *Counter) Get(ns string) Usage {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.use[ns]
}

func (c *Counter) Set(ns string, u Usage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if u.Keys == 0 && u.Bytes == 0 {
		delete(c.use, ns)
		return
	}
	c.use[ns] = u
}

func (c *Counter) Apply(ns string, keys int, bytes int64) Usage {
	c.mu.Lock()
	defer c.mu.Unlock()
	u := c.use[ns].Add(keys, bytes)
	c.use[ns] = u
	return u
}

func (c *Counter) DeleteNS(ns string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.use, ns)
}

func (c *Counter) Export() map[string]Usage {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]Usage, len(c.use))
	for k, v := range c.use {
		out[k] = v
	}
	return out
}

func (c *Counter) Import(m map[string]Usage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.use = make(map[string]Usage, len(m))
	for k, v := range m {
		c.use[k] = v
	}
}
