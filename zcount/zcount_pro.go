package zcount

import (
  "fmt"
  "github.com/RexLetRock/zlib/zgoid"
)

var (
  size = 1_000_000
  trunkSize = 20_000_000
)

type ZC struct {
  n []int
  m [][]int
}

func New() *ZC {
  tr := new(ZC)
  tr.n = make([]int, size)
  tr.m = make([][]int, size)
  return tr
}

func (tr *ZC) Add(item int) {
  i := zgoid.Get()
  tr.n[i] += 1
  index := tr.n[i]-1
  if index == 0 {
    tr.m[i] = make([]int, trunkSize)
  }
  tr.m[i][index] = item + 1
}

func Len(items []int) int {
  n := 0
  for _, v := range items {
    if v != 0 {
      n++
    }
  }
  return n
}

func (tr *ZC) Get() {
  m := 0
  for i := range tr.m {
    n := tr.n[i]
    m += n
    if n != 0 {
      // fmt.Printf("- %v %v \n", i, n)
      // fmt.Printf("%v \n", v)
    }
  }
  fmt.Printf("Total %v \n", m)
}
