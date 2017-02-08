package godnf_test

import (
	"fmt"

	dnf "github.com/brg-liuwei/godnf"
)

type adAttr struct {
	h        int
	w        int
	duration int
}

func (attr adAttr) ToString() string {
	return fmt.Sprintf("{height: %d, width: %d, duration: %d}", attr.h, attr.w, attr.duration)
}

func (attr adAttr) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"height":   attr.h,
		"width":    attr.w,
		"duration": attr.duration,
	}
}

func Example() {
	dnf.SetDebug(true)
	dnf.SetHandler(dnf.NewHandler())
	h := dnf.GetHandler()
	var err error
	err = h.AddDoc("ad0", "0", "(region in {ShangHai, Beijing} and age not in {3, 4})", adAttr{300, 250, 20})
	if err != nil {
		panic(err)
	}
	err = h.AddDoc("ad1", "1", "(region in {ShenZhen, ShangHai}) or (age not in {4, 6})", adAttr{300, 250, 15})
	if err != nil {
		panic(err)
	}
	err = h.AddDoc("ad2", "2", "(region in {ShangHai, NanJing} and age not in {3, 5, 6})", adAttr{300, 250, 10})
	if err != nil {
		panic(err)
	}
	err = h.AddDoc("ad3", "3", "(region in {ChengDu, Beijing, WuHan}) or (age not in {4, 3})", adAttr{300, 250, 30})
	if err != nil {
		panic(err)
	}
	err = h.AddDoc("ad4", "4", "(age not in {3, 4})", adAttr{300, 250, 35})
	if err != nil {
		panic(err)
	}

	conds := []dnf.Cond{
		{"region", "NanJing"},
		{"age", "5"},
	}
	var docs []int
	docs, err = h.Search(conds, func(a dnf.DocAttr) bool { return a.(adAttr).duration <= 30 })
	if err != nil {
		panic(err)
	}
	fmt.Println("docs:", docs)
	for _, doc := range docs {
		fmt.Println(h.DocId2Map(doc))
	}
	h.DisplayDocs()

	fmt.Println()
	h.DisplayConjRevs()
	fmt.Println()
	h.DisplayConjRevs2()

	// Output:
	// docs: [1 3]
	// map[height:300 width:250 duration:15]
	// map[height:300 width:250 duration:30]
	// len docs ==  5
	// Doc[ 0 ]: ( region ∈ { ShangHai, Beijing } ∩ age ∉ { 3, 4 } )
	// Attr: {height: 300, width: 250, duration: 20}
	// Doc[ 1 ]: ( region ∈ { ShangHai, ShenZhen } ) ∪ ( age ∉ { 4, 6 } )
	// Attr: {height: 300, width: 250, duration: 15}
	// Doc[ 2 ]: ( region ∈ { ShangHai, NanJing } ∩ age ∉ { 3, 6, 5 } )
	// Attr: {height: 300, width: 250, duration: 10}
	// Doc[ 3 ]: ( region ∈ { Beijing, ChengDu, WuHan } ) ∪ ( age ∉ { 3, 4 } )
	// Attr: {height: 300, width: 250, duration: 30}
	// Doc[ 4 ]: ( age ∉ { 3, 4 } )
	// Attr: {height: 300, width: 250, duration: 35}
	//
	// reverse list 1:
	// conj[0]: -> 0
	// conj[1]: -> 1
	// conj[2]: -> 1
	// conj[3]: -> 2
	// conj[4]: -> 3
	// conj[5]: -> 3 4
	//
	// reverse list 2:
	// ***** size: 0 *****
	//      ∅  -> (2 ∈) (5 ∈)
	//     ( age  3 ) -> (5 ∉)
	//     ( age  4 ) -> (2 ∉) (5 ∉)
	//     ( age  6 ) -> (2 ∉)
	// ***** size: 1 *****
	//     ( region  ShangHai ) -> (0 ∈) (1 ∈) (3 ∈)
	//     ( region  Beijing ) -> (0 ∈) (4 ∈)
	//     ( age  3 ) -> (0 ∉) (3 ∉)
	//     ( age  4 ) -> (0 ∉)
	//     ( region  ShenZhen ) -> (1 ∈)
	//     ( age  6 ) -> (3 ∉)
	//     ( region  NanJing ) -> (3 ∈)
	//     ( age  5 ) -> (3 ∉)
	//     ( region  ChengDu ) -> (4 ∈)
	//     ( region  WuHan ) -> (4 ∈)
}
