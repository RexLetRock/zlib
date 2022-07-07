package main

import (
  "fmt"
  "time"
  "github.com/kpango/fastime"
  "github.com/RexLetRock/zlib/zbench"
)

const (
  NRun = 10_000_000
  NCpu = 12
)

func main() {
  fmt.Printf("\n\n=== FAST TIME ===\n")
  fmt.Printf("Time %v\n", fastime.UnixNanoNow())
  zbench.Run(NRun, NCpu, func(_, _ int) {
    fmt.Printf("%v\n", time.Now().UnixNano() - fastime.UnixNanoNow())
  })
}
