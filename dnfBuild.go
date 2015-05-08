package godnf

import (
	"errors"
	"sort"
	"sync"
)

type DocAttr interface {
	ToString() string
	ToMap() map[string]interface{}
}

/* add new doc and insert infos into reverse lists */
func (h *Handler) AddDoc(name string, docid string, dnfDesc string, attr DocAttr) error {
	f := func() error {
		h.docs_.RLock()
		defer h.docs_.RUnlock()
		for _, doc := range h.docs_.docs {
			if doc.docid == docid {
				return errors.New("doc " + docid + "has been added before")
			}
		}
		return nil
	}

	if err := f(); err != nil {
		return err
	}

	if err := DnfCheck(dnfDesc); err != nil {
		return err
	}
	h.doAddDoc(name, docid, dnfDesc, attr)
	return nil
}

func (h *Handler) doAddDoc(name string, docid string, dnf string, attr DocAttr) {
	doc := &Doc{
		docid: docid,
		name:  name,
		dnf:   dnf,
		conjs: make([]int, 0),
		attr:  attr,
	}

	var conjId int
	var orStr string

	i := skipSpace(&dnf, 0)
	for {
		i, conjId = h.conjParse(&dnf, i)
		doc.conjs = append(doc.conjs, conjId)
		i = skipSpace(&dnf, i+1)
		if i >= len(dnf) {
			break
		}
		orStr, i = getString(&dnf, i)
		ASSERT(orStr == "or")
		i = skipSpace(&dnf, i+1)
	}
	docInternalId := h.docs_.Add(doc, h)
	h.conjReverse1(docInternalId, doc.conjs)
}

/*
conj: ( age in {3, 4} and state not in {CA, NY } )
*/
func (h *Handler) conjParse(dnf *string, i int) (endIndex int, conjId int) {
	var key, val string
	var vals []string
	var belong bool
	var op string /* "in" or "not in" */

	conj := &Conj{amts: make([]int, 0)}

	ASSERT((*dnf)[i] == '(')

	for {
		/* get assignment key */
		i = skipSpace(dnf, i+1)
		key, i = getString(dnf, i)

		/* get assignment op */
		i = skipSpace(dnf, i)
		op, i = getString(dnf, i)
		if op == "in" {
			belong = true
		} else {
			ASSERT(op == "not")
			i = skipSpace(dnf, i)
			op, i = getString(dnf, i)
			ASSERT(op == "in")
			belong = false
		}

		/* get assignment vals */
		i = skipSpace(dnf, i)
		ASSERT((*dnf)[i] == '{')
		vals = make([]string, 0, 1)
		for {
			i = skipSpace(dnf, i+1)
			val, i = getString(dnf, i)
			vals = append(vals, val)
			i = skipSpace(dnf, i)
			if (*dnf)[i] == '}' {
				break
			}
			ASSERT((*dnf)[i] == ',')
		}
		amtId := h.amtBuild(key, vals, belong)
		conj.amts = append(conj.amts, amtId)
		if belong {
			conj.size++
		}

		/* get next assignment or end of this conjunction */
		i = skipSpace(dnf, i+1)
		if (*dnf)[i] == ')' {
			conjId = h.conjs_.Add(conj, h)
			endIndex = i

			/* reverse list insert */
			h.conjReverse2(conj)
			return
		}

		val, i = getString(dnf, i)
		ASSERT(val == "and")
	}
}

func (h *Handler) amtBuild(key string, vals []string, belong bool) (amtId int) {
	amt := &Amt{terms: make([]int, 0), belong: belong}
	for _, val := range vals {
		term := &Term{key: key, val: val}
		tid := h.terms_.Add(term, h)
		amt.terms = append(amt.terms, tid)
	}
	return h.amts_.Add(amt, h)
}

/*
   doc: (age ∈ { 3, 4 } and state ∈ { NY } ) or ( state ∈ { CA } and gender ∈ { M } ) -->

       conj1: (age ∈ { 3, 4 } and state ∈ { NY } )
       conj2: ( state ∈ { CA } and gender ∈ { M } )
*/
type Doc struct {
	id         int    /* unique id */
	docid      string /* sent by doc adder */
	name       string /* name of doc, for ad management */
	dnf        string /* dnf decription */
	conjSorted bool   /* is conjs slice sorted? */
	conjs      []int  /* conjunction ids */
	active     bool
	attr       DocAttr /* ad attr */
}

func NewDoc(docid string, dnf string, name string, active bool, attr DocAttr) *Doc {
	return &Doc{
		docid:  docid,
		dnf:    dnf,
		name:   name,
		active: active,
		attr:   attr,
	}
}

func (doc *Doc) GetName() string {
	return doc.name
}

func (doc *Doc) GetDocId() string {
	return doc.docid
}

func (doc *Doc) GetDnf() string {
	return doc.dnf
}

func (doc *Doc) GetAttr() DocAttr {
	return doc.attr
}

