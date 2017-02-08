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

/* dnf handler used to save docs and search docs */
type Handler struct {
	docs_       *docList
	conjs_      *conjList
	amts_       *amtList
	terms_      *termList
	termMap     map[string]int
	termMapLock *rwLockWrapper

	conjRvs     [][]int
	conjRvsLock *rwLockWrapper

	conjSzRvs     [][]termRvs
	conjSzRvsLock *rwLockWrapper
}

var currentHandler *Handler = nil

/* create a dnf handler which is safe for concurrent use by multiple goroutines */
func NewHandler() *Handler {
	return newHandler(true)
}

/* create a dnf handler which is unsafe for concurrent use by multiple goroutines */
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
		docs_: &docList{
			docs:   make([]Doc, 0, 16),
			locker: newRwLockWrapper(useLock),
		},
		conjs_: &conjList{
			conjs:  make([]Conj, 0, 16),
			locker: newRwLockWrapper(useLock),
		},
		amts_: &amtList{
			amts:   make([]Amt, 0, 16),
			locker: newRwLockWrapper(useLock),
		},
		terms_: &termList{
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
	h.docs_.h = h
	h.conjs_.h = h
	h.amts_.h = h
	h.terms_.h = h
	return h
}

/* get global handler */
func GetHandler() *Handler {
	return currentHandler
}

/* set `handler` to global handler */
func SetHandler(handler *Handler) {
	currentHandler = handler
}
