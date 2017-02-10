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
	positive *sparseArray // use sparseArray instead of golang native map for better performance
	negetive *sparseArray
	result   *sparseArray
}

// NewCountSet creates a count set with positive `count`
func NewCountSet(count uint8, setSize int) *CountSet {
	return &CountSet{
		count:    count,
		positive: newSparseArray(setSize),
		negetive: newSparseArray(setSize),
		result:   newSparseArray(setSize),
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
	for i := range set.negetive.links {
		for j, val := range set.negetive.links[i] {
			if val > 0 {
				set.result.Set(i*16+j, 0)
			}
		}
	}
	unlock()

	rlock()
	rc := make([]int, 0, 8)
	for i := range set.result.links {
		for j, val := range set.result.links[i] {
			if val > 0 {
				rc = append(rc, i*16+j)
			}
		}
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
