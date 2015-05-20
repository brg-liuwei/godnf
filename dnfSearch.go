package godnf

import (
	"errors"
	"sort"
	"sync"

	"godnf/set"
)

type Cond struct {
	Key string
	Val string
}

func searchCondCheck(conds []Cond) error {
	if conds == nil || len(conds) == 0 {
		return errors.New("no conds to search")
	}
	m := make(map[string]bool)
	for _, cond := range conds {
		if _, ok := m[cond.Key]; ok {
			return errors.New("duplicate keys: " + cond.Key)
		}
		m[cond.Key] = true
	}
	return nil
}

func (h *Handler) Search(conds []Cond, attrFilter func(DocAttr) bool) (docs []int, err error) {
	if err := searchCondCheck(conds); err != nil {
		return nil, err
	}
	termids := make([]int, 0)
	for i := 0; i < len(conds); i++ {
		if id, ok := h.termMap[conds[i].Key+"%"+conds[i].Val]; ok {
			termids = append(termids, id)
		}
	}
	return h.doSearch(termids, attrFilter), nil
}

func (h *Handler) doSearch(terms []int, attrFilter func(DocAttr) bool) (docs []int) {
	conjs := h.getConjs(terms)
	if len(conjs) == 0 {
		return nil
	}
	return h.getDocs(conjs, attrFilter)
}

func (h *Handler) getDocs(conjs []int, attrFilter func(DocAttr) bool) (docs []int) {
	h.conjRvsLock.RLock()
	defer h.conjRvsLock.RUnlock()

	set := set.NewIntSet()

	var wg sync.WaitGroup

	for _, conj := range conjs {
		ASSERT(conj < len(h.conjRvs))
		doclist := h.conjRvs[conj]
		if doclist == nil {
			continue
		}
		for _, doc := range doclist {
			h.docs_.RLock()
			ok := attrFilter(h.docs_.docs[doc].attr)
			h.docs_.RUnlock()
			if !ok {
				continue
			}
			wg.Add(1)
			go func(h *Handler, docid int, w *sync.WaitGroup) {
				set.Add(docid)
				w.Done()
			}(h, doc, &wg)
		}
	}
	wg.Wait()
	return set.ToSlice()
}

func (h *Handler) getConjs(terms []int) (conjs []int) {
	h.conjSzRvsLock.RLock()
	defer h.conjSzRvsLock.RUnlock()

	n := len(terms)
	ASSERT(len(h.conjSzRvs) > 0)
	if n >= len(h.conjSzRvs) {
		n = len(h.conjSzRvs) - 1
	}

	conjSet := set.NewIntSet()

	for i := 0; i <= n; i++ {
		termlist := h.conjSzRvs[i]
		if termlist == nil || len(termlist) == 0 {
			continue
		}

		countSet := set.NewCountSet(i)

		for _, tid := range terms {
			idx := sort.Search(len(termlist), func(i int) bool {
				return termlist[i].termId >= tid
			})
			if idx < len(termlist) && termlist[idx].termId == tid &&
				termlist[idx].cList != nil {

				for _, pair := range termlist[idx].cList {
					countSet.Add(pair.conjId, pair.belong)
				}
			}
		}

		/* 处理∅ */
		if i == 0 {
			for _, pair := range termlist[0].cList {
				ASSERT(pair.belong == true)
				countSet.Add(pair.conjId, pair.belong)
			}
		}

		conjSet.AddSlice(countSet.ToSlice())
	}

	return conjSet.ToSlice()
}
