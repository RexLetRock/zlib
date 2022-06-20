package main

import (
  "bufio"
  "os"
  "fmt"

  "github.com/RexLetRock/zlib/zlog"
  "github.com/RexLetRock/zlib/zbench"
)

var (
  NRun = 20_000_000
  NCpu = 12
)

func main() {
  benchZID()

  fmt.Printf("\n\nStop with ctrl + c \n\n")
  input := bufio.NewScanner(os.Stdin)
  input.Scan()
}

func benchZID() {
  fmt.Printf("\n\n=== ZCOUNT ===\n")
  fmt.Printf("\n== RUN %v threads\n", NCpu)
  a := zlog.New()
  zbench.Run(NRun, 1, func(i, _ int) {
    a.Add(i)
  })
  a.Get()
}
