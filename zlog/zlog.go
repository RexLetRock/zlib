package zlog

import (
  "os"
  "bufio"
  "fmt"
  "sync"
  "strconv"
  "time"
  "github.com/RexLetRock/zlib/zgoid"
)

var (
  size = 1_000
  trunkSize = 1_000_000
  trunkSizeLog = 1_000_000
  itemToFlush = 1_000
  logPath = "extra/data"
)

type ZLogQueue struct {
  n []int
  m [][]int
  x [][]int
  mu sync.Mutex
}

func New() *ZLogQueue {
  tr := new(ZLogQueue)
  tr.n = make([]int, size)
  tr.m = make([][]int, size)
  tr.x = make([][]int, size)
  return tr
}

func (tr *ZLogQueue) Add(item int) {
  i := zgoid.Get()
  tr.n[i] += 1
  index := tr.n[i]-1
  if index == 0 {
    tr.m[i] = make([]int, trunkSize)
  }
  if index >= trunkSizeLog {
    go writeLog(tr.m[i], i)
    tr.n[i] = 0
    index = 0
  }
  tr.m[i][index] = item + 1
}

func writeLog(items []int, cpuName int64) {
  timeseed := time.Now().UnixNano()
  filePath := logPath + "/logfile_" + strconv.Itoa(int(cpuName)) + strconv.Itoa(int(timeseed))
  f, _ := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  w := bufio.NewWriter(f)

  count := 0
  for _, item := range items {
    count++
    w.WriteString(strconv.Itoa(item) + "\n")
    if count % itemToFlush == 0 {
      w.Flush()
    }
  }
  w.Flush()
}

func (tr *ZLogQueue) GetAll() (result [][]int) {
  tr.mu.Lock()
  result = tr.m
  tr.m = make([][]int, size)
  tr.n = make([]int, size)
  tr.mu.Unlock()
  return result
}

func (tr *ZLogQueue) Get() {
  m := 0
  for i := range tr.m {
    n := tr.n[i]
    m += n
    if n != 0 {
    }
  }
  fmt.Printf("Total %v \n", m)
}
