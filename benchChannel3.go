package main

import (
  "os"
  "bufio"
	"fmt"
  "time"

  "github.com/RexLetRock/zlib/zbench"
  "github.com/RexLetRock/zlib/zgoid"
)

const NChannel = 12
const NCpuname = 10000

var (
  NRun = 10_000_000
  NCpu = 12
  ACount [NCpuname]int

  timeStart = int64(0)
  timeNow = int64(0)
)

func benchChannel(threadName int) {
  // Prebuffer channel
  c := make(chan string, 1000)

  // Producer
	go func() { for { c <- "How is the weather like today ? hope you okie" } }()

  // Consumer
  for r := range c {
    _ = r
    ACount[threadName] += 1
	}
}

func benchZID() {
  fmt.Printf("\n== ZGOID %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    zgoid.Get()
  })
}

func showResult() {
  ticker := time.NewTicker(1 * time.Second)
  quit := make(chan struct{})
  go func() {
    for {
      select {
        case <- ticker.C:
          timeNow = time.Now().Unix()
          total := 0
          for i, v := range ACount {
        		if v != 0 {
              total += v
              ACount[i] = 0
            }
        	}
          fmt.Printf("Threadnum %v - Msg/s %v \n", NCpu, int64(float64(total) / float64(timeNow - timeStart)))
          timeStart = timeNow
        case <- quit:
          ticker.Stop()
          return
      }
    }
  }()
}

func main() {
  timeStart = time.Now().Unix()

  // Bench
  for i := 0; i <= NCpu; i++ {
    go benchChannel(i)
  }

  // Result
  go showResult()

  fmt.Printf("\n\nStop with ctrl + c \n\n")
  input := bufio.NewScanner(os.Stdin)
  input.Scan()
}
