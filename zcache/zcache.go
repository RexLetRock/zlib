// FAST CACHE < KEY : VALUE > : Superfast
package zcache

import (
  "github.com/segmentio/fasthash/fnv1a"
)

const mapSize = 20_000_000
var (
  fastMod = NewUint32(mapSize)
)

// ZCache General Type
type ZCache struct {
  M [mapSize]interface{}
  MFlag [mapSize]bool
  MBack *MapOf[string, interface{}]
  MEmpty interface{}
}

func New() (tr *ZCache) {
  return ZCacheCreate()
}

func ZCacheCreate() (tr *ZCache) {
  tr = new(ZCache)
  tr.MBack = new(MapOf[string, interface{}])
  return
}

func (tr *ZCache) Set(name string, item interface{}) bool {
  index := hash(name)
  indexFix := fastMod.Mod(index)
  if tr.MFlag[indexFix] {
    tr.MBack.Store(name, item)
    return true
  } else {
    tr.M[indexFix] = item
    tr.MFlag[indexFix] = true
    return false
  }
}

func (tr *ZCache) Get(name string) (result interface{}) {
  result = 0
  index := hash(name)
  indexFix := fastMod.Mod(index)
  if tr.MFlag[indexFix] {
    if tmp, ok := tr.MBack.Load(name); ok {
  		result = tmp
  	} else {
      result = tr.M[indexFix]
    }
  } else {
    result = tr.M[indexFix]
  }
  return
}

func hash(s string) uint32 {
  // h := fnv.New32a()
  // h.Write([]byte(s))
  // return h.Sum32()
  return fnv1a.HashString32(s)
}
