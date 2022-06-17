package zcount

import (
  "github.com/RexLetRock/goid"
)

const (
  size = 1_000
)

type ZC struct {
  m []int
}

func New() *ZC {
  tr := new(ZC)
  tr.m = make([]int, size)
  return tr
}

func (tr *ZC) Inc() {
  i := goid.Get()
  tr.m[i] += 1
}

func (tr *ZC) Get() int {
  result := 0
  for _, v := range tr.m {
    result += v
  }
  return result
}
