package set

import (
	"fmt"
)

type sparseArray struct {
	links [][]uint8
	size  int
}

func newSparseArray(size int) *sparseArray {
	nlink := (size / 16)
	if size%16 > 0 {
		nlink += 1
	}
	return &sparseArray{
		links: make([][]uint8, nlink),
		size:  size,
	}
}

func (arr *sparseArray) getPos(flatPos int) (i, j int) {
	if flatPos < 0 || flatPos >= arr.size {
		panic(fmt.Sprintf(
			"sparseArray[%d] out of range, max index: %d\n",
			flatPos, arr.size))
	}
	return flatPos / 16, flatPos % 16
}

func (arr *sparseArray) getPosWithAlloc(flatPos int) (i, j int) {
	i, j = arr.getPos(flatPos)
	if arr.links[i] == nil {
		arr.links[i] = make([]uint8, 16)
	}
	return
}

func (arr *sparseArray) Add(pos int, val uint8) (newVal uint8) {
	i, j := arr.getPosWithAlloc(pos)
	newVal = arr.links[i][j] + val
	arr.links[i][j] = newVal
	return
}

func (arr *sparseArray) Set(pos int, val uint8) (oldVal uint8) {
	i, j := arr.getPosWithAlloc(pos)
	oldVal = arr.links[i][j]
	arr.links[i][j] = val
	return
}

func (arr *sparseArray) Get(pos int) uint8 {
	i, j := arr.getPos(pos)
	if arr.links[i] != nil {
		return arr.links[i][j]
	}
	return uint8(0)
}
