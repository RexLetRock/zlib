package main

import (
  "bufio"
  "os"
  "fmt"

  "github.com/RexLetRock/zlib/zid"
  "github.com/RexLetRock/zlib/zbench"
)

var (
  NRun = 10_000_000
  NCpu = 12
  n = zid.NewAtomic()
)

func main() {
  benchZID()

  fmt.Printf("\n\nStop with ctrl + c \n\n")
  input := bufio.NewScanner(os.Stdin)
  input.Scan()
}

func benchZID() {
  fmt.Printf("\n\n=== ATOMIC ===\n")

  fmt.Printf("\n== RUN 1 threads\n")
  n.Reset()
  zbench.Run(NRun, 1, func(_, _ int) {
    n.Next()
  })

  fmt.Printf("\n== RUN %v threads - race\n", NCpu)
  n.Reset()
  zbench.Run(NRun, NCpu, func(_, _ int) {
    n.Next()
  })

  fmt.Printf("\n\n=== ZID ===\n")
  fmt.Printf("\n== RUN 1 threads\n")
  ZID := zid.New("default")
  zbench.Run(NRun, 1, func(_, _ int) {
    _ = ZID.Next()
  })
  fmt.Printf("\n== RUN %v threads - race\n", NCpu)
  zbench.Run(NRun, NCpu, func(_, _ int) {
    _ = ZID.Next()
  })
}
