package godnf

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var debug bool

// DEBUG function, output msg to stdout when called SetDebug(true)
var DEBUG func(msg ...interface{})

// ASSERT function, enabled by calling SetDebug(true)
var ASSERT func(expression bool)

func doDEBUG(msg ...interface{}) {
	if debug {
		fmt.Println(msg...)
	}
}

func noDEBUG(msg ...interface{}) {}

func doASSERT(expression bool) {
	if !(expression) {
		panic("Assert Fail")
	}
}

func noASSERT(expression bool) {}

// SetDebug: enable/disable debug interface
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
	} else {
		DEBUG = noDEBUG
		ASSERT = noASSERT
	}
}

func init() {
	debug = false
	setDebug()
}

// Term to string
func (term *Term) ToString() string {
	if term.id == 0 {
		// empty set
		return "∅"
	}
	return fmt.Sprintf("( %s  %s )", term.key, term.val)
}

// Amt to string
func (amt *Amt) ToString(h *Handler) string {
	if len(amt.terms) == 0 {
		return ""
	}

	h.terms.RLock()
	defer h.terms.RUnlock()

	var key, op string

	if amt.belong {
		op = "∈"
	} else {
		op = "∉"
	}
	key = h.terms.terms[amt.terms[0]].key
	s := fmt.Sprintf("%s %s { ", key, op)
	for i, idx := range amt.terms {
		s += h.terms.terms[idx].val
		if i+1 < len(amt.terms) {
			s += ", "
		}
	}
	return s + " }"
}

// Conj to string
func (conj *Conj) ToString(h *Handler) string {
	if len(conj.amts) == 0 {
		return ""
	}
	h.amts.RLock()
	defer h.amts.RUnlock()
	s := "( "
	for i, idx := range conj.amts {
		s += h.amts.amts[idx].ToString(h)
		if i+1 < len(conj.amts) {
			s += " ∩ "
		}
	}
	return s + " )"
}

// Doc to string
func (doc *Doc) ToString(h *Handler) (s string) {
	if len(doc.conjs) == 0 {
		s = "len(conjs == 0)"
	}
	h.conjs.RLock()
	defer h.conjs.RUnlock()
	for i, idx := range doc.conjs {
		s += h.conjs.conjs[idx].ToString(h)
		if i+1 < len(doc.conjs) {
			s += " ∪ "
		}
	}
	s += "\nAttr: "
	s += doc.attr.ToString()
	return
}

func (dl *docList) display() {
	dl.RLock()
	defer dl.RUnlock()
	DEBUG("len(docs):", len(dl.docs))
	for i, doc := range dl.docs {
		if !doc.active {
			DEBUG("Doc[", i, "](del):", doc.ToString(dl.h))
		} else {
			DEBUG("Doc[", i, "]:", doc.ToString(dl.h))
		}
	}
}

func (dl *docList) docId2Attr(docid int) (DocAttr, error) {
	if len(dl.docs) <= docid {
		return nil, errors.New("docid over flow")
	}
	dl.RLock()
	defer dl.RUnlock()
	doc := &dl.docs[docid]
	return doc.attr, nil
}

// DocId2Attr: get DocAttr by docid
func (h *Handler) DocId2Attr(docid int) (DocAttr, error) {
	return h.docs.docId2Attr(docid)
}

func (dl *docList) docId2Map(docid int) map[string]interface{} {
	if len(dl.docs) <= docid {
		return nil
	}
	dl.RLock()
	defer dl.RUnlock()
	doc := &dl.docs[docid]
	return doc.attr.ToMap()
}

// DocId2Map: get DocAttr and convert the attr to map by docid
func (h *Handler) DocId2Map(docid int) map[string]interface{} {
	return h.docs.docId2Map(docid)
}

