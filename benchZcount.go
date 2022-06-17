package main

import (
  "bufio"
  "os"
  "fmt"

  "github.com/RexLetRock/zlib/zcount"
  "github.com/RexLetRock/zlib/zbench"
)

var (
  NRun = 10_000_000
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
  zbench.Run(NRun, NCpu, func(_, _ int) {
    a.Inc()
  })

  fmt.Printf("COUNT %v \n", a.Get())

  fmt.Printf("\n== RUN %v threads - race\n", NCpu)
  n.Reset()
  zbench.Run(NRun, NCpu, func(_, _ int) {
    n.Inc()
  })

  zbench.Run(NRun, NCpu, func(_, _ int) {
    n.Value()
  })
}
