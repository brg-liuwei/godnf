package set

import (
	"sort"
	"sync"
)

type CountSet struct {
	sync.RWMutex
	count    int
	positive map[int]int
	negetive map[int]int
	result   map[int]bool
}

func NewCountSet(count int) *CountSet {
	return &CountSet{
		count:    count,
		positive: make(map[int]int),
		negetive: make(map[int]int),
		result:   make(map[int]bool),
	}
}

func (set *CountSet) Add(id int, post bool, useMutex bool) {
	if useMutex {
		set.Lock()
		defer set.Unlock()
	}

	if !post {
		set.negetive[id] = 1
	} else {
		val := set.positive[id]
		val++
		if val >= set.count {
			set.result[id] = true
		} else {
			set.positive[id] = val
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
	for k, _ := range set.negetive {
		if _, ok := set.result[k]; ok {
			delete(set.result, k)
		}
	}
	unlock()

	rlock()
	rc := make([]int, 0, len(set.result))
	for k, _ := range set.result {
		rc = append(rc, k)
	}
	runlock()

	if !sort.IntsAreSorted(rc) {
		sort.IntSlice(rc).Sort()
	}
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
