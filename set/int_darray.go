package set

type intDArray struct {
	useMap  bool
	array   []uint8
	arrSize int
	m       map[int]uint8
}

func newIntDArray(arrSize int) *intDArray {
	return &intDArray{
		useMap:  false,
		array:   make([]uint8, arrSize),
		arrSize: arrSize,
	}
}

func (arr *intDArray) Add(pos int, val uint8) (newVal uint8) {
	if pos < arr.arrSize {
		arr.array[pos] += val
		return val
	}
	if !arr.useMap {
		arr.useMap = true
		arr.m = make(map[int]uint8, 8)
	}
	newVal = arr.m[pos]
	newVal += val
	arr.m[pos] = newVal
	return
}

func (arr *intDArray) Set(pos int, val uint8) (oldVal uint8) {
	if pos < arr.arrSize {
		oldVal = arr.array[pos]
		arr.array[pos] = val
		return
	}
	if !arr.useMap {
		arr.useMap = true
		arr.m = make(map[int]uint8, 8)
		oldVal = 0
	} else {
		oldVal = arr.m[pos]
	}
	arr.m[pos] = val
	return
}

func (arr *intDArray) Get(pos int) uint8 {
	if pos < arr.arrSize {
		return arr.array[pos]
	}
	if arr.useMap {
		return arr.m[pos]
	}
	return 0
}
