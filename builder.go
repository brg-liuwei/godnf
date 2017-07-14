package godnf

import (
	"errors"
	"sort"
)

var conjSizeTooLargeError error = errors.New("conjunction size too large(max: 255)")

// attribute interface of doc
type DocAttr interface {
	ToString() string
	ToMap() map[string]interface{}
}

// delete(lazy delete) doc from Handler by id
func (h *Handler) DeleteDoc(docid, comment string) bool {
	h.docs.Lock()
	defer h.docs.Unlock()
	for i := 0; i != len(h.docs.docs); i++ {
		pdoc := &h.docs.docs[i]
		if pdoc.docid == docid {
			rc := pdoc.active
			pdoc.active = false
			pdoc.comment = comment
			return rc
		}
	}
	return false
}

// add new doc and insert infos into reverse lists
func (h *Handler) AddDoc(name string, docid string, dnfDesc string, attr DocAttr) error {
	f := func() error {
		h.docs.RLock()
		defer h.docs.RUnlock()
		if _, ok := h.docs.docMap[docid]; ok {
			return errors.New("doc " + docid + " has been added before")
		}
		return nil
	}

	if err := f(); err != nil {
		return err
	}

	if err := DnfCheck(dnfDesc); err != nil {
		return err
	}
	return h.doAddDoc(name, docid, dnfDesc, attr)
}

func (h *Handler) doAddDoc(name string, docid string, dnf string, attr DocAttr) error {
	doc := &Doc{
		docid:   docid,
		name:    name,
		dnf:     dnf,
		conjs:   make([]int, 0),
		attr:    attr,
		active:  true,
		comment: "",
	}

	var conjId int
	var orStr string

	i := skipSpace(&dnf, 0)
	var err error
	for {
		if i, conjId, err = h.conjParse(&dnf, i); err != nil {
			return err
		}
		doc.conjs = append(doc.conjs, conjId)
		i = skipSpace(&dnf, i+1)
		if i >= len(dnf) {
			break
		}
		orStr, i = getString(&dnf, i)
		ASSERT(orStr == "or")
		i = skipSpace(&dnf, i+1)
	}
	docInternalId := h.docs.Add(doc, h)
	h.conjReverse1(docInternalId, doc.conjs)
	return nil
}

