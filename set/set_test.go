package set_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	set "github.com/brg-liuwei/godnf/set"
)

func runIntSetSerial(s *set.IntSet, loop int) time.Duration {
	now := time.Now()
	for i := 0; i < loop; i++ {
		s.Add(i, false)
	}
	return time.Since(now)
}

func runIntSetParallel(s *set.IntSet, loop int) time.Duration {
	var wg sync.WaitGroup
	wg.Add(loop)
	now := time.Now()
	for i := 0; i < loop; i++ {
		go func(i int, wg *sync.WaitGroup) {
			s.Add(i, true)
			wg.Done()
		}(i, &wg)
	}
	wg.Wait()
	return time.Since(now)
}

func runIntSetParallelWithCpuCount(s *set.IntSet, loop int) time.Duration {
	var wg sync.WaitGroup
	f := func(s *set.IntSet, ch <-chan int, wg *sync.WaitGroup) {
		for n := range ch {
			s.Add(n, true)
		}
		wg.Done()
	}
	ch := make(chan int, 1024)
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go f(s, ch, &wg)
	}

	runtime.Gosched()

	now := time.Now()
	for i := 0; i < loop; i++ {
		ch <- i
	}
	close(ch)

	wg.Wait()
	return time.Since(now)
}

func TestRunCase(t *testing.T) {
	loop := 1
	for i := 0; i < 6; i++ {
		loop = loop * 4
		t.Log("=== loop: ", loop)
		s1, s2, s3 := set.NewIntSet(), set.NewIntSet(), set.NewIntSet()
		t.Log("     serial: ", runIntSetSerial(s1, loop))
		t.Log("     parallel: ", runIntSetParallel(s2, loop))
		t.Log("     useChannel: ", runIntSetParallelWithCpuCount(s3, loop))
	}
}
