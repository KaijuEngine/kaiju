/******************************************************************************/
/* instance_buffer_capacity.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

const minInstanceBufferCapacity = 4

type InstanceBufferCapacity struct {
	capacity int
}

func (c InstanceBufferCapacity) Capacity() int {
	return c.capacity
}

func (c *InstanceBufferCapacity) Ensure(instanceCount int) (int, bool) {
	next, changed := c.Next(instanceCount)
	if changed {
		c.capacity = next
	}
	return c.capacity, changed
}

func (c InstanceBufferCapacity) Next(instanceCount int) (int, bool) {
	if instanceCount <= c.capacity {
		return c.capacity, false
	}
	next := c.capacity
	if next < minInstanceBufferCapacity {
		next = minInstanceBufferCapacity
	}
	for next < instanceCount {
		next *= 2
	}
	return next, true
}

func (c *InstanceBufferCapacity) Commit(capacity int) {
	c.capacity = capacity
}

func (c *InstanceBufferCapacity) Reset() {
	c.capacity = 0
}
