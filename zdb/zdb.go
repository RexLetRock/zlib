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

type Table[T Item] struct {
  name string
  index []string

  Tree *Btree[T]
  Data *zmap.ZMap[T]
  log  *zlog.ZLogQueueGenericV2[T]
}

var DB = make(map[string]interface{})

func NewTable[T Item](dbname string, less func(a, b T) bool) *Table[T] {
  tr := getTable[T](dbname, less)
  return tr
}

func (tr *Table[T]) Set(index int, item T) (T, bool) {
  tr.Tree.Set(item) // tr.log.Add(item)
  return item, true // tr.Data.SetAt(index, item)
}

func (tr *Table[T]) Get(index int) (T, bool) {
  return tr.Data.GetAt(index)
}

func (tr *Table[T]) GetByIndex(item T) (T, bool) {
  return tr.Tree.Get(item)
}

func (tr *Table[T]) GetAtByIndex(index int) (T, bool) {
  return tr.Tree.GetAt(index)
}

// Fast get functions start with Z prefix
func (tr *Table[T]) ZGet(index int) T {
  i, _ := tr.Data.GetAt(index)
  return i
}

func (tr *Table[T]) ZGetByIndex(item T) T {
  i, _ := tr.Tree.Get(item)
  return i
}

func (tr *Table[T]) ZGetAtByIndex(index int) T {
  i, _ := tr.Tree.GetAt(index)
  return i
}

func newTable[T Item](dbname string, less func(a, b T) bool) *Table[T] {
  tr := new(Table[T])
  tr.Tree = NewBtree[T](less)
  tr.Data = zmap.New[T](dataSize)
  tr.log = zlog.NewV2[T](logSize)
  tr.name = dbname
  return tr
}

func getTable[T Item](dbname string, less func(a, b T) bool) *Table[T] {
  if _, ok := DB[dbname]; !ok {
    DB[dbname] = newTable[T](dbname, less)
  }
  return DB[dbname].(*Table[T])
}
