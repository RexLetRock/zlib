package main

import (
	"fmt"
	"os"
	"os/signal"
	// "time"

  "github.com/RexLetRock/zlib/ztime"
)

type source struct {
	title string
	url   string
}

func main() {
	numWorker := 100

	// Puller
	sources := make(chan source)
  for index := 0; index < numWorker+1; index++ {
		go func(_ int) {
		  for { sources <- source{"test", "test"} }
    }(index)
  }

	// Consumers
	results := make(chan string)
	for index := 0; index < numWorker+1; index++ {
		go func(i int) {
			for s := range sources {
				results <- fmt.Sprintf("%d, %s", i, s.title)
			}
		}(index)
	}

	// Safe interrupt
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		close(sources)
		close(results)
	}()

	// Print the result
  n := int64(0)
  time := ztime.UnixNanoNow()
  for index := 0; index < numWorker+1; index++ {
		go func(_ int) {
    	for r := range results {
        if r != "" {
          n += 1
          if n % 100000 == 0 {
            nTime := ztime.UnixNanoNow()
            fmt.Printf("%v (msg/s) - Total %v - Txt %v \n", 1000 * n/((nTime - time)/1_000_000), n, r)
          }
        }
    	}
    }(index)
  }

  for r := range results {
    if r != "" {
      n += 1
      if n % 100000 == 0 {
        nTime := ztime.UnixNanoNow()
        fmt.Printf("%v (msg/s) - Total %v - Txt %v \n", 1000 * n/((nTime - time)/1_000_000), n, r)
      }
    }
  }
}
