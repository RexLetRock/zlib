// FAST CACHE < KEY : VALUE > : Superfast
package zcache

// ZCache Int Type
type ZCacheInt struct {
  M [mapSize]int
  MFlag [mapSize]int
  MBack *MapOf[string, int]
  MEmpty int
}

func ZCacheIntCreate() (tr *ZCacheInt) {
  tr = new(ZCacheInt)
  tr.MBack = new(MapOf[string, int])
  tr.MEmpty = 0
  return
}

func (tr *ZCacheInt) Set(name string, item int) {
  index := hash(name)
  indexFix := index % mapSize
  if tr.MFlag[indexFix] == 1 || tr.M[indexFix] != tr.MEmpty {
    tr.MFlag[indexFix] = 1
    tr.MBack.Store(name, item)
  } else {
    tr.M[indexFix] = item
  }
}

func (tr *ZCacheInt) Get(name string) (result int) {
  result = 0
  index := hash(name)
  indexFix := index % mapSize
  if tr.MFlag[indexFix] == 1 {
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

// ZCache String Type
type ZCacheString struct {
  M [mapSize]string
  MFlag [mapSize]int
  MBack *MapOf[string, string]
  MEmpty string
}

func ZCacheStringCreate() (tr *ZCacheString) {
  tr = new(ZCacheString)
  tr.MBack = new(MapOf[string, string])
  tr.MEmpty = ""
  return
}

func (tr *ZCacheString) Set(name string, item string) {
  index := hash(name)
  indexFix := index % mapSize
  if tr.MFlag[indexFix] == 1 || tr.M[indexFix] != tr.MEmpty {
    tr.MFlag[indexFix] = 1
    tr.MBack.Store(name, item)
  } else {
    tr.M[indexFix] = item
  }
}

func (tr *ZCacheString) Get(name string) (result string) {
  index := hash(name)
  indexFix := index % mapSize
  if tr.MFlag[indexFix] == 1 {
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
