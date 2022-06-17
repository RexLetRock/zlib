package zid

import (
  "sync/atomic"
)

type count32 int32
func (c *count32) Next() int {
  return int(atomic.AddInt32((*int32)(c), 1))
}

func (c *count32) Get() int {
  return int(atomic.LoadInt32((*int32)(c)))
}

func (c *count32) Reset() int {
  return int(atomic.SwapInt32((*int32)(c), 0))
}

func NewAtomic() *count32 {
  return new(count32)
}
