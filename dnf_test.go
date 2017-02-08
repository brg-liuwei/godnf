package godnf_test

import (
	"fmt"
	"strconv"
	"testing"

	dnf "github.com/brg-liuwei/godnf"
)

type attr struct {
	docId   int
	docName string
}

func (this attr) ToString() string {
	return fmt.Sprintf("( %d -> %s )",
		this.docId, this.docName)
}

func (this attr) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"DocId":   this.docId,
		"DocName": this.docName,
	}
}

func DocFilter(attr dnf.DocAttr) bool { return true }

var dnfDesc []string = []string{
	"(region in {SH, BJ } and age not in {3,4} )",                  // docid: 0
	"(region in { HZ , SZ } and gender in { male })",               // docid: 1
	"(region not in {WH, BJ} and age in {4, 5})",                   // docid: 2
	"(region in {CD, BJ} and age in {3} and gender in { female })", // docid: 3
	"(region in {GZ, SH} and age in {4})",                          // docid: 4
	"(region in {BJ} and age in {3, 4 ,5})",                        // docid: 5
	"(region not in {CD} and age not in {3})",                      // docid: 6
	"(gender in {male} and age not in {2, 3, 4})",                  // docid: 7
	"(region in {SH, BJ, CD, GZ} and age in {2, 3})",               // docid: 8
	"(region not in {SH, BJ} and age not in {4})",                  // docid: 9
	"(OS in {Windows, MacOS} and region not in {SH})",              // docid: 10
}

var conds []dnf.Cond = []dnf.Cond{
	{Key: "region", Val: "BJ"},
	{Key: "age", Val: "3"},
	{Key: "OS", Val: "MacOS"},
}

func createDnfHandler(descs []string, useLock bool) *dnf.Handler {
	dnf.SetDebug(true)

	var h *dnf.Handler
	if useLock {
		h = dnf.NewHandler()
	} else {
		h = dnf.NewHandlerWithoutLock()
	}

	for i, desc := range descs {
		name := "doc-" + strconv.Itoa(i)
		err := h.AddDoc(name, strconv.Itoa(i), desc, attr{
			docName: name,
			docId:   i,
		})
		if err != nil {
			panic("AddDoc[" + strconv.Itoa(i) + "] err:" + err.Error())
		}
	}
	return h
}

func ExampleRetrieval() {
	retrievalHelper := func(h *dnf.Handler) {
		docs, err := h.Search(conds, DocFilter)
		if err != nil {
			panic(err)
		}
		for _, doc := range docs {
			attr, err := h.DocId2Attr(doc)
			if err != nil {
				panic(err)
			}
			fmt.Println(attr.ToString())
		}
	}

	h := createDnfHandler(dnfDesc, true)

	fmt.Println("before delete:")
	// expected result: 5, 8, 10
	retrievalHelper(h)

	h.DeleteDoc("5")
	fmt.Println("after delete [5]:")
	// expected result: 8, 10
	retrievalHelper(h)

	h.DeleteDoc("10")
	// expected result: 8
	fmt.Println("after delete [10]:")
	retrievalHelper(h)

	// Output:
	// before delete:
	// ( 5 -> doc-5 )
	// ( 8 -> doc-8 )
	// ( 10 -> doc-10 )
	// after delete [5]:
	// ( 8 -> doc-8 )
	// ( 10 -> doc-10 )
	// after delete [10]:
	// ( 8 -> doc-8 )
}

func BenchmarkRetrieval(b *testing.B) {
	h := createDnfHandler(dnfDesc, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := h.Search(conds, DocFilter)
		if err != nil {
			b.Error("Search error: ", err)
		}
	}

	b.ReportAllocs()
}

func BenchmarkRetrievalWithoutLock(b *testing.B) {
	h := createDnfHandler(dnfDesc, false)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := h.Search(conds, DocFilter)
		if err != nil {
			b.Error("Search error: ", err)
		}
	}

	b.ReportAllocs()
}

func BenchmarkParallelRetrieval(b *testing.B) {
	h := createDnfHandler(dnfDesc, true)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := h.Search(conds, DocFilter)
			if err != nil {
				b.Error("Search error: ", err)
			}
		}
	})

	b.ReportAllocs()
}

func BenchmarkParallelRetrievalWithoutLock(b *testing.B) {
	h := createDnfHandler(dnfDesc, false)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := h.Search(conds, DocFilter)
			if err != nil {
				b.Error("Search error: ", err)
			}
		}
	})

	b.ReportAllocs()
}

/*
test delim:

    lconj, rconj := GetDelimOfConj()
    lset, rset := GetDelimOfSet()
    SetDelimOfConj('<', '>')
    SetDelimOfConj('[', ']')

    // run test func

    SetDelimOfConj(lconj, rconj)
    SetDelimOfConj(lset, rset)
*/
var dnfDescWithCustomizedDelim []string = []string{
	"< region in [SH. BJ ]   and age not in [3] >",
	"< region in [ HZ. SZ ] and gender in [ male ] >",
	"< region not in [ WH. BJ ] and age in [ 4. 5 ] >",
	"< region in [ CD. BJ ] and age in [ 3 ] and gender in [ female ] >",
	"< region in [ GZ. SH ] and age in [ 4 ] >",
	"< region in [ BJ ] and age in [ 3. 4 .5 ] >",
	"< region not in [ CD ] and age not in [ 3 ] >",
	"< gender in [ male ] and age not in [ 2. 3. 4 ] >",
	"< region in [ SH. BJ. CD. GZ ] and age in [ 2. 3 ] >",
	"< region not in [ SH. BJ ] and age not in [ 4 ] >",
	"< OS in [ Windows. MacOS ] and region not in [ SH ] >",
	"< size in [300*250] and adx in [1,2,3] and region in [1156110000,1156120000,1156130000,1156140000,1156150000,1156210000,1156220000,1156230000,1156310000,1156330000,1156340000,1156350000,1156360000,1156370000]>",
}

func ExampleRetrievalWithCustomizedDelim() {
	lconj, rconj := dnf.GetDelimOfConj()
	lset, rset := dnf.GetDelimOfSet()
	sep := dnf.GetSeparatorOfSet()

	dnf.SetDelimOfConj('<', '>')
	dnf.SetDelimOfSet('[', ']')
	dnf.SetSeparatorOfSet('.')

	h := createDnfHandler(dnfDescWithCustomizedDelim, true)

	dnf.SetDelimOfSet(lset, rset)
	dnf.SetDelimOfConj(lconj, rconj)
	dnf.SetSeparatorOfSet(sep)

	docs, err := h.Search(conds, DocFilter)
	if err != nil {
		panic(err)
	}
	for _, doc := range docs {
		attr, err := h.DocId2Attr(doc)
		if err != nil {
			panic(err)
		}
		fmt.Println(attr.ToString())
	}
	// Output:
	// ( 5 -> doc-5 )
	// ( 8 -> doc-8 )
	// ( 10 -> doc-10 )
}
