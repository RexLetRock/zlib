// WriteLog Parallel Superfast
package zlog

import (
  "os"
  "bufio"
  "fmt"
  "sync"
  "time"
  "github.com/RexLetRock/zlib/zgoid"
)

const (
  size = 1000
  itemToFlush = 500
  logPath = "extra/data"
)

type Item interface {
  IID() int
  ToString() string
}

type ZLogQueueGeneric[T Item] struct {
  trunkSize int
  count []int
  countAll []int
  countWrite []int
  countWriteTime []int64

  items [][]T

  mu sync.Mutex
}

func NewGeneric[T Item](trunkSize int) *ZLogQueueGeneric[T] {
  tr := new(ZLogQueueGeneric[T])
  tr.trunkSize = trunkSize
  tr.count = make([]int, size)
  tr.countAll = make([]int, size)

  tr.countWrite = make([]int, size)
  tr.countWriteTime = make([]int64, size)

  tr.items = make([][]T, size)
  go tr.writeLogCheckTime()
  return tr
}

func (tr *ZLogQueueGeneric[T]) Add(item T) {
  i := zgoid.Get()
  tr.count[i] += 1
  tr.countAll[i] += 1
  index := tr.count[i]-1
  if index == 0 {
    tr.items[i] = make([]T, tr.trunkSize)
  }
  tr.items[i][index] = item
  if tr.count[i] >= tr.trunkSize {
    go tr.writeLog(tr.items[i][0:tr.trunkSize], i)
    tr.countWriteTime[i] = time.Now().Unix()
    tr.count[i] = 0
  }
}

func (tr *ZLogQueueGeneric[T]) writeLog(items []T, cpuName int64) {
  timeseed := time.Now().UnixNano()
  filePath := logPath + "/logfile_" + fmt.Sprintf("%d", timeseed) + "_" + fmt.Sprintf("%d", cpuName)
  f, _ := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  defer f.Close()

  w := bufio.NewWriter(f)
  count := 0
  for _, item := range items {
    count++
    w.WriteString(item.ToString())
    if count % itemToFlush == 0 {
      w.Flush()
    }
  }
  w.Flush()
}

func (tr *ZLogQueueGeneric[T]) writeLogCheckTime() {
  ticker := time.NewTicker(5 * time.Second)
  quit := make(chan struct{})
  go func() {
    for {
      select {
      case <- ticker.C:
        curTime := time.Now().Unix()
        for iAll, vAll := range tr.countAll {
          if vAll != 0 {
            if curTime != tr.countWriteTime[iAll] && tr.count[iAll] != 0 {
              go tr.writeLog(tr.items[iAll][:tr.count[iAll]], int64(iAll))
              tr.countWriteTime[iAll] = curTime
              tr.count[iAll] = 0
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
