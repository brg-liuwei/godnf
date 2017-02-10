package godnf

import (
	"errors"
	"fmt"
	"sort"

	"github.com/brg-liuwei/godnf/set"
)

// Cond: retrieval condition
//
// A dnf like ({Country in {CN, RU, US}) can be retrievaled by Cond as follows:
//
// Cond{
//     Key: "Country",
//     Val: "CN",
// }
//
// or
//
// Cond{
//     Key: "Country",
//     Val: "RU",
// }
//
// or
//
// Cond{
//     Key: "Country",
//     Val: "US",
// }
type Cond struct {
	Key string
	Val string
}

func (c *Cond) ToString() string {
	return fmt.Sprintf("(%s: %s)", c.Key, c.Val)
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

// Search docs which match conds and passed by attrFilter
func (h *Handler) Search(conds []Cond, attrFilter func(DocAttr) bool) (docs []int, err error) {
	if err := searchCondCheck(conds); err != nil {
		return nil, err
	}
	termids := make([]int, 0)
	h.termMapLock.RLock()
	for i := 0; i < len(conds); i++ {
		if id, ok := h.termMap[conds[i].Key+"%"+conds[i].Val]; ok {
			termids = append(termids, id)
		}
	}
	h.termMapLock.RUnlock()
	return h.doSearch(termids, attrFilter), nil
}

// SearchAll searches all docs which match conds
func (h *Handler) SearchAll(conds []Cond) (docs []int, err error) {
	return h.Search(conds, func(DocAttr) bool { return true })
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

	for _, conj := range conjs {
		ASSERT(conj < len(h.conjRvs))
		doclist := h.conjRvs[conj]
		if doclist == nil {
			continue
		}
		for _, doc := range doclist {
			h.docs.RLock()
			ok := h.docs.docs[doc].active && attrFilter(h.docs.docs[doc].attr)
			h.docs.RUnlock()
			if !ok {
				continue
			}
			set.Add(doc, false)
		}
	}
	return set.ToSlice(false)
}

func (h *Handler) getConjs(terms []int) (conjs []int) {
	h.conjSzRvsLock.RLock()
	defer h.conjSzRvsLock.RUnlock()

	n := len(terms)
	ASSERT(len(h.conjSzRvs) > 0)
	if n >= len(h.conjSzRvs) {
		n = len(h.conjSzRvs) - 1
	}

	ASSERT(n <= 255) // max(uint8) == 255

	conjSet := set.NewIntSet()

	for i := 0; i <= n; i++ {
		termlist := h.conjSzRvs[i]
		if termlist == nil || len(termlist) == 0 {
			continue
		}

		countSet := set.NewCountSet(uint8(i), h.conjs.size())

		for _, tid := range terms {
			idx := sort.Search(len(termlist), func(i int) bool {
				return termlist[i].termId >= tid
			})
			if idx < len(termlist) && termlist[idx].termId == tid &&
				termlist[idx].cList != nil {

				for _, pair := range termlist[idx].cList {
					countSet.Add(pair.conjId, pair.belong, false)
				}
			}
		}

		// 处理∅
		if i == 0 {
			for _, pair := range termlist[0].cList {
				ASSERT(pair.belong == true)
				countSet.Add(pair.conjId, pair.belong, false)
			}
		}

		conjSet.AddSlice(countSet.ToSlice(false), false)
	}

	return conjSet.ToSlice(false)
}
