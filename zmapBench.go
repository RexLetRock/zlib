package main

import (
  "bufio"
  "sync/atomic"
  "os"
  "fmt"

  "github.com/RexLetRock/zlib/zbench"
  "github.com/RexLetRock/zlib/zmap"
)

var (
  n = new(count32)
  m = new(count32)
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
  SRC := zmap.New[User](NRun)
  fmt.Printf("\n\n=== ZMAP ===\n")
  fmt.Printf("\n== SET %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    SRC.SetAt(i, User{ ID: i, Name: "Le Vo Huu Tai" })
  })
  fmt.Printf("\n== GET %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    SRC.ZGetAt(i)
  })

  fmt.Printf("\n== ZMAP ADD 1 THREAD \n")
  S1 := zmap.New[User](NRun)
  zbench.Run(NRun, 1, func(i, _ int) {
    S1.Add(SRC.ZGetAt(i))
  })

  n.reset()
  zbench.Run(NRun, NCpu, func(i, _ int) {
    u, _ := S1.GetAt(i)
    if i != u.ID {
      n.inc()
    }
  })
  fmt.Printf(" ↳ ERROR %v \n", n.get())

  fmt.Printf("\n== ZMAP ADD 12 THREAD & GETALL \n")
  S2 := zmap.New[User](NRun)
  zbench.Run(NRun / 2, NCpu, func(i, _ int) {
    S2.Add(SRC.ZGetAt(i))
  })
  S2A := S2.GetAll()
  SARR := zmap.New[User](NRun)
  for _, vS2A := range S2A {
    SARR.SetAt(vS2A.IID(), SRC.ZGetAt(vS2A.IID()))
  }

  n.reset()
  zbench.Run(NRun, NCpu, func(i, _ int) {
    u, _ := SARR.GetAt(i)
    if i == u.IID()  {
      m.inc()
    }
    if i != u.IID() && u.IID() != 0 {
      n.inc()
    }
  })
  fmt.Printf(" ↳ ERROR %v - CORRECT %v \n", n.get(), m.get())
  m.reset()
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

type User struct {
  ID int
  Name string
  Extra string
  DID string
}

func User_indexID (a, b User) bool {
  return a.ID < b.ID
}

func (p User) IID() int {
  return p.ID
}