// conj: ( age in {3, 4} and state not in {CA, NY } )
func (h *Handler) conjParse(dnf *string, i int) (endIndex int, conjId int, err error) {
	var key, val string
	var vals []string
	var belong bool
	var op string // "in" or "not in"

	conj := &Conj{amts: make([]int, 0)}

	ASSERT((*dnf)[i] == leftDelimOfConj)

	for {
		// get assignment key
		i = skipSpace(dnf, i+1)
		key, i = getString(dnf, i)

		// get assignment op
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

		// get assignment vals
		i = skipSpace(dnf, i)
		ASSERT((*dnf)[i] == leftDelimOfSet)
		vals = make([]string, 0, 1)
		for {
			i = skipSpace(dnf, i+1)
			val, i = getString(dnf, i)
			vals = append(vals, val)
			i = skipSpace(dnf, i)
			if (*dnf)[i] == rightDelimOfSet {
				break
			}
			ASSERT((*dnf)[i] == separatorOfSet)
		}
		amtId := h.amtBuild(key, vals, belong)
		conj.amts = append(conj.amts, amtId)
		if belong {
			conj.size++
			if conj.size > 255 { // 255 == max(uint8)
				return -1, -1, conjSizeTooLargeError
			}
		}

		// get next assignment or end of this conjunction
		i = skipSpace(dnf, i+1)
		if (*dnf)[i] == rightDelimOfConj {
			conjId = h.conjs.Add(conj, h)
			endIndex = i

			// reverse list insert
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
		tid := h.terms.Add(term, h)
		amt.terms = append(amt.terms, tid)
	}
	return h.amts.Add(amt, h)
}

// Doc: (age ∈ { 3, 4 } and state ∈ { NY } ) or ( state ∈ { CA } and gender ∈ { M } ) -->
//     conj1: (age ∈ { 3, 4 } and state ∈ { NY } )
//     conj2: ( state ∈ { CA } and gender ∈ { M } )
type Doc struct {
	id         int     // unique id
	docid      string  // sent by doc adder
	name       string  // name of doc, for ad management
	dnf        string  // dnf decription
	conjSorted bool    // is conjs slice sorted
	conjs      []int   // conjunction ids
	attr       DocAttr // ad attr
	active     bool    // for lazy delete
	comment    string  // comment for active
}

// GetName returns name of this doc
func (doc *Doc) GetName() string {
	return doc.name
}

// GetDocId returns id of this doc
func (doc *Doc) GetDocId() string {
	return doc.docid
}

// GetDnf returns dnf statement of this doc
func (doc *Doc) GetDnf() string {
	return doc.dnf
}

// GetAttr returns attribute of this doc
func (doc *Doc) GetAttr() DocAttr {
	return doc.attr
}

// Conj(conjunction): age ∈ { 3, 4 } and state ∈ { NY } -->
//     assignment1: age ∈ { 3, 4 }
//     assignment2: state ∈ { NY }
type Conj struct {
	id        int   // unique id
	size      int   // conj size: number of ∈
	amtSorted bool  // is amts slice sorted
	amts      []int // assignments ids
}

func (c *Conj) Equal(conj *Conj) bool {
	if !c.amtSorted {
		sort.IntSlice(c.amts).Sort()
		c.amtSorted = true
	}
	if !conj.amtSorted {
		sort.IntSlice(conj.amts).Sort()
		conj.amtSorted = true
	}
	if c.size != conj.size {
		return false
	}
	if len(c.amts) != len(conj.amts) {
		return false
	}
	for i, amtId := range c.amts {
		if amtId != conj.amts[i] {
			return false
		}
	}
	return true
}

// Amt(assignment): age ∈ { 3, 4 } -->
//     term1: age ∈ { 3 }
//     term2: age ∈ { 4 }
type Amt struct {
	id         int   // unique id
	belong     bool  // ∈ or ∉
	termSorted bool  // is terms slice sorted
	terms      []int // terms ids
}

func (a *Amt) Equal(amt *Amt) bool {
	if !a.termSorted {
		sort.IntSlice(a.terms).Sort()
		a.termSorted = true
	}
	if !amt.termSorted {
		sort.IntSlice(amt.terms).Sort()
		amt.termSorted = true
	}
	if len(a.terms) != len(amt.terms) {
		return false
	}
	if a.belong != amt.belong {
		return false
	}
	for i, term := range a.terms {
		if term != amt.terms[i] {
			return false
		}
	}
	return true
}

// A Term like state ∉ { CA } reprensents the following value:
// Term{id: xxx, key: state, val: CA, belong: false}
type Term struct {
	id  int
	key string
	val string
}

// A term Equal iff key and val equal
func (t *Term) Equal(term *Term) bool {
	if t.key == term.key && t.val == term.val {
		return true
	}
	return false
}

// post lists
type docList struct {
	locker *rwLockWrapper
	docs   []Doc
	docMap map[string]bool
	h      *Handler
}

func (dl *docList) RLock()   { dl.locker.RLock() }
func (dl *docList) RUnlock() { dl.locker.RUnlock() }
func (dl *docList) Lock()    { dl.locker.Lock() }
func (dl *docList) Unlock()  { dl.locker.Unlock() }

type conjList struct {
	locker *rwLockWrapper
	conjs  []Conj
	h      *Handler
}

func (cl *conjList) RLock()   { cl.locker.RLock() }
func (cl *conjList) RUnlock() { cl.locker.RUnlock() }
func (cl *conjList) Lock()    { cl.locker.Lock() }
func (cl *conjList) Unlock()  { cl.locker.Unlock() }

func (cl *conjList) size() int {
	cl.locker.RLock()
	defer cl.locker.RUnlock()
	return len(cl.conjs)
}

type amtList struct {
	locker *rwLockWrapper
	amts   []Amt
	h      *Handler
}

func (al *amtList) RLock()   { al.locker.RLock() }
func (al *amtList) RUnlock() { al.locker.RUnlock() }
func (al *amtList) Lock()    { al.locker.Lock() }
func (al *amtList) Unlock()  { al.locker.Unlock() }

type termList struct {
	locker *rwLockWrapper
	terms  []Term
	h      *Handler
}

func (tl *termList) RLock()   { tl.locker.RLock() }
func (tl *termList) RUnlock() { tl.locker.RUnlock() }
func (tl *termList) Lock()    { tl.locker.Lock() }
func (tl *termList) Unlock()  { tl.locker.Unlock() }

func (dl *docList) Add(doc *Doc, h *Handler) int {
	dl.Lock()
	defer dl.Unlock()
	dl.docMap[doc.docid] = true
	doc.id = len(dl.docs)
	if !doc.conjSorted {
		sort.IntSlice(doc.conjs).Sort()
		doc.conjSorted = true
	}
	dl.docs = append(dl.docs, *doc)
	return doc.id
}

func (cl *conjList) Add(conj *Conj, h *Handler) (conjId int) {
	cl.Lock()
	defer cl.Unlock()
	for i, c := range cl.conjs {
		if c.Equal(conj) {
			conj.id = c.id
			return i
		}
	}
	conj.id = len(cl.conjs)

	// append post list
	cl.conjs = append(cl.conjs, *conj)

	// append reverse list
	h.conjRvsLock.Lock()
	defer h.conjRvsLock.Unlock()

	h.conjRvs = append(h.conjRvs, make([]int, 0))

	return conj.id
}

func (al *amtList) Add(amt *Amt, h *Handler) (amtId int) {
	al.Lock()
	defer al.Unlock()
	for i, a := range al.amts {
		if a.Equal(amt) {
			amt.id = a.id
			return i
		}
	}
	amt.id = len(al.amts)
	al.amts = append(al.amts, *amt)
	return amt.id
}

func (tl *termList) Add(term *Term, h *Handler) (termId int) {
	h.termMapLock.RLock()
	if tid, ok := h.termMap[term.key+"%"+term.val]; ok {
		h.termMapLock.RUnlock()
		term.id = tid
		return term.id
	}
	h.termMapLock.RUnlock()

	tl.Lock()
	term.id = len(tl.terms)
	tl.terms = append(tl.terms, *term)
	tl.Unlock()

	h.termMapLock.Lock()
	h.termMap[term.key+"%"+term.val] = term.id
	h.termMapLock.Unlock()
	return term.id
}

// reverse lists 1
/*
             | <-- sizeof conjs --> |
   conjRvs:  +--+--+--+--+--+--+--+--+
             |0 |1 |2 | ...    ...   |
             +--+--+--+--+--+--+--+--+
                 |
                 +--> doc1.id --> doc3.id --> docN.id
*/

// build the first layer reverse list
func (h *Handler) conjReverse1(docId int, conjIds []int) {
	h.conjRvsLock.Lock()
	defer h.conjRvsLock.Unlock()

	rvsLen := len(h.conjRvs)
	for _, conjId := range conjIds {
		ASSERT(rvsLen > conjId)
		rvsDocList := h.conjRvs[conjId]

		// append docId to rvsDocList and promise rvsDocList sorted
		pos := sort.IntSlice(rvsDocList).Search(docId)
		if pos < len(rvsDocList) && rvsDocList[pos] == docId {
			// doc id exists
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

// reverse lists 2
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

// for sort interface
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

// for sort interface
type termRvsSlice []termRvs

func (p termRvsSlice) Len() int           { return len(p) }
func (p termRvsSlice) Less(i, j int) bool { return p[i].termId < p[j].termId }
func (p termRvsSlice) Swap(i, j int) {
	p[i].termId, p[j].termId = p[j].termId, p[i].termId
	p[i].cList, p[j].cList = p[j].cList, p[i].cList
}

// build the second layer reverse list
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

	h.amts.RLock()
	defer h.amts.RUnlock()

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
	ASSERT(a > 1) // to avoid overflow
	return a
}

func (h *Handler) insertTermRvsList(conjId int, amtId int, list []termRvs) []termRvs {
	amt := &h.amts.amts[amtId]

	for _, tid := range amt.terms {
		idx := sort.Search(len(list), func(i int) bool { return list[i].termId >= tid })
		if idx < len(list) && list[idx].termId == tid {
			//term found
			clist := list[idx].cList
			if clist == nil {
				clist = make([]cPair, 0)
			}
			clist = insertClist(conjId, amt.belong, clist)
			list[idx].cList = clist
		} else {
			// term has not been found
			clist := make([]cPair, 0, 1)
			clist = append(clist, cPair{conjId: conjId, belong: amt.belong})
			list = append(list, termRvs{termId: tid, cList: clist})
			n := len(list)
			if n > 1 && list[n-1].termId < list[n-2].termId {
				// sort this list
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
		// found
		return l
	}
	l = append(l, cPair{conjId: conjId, belong: belong})
	n := len(l)
	if n > 1 && !cPairSlice(l).Less(n-2, n-1) {
		sort.Sort(cPairSlice(l))
	}
	return l
}
