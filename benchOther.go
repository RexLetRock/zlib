package main

import (
	// "sync"
  "time"
  "bytes"
	"bufio"
	"fmt"
	"os"
	"strconv"

	"hash/fnv"
  "github.com/segmentio/fasthash/fnv1a"

	_ "github.com/alphadose/zenq"
	"golang.design/x/lockfree"
	// "github.com/kpango/fastime"
	// "github.com/Avalanche-io/sled"
	"github.com/hlts2/gfreequeue"
	tbtree "github.com/tidwall/btree"

	"github.com/kpango/gache"
	ccmap "github.com/orcaman/concurrent-map"
	"github.com/sigurn/crc16"

  "github.com/fengyoulin/shm"
  "github.com/alphadose/haxmap"

	"github.com/RexLetRock/zlib/zbench"
	"github.com/RexLetRock/zlib/zcache"
	"github.com/RexLetRock/zlib/ztime"
)

func newBTreeM() *tbtree.Map[int, string] {
	return new(tbtree.Map[int, string])
}

var (
	NRun = 5_000_000
	NCpu = 12
)

func main() {
	benchZID()

	fmt.Printf("\n\nStop with ctrl + c \n\n")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}

func hashStrToInt(s string) uint32 {
  h := fnv.New32a()
  h.Write([]byte(s))
  return h.Sum32()
}

