package main

import (
  "bufio"
  "os"
  "fmt"

  "github.com/RexLetRock/zlib/zcount"
  "github.com/RexLetRock/zlib/zbench"
)

var (
  NRun = 20_000_000
  NCpu = 12
  n = zcount.Counter{}
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
  a := zcount.New()
  zbench.Run(NRun, NCpu, func(_, i int) {
    a.Add(i)
  })
  a.Get()

  fmt.Printf("\n\n=== LONGADDER ===\n")
  fmt.Printf("\n== RUN %v threads - race\n", NCpu)
  n.Reset()
  zbench.Run(NRun, NCpu, func(_, _ int) {
    n.IncZ()
  })

  zbench.Run(NRun, NCpu, func(_, _ int) {
    n.Value()
  })
}
