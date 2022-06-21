package zlog

import (
  "sync"
  "time"
  "github.com/RexLetRock/zlib/zgoid"
)

type ZLogQueueGenericV2[T Item] struct {
  trunkSize int
  count []int
  countAll []int
  countWrite []int
  countWriteTime []int64

  items [][]T

  mu sync.Mutex
}

func NewGenericv2[T Item](trunkSize int) *ZLogQueueGenericV2[T] {
  tr := new(ZLogQueueGenericV2[T])
  tr.trunkSize = trunkSize
  tr.count = make([]int, size)
  tr.countAll = make([]int, size)

  tr.countWrite = make([]int, size)
  tr.countWriteTime = make([]int64, size)

  tr.items = make([][]T, size)
  go tr.writeLogCheckTime()
  return tr
}

func (tr *ZLogQueueGenericV2[T]) Add(item T) {
  iCpu := zgoid.Get()
  tr.count[iCpu] += 1
  tr.countAll[iCpu] += 1
  index := tr.count[iCpu]-1
  if index == 0 {
    tr.items[iCpu] = make([]T, tr.trunkSize)
  }
  tr.items[iCpu][index] = item
  if tr.count[iCpu] >= tr.trunkSize {
    go tr.writeLog(tr.items[iCpu][0:tr.trunkSize], iCpu)
    tr.countWriteTime[iCpu] = time.Now().Unix()
    tr.count[iCpu] = 0
  }
}

func (tr *ZLogQueueGenericV2[T]) writeLogCheckTime() {
  ticker := time.NewTicker(2_000 * time.Millisecond)
  quit := make(chan struct{})
  go func() {
    for {
      select {
      case <- ticker.C:
        curTime := time.Now().Unix()
        for iCpu, vAll := range tr.countAll {
          if vAll != 0 {
            if curTime != tr.countWriteTime[iCpu] && tr.count[iCpu] != 0 {
              go tr.writeLog(tr.items[iCpu][:tr.count[iCpu]], int64(iCpu))
              tr.countWriteTime[iCpu] = curTime
              tr.count[iCpu] = 0
            }
          }
        }
      case <- quit:
        ticker.Stop()
        return
      }
    }
  }()
}
