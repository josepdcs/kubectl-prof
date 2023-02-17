package main

import (
	"fmt"
	"sync"
	"time"
)

func work(n, t int, f string) {
	for k := 0; k < n; k++ {
		fmt.Printf("Function: %s, Output: %d\n", f, k)
		time.Sleep(time.Duration(t) * time.Millisecond)
	}
}

func slowFunction(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		work(100, 500, "slow_function")
	}
}

func fastFunction(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		work(100, 50, "fast_function")
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go fastFunction(&wg)
	go slowFunction(&wg)
	wg.Wait()
}
