package set

import (
	"unsafe"
)

var _x int
var intSize int = int(unsafe.Sizeof(_x))
var arrSize int = 1024 / intSize

type intDArray struct {
	useMap bool
	array  []int
	m      map[int]int
}

func newIntDArray() *intDArray {
	return &intDArray{
		useMap: false,
		array:  make([]int, arrSize),
	}
}

func (arr *intDArray) Add(pos, val int) (newVal int) {
	if pos < arrSize {
		arr.array[pos] += val
		return val
	}
	if !arr.useMap {
		arr.useMap = true
		arr.m = make(map[int]int, 8)
	}
	newVal = arr.m[pos]
	newVal += val
	arr.m[pos] = newVal
	return
}

func (arr *intDArray) Set(pos, val int) (oldVal int) {
	if pos < arrSize {
		oldVal = arr.array[pos]
		arr.array[pos] = val
		return
	}
	if !arr.useMap {
		arr.useMap = true
		arr.m = make(map[int]int, 8)
		oldVal = 0
	} else {
		oldVal = arr.m[pos]
	}
	arr.m[pos] = val
	return
}

func (arr *intDArray) Get(pos int) int {
	if pos < arrSize {
		return arr.array[pos]
	}
	if arr.useMap {
		return arr.m[pos]
	}
	return 0
}
