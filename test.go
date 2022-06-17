package main

import (
	"fmt"
	"time"

  "github.com/RexLetRock/zlib/zbench"
	ga "go.linecorp.com/garr/adder"
)

func main() {
  NRun := 10_000_000
	// or ga.DefaultAdder() which uses jdk long-adder as default
	adder := ga.NewLongAdder(ga.JDKAdderType)
  zbench.Run(NRun, 12, func(_, _ int) {
    adder.Add(1)
  })

	time.Sleep(3 * time.Second)

	// get total added value
	fmt.Println(adder.Sum())
}
