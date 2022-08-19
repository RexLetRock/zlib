package main

import (
  "bufio"
  "os"
  "fmt"
  "sync/atomic"

  "github.com/panjf2000/ants/v2"

  "github.com/RexLetRock/zlib/zbench"
  "github.com/RexLetRock/zlib/zmap"

  "github.com/RexLetRock/zlib/extra/model"
)

var (
  n = new(count32)
  m = new(count32)
  NRun = 10_000_000
  NCpu = 12
)

func main() {
  ants.NewPool(1000000, ants.WithPreAlloc(true))
  benchZMap()

  fmt.Printf("\n\nStop with ctrl + c \n\n")
  input := bufio.NewScanner(os.Stdin)
  input.Scan()
}

func benchZMap() {
  SRC := zmap.New[model.User](NRun)
  fmt.Printf("\n\n=== ZMAP ===\n")

  fmt.Printf("\n== SET %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    ants.Submit(func() {
      SRC.SetAt(i, model.User{ ID: i, Name: "Le Vo Huu Tai" })
    })
  })

  zbench.Run(NRun, NCpu, func(i, _ int) {
    u, _ := SRC.GetAt(i)
    if i == u.IID()  {
      m.inc()
    }
    if i != u.IID() && u.IID() != 0 {
      n.inc()
    }
  })
  fmt.Printf(" ↳ ERROR %v - CORRECT %v \n", n.get(), m.get())
}


type count32 int32
func (c *count32) inc() int {
  return int(atomic.AddInt32((*int32)(c), 1))
}
func (c *count32) get() int {
  return int(atomic.LoadInt32((*int32)(c)))
}
func (c *count32) set(item int) int {
  return int(atomic.SwapInt32((*int32)(c), int32(item)))
}
func (c *count32) reset() int {
  return int(atomic.SwapInt32((*int32)(c), 0))
}
