package godnf

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var debug bool
var output *log.Logger

var DEBUG func(msg ...interface{})
var ASSERT func(expression bool)

func doDEBUG(msg ...interface{}) {
	if debug {
		output.Println(msg)
	}
}

func noDEBUG(msg ...interface{}) {}

func doASSERT(expression bool) {
	if !(expression) {
		panic("Assert Fail")
	}
}

func noASSERT(expression bool) {}

func SetDebug(flag bool) {
	if debug != flag {
		debug = flag
		setDebug()
	}
}

func setDebug() {
	if debug {
		DEBUG = doDEBUG
		ASSERT = doASSERT
		output = log.New(os.Stderr, "", log.LstdFlags)
	} else {
		DEBUG = noDEBUG
		ASSERT = noASSERT
		output = nil
	}
}

func init() {
	debug = false
	setDebug()
}

func (this *Term) ToString() string {
	if this.id == 0 {
		/* empty set */
		return " ∅ "
	}
	return fmt.Sprintf("( %s  %s )", this.key, this.val)
}

func (this *Amt) ToString(h *Handler) string {
	if len(this.terms) == 0 {
		return ""
	}

	h.terms_.RLock()
	defer h.terms_.RUnlock()

	var key, op string

	if this.belong {
		op = "∈"
	} else {
		op = "∉"
	}
	key = h.terms_.terms[this.terms[0]].key
	s := fmt.Sprintf("%s %s { ", key, op)
	for i, idx := range this.terms {
		s += h.terms_.terms[idx].val
		if i+1 < len(this.terms) {
			s += ", "
		}
	}
	return s + " }"
}

func (this *Conj) ToString(h *Handler) string {
	if len(this.amts) == 0 {
		return ""
	}
	h.amts_.RLock()
	defer h.amts_.RUnlock()
	s := "( "
	for i, idx := range this.amts {
		s += h.amts_.amts[idx].ToString(h)
		if i+1 < len(this.amts) {
			s += " ∩ "
		}
	}
	return s + " )"
}

func (this *Doc) ToString(h *Handler) (s string) {
	if len(this.conjs) == 0 {
		s = "len(conjs == 0)"
	}
	h.conjs_.RLock()
	defer h.conjs_.RUnlock()
	for i, idx := range this.conjs {
		s += h.conjs_.conjs[idx].ToString(h)
		if i+1 < len(this.conjs) {
			s += " ∪ "
		}
	}
	s += "\nAttr: "
	s += this.attr.ToString()
	return
}

func (this *docList) display() {
	this.RLock()
	defer this.RUnlock()
	DEBUG("len docs == ", len(this.docs))
	for i, doc := range this.docs {
		if !doc.active {
			DEBUG("Doc[", i, "](del):", doc.ToString(this.h))
		} else {
			DEBUG("Doc[", i, "]:", doc.ToString(this.h))
		}
	}
}

func (this *docList) docId2Attr(docid int) (DocAttr, error) {
	if len(this.docs) <= docid {
		return nil, errors.New("docid over flow")
	}
	this.RLock()
	defer this.RUnlock()
	doc := &this.docs[docid]
	return doc.attr, nil
}

func (h *Handler) DocId2Attr(docid int) (DocAttr, error) {
	return h.docs_.docId2Attr(docid)
}

func (this *docList) docId2Map(docid int) map[string]interface{} {
	if len(this.docs) <= docid {
		return nil
	}
	this.RLock()
	defer this.RUnlock()
	doc := &this.docs[docid]
	return doc.attr.ToMap()
}

func (h *Handler) DocId2Map(docid int) map[string]interface{} {
	return h.docs_.docId2Map(docid)
}

func (h *Handler) DumpByPage(pageNum, pageSize int) []byte {
	h.docs_.RLock()
	defer h.docs_.RLock()

	totalRcd := len(h.docs_.docs)
	start := (pageNum - 1) * pageSize

	if totalRcd == 0 || start >= totalRcd {
		b, _ := json.Marshal(map[string]interface{}{
			"total_records": 0,
			"data":          []interface{}{},
		})
		return b
	}

	if start < 0 {
		start = 0
	}

	end := start + pageSize
	if end > totalRcd {
		end = totalRcd
	}

	// DumpByPage(0, 0) means dump all
	if pageSize == 0 {
		end = totalRcd
	}

	s := make([]interface{}, 0, len(h.docs_.docs[start:end]))
	for _, doc := range h.docs_.docs[start:end] {
		s = append(s, map[string]interface{}{
			"name":   doc.name,
			"docid":  doc.docid,
			"active": doc.active,
			"dnf":    doc.dnf,
			"attr":   doc.attr.ToMap(),
		})
	}

	b, _ := json.Marshal(map[string]interface{}{
		"total_records": totalRcd,
		"data":          s,
	})
	return b
}

