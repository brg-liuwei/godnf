package set_test

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"

	set "github.com/brg-liuwei/godnf/set"
)

func TestIntSet(t *testing.T) {
	s := set.NewIntSet()

	s.Add(1, false)
	s.Add(2, false)
	s.Add(3, false)

	s.AddSlice([]int{3, 4, 5}, false)

	expected := []int{1, 2, 3, 4, 5}

	slice := s.ToSlice(false)
	if len(expected) != len(slice) {
		t.Error("slice size error")
	}
	for i := range slice {
		if slice[i] != expected[i] {
			t.Errorf("slice[%d] error", i)
		}
	}

	slice2 := s.ToSlice(true)
	if len(expected) != len(slice2) {
		t.Error("slice2 size error")
	}
	for i := range slice2 {
		if slice2[i] != expected[i] {
			t.Errorf("slice2[%d] error", i)
		}
	}
}

func BenchmarkIntSetSerial(b *testing.B) {
	s := set.NewIntSet()
	for i := 0; i < b.N; i++ {
		s.Add(i, false)
	}
	b.ReportAllocs()
}

func BenchmarkIntParallel(b *testing.B) {
	s := set.NewIntSet()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Add(rand.Intn(b.N), true)
		}
	})
	b.ReportAllocs()
}

func BenchmarkIntParallelWithNCpu(b *testing.B) {
	s := set.NewIntSet()
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

	for i := 0; i < b.N; i++ {
		ch <- i
	}
	close(ch)

	wg.Wait()
	b.ReportAllocs()
}

func TestCountSet(t *testing.T) {
	c := set.NewCountSet(3)

	c.Add(10, true, false)
	c.Add(10, true, false)
	c.Add(10, true, false)

	c.Add(9, false, false)

	for i := 100; i != 2000; i++ {
		c.Add(i, true, false)
	}

	slice := c.ToSlice(false)
	if len(slice) != 1 && slice[0] != 10 {
		t.Error("test count set fail")
	}
}

func TestCountSetNegetive(t *testing.T) {
	c := set.NewCountSet(3)

	for i := 0; i != 2000; i++ {
		// disable all slot
		c.Add(i, false, false)
	}

	for i := 0; i != 100; i++ {
		c.Add(rand.Intn(2000), true, false)
	}

	slice := c.ToSlice(false)
	if len(slice) != 0 {
		t.Error("test count set fail")
	}
}

func BenchmarkCountSetSerial(b *testing.B) {
	s := set.NewCountSet(2)
	for i := 0; i < b.N; i++ {
		s.Add(i, true, false)
		s.Add(i, true, false)
	}
	b.ReportAllocs()
}

func BenchmarkCountSetParallel(b *testing.B) {
	s := set.NewCountSet(3)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Add(rand.Intn(10), true, true)
			s.Add(rand.Intn(10), true, true)
		}
	})
	b.ReportAllocs()
}
