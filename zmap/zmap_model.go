package zmap

import (
  "sync"
  "sync/atomic"
)

type count32 int32
func (c *count32) inc() int {
  return int(atomic.AddInt32((*int32)(c), 1))
}
func (c *count32) get() int {
  return int(atomic.LoadInt32((*int32)(c)))
}
func (c *count32) set(item int) int {
  return int(atomic.SwapInt32((*int32)(c), int32(item)))
}
func (c *count32) reset() int {
  return int(atomic.SwapInt32((*int32)(c), 0))
}

type Item interface {
  // IID() int
}

type ZMap[T Item] struct {
  items []T
  mutex sync.Mutex
  count count32
  size int
}
