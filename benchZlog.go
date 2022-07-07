package main

import (
  "bufio"
  "os"
  "fmt"

  "github.com/RexLetRock/zlib/zlog"
  "github.com/RexLetRock/zlib/zbench"

  "github.com/RexLetRock/zlib/extra/model"
)

var (
  NRun = 40_311_123
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
  a := zlog.NewV2[model.User](1_000_000)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    a.Add(model.User{ ID: i + 1, Name: "Le Vo Huu Tai" })
  })

  // b := new(zlog.ZCount)
  // zbench.Run(NRun, NCpu, func(_, _ int) {
  //   b.Add()
  // })
  // zbench.Run(NRun, NCpu, func(_, _ int) {
  //   b.Retrieve()
  // })
}
