package main

import (
  "fmt"
  "time"
  "github.com/RexLetRock/zlib/zbandwidth"
  "github.com/RexLetRock/zlib/zbench"
)

var (
	NRun = 5_000_000
	NCpu = 12
)

func main() {
  zBandwidth := zbandwidth.New("eth0")
  time.Sleep(1 * time.Second)

  fmt.Printf("\n\n=== GCACHE ===\n")
	zbench.Run(NRun, NCpu, func(i, _ int) {
		a, b := zBandwidth.Get()
    fmt.Printf("%v %v \n", a, b)
	})
}
