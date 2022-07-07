package main

import (
	"fmt"

	"github.com/RexLetRock/zlib/zbench"
	tbtree "github.com/tidwall/btree"
)

var (
	NRun = 5_000_000
	NCpu = 12
)

func main() {
	benchZID()
}

func benchZID() {
	fmt.Printf("\n\n=== BTREE ===\n")
	print("tidwall(M): load-seq \n")
	ttrM := newBTreeM()
	zbench.Run(NRun, 1, func(i, _ int) {
		ttrM.Load(i, i)
	})

	print("\ntidwall(M): set \n")
	ttrM = newBTreeM()
	zbench.Run(NRun, 1, func(i, _ int) {
		ttrM.Set(i, i)
	})

	print("\n")
}

func newBTreeM() *tbtree.Map[int, int] {
	return new(tbtree.Map[int, int])
}
