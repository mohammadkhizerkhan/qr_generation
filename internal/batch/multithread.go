package batch

import (
	"runtime"
	"sync"
)

// runWorkers runs fn(i) for i in [0, n) across a bounded worker pool.
// For tiny batches (n == 1 or concurrency == 1) it runs synchronously to
// avoid goroutine overhead. concurrency <= 0 defaults to runtime.NumCPU().
func runWorkers(n, concurrency int, fn func(int)) {
if n == 0 {
return
}
if concurrency <= 0 {
concurrency = runtime.NumCPU()
}
if concurrency > n {
concurrency = n
}
if concurrency == 1 {
for i := 0; i < n; i++ {
fn(i)
}
return
}

jobs := make(chan int, concurrency)
var wg sync.WaitGroup
for w := 0; w < concurrency; w++ {
wg.Add(1)
go func() {
defer wg.Done()
for i := range jobs {
fn(i)
}
}()
}
for i := 0; i < n; i++ {
jobs <- i
}
close(jobs)
wg.Wait()
}
