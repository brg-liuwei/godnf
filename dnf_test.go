package godnf_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

	h.DeleteDoc("5", "1")
	fmt.Println("after delete [5]:")
	// expected result: 8, 10
	retrievalHelper(h)

	h.DeleteDoc("10", "2")
	// expected result: 8
	fmt.Println("after delete [10]:")
	retrievalHelper(h)

	if h.DeleteDoc("11", "3") {
		fmt.Println("delete un-exist doc, expected return false")
	}

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

func ExampleDumpByPage() {
	h := createDnfHandler(dnfDesc, false)

	var m map[string]interface{}
	json.Unmarshal(h.DumpByPage(10, 10, func(dnf.DocAttr) bool { return true }), &m) // start out of range
	fmt.Println(m["total_records"], len(m["data"].([]interface{})))

	json.Unmarshal(h.DumpByPage(1, 0, func(dnf.DocAttr) bool { return true }), &m) // page size 0: dump all
	fmt.Println(m["total_records"], len(m["data"].([]interface{})))

	json.Unmarshal(h.DumpByPage(1, 5, func(dnf.DocAttr) bool { return true }), &m) // from page 1, page_size 5, return 5 elems
	fmt.Println(m["total_records"], len(m["data"].([]interface{})))

	json.Unmarshal(h.DumpByPage(3, 4, func(dnf.DocAttr) bool { return true }), &m) // page num 5: page_size 4, 11 elem totally, return last 3 elems
	fmt.Println(m["total_records"], len(m["data"].([]interface{})))

	// Output:
	// 11 0
	// 11 11
	// 11 5
	// 11 3
}

func ExampleDumpByFilter() {
	h := createDnfHandler(dnfDesc, false)

	var m map[string]interface{}
	json.Unmarshal(h.DumpByFilter(func(docAttr dnf.DocAttr) bool {
		a := docAttr.(attr)
		if a.docId < 3 {
			return true
		}
		return false
	}), &m)
	fmt.Println(m["total_records"])

	// Output:
	// 3
}

func ExampleDumpById() {
	h := createDnfHandler(dnfDesc, false)

	var slice []interface{}
	json.Unmarshal(h.DumpById(), &slice)
	m2 := slice[2].(map[string]interface{})
	fmt.Println(m2["name"])
	fmt.Println(m2["docid"])

	// Output:
	// doc-2
	// 2
}

func ExampleDumpByDocId() {
	h := createDnfHandler(dnfDesc, false)

	var m map[string]interface{}
	json.Unmarshal(h.DumpByDocId(), &m)
	m2 := m["2"].(map[string]interface{})
	fmt.Println(m2["id"])
	fmt.Println(m2["name"])

	// Output:
	// 2
	// doc-2
}

func ExampleDumpByName() {
	h := createDnfHandler(dnfDesc, false)

	var m map[string]interface{}
	json.Unmarshal(h.DumpByName(), &m)
	mm := m["doc-2"].(map[string]interface{})
	fmt.Println(mm["id"])
	fmt.Println(mm["docid"])
	fmt.Println(mm["active"])
	fmt.Println(mm["dnf"])

	// Output:
	// 2
	// 2
	// true
	// (region not in {WH, BJ} and age in {4, 5})
}

func ExampleDisplayMetaData() {
	// dnfDesc[:2]:
	// ( region ∈ { SH, BJ } ∩ age ∉ { 3, 4 } )
	// ( region ∈ { HZ, SZ } ∩ gender ∈ { male } )
	h := createDnfHandler(dnfDesc[:2], false)
	dnf.SetDebug(true)

	h.DeleteDoc("0", "4")
	h.DisplayDocs()
	h.DisplayConjs()
	h.DisplayAmts()
	h.DisplayTerms()

	// Output:
	// len(docs): 2
	// Doc[ 0 ](del): ( region ∈ { SH, BJ } ∩ age ∉ { 3, 4 } )
	// Attr: ( 0 -> doc-0 )
	// Doc[ 1 ]: ( region ∈ { HZ, SZ } ∩ gender ∈ { male } )
	// Attr: ( 1 -> doc-1 )
	// Conj[ 0 ] size: 1 , ( region ∈ { SH, BJ } ∩ age ∉ { 3, 4 } )
	// Conj[ 1 ] size: 2 , ( region ∈ { HZ, SZ } ∩ gender ∈ { male } )
	// Amt[ 0 ]: region ∈ { SH, BJ }
	// Amt[ 1 ]: age ∉ { 3, 4 }
	// Amt[ 2 ]: region ∈ { HZ, SZ }
	// Amt[ 3 ]: gender ∈ { male }
	// Term[ 0 ]: ∅
	// Term[ 1 ]: ( region  SH )
	// Term[ 2 ]: ( region  BJ )
	// Term[ 3 ]: ( age  3 )
	// Term[ 4 ]: ( age  4 )
	// Term[ 5 ]: ( region  HZ )
	// Term[ 6 ]: ( region  SZ )
	// Term[ 7 ]: ( gender  male )
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

func TestTooLargeConjunctions(t *testing.T) {
	terms := make([]string, 0, 256)
	for i := 0; i != 255; i++ {
		terms = append(terms, fmt.Sprintf("key-%d in {val-%d}", i, i))
	}
	rightDnf := "(" + strings.Join(terms, " and ") + ")"

	terms = append(terms, "key-256 in {val-256}")
	wrongDnf := "(" + strings.Join(terms, " and ") + ")"

	h := dnf.NewHandlerWithoutLock()
	if err := h.AddDoc("rightDnf", "0", rightDnf, attr{0, ""}); err != nil {
		t.Error("unexpected error when AddDoc: ", err)
	}
	if err := h.AddDoc("wrongDnf", "1", wrongDnf, attr{1, ""}); err == nil {
		t.Error("Test too large conjunctions fail")
	} else if err.Error() != "conjunction size too large(max: 255)" {
		t.Error("unexpected error message: ", err)
	}
}

func TestDocAdded(t *testing.T) {
	h := dnf.NewHandlerWithoutLock()
	if err := h.AddDoc("doc-0", "0", dnfDesc[0], attr{0, "doc-0"}); err != nil {
		t.Error("unexpected error when AddDoc: ", err)
	}
	// re-add
	if err := h.AddDoc("doc-0", "0", dnfDesc[0], attr{0, "doc-0"}); err == nil {
		t.Error("Test add duplicate doc fail")
	} else if err.Error() != "doc 0 has been added before" {
		t.Error("unexpected error message: ", err)
	}
}
