package godnf

import (
	"fmt"
	"strconv"
	"testing"
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

func DocFilter(attr DocAttr) bool { return true }

var dnfDesc []string = []string{
	"(region in {SH, BJ} and age not in {3, 4})",
	"(region in {HZ, SZ} and sex in { male })",
	"(region not in {WH, BJ} and age in {4, 5})",
	"(region in {CD, BJ} and age in {3} and sex in { female })",
	"(region in {GZ, SH} and age in {4})",
	"(region in {BJ} and age in {3, 4 ,5})",
	"(region not in {CD} and age not in {3})",
	"(sex in {male} and age not in {2, 3, 4})",
	"(region in {SH, BJ, CD, GZ} and age in {2, 3})",
	"(region not in {SH, BJ} and age not in {4})",
	"(OS in {Windows, MacOS} and region not in {SH})",
}

var conds []Cond = []Cond{
	{"region", "BJ"},
	{"age", "3"},
	{"OS", "MacOS"},
}

func createDnfHandler() *Handler {
	h := NewHandler()
	for i, desc := range dnfDesc {
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
	h := createDnfHandler()
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

func BenchmarkRetrieval(b *testing.B) {
	h := createDnfHandler()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := h.Search(conds, DocFilter)
		if err != nil {
			b.Error("Search error: ", err)
		}
	}
}
