package main

import (
  "bufio"
  "os"
  "fmt"
  "sync/atomic"

  "gopkg.in/OlexiyKhokhlov/avltree.v2"

  "github.com/RexLetRock/zlib/extra/model"
  "github.com/RexLetRock/zlib/zbench"
  "github.com/RexLetRock/zlib/zdb"
  "github.com/RexLetRock/zlib/zmap"
)

var (
  NRun = 5_000_000
  NCpu = 12
)

func main() {
  benchZMap()

  fmt.Printf("\n\nStop with ctrl + c \n\n")
  input := bufio.NewScanner(os.Stdin)
  input.Scan()
}

func benchZMap() {
  fmt.Printf("\n\n=== AVLTREE ===\n")
  tree := avltree.NewAVLTreeOrderedKey[int, string]()
  zbench.Run(NRun, 1, func(i, _ int) {
    tree.Insert(i, "")
  })

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
  Zdb := zdb.NewTable[model.User]("User", model.User_Index_Name)
  fmt.Printf("\n== SET %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    Zdb.Set(i, model.User{ ID: i, Name: "Le Vo Huu Tai" })
  })

  fmt.Printf("\n== GET %v threads\n", NCpu)
  n.reset()
  zbench.Run(NRun, NCpu, func(i, _ int) {
    j := Zdb.ZGet(i)
    if i != j.ID {
      n.inc()
    }
  })
  fmt.Printf("SUCCESS - ERROR %v\n", n.get())


  n.reset()
  fmt.Printf("\n== GET WITH INDEX %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    j := Zdb.ZGetByIndex(model.User{ ID: i, Name: "Le Vo Huu Tai" })
    if i != j.ID {
      n.inc()
    }
  })
  fmt.Printf("SUCCESS - ERROR %v\n", n.get())


  n.reset()
  fmt.Printf("\n== GET AT WITH INDEX %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    j := Zdb.ZGetAtByIndex(i)
    if i != j.ID {
      n.inc()
    }
  })
  fmt.Printf("SHOW VALUE %v %v \n", 10, Zdb.ZGetAtByIndex(10))
  fmt.Printf("SUCCESS - ERROR %v\n", n.get())


  fmt.Printf("\n== ASCEND SCAN 10 %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    i = 0
    Zdb.Tree.Ascend(model.User{ ID: i }, func(item model.User) bool {
      i += 1
      if i < 10 {
        return true
      } else {
        return false
      }
  	})
  })


  fmt.Printf("\n== RANK %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    GetRank(Zdb, model.User{ ID: i, Name: "Le Vo Huu Tai" })
  })
  rank, user := GetRank(Zdb, model.User{ ID: 237468, Name: "Le Vo Huu Tai" })
  fmt.Printf("USER: %v - RANK: %v \n", user, rank)
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

// GetRank
func GetRank(tr *zdb.Table[model.User], item model.User) (int, model.User) {
  keys := tr.Tree
  searchItem := item // , _ := keys.Get(item)
  itemB, _ := keys.GetAt(0)

  low := 0
  high := keys.Len() - 1
  predicted := true
  for low < high {
    mid := 0
    // Predict
    if !predicted {
      predicted = true
      valueB, _ := keys.GetAt(high)
      valueA, _ := keys.GetAt(int(searchItem.ID))
      valuePredict := (high * int(valueA.ID)) / int(valueB.ID)
      mid = valuePredict
      valueCompare, _ := keys.GetAt(valuePredict)
      if valueCompare.ID >= searchItem.ID {
        high = valuePredict + 1
      } else {
        low = valuePredict
      }
      // fmt.Printf("Predict %v - Mid %v - Low %v - High %v \n", valuePredict, mid, low, high)
    } else {
      mid = (high + low) / 2
    }

    itemB, _ = keys.GetAt(mid)
    if searchItem.ID > itemB.ID {
      low = mid + 1
    } else if searchItem.ID < itemB.ID {
      high = mid
    } else if searchItem.ID == itemB.ID {
      low = mid
      high = low - 1 // exit
    }
  }

  // fmt.Printf("Low %v - High %v \n", low, high)
  result := keys.Len() - low
  return result, searchItem
}