// DumpByPage: dump all docs by page_num and page_size for debug
func (h *Handler) DumpByPage(pageNum, pageSize int) []byte {
	h.docs.RLock()
	defer h.docs.RUnlock()

	totalRcd := len(h.docs.docs)
	start := (pageNum - 1) * pageSize

	if totalRcd == 0 || start >= totalRcd {
		b, _ := json.Marshal(map[string]interface{}{
			"total_records": totalRcd,
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

	s := make([]interface{}, 0, len(h.docs.docs[start:end]))
	for _, doc := range h.docs.docs[start:end] {
		s = append(s, map[string]interface{}{
			"name":    doc.name,
			"docid":   doc.docid,
			"active":  doc.active,
			"comment": doc.comment,
			"dnf":     doc.dnf,
			"attr":    doc.attr.ToMap(),
		})
	}

	b, _ := json.Marshal(map[string]interface{}{
		"total_records": totalRcd,
		"data":          s,
	})
	return b
}

// DumpByFilter: dump docs by funnel func for debug
func (h *Handler) DumpByFilter(filter func(DocAttr) bool) []byte {
	h.docs.RLock()
	defer h.docs.RUnlock()

	var s []interface{}
	for _, doc := range h.docs.docs {
		if filter(doc.attr) {
			s = append(s, map[string]interface{}{
				"name":    doc.name,
				"docid":   doc.docid,
				"active":  doc.active,
				"comment": doc.comment,
				"dnf":     doc.dnf,
				"attr":    doc.attr.ToMap(),
			})
		}
	}
	b, _ := json.Marshal(map[string]interface{}{
		"total_records": len(s),
		"data":          s,
	})
	return b
}

// DumpById: dump all docs by id for debug
func (h *Handler) DumpById() []byte {
	h.docs.RLock()
	defer h.docs.RUnlock()

	var s []interface{}
	for _, doc := range h.docs.docs {
		s = append(s, map[string]interface{}{
			"name":    doc.name,
			"docid":   doc.docid,
			"active":  doc.active,
			"comment": doc.comment,
			"dnf":     doc.dnf,
			"attr":    doc.attr.ToMap(),
		})
	}
	b, _ := json.Marshal(s)
	return b
}

// DumpByDocId: dump all docs by docid for debug
func (h *Handler) DumpByDocId() []byte {
	h.docs.RLock()
	defer h.docs.RUnlock()

	m := make(map[string]interface{})
	for _, doc := range h.docs.docs {
		m[doc.docid] = map[string]interface{}{
			"id":      doc.id,
			"name":    doc.name,
			"active":  doc.active,
			"comment": doc.comment,
			"dnf":     doc.dnf,
			"attr":    doc.attr.ToMap(),
		}
	}
	b, _ := json.Marshal(m)
	return b
}

// DumpByName: dump all docs by name for debug
func (h *Handler) DumpByName() []byte {
	h.docs.RLock()
	defer h.docs.RUnlock()

	m := make(map[string]interface{})
	for _, doc := range h.docs.docs {
		m[doc.name] = map[string]interface{}{
			"id":      doc.id,
			"docid":   doc.docid,
			"active":  doc.active,
			"comment": doc.comment,
			"dnf":     doc.dnf,
			"attr":    doc.attr.ToMap(),
		}
	}
	b, _ := json.Marshal(m)
	return b
}

func (cl *conjList) display() {
	cl.RLock()
	defer cl.RUnlock()
	for i, conj := range cl.conjs {
		DEBUG("Conj[", i, "]", "size:", conj.size, ",", conj.ToString(cl.h))
	}
}

func (al *amtList) display() {
	al.RLock()
	defer al.RUnlock()
	for i, amt := range al.amts {
		DEBUG("Amt[", i, "]:", amt.ToString(al.h))
	}
}

func (tl *termList) display() {
	tl.RLock()
	defer tl.RUnlock()
	for i, term := range tl.terms {
		DEBUG("Term[", i, "]:", term.ToString())
	}
}

type displayer interface {
	display()
}

func display(obj displayer) {
	obj.display()
}

// display doc list for debug
func (h *Handler) DisplayDocs() {
	display(h.docs)
}

// display conj list for debug
func (h *Handler) DisplayConjs() {
	display(h.conjs)
}

// display amt list for debug
func (h *Handler) DisplayAmts() {
	display(h.amts)
}

// display terms list for debug
func (h *Handler) DisplayTerms() {
	display(h.terms)
}

// DisplayConjRevs: display conj inverse list1 for debug
func (h *Handler) DisplayConjRevs() {
	DEBUG("reverse list 1:")

	h.conjRvsLock.RLock()
	defer h.conjRvsLock.RUnlock()

	for i, docs := range h.conjRvs {
		s := fmt.Sprint("conj[", i, "]: ->")
		for _, id := range docs {
			s += " " + strconv.Itoa(id)
		}
		DEBUG(s)
	}
}

// DisplayConjRevs2: display conj inverse list2 for debug
func (h *Handler) DisplayConjRevs2() {
	DEBUG("reverse list 2:")

	h.conjSzRvsLock.RLock()
	defer h.conjSzRvsLock.RUnlock()

	h.terms.RLock()
	defer h.terms.RUnlock()

	for i := 0; i < len(h.conjSzRvs); i++ {
		termlist := h.conjSzRvs[i]
		if termlist == nil || len(termlist) == 0 {
			continue
		}
		DEBUG("***** size:", i, "*****")
		for _, termrvs := range termlist {
			s := fmt.Sprint(h.terms.terms[termrvs.termId].ToString(), " ->")
			for _, cpair := range termrvs.cList {
				var op string
				if cpair.belong {
					op = "∈"
				} else {
					op = "∉"
				}
				s += fmt.Sprintf(" (%d %s)", cpair.conjId, op)
			}
			DEBUG("   ", s)
		}
	}
}

// ConditionsToString: convert Cond slice to string for debug
func ConditionsToString(conds []Cond) string {
	ss := make([]string, 0, len(conds))
	for i := 0; i != len(conds); i++ {
		ss = append(ss, conds[i].ToString())
	}
	return "{ " + strings.Join(ss, ", ") + " }"
}
