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
  zBandwidth := zbandwidth.New("enp4s0f1")
  time.Sleep(1 * time.Second)

  fmt.Printf("\n\n=== GCACHE ===\n")
	zbench.Run(NRun, NCpu, func(i, _ int) {
		zBandwidth.Get()
	})
  fmt.Printf("\nRx: %v, Tx: %v \n", zBandwidth.GetRx() / float64(2<<19), zBandwidth.GetTx() / float64(2<<19))
  fmt.Printf("%v", zBandwidth.GetString())
}
