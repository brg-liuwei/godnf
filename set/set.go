package set

import (
	"sort"
	"sync"
)

const maxArraySize = 4096

// A CountSet is a set if an elem with enough positive count(>= set.count)
// and with no negetive count, we define this elem is in the set
type CountSet struct {
	sync.RWMutex
	count    uint8
	positive *intDArray // use intDArray instead of golang native map for better performance
	negetive *intDArray
	result   *intDArray
}

// NewCountSet creates a count set with positive `count`
func NewCountSet(count uint8, arraySize ...int) *CountSet {
	var arrSize int
	if len(arraySize) > 0 {
		arrSize = arraySize[0]
	} else {
		arrSize = 256
	}
	return &CountSet{
		count:    count,
		positive: newIntDArray(arrSize),
		negetive: newIntDArray(arrSize),
		result:   newIntDArray(arrSize),
	}
}

// Add an elem into set
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

// ToSlice returns a CountSet slice contains all elems(with enough count) of set in order
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
	for pos := range set.negetive.m {
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
	for pos := range set.result.m {
		rc = append(rc, pos)
	}
	runlock()

	// rc is a sorted array, absolutely
	return rc
}

// IntSet: a set whose elems are integer
type IntSet struct {
	sync.RWMutex
	data map[int]bool
}

// NewIntSet creates an int set
func NewIntSet() *IntSet {
	return &IntSet{data: make(map[int]bool)}
}

// Add an elem into set
func (set *IntSet) Add(elem int, useMutex bool) {
	if useMutex {
		set.Lock()
		defer set.Unlock()
	}
	set.data[elem] = true
}

// Add elems into set
func (set *IntSet) AddSlice(elems []int, useMutex bool) {
	if useMutex {
		set.Lock()
		defer set.Unlock()
	}
	for _, elem := range elems {
		set.data[elem] = true
	}
}

// ToSlice returns a int slice contains all elems of set in order
func (set *IntSet) ToSlice(useMutex bool) []int {
	var rlock, runlock func()
	if useMutex {
		rlock, runlock = set.RLock, set.RUnlock
	} else {
		rlock, runlock = func() {}, func() {}
	}

	rlock()
	rc := make([]int, 0, len(set.data))
	for k := range set.data {
		rc = append(rc, k)
	}
	runlock()

	if !sort.IntsAreSorted(rc) {
		sort.IntSlice(rc).Sort()
	}
	return rc
}
