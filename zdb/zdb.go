// Simple Database With Btree, Array, Map
package zdb

import (
  "github.com/RexLetRock/zlib/zmap"
  "github.com/RexLetRock/zlib/zlog"

  // "github.com/tidwall/btree"
)

const (
  dataSize = 10_000_000
  logSize = 1_000_000
)

type Item interface {
  IID() int
  ToString() string
}

// type SItem struct {
//   ID int
// }

type Table[T Item] struct {
  name string
  index []string

  tree *Btree[T]
  data *zmap.ZMap[T]
  log  *zlog.ZLogQueueGenericV2[T]
}

var DB = make(map[string]interface{})

func NewTable[T Item](dbname string, less func(a, b T) bool) *Table[T] {
  tr := getTable[T](dbname, less)
  return tr
}

func (tr *Table[T]) Set(_ int, item T) (T, bool) {
  tr.tree.Set(item)
  // tr.log.Add(item)
  return item, true
  // return tr.data.SetAt(index, item)
}

func (tr *Table[T]) Get(key T) (T, bool) {
  return tr.data.GetAt(key.IID())
}

func (tr *Table[T]) GetAt(index int) (T, bool) {
  return tr.data.GetAt(index)
}

func (tr *Table[T]) ZGet(key T) T {
  i, _ := tr.data.GetAt(key.IID())
  return i
}

func newTable[T Item](dbname string, less func(a, b T) bool) *Table[T] {
  tr := new(Table[T])
  tr.tree = NewBtree[T](less)
  tr.data = zmap.New[T](dataSize)
  tr.log = zlog.NewGenericv2[T](logSize)
  tr.name = dbname
  return tr
}

func getTable[T Item](dbname string, less func(a, b T) bool) *Table[T] {
  if _, ok := DB[dbname]; !ok {
    DB[dbname] = newTable[T](dbname, less)
  }
  return DB[dbname].(*Table[T])
}
