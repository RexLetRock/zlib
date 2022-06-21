package main

import (
  "bufio"
  "os"
  "fmt"
  "sync/atomic"

  "github.com/RexLetRock/zlib/extra/model"

  "github.com/RexLetRock/zlib/zbench"
  "github.com/RexLetRock/zlib/zdb"
  "github.com/RexLetRock/zlib/zmap"
)

var (
  NRun = 10_000_000
  NCpu = 12
)

func main() {
  benchZMap()

  fmt.Printf("\n\nStop with ctrl + c \n\n")
  input := bufio.NewScanner(os.Stdin)
  input.Scan()
}

func benchZMap() {
  n := new(count32)
  m := new(count32)

  fmt.Printf("\n\n=== ZMAP ===\n")
  SRC := zmap.New[model.User](NRun)
  fmt.Printf("\n== SET %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    SRC.SetAt(i, model.User{ ID: i, Name: "Le Vo Huu Tai" })
  })
  fmt.Printf("\n== GET %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    j := SRC.ZGetAt(i)
    if i != j.IID() {
      n.inc()
    } else {
      m.inc()
    }
  })
  fmt.Printf("SUCCESS %v - ERROR %v\n", m.get(), n.get())



  fmt.Printf("\n\n=== ZDB ===\n")
  Zdb := zdb.NewTable[model.User]("User", model.User_indexID)
  fmt.Printf("\n== SET %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    Zdb.Set(i, model.User{ ID: i, Name: "Le Vo Huu Tai" })
  })

  n.reset()
  m.reset()
  zbench.Run(NRun, NCpu, func(i, _ int) {
    j := Zdb.ZGet(model.User{ ID: i, Name: "Le Vo Huu Tai" })
    if i != j.IID() {
      n.inc()
    } else {
      m.inc()
    }
  })
  fmt.Printf("SUCCESS %v - ERROR %v\n", m.get(), n.get())
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
