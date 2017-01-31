package set

import (
	"sort"
	"sync"
)

type CountSet struct {
	sync.RWMutex
	count    int
	positive *IntDArray // use IntDArray instead of golang native map for better performance
	negetive *IntDArray
	result   *IntDArray
}

func NewCountSet(count int) *CountSet {
	return &CountSet{
		count:    count,
		positive: NewIntDArray(),
		negetive: NewIntDArray(),
		result:   NewIntDArray(),
	}
}

func (set *CountSet) Add(id int, post bool, useMutex bool) {
	if useMutex {
		set.Lock()
		defer set.Unlock()
	}

	if !post {
		set.negetive.Set(id, 1)
	} else {
		val := set.positive.Get(id)
		if val+1 >= set.count {
			set.result.Set(id, 1)
		} else {
			set.positive.Add(id, 1)
		}
	}
}

func (set *CountSet) ToSlice(useMutex bool) []int {
	var lock, unlock, rlock, runlock func()
	if useMutex {
		lock, unlock, rlock, runlock = set.Lock, set.Unlock, set.RLock, set.Unlock
	} else {
		nop := func() {}
		lock, unlock, rlock, runlock = nop, nop, nop, nop
	}
	lock()
	for pos, val := range set.negetive.array {
		if val > 0 {
			set.result.Set(pos, 0)
		}
	}
	for pos, _ := range set.negetive.m {
		delete(set.result.m, pos)
	}
	unlock()

	rlock()
	rc := make([]int, 0, 8)
	for pos, val := range set.result.array {
		if val > 0 {
			rc = append(rc, pos)
		}
	}
	for pos, _ := range set.result.m {
		rc = append(rc, pos)
	}
	runlock()

	// rc is a sorted array, absolutely
	return rc
}

type IntSet struct {
	sync.RWMutex
	data map[int]bool
}

func NewIntSet() *IntSet {
	return &IntSet{data: make(map[int]bool)}
}

func (set *IntSet) Add(elem int, useMutex bool) {
	if useMutex {
		set.Lock()
		defer set.Unlock()
	}
	set.data[elem] = true
}

func (set *IntSet) AddSlice(elems []int, useMutex bool) {
	if useMutex {
		set.Lock()
		defer set.Unlock()
	}
	for _, elem := range elems {
		set.data[elem] = true
	}
}

func (set *IntSet) ToSlice(useMutex bool) []int {
	var rlock, runlock func()
	if useMutex {
		rlock, runlock = set.RLock, set.RUnlock
	} else {
		rlock, runlock = func() {}, func() {}
	}

	rlock()
	rc := make([]int, 0, len(set.data))
	for k, _ := range set.data {
		rc = append(rc, k)
	}
	runlock()

	if !sort.IntsAreSorted(rc) {
		sort.IntSlice(rc).Sort()
	}
	return rc
}
