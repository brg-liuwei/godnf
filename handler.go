package godnf

import (
	"sync"
)

type rwLockWrapper struct {
	RLock   func()
	RUnlock func()
	Lock    func()
	Unlock  func()
	rwlock  sync.RWMutex
}

func newRwLockWrapper(useLock bool) *rwLockWrapper {
	locker := &rwLockWrapper{}
	if useLock {
		locker.RLock = func() { locker.rwlock.RLock() }
		locker.RUnlock = func() { locker.rwlock.RUnlock() }
		locker.Lock = func() { locker.rwlock.Lock() }
		locker.Unlock = func() { locker.rwlock.Unlock() }
	} else {
		nop := func() {}
		locker.RLock, locker.RUnlock = nop, nop
		locker.Lock, locker.Unlock = nop, nop
	}
	return locker
}

// Handler is used to save docs and search docs
type Handler struct {
	docs        *docList
	conjs       *conjList
	amts        *amtList
	terms       *termList
	termMap     map[string]int
	termMapLock *rwLockWrapper

	conjRvs     [][]int
	conjRvsLock *rwLockWrapper

	conjSzRvs     [][]termRvs
	conjSzRvsLock *rwLockWrapper
}

var currentHandler *Handler = nil

// NewHandler creates a dnf handler which is safe for concurrent use by multiple goroutines
func NewHandler() *Handler {
	return newHandler(true)
}

// NewHandlerWithoutLock creates a dnf handler
// which is unsafe for concurrent use by multiple goroutines
func NewHandlerWithoutLock() *Handler {
	return newHandler(false)
}

func newHandler(useLock bool) *Handler {
	terms := make([]Term, 0, 16)
	terms = append(terms, Term{id: 0, key: "", val: ""})

	termrvslist := make([]termRvs, 0, 1)
	termrvslist = append(termrvslist, termRvs{termId: 0, cList: make([]cPair, 0)})
	conjSzRvs := make([][]termRvs, 16)
	conjSzRvs[0] = termrvslist

	h := &Handler{
		docs: &docList{
			docs:   make([]Doc, 0, 16),
			locker: newRwLockWrapper(useLock),
		},
		conjs: &conjList{
			conjs:  make([]Conj, 0, 16),
			locker: newRwLockWrapper(useLock),
		},
		amts: &amtList{
			amts:   make([]Amt, 0, 16),
			locker: newRwLockWrapper(useLock),
		},
		terms: &termList{
			terms:  terms,
			locker: newRwLockWrapper(useLock),
		},
		termMap:     make(map[string]int),
		termMapLock: newRwLockWrapper(useLock),

		conjRvs:       make([][]int, 0),
		conjRvsLock:   newRwLockWrapper(useLock),
		conjSzRvs:     conjSzRvs,
		conjSzRvsLock: newRwLockWrapper(useLock),
	}
	h.docs.h = h
	h.conjs.h = h
	h.amts.h = h
	h.terms.h = h
	return h
}

// GetHandler returns current global handler
func GetHandler() *Handler {
	return currentHandler
}

// SetHandler set parameter handler to global handler
func SetHandler(handler *Handler) {
	currentHandler = handler
}
