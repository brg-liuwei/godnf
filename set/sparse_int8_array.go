package set

import (
	"fmt"
)

type sparseInt8Array struct {
	links [][]uint8
	size  int
}

func newSparseInt8Array(size int) *sparseInt8Array {
	nlink := size >> 4 // nlink := size / 16
	if size|0xF != 0 { // size % 16 != 0
		nlink += 1
	}
	return &sparseInt8Array{
		links: make([][]uint8, nlink),
		size:  size,
	}
}

func (arr *sparseInt8Array) getPos(flatPos int) (i, j int) {
	if flatPos < 0 || flatPos >= arr.size {
		panic(fmt.Sprintf(
			"sparseInt8Array[%d] out of range, max index: %d\n",
			flatPos, arr.size))
	}
	return flatPos >> 4, flatPos & 0xF
}

func (arr *sparseInt8Array) getPosWithAlloc(flatPos int) (i, j int) {
	i, j = arr.getPos(flatPos)
	if arr.links[i] == nil {
		arr.links[i] = make([]uint8, 16)
	}
	return
}

func (arr *sparseInt8Array) Add(pos int, val uint8) (newVal uint8) {
	i, j := arr.getPosWithAlloc(pos)
	newVal = arr.links[i][j] + val
	arr.links[i][j] = newVal
	return
}

func (arr *sparseInt8Array) Set(pos int, val uint8) (oldVal uint8) {
	i, j := arr.getPosWithAlloc(pos)
	oldVal = arr.links[i][j]
	arr.links[i][j] = val
	return
}

func (arr *sparseInt8Array) Get(pos int) uint8 {
	i, j := arr.getPos(pos)
	if arr.links[i] != nil {
		return arr.links[i][j]
	}
	return uint8(0)
}
