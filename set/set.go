package set

import (
	"sort"
	"sync"
)

// A CountSet is a set if an elem with enough positive count(>= set.count)
// and with no negetive count, we define this elem is in the set
type CountSet struct {
	sync.RWMutex
	count    uint8
	positive *sparseInt8Array // use sparseInt8Array instead of golang native map for better performance
	negetive *sparseBoolArray // use sparseBoolArray instead of golang native map for better performance
	result   *sparseBoolArray // use sparseBoolArray instead of golang native map for better performance
}

// NewCountSet creates a count set with positive `count`
func NewCountSet(count uint8, setSize int) *CountSet {
	return &CountSet{
		count:    count,
		positive: newSparseInt8Array(setSize),
		negetive: newSparseBoolArray(setSize),
		result:   newSparseBoolArray(setSize),
	}
}

// Add an elem into set
func (set *CountSet) Add(id int, post bool, useMutex bool) {
	if useMutex {
		set.Lock()
		defer set.Unlock()
	}

	if !post {
		set.negetive.Set(id)
	} else {
		val := set.positive.Get(id)
		if val+1 >= set.count {
			set.result.Set(id)
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
	for i := range set.negetive.links {
		for j, bBlock := range set.negetive.links[i] {
			flatBase := (i<<4 + j) << 3
			for k, bVal := range bBlock.ToSlice() {
				if bVal {
					set.result.Reset(flatBase + k)
				}
			}
		}
	}
	unlock()

	rlock()
	rc := make([]int, 0, 8)
	for i := range set.result.links {
		for j, bBlock := range set.result.links[i] {
			flatBase := (i<<4 + j) << 3
			for k, bVal := range bBlock.ToSlice() {
				if bVal {
					rc = append(rc, flatBase+k)
				}
			}
		}
	}
	runlock()

	// rc is a sorted array, absolutely
	return rc
}

// An IntSet is a set whose elems are integer
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