/*
   conjunction: age ∈ { 3, 4 } and state ∈ { NY } -->

       assignment1: age ∈ { 3, 4 }
       assignment2: state ∈ { NY }
*/
type Conj struct {
	id        int   /* unique id */
	size      int   /* conj size: number of ∈ */
	amtSorted bool  /* is amts slice sorted? */
	amts      []int /* assignments ids */
}

func (this *Conj) Equal(c *Conj) bool {
	if !this.amtSorted {
		sort.IntSlice(this.amts).Sort()
		this.amtSorted = true
	}
	if !c.amtSorted {
		sort.IntSlice(c.amts).Sort()
		c.amtSorted = true
	}
	if this.size != c.size {
		return false
	}
	if len(this.amts) != len(c.amts) {
		return false
	}
	for i, amtId := range this.amts {
		if amtId != c.amts[i] {
			return false
		}
	}
	return true
}

/*
   assignment: age ∈ { 3, 4 } -->

       term1: age ∈ { 3 }
       term2: age ∈ { 4 }
*/
type Amt struct {
	id         int   /* unique id */
	belong     bool  /* ∈ or ∉ */
	termSorted bool  /* is terms slice sorted? */
	terms      []int /* terms ids */
}

func (this *Amt) Equal(amt *Amt) bool {
	if !this.termSorted {
		sort.IntSlice(this.terms).Sort()
		this.termSorted = false
	}
	if !amt.termSorted {
		sort.IntSlice(amt.terms).Sort()
		amt.termSorted = false
	}
	if len(this.terms) != len(amt.terms) {
		return false
	}
	if this.belong != amt.belong {
		return false
	}
	for i, term := range this.terms {
		if term != amt.terms[i] {
			return false
		}
	}
	return true
}

/*
   term: state ∉ { CA }
   Term{id: xxx, key: state, val: CA, belong: false}
*/
type Term struct {
	id  int
	key string
	val string
}

func (this *Term) Equal(term *Term) bool {
	if this.key == term.key && this.val == term.val {
		return true
	}
	return false
}

/* post lists */
type docList struct {
	sync.RWMutex
	docs []Doc
}

type conjList struct {
	sync.RWMutex
	conjs []Conj
}

type amtList struct {
	sync.RWMutex
	amts []Amt
}

type termList struct {
	sync.RWMutex
	terms []Term
}

func (this *docList) Add(doc *Doc, h *Handler) int {
	this.Lock()
	defer this.Unlock()
	doc.id = len(this.docs)
	doc.active = true
	if !doc.conjSorted {
		sort.IntSlice(doc.conjs).Sort()
		doc.conjSorted = true
	}
	this.docs = append(this.docs, *doc)
	return doc.id
}

func (this *conjList) Add(conj *Conj, h *Handler) (conjId int) {
	this.Lock()
	defer this.Unlock()
	for i, c := range this.conjs {
		if c.Equal(conj) {
			conj.id = c.id
			return i
		}
	}
	conj.id = len(this.conjs)

	/* append post list */
	this.conjs = append(this.conjs, *conj)

	/* append reverse list */
	h.conjRvsLock.Lock()
	defer h.conjRvsLock.Unlock()

	h.conjRvs = append(h.conjRvs, make([]int, 0))

	return conj.id
}

func (this *amtList) Add(amt *Amt, h *Handler) (amtId int) {
	this.Lock()
	defer this.Unlock()
	for i, a := range this.amts {
		if a.Equal(amt) {
			amt.id = a.id
			return i
		}
	}
	amt.id = len(this.amts)
	this.amts = append(this.amts, *amt)
	return amt.id
}

func (this *termList) Add(term *Term, h *Handler) (termId int) {
	if tid, ok := h.termMap[term.key+"%"+term.val]; ok {
		term.id = tid
		return term.id
	}
	this.Lock()
	defer this.Unlock()
	term.id = len(this.terms)
	this.terms = append(this.terms, *term)
	h.termMap[term.key+"%"+term.val] = term.id
	return term.id
}

/* reverse lists 1 */
/*
             | <-- sizeof conjs_ --> |
   conjRvs:  +--+--+--+--+--+--+--+--+
             |0 |1 |2 | ...    ...   |
             +--+--+--+--+--+--+--+--+
                 |
                 +--> doc1.id --> doc3.id --> docN.id
*/

/* build the first layer reverse list */
func (h *Handler) conjReverse1(docId int, conjIds []int) {
	h.conjRvsLock.Lock()
	defer h.conjRvsLock.Unlock()

	rvsLen := len(h.conjRvs)
	for _, conjId := range conjIds {
		ASSERT(rvsLen > conjId)
		rvsDocList := h.conjRvs[conjId]

		/* append docId to rvsDocList and promise rvsDocList sorted */
		pos := sort.IntSlice(rvsDocList).Search(docId)
		if pos < len(rvsDocList) && rvsDocList[pos] == docId {
			/* doc id exists */
			return
		}

		rvsDocList = append(rvsDocList, docId)
		if len(rvsDocList) > 1 {
			if docId < rvsDocList[len(rvsDocList)-2] {
				sort.IntSlice(rvsDocList).Sort()
			}
		}
		h.conjRvs[conjId] = rvsDocList
	}
}