func benchZID() {
	table := crc16.MakeTable(crc16.CRC16_MAXIM)
	crc := crc16.Checksum([]byte("Hello world!"), table)
	fmt.Printf("CRC-16 MAXIM: %X\n", crc)

	strArr := make([]string, NRun+1)
	for i := 0; i <= NRun; i++ {
		strArr[i] = strconv.Itoa(i)
	}

  fmt.Printf("\n\n=== HAX MAP ===\n")
  haxMap := haxmap.New[int, int]()
  zbench.Run(NRun, NCpu, func(i, _ int) {
    haxMap.Set(1, 123123)
  })

  m, _ := shm.Create("map.db", 4096, 40, 32, 20, time.Second)
  defer m.Close()
  fmt.Printf("\n\n=== SHM MAP ===\n")
  zbench.Run(NRun, NCpu, func(i, _ int) {
    m.Get("1a2b3c4d5e6f", true)
  })

  fmt.Printf("\n\n=== JOIN BYTES ===\n")
  name := [][]byte{[]byte("Sumit"), []byte("Kumar")}
  sep := []byte("-")
  zbench.Run(NRun, NCpu, func(i, _ int) {
    _ = bytes.Join(name, sep)
  })
  zbench.Run(NRun, NCpu, func(i, _ int) {
    _ = "Sumit" + "Kumar"
  })
  zbench.Run(NRun, NCpu, func(i, _ int) {
    _ = []byte("Sumit - Kumar ¶")
  })

	fmt.Printf("\n\n=== GCACHE ===\n")
	ExGache := gache.GetGache()
	zbench.Run(NRun, NCpu, func(i, _ int) {
		ExGache.Set(strArr[i], i)
	})

	fmt.Printf("\n\n=== LOCKFREE ===\n")
	q := lockfree.NewQueue()
	zbench.Run(NRun, NCpu, func(i, _ int) {
		q.Enqueue(i)
	})

	fmt.Printf("\n\n=== ZTIME ===\n")
	fmt.Printf("Time %v\n", ztime.UnixNanoNow())
	zbench.Run(NRun, NCpu, func(_, _ int) {
		_ = ztime.UnixNanoNow()
	})

	// fmt.Printf("\n\n=== SLED MAP ===\n")
	// sl := sled.New()
	// zbench.Run(NRun, NCpu, func(i, _ int) {
	// 	sl.Set(strconv.Itoa(i), i)
	// })

	fmt.Printf("\n\n=== GFREEQUEUE ===\n")
	ExQueue := gfreequeue.New()
	zbench.Run(NRun, NCpu, func(i, _ int) {
		ExQueue.Enqueue(i)
	})

	fmt.Printf("\n\n=== BTREE M ===\n")
	ttrM := newBTreeM()
	zbench.Run(NRun, 1, func(i, _ int) {
		ttrM.Load(i, "Haha")
	})
	tmpV, _ := ttrM.Get(2)
	fmt.Printf("DATA %v \n", tmpV)

  fmt.Printf("\n\n")
  fmt.Printf("========================\n")
  fmt.Printf("===     FAST  MAP    ===\n")
  fmt.Printf("========================\n")

  zbench.Run(NRun, NCpu, func(i, _ int) {
    fnv1a.HashString32(strArr[i])
	})

  zbench.Run(NRun, NCpu, func(i, _ int) {
    hashStrToInt(strArr[i])
	})

	fmt.Printf("\n\n=== FAST MAP GENERIC ===\n")
	a := 0
	ExZcache := zcache.ZCacheCreate()
	zbench.Run(NRun, NCpu, func(i, _ int) {
    ExZcache.Set(strArr[i], i)
	})
	zbench.Run(NRun, NCpu, func(i, _ int) {
    a := ExZcache.Get(strArr[i])
    if a != nil && a.(int) != i {
			fmt.Printf("Error %v %v %v \n", a, i, strArr[i])
		}
	})
  fmt.Printf("SLOW %v \n", a)

	fmt.Printf("\n\n=== FAST MAP INT ===\n")
	ExZcacheInt := zcache.ZCacheIntCreate()
	zbench.Run(NRun, NCpu, func(i, _ int) {
		in := i + 1
		ExZcacheInt.Set(strArr[in], in)
	})
	zbench.Run(NRun, NCpu, func(i, _ int) {
		in := i + 1
		a := ExZcacheInt.Get(strArr[in])
		if a != in {
			fmt.Printf("Error %v %v %v \n", a, in, strArr[in])
		}
	})

	fmt.Printf("\n\n=== FAST MAP STRING ===\n")
	ExZcacheString := zcache.ZCacheStringCreate()
	zbench.Run(NRun, NCpu, func(i, _ int) {
		ExZcacheString.Set(strArr[i]+"FAST MAP", strArr[i])
	})
	zbench.Run(NRun, NCpu, func(i, _ int) {
		a := ExZcacheString.Get(strArr[i] + "FAST MAP")
		if a != strArr[i] {
			fmt.Printf("Error %v %v %v \n", a, i, strArr[i])
		}
	})
	fmt.Printf("VAL %v \n", ExZcacheString.Get(strArr[10]+"FAST MAP"))

	fmt.Printf("\n\n=== CCMAP ===\n")
	ExMap1 := ccmap.New()
	zbench.Run(NRun, NCpu, func(i, _ int) {
		ExMap1.Set(strArr[i], i)
	})
	zbench.Run(NRun, NCpu, func(i, _ int) {
		ExMap1.Get(strArr[i])
	})

	fmt.Printf("\n\n=== MAP FIX SIZE ===\n")
	ExMap := make(map[int]int, 1_000_000)
	zbench.Run(NRun, 1, func(i, _ int) {
		ExMap[i] = i
	})
}

const trunkNumb = 1_000_000
const trunkSize = 1_000_000

type BigTrunk struct {
	I []int
}

func BigTrunkCreate() (tr *BigTrunk) {
	tr = new(BigTrunk)
	tr.I = make([]int, trunkSize)
	return
}

type BigSlice struct {
	I [trunkNumb](*BigTrunk)
}

func BigSliceCreate() (tr *BigSlice) {
	tr = new(BigSlice)
	tr.I[0] = BigTrunkCreate()
	return
}

func (tr *BigSlice) Set(index int, value int) {
	q, r := 0, index
	if index >= trunkSize {
		q, r = index/trunkSize, index%trunkSize
		if tr.I[q] == nil {
			tr.I[q] = BigTrunkCreate()
		}
	}
	tr.I[q].I[r] = value
}

func (tr *BigSlice) Get(index int) (int, bool) {
	q, r := 0, index
	if index >= trunkSize {
		q, r = index/trunkSize, index%trunkSize
		if tr.I[q] == nil {
			return 0, false
		}
	}
	return tr.I[q].I[r], true
}

func (tr *BigSlice) ZGet(index int) (result int) {
	result, _ = tr.Get(index)
	return
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

type user struct {
	name string
}

type payload struct {
	alpha int
	beta  string
}
