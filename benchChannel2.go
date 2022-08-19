package main

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
)

func main() {
    t := time.Now()
    cs := make([]<-chan int, 1000)
    for i := 0; i < len(cs); i++ {
        cs[i] = generator(rand.Perm(10000)...)
    }
    ch := fanIn(cs...)
    fmt.Println(time.Now().Sub(t))

    is := make([]int, 0, len(ch))
    for v := range ch {
        is = append(is, v)
    }
    fmt.Println("len=", len(is))
}

func generator(nums ...int) <-chan int {
    out := make(chan int, len(nums))
    go func() {
        defer close(out)
        for _, v := range nums {
            out <- v
        }
    }()
    return out
}

func fanIn(in ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int, 10)
    wg.Add(len(in))

    go func() {
        for _, v := range in {
            go func(ch <-chan int) {
                defer wg.Done()
                for val := range ch {
                    out <- val
                }
            }(v)
        }

    }()
    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}
