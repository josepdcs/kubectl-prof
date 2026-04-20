package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
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
	// Start pprof HTTP server on port 6060 for profiling
	go func() {
		fmt.Println("Starting pprof server on :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			fmt.Printf("pprof server error: %v\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go fastFunction(&wg)
	go slowFunction(&wg)
	wg.Wait()
}