/* reverse lists 2 */
/*
                 +----- sizeof (conj)
                 |
 conjSzRvs:  +--+--+--+--+--+--+
             |0 |1 | ...  ...  |
             +--+--+--+--+--+--+
                 |
                 +--> +-------+-------+-------+-------+
                      |termId |termId |termId |termId |
      []termRvs:      +-------+-------+-------+-------+
                      | clist | clist | clist | clist |
                      +-------+-------+-------+-------+
                         |
                         +--> +-----+-----+-----+-----+-----+
                              |cId:1|cId:4|cId:4|cId:8|cId:9|
              []cPair:        +-----+-----+-----+-----+-----+
                              |  ∈  |  ∈  |  ∉  |  ∉  |  ∈  |
                              +-----+-----+-----+-----+-----+
*/
type cPair struct {
	conjId int
	belong bool
}

/* for sort interface */
type cPairSlice []cPair

func (p cPairSlice) Len() int { return len(p) }
func (p cPairSlice) Less(i, j int) bool {
	if p[i].conjId == p[j].conjId {
		return p[i].belong
	}
	return p[i].conjId < p[j].conjId
}
func (p cPairSlice) Swap(i, j int) {
	p[i].conjId, p[j].conjId = p[j].conjId, p[i].conjId
	p[i].belong, p[j].belong = p[j].belong, p[i].belong
}

type termRvs struct {
	termId int
	cList  []cPair
}

/* for sort interface */
type termRvsSlice []termRvs

func (p termRvsSlice) Len() int           { return len(p) }
func (p termRvsSlice) Less(i, j int) bool { return p[i].termId < p[j].termId }
func (p termRvsSlice) Swap(i, j int) {
	p[i].termId, p[j].termId = p[j].termId, p[i].termId
	p[i].cList, p[j].cList = p[j].cList, p[i].cList
}

/* build the second layer reverse list */
func (h *Handler) conjReverse2(conj *Conj) {
	h.conjSzRvsLock.Lock()
	defer h.conjSzRvsLock.Unlock()

	if conj.size >= len(h.conjSzRvs) {
		h.resizeConjSzRvs(conj.size + 1)
	}

	termRvsList := h.conjSzRvs[conj.size]
	defer func() { h.conjSzRvs[conj.size] = termRvsList }()

	if termRvsList == nil {
		termRvsList = make([]termRvs, 0)
	}

	h.amts_.RLock()
	defer h.amts_.RUnlock()

	for _, amtId := range conj.amts {
		termRvsList = h.insertTermRvsList(conj.id, amtId, termRvsList)
	}
	if conj.size == 0 {
		termRvsList[0].cList = insertClist(conj.id, true, termRvsList[0].cList)
	}
}

func (h *Handler) resizeConjSzRvs(size int) {
	ASSERT(size >= len(h.conjSzRvs))
	size = upperPowerOfTwo(size)
	tmp := make([][]termRvs, size)
	copy(tmp[:len(h.conjSzRvs)], h.conjSzRvs[:])
	h.conjSzRvs = tmp
	ASSERT(len(h.conjSzRvs) == size)
}

func upperPowerOfTwo(size int) int {
	a := 4
	for a < size && a > 1 {
		a = a << 1
	}
	ASSERT(a > 1) /* to avoid overflow */
	return a
}

func (h *Handler) insertTermRvsList(conjId int, amtId int, list []termRvs) []termRvs {
	amt := &h.amts_.amts[amtId]

	for _, tid := range amt.terms {
		idx := sort.Search(len(list), func(i int) bool { return list[i].termId >= tid })
		if idx < len(list) && list[idx].termId == tid {
			/* term found */
			clist := list[idx].cList
			if clist == nil {
				clist = make([]cPair, 0)
			}
			clist = insertClist(conjId, amt.belong, clist)
			list[idx].cList = clist
		} else {
			/* term has not been found */
			clist := make([]cPair, 0, 1)
			clist = append(clist, cPair{conjId: conjId, belong: amt.belong})
			list = append(list, termRvs{termId: tid, cList: clist})
			n := len(list)
			if n > 1 && list[n-1].termId < list[n-2].termId {
				/* sort this list */
				sort.Sort(termRvsSlice(list))
			}
		}
	}
	return list
}

func insertClist(conjId int, belong bool, l []cPair) []cPair {
	idx := sort.Search(len(l), func(i int) bool {
		if l[i].conjId == conjId {
			return !l[i].belong || l[i].belong == belong
		}
		return l[i].conjId >= conjId
	})
	if idx < len(l) && (l[idx].conjId == conjId && l[idx].belong == belong) {
		/* found */
		return l
	}
	l = append(l, cPair{conjId: conjId, belong: belong})
	n := len(l)
	if n > 1 && !cPairSlice(l).Less(n-2, n-1) {
		sort.Sort(cPairSlice(l))
	}
	return l
}