func (h *Handler) DumpByFilter(filter func(DocAttr) bool) []byte {
	var s []interface{}
	for _, doc := range h.docs_.docs {
		if filter(doc.attr) {
			s = append(s, map[string]interface{}{
				"name":   doc.name,
				"docid":  doc.docid,
				"active": doc.active,
				"dnf":    doc.dnf,
				"attr":   doc.attr.ToMap(),
			})
		}
	}
	b, _ := json.Marshal(map[string]interface{}{
		"total_records": len(s),
		"data":          s,
	})
	return b
}

func (h *Handler) DumpById() []byte {
	var s []interface{}
	for _, doc := range h.docs_.docs {
		s = append(s, map[string]interface{}{
			"name":   doc.name,
			"docid":  doc.docid,
			"active": doc.active,
			"dnf":    doc.dnf,
			"attr":   doc.attr.ToMap(),
		})
	}
	b, _ := json.Marshal(s)
	return b
}

func (h *Handler) DumpByDocId() []byte {
	m := make(map[string]interface{})
	for _, doc := range h.docs_.docs {
		m[doc.docid] = map[string]interface{}{
			"id":     doc.id,
			"name":   doc.name,
			"active": doc.active,
			"dnf":    doc.dnf,
			"attr":   doc.attr.ToMap(),
		}
	}
	b, _ := json.Marshal(m)
	return b
}

func (h *Handler) DumpByName() []byte {
	m := make(map[string]interface{})
	for _, doc := range h.docs_.docs {
		m[doc.name] = map[string]interface{}{
			"id":     doc.id,
			"docid":  doc.docid,
			"active": doc.active,
			"dnf":    doc.dnf,
			"attr":   doc.attr.ToMap(),
		}
	}
	b, _ := json.Marshal(m)
	return b
}

func (this *conjList) display() {
	this.RLock()
	defer this.RUnlock()
	for i, conj := range this.conjs {
		DEBUG("Conj[", i, "]", "size:", conj.size, conj.ToString(this.h))
	}
}

func (this *amtList) display() {
	this.RLock()
	defer this.RUnlock()
	for i, amt := range this.amts {
		DEBUG("Amt[", i, "]:", amt.ToString(this.h))
	}
}

func (this *termList) display() {
	this.RLock()
	defer this.RUnlock()
	for i, term := range this.terms {
		DEBUG("Term[", i, "]", term.ToString())
	}
}

type displayer interface {
	display()
}

func display(obj displayer) {
	obj.display()
}

func (h *Handler) DisplayDocs() {
	display(h.docs_)
}

func (h *Handler) DisplayConjs() {
	display(h.conjs_)
}

func (h *Handler) DisplayAmts() {
	display(h.amts_)
}

func (h *Handler) DisplayTerms() {
	display(h.terms_)
}

func (h *Handler) DisplayConjRevs() {
	DEBUG("reverse list 1: ")
	h.conjRvsLock.RLock()
	defer h.conjRvsLock.RUnlock()
	for i, docs := range h.conjRvs {
		s := fmt.Sprint("conj[", i, "]: -> ")
		for _, id := range docs {
			s += strconv.Itoa(id) + " "
		}
		DEBUG(s)
	}
}

func (h *Handler) DisplayConjRevs2() {
	DEBUG("reverse list 2: ")

	h.conjSzRvsLock.RLock()
	defer h.conjSzRvsLock.RUnlock()

	h.terms_.RLock()
	defer h.terms_.RUnlock()

	for i := 0; i < len(h.conjSzRvs); i++ {
		termlist := h.conjSzRvs[i]
		if termlist == nil || len(termlist) == 0 {
			continue
		}
		DEBUG("***** size:", i, "*****")
		for _, termrvs := range termlist {
			s := fmt.Sprint(h.terms_.terms[termrvs.termId].ToString(), " -> ")
			for _, cpair := range termrvs.cList {
				var op string
				if cpair.belong {
					op = "∈"
				} else {
					op = "∉"
				}
				s += fmt.Sprintf("(%d %s) ", cpair.conjId, op)
			}
			DEBUG("   ", s)
		}
	}
}

func ConditionsToString(conds []Cond) string {
	ss := make([]string, 0, len(conds))
	for i := 0; i != len(conds); i++ {
		ss = append(ss, conds[i].ToString())
	}
	return "{ " + strings.Join(ss, ", ") + " }"
}
