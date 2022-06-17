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
)

func main() {
  benchZID()

  fmt.Printf("\n\nStop with ctrl + c \n\n")
  input := bufio.NewScanner(os.Stdin)
  input.Scan()
}

func benchZID() {
  fmt.Printf("\n\n=== ZID ===\n")
  ZID := zid.New("default")

  fmt.Printf("\n== RUN 1 threads\n")
  zbench.Run(NRun, 1, func(_, _ int) {
    _ = ZID.Next()
  })
  fmt.Printf("\n== RUN %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(_, _ int) {
    _ = ZID.Next()
  })
}
