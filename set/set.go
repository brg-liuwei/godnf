package set

import (
	"sort"
	"sync"
)

/*
   count set:

       if an elem with enough positive count(>= set.count) and with no negetive count,
       we define this elem is in the set
*/
type CountSet struct {
	sync.RWMutex
	count    int
	positive *intDArray // use intDArray instead of golang native map for better performance
	negetive *intDArray
	result   *intDArray
}

/* create a count set */
func NewCountSet(count int) *CountSet {
	return &CountSet{
		count:    count,
		positive: newIntDArray(),
		negetive: newIntDArray(),
		result:   newIntDArray(),
	}
}

/* add an elem into set */
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

/* return a int slice contains all elems(with enough count) of set in order */
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

/* IntSet: a set whose elems are integer */
type IntSet struct {
	sync.RWMutex
	data map[int]bool
}

/* create an int set */
func NewIntSet() *IntSet {
	return &IntSet{data: make(map[int]bool)}
}

/* add an elem into set */
func (set *IntSet) Add(elem int, useMutex bool) {
	if useMutex {
		set.Lock()
		defer set.Unlock()
	}
	set.data[elem] = true
}

/* add elems into set */
func (set *IntSet) AddSlice(elems []int, useMutex bool) {
	if useMutex {
		set.Lock()
		defer set.Unlock()
	}
	for _, elem := range elems {
		set.data[elem] = true
	}
}

/* return a int slice contains all elems of set in order */
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
