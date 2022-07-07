package main

import (
	"bufio"
	"fmt"
	"os"
	"sync/atomic"

	// "github.com/RexLetRock/zlib/zlog"
	"github.com/RexLetRock/zlib/zbench"
	"github.com/RexLetRock/zlib/zgoid"
	// "github.com/RexLetRock/zlib/extra/model"
)

const (
	BufferSize = 10_000
	CpuMax     = 13
	CpuNameMax = 50_000
)

type QueueShard struct {
	iCpuCount [CpuMax]int
	iCpuMap   [CpuNameMax]int64
	Buffer    [BufferSize][CpuMax]int64
	n         count32
}

func (tr *QueueShard) Reset() {
	tr.iCpuCount = [CpuMax]int{}
	// tr.iCpuMap = [CpuNameMax]int64{}
	// tr.Buffer = [BufferSize][CpuMax]int64{}
	tr.n.reset()
}

func (tr *QueueShard) Set(i int64) (isFlip bool) {
	isFlip = false
	id := zgoid.Get()
	if tr.iCpuMap[id] == 0 {
		tr.iCpuMap[id] = tr.n.inc()
	}
	idReal := tr.iCpuMap[id]

	tr.iCpuCount[idReal] += 1
	if tr.iCpuCount[idReal] >= BufferSize {
		isFlip = true
		fmt.Printf("CPU-R %v - CPU %v - COUNT %v \n", idReal, id, tr.iCpuCount[idReal])
		tr.iCpuCount[idReal] = 0
	}

	tr.Buffer[tr.iCpuCount[idReal]][idReal] = i
	return
}

type Queue struct {
	M1 QueueShard
	M2 QueueShard
	M3 QueueShard
	pM *QueueShard
	n  count32
}

func QueueCreate() *Queue {
	tr := new(Queue)
	tr.M1 = QueueShard{}
	tr.M2 = QueueShard{}
	tr.M3 = QueueShard{}
	tr.pM = &(tr.M1)
	return tr
}

func (tr *Queue) Flip() {
	tmp := tr.n.inc()
	// TMP = 0
	if tmp > 1 {
		tmp = 0
		tr.n.set(tmp)
		// tr.M2.Reset()
		tr.pM = &tr.M1
		// TMP = 1
	} else {
		// tr.M1.Reset()
		tr.pM = &tr.M2
	}
	fmt.Printf("FLIP INFO %v %v \t%p \n", len(tr.M1.Buffer), len(tr.M2.Buffer), tr.pM)
}

func (tr *Queue) Set(i int64) (isFlip bool) {
	return tr.pM.Set(i)
}

var (
	NRun = 100_311_123
	NCpu = 12
	M    = QueueCreate()
)

type count32 int64

func (c *count32) inc() int64 {
	return atomic.AddInt64((*int64)(c), 1)
}
func (c *count32) get() int64 {
	return atomic.LoadInt64((*int64)(c))
}
func (c *count32) set(item int64) int64 {
	return atomic.SwapInt64((*int64)(c), int64(item))
}
func (c *count32) reset() int64 {
	return atomic.SwapInt64((*int64)(c), 0)
}

func main() {
	benchZID()

	fmt.Printf("\n\nStop with ctrl + c \n\n")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}

func benchZID() {
	fmt.Printf("\n\n=== PARALLEL WRITE ===\n")
	fmt.Printf("\n== RUN %v threads\n", NCpu)
	zbench.Run(NRun, NCpu, func(i, _ int) {
		doBench(i)
	})
}

func doBench(i int) {
	if flip := M.Set(int64(i)); flip {
		M.Flip()
	}
}
