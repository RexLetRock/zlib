package zlog

import (
  "time"
  "fmt"
  "os"

  "bufio"
)

func (tr *ZLogQueueGenericV2[T]) writeLog(items []T, iCpu int64) {
  timeseed := time.Now().UnixNano() // / 100_000_000
  filePath := logPath + "/logfile_" + fmt.Sprintf("%012d_%d_%08d", items[0].IID(), timeseed, iCpu)
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
