package set

import (
	"fmt"
)

var bit [8]uint8

func init() {
	for i := 0; i != 8; i++ {
		bit[i] = 0x1 << uint8(7-i)
	}
}

type boolBlock uint8

func (bb boolBlock) ToSlice() (slice []bool) {
	slice = make([]bool, 8)
	for i := 0; i != 8; i++ {
		slice[i] = (bit[i] & uint8(bb)) != 0
	}
	return
}

type sparseBoolArray struct {
	links    [][]boolBlock
	size     int
	realSize int
}

func newSparseBoolArray(size int) *sparseBoolArray {
	realSize := size >> 3
	if size&0x7 != 0 {
		realSize += 1
	}
	nlink := realSize >> 4
	if realSize&0xF != 0 {
		nlink += 1
	}
	return &sparseBoolArray{
		links:    make([][]boolBlock, nlink),
		size:     size,
		realSize: realSize,
	}
}

func (arr *sparseBoolArray) getPos(flatPos int) (i, j int) {
	if flatPos < 0 || flatPos >= arr.size {
		panic(fmt.Sprintf(
			"sparseBoolArray[%d] out of range, max index: %d\n",
			flatPos, arr.size))
	}
	realPos := flatPos >> 3
	return realPos >> 4, realPos & 0xF
}

func (arr *sparseBoolArray) getPosWithAlloc(flatPos int) (i, j int) {
	i, j = arr.getPos(flatPos)
	if arr.links[i] == nil {
		arr.links[i] = make([]boolBlock, 16)
	}
	return
}

// func (arr *sparseBoolArray) Get(pos int) bool {
// 	i, j := arr.getPos(pos)
// 	if arr.links[i] != nil {
// 		return (uint8(arr.links[i][j]) & bit[pos&0x7]) != 0
// 	}
// 	return false
// }

func (arr *sparseBoolArray) Set(pos int) (oldVal bool) {
	i, j := arr.getPosWithAlloc(pos)
	oldVal = (uint8(arr.links[i][j]) & bit[pos&0x7]) != 0
	arr.links[i][j] |= boolBlock(bit[pos&0x7])
	return
}

func (arr *sparseBoolArray) Reset(pos int) (oldVal bool) {
	i, j := arr.getPos(pos)
	if arr.links[i] != nil {
		oldVal = (uint8(arr.links[i][j]) & bit[pos&0x7]) != 0
		arr.links[i][j] &= boolBlock(^bit[pos&0x7])
	}
	return
}
