package godnf

import (
	"sync"
)

type Handler struct {
	docs_   *docList
	conjs_  *conjList
	amts_   *amtList
	terms_  *termList
	termMap map[string]int

	conjRvs     [][]int
	conjRvsLock sync.RWMutex

	conjSzRvs     [][]termRvs
	conjSzRvsLock sync.RWMutex
}

var currentHandler *Handler = nil

func NewHandler() *Handler {
	terms := make([]Term, 0, 16)
	terms = append(terms, Term{id: 0, key: "", val: ""})

	termrvslist := make([]termRvs, 0, 1)
	termrvslist = append(termrvslist, termRvs{termId: 0, cList: make([]cPair, 0)})
	conjSzRvs_ := make([][]termRvs, 16)
	conjSzRvs_[0] = termrvslist

	h := &Handler{
		docs_:   &docList{docs: make([]Doc, 0, 16)},
		conjs_:  &conjList{conjs: make([]Conj, 0, 16)},
		amts_:   &amtList{amts: make([]Amt, 0, 16)},
		terms_:  &termList{terms: terms},
		termMap: make(map[string]int),

		conjRvs:   make([][]int, 0),
		conjSzRvs: conjSzRvs_,
	}
	h.docs_.h = h
	h.conjs_.h = h
	h.amts_.h = h
	h.terms_.h = h
	return h
}

func GetHandler() *Handler {
	return currentHandler
}

func SetHandler(handler *Handler) {
	currentHandler = handler
}
