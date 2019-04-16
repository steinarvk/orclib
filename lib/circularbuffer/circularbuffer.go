package circularbuffer

type Circular struct {
	Capacity int
	Elements int
	Start    int
	End      int
}

func New(capacity int) *Circular {
	return &Circular{Capacity: capacity}
}

func (c *Circular) AppendIndex() int {
	rv := c.End

	c.End += 1
	if c.End >= c.Capacity {
		c.End = 0
	}

	if c.Elements < c.Capacity {
		c.Elements += 1
	} else {
		c.Start = c.End
	}

	return rv
}

// SliceIndices returns indices (a, b, c, d) such that append(buf[a:b], buf[c:d]...) contains the correct values.
func (c *Circular) SliceIndices() (int, int, int, int) {
	if c.Elements == 0 {
		return 0, 0, 0, 0
	}
	end := c.Start + c.Elements
	if end > c.Capacity {
		end = c.Capacity
	}
	leftover := c.Elements - (end - c.Start)
	return c.Start, end, 0, leftover
}
