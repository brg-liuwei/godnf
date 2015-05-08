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

func (set *CountSet) Add(id int, post bool) {
	set.Lock()
	defer set.Unlock()

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

func (set *CountSet) ToSlice() []int {
	set.Lock()
	for k, _ := range set.negetive {
		if _, ok := set.result[k]; ok {
			delete(set.result, k)
		}
	}
	set.Unlock()

	set.RLock()
	rc := make([]int, 0, len(set.result))
	for k, _ := range set.result {
		rc = append(rc, k)
	}
	set.RUnlock()

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

func (set *IntSet) Add(elem int) {
	set.Lock()
	defer set.Unlock()
	set.data[elem] = true
}

func (set *IntSet) AddSlice(elems []int) {
	set.Lock()
	defer set.Unlock()
	for _, elem := range elems {
		set.data[elem] = true
	}
}

func (set *IntSet) ToSlice() []int {
	set.RLock()
	rc := make([]int, 0, len(set.data))
	for k, _ := range set.data {
		rc = append(rc, k)
	}
	set.RLock()
	if !sort.IntsAreSorted(rc) {
		sort.IntSlice(rc).Sort()
	}
	return rc
}
