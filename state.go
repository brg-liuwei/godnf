package godnf

import (
	"errors"
)

var leftDelimOfConj, rightDelimOfConj byte = byte('('), byte(')')
var leftDelimOfSet, rightDelimOfSet byte = byte('{'), byte('}')
var separatorOfSet byte = byte(',')

/* set delim of conjunction */
func SetDelimOfConj(left, right rune) {
	leftDelimOfConj, rightDelimOfConj = byte(left), byte(right)
}

/*
   get delim of conjunction

   eg: if a dnf is like (Country in {CN, RU, US}),
   this func will return rune('('), rune(')')
*/
func GetDelimOfConj() (left, right rune) {
	return rune(leftDelimOfConj), rune(rightDelimOfConj)
}

/* set delim of set */
func SetDelimOfSet(left, right rune) {
	leftDelimOfSet, rightDelimOfSet = byte(left), byte(right)
}

/*
   get delim of set

   eg: if a dnf is like (Country in {CN, RU, US}),
   this func will return rune('{'), rune('}')
*/
func GetDelimOfSet() (left, right rune) {
	return rune(leftDelimOfSet), rune(rightDelimOfSet)
}

/* set separator of set */
func SetSeparatorOfSet(sep rune) {
	separatorOfSet = byte(sep)
}

/*
   get separator of set
   eg, the separator of set of a dnf like (Country in {CN, RU, US}) is rune(',')
*/
func GetSeparatorOfSet() rune {
	return rune(separatorOfSet)
}

/* dnf format error */
var DnfFmtError error = errors.New("dnf format error")

func skipSpace(s *string, i int) int {
	for i < len(*s) {
		if (*s)[i] != ' ' {
			break
		}
		i++
	}
	return i
}

func skipString(s *string, i int) (j int) {
	if i >= len(*s) {
		return len(*s)
	}
	for j = i + 1; j < len(*s); j++ {
		switch (*s)[j] {
		case ' ':
			return
		case separatorOfSet:
			return
		case leftDelimOfSet:
			return
		case rightDelimOfSet:
			return
		}
	}
	return
}

func getString(s *string, i int) (subStr string, endIndex int) {
	j := skipString(s, i)
	ASSERT(i < j)
	return string((*s)[i:j]), j
}

func dnfIdxCheck(dnf *string, i int, errmsg string) bool {
	if i >= len(*dnf) || i < 0 {
		DEBUG(errmsg)
		return false
	}
	return true
}

/* check dnf syntax */
func DnfCheck(dnf string) error {
	return dnfStart(&dnf, skipSpace(&dnf, 0))
}

/* start: get leftDelimOfConj, default is '(' */
func dnfStart(dnf *string, i int) error {
	if !dnfIdxCheck(dnf, i, "start dnf idx error") {
		return DnfFmtError
	}
	if (*dnf)[i] != leftDelimOfConj {
		DEBUG("dnf start error, dnf[", i, "] =", string((*dnf)[i]))
		return DnfFmtError
	}

	m := make(map[string]bool)
	return dnfState1(dnf, skipSpace(dnf, i+1), m)
}

/* state1: get key */
func dnfState1(dnf *string, i int, m map[string]bool) error {
	if !dnfIdxCheck(dnf, i, "dnfState1 check error") {
		return DnfFmtError
	}
	j := skipString(dnf, i)
	if j >= len(*dnf) {
		DEBUG("state 1 internal error")
		return DnfFmtError
	}
	key := string((*dnf)[i:j])
	if _, ok := m[key]; ok {
		return errors.New("conjunction key " + key + " duplicate")
	}
	m[key] = true
	return dnfState2(dnf, skipSpace(dnf, j+1), m)
}

/* state2: get 'not' or get 'in' */
func dnfState2(dnf *string, i int, m map[string]bool) error {
	if !dnfIdxCheck(dnf, i, "dnfState2 check error") {
		return DnfFmtError
	}
	if i+3 <= len(*dnf) && string((*dnf)[i:i+3]) == "not" {
		return dnfState3(dnf, skipSpace(dnf, i+3), m)
	}
	if i+2 <= len(*dnf) && string((*dnf)[i:i+2]) == "in" {
		return dnfState4(dnf, skipSpace(dnf, i+2), m)
	}
	DEBUG("state2 internal error")
	return DnfFmtError
}

/* state3: get 'in' */
func dnfState3(dnf *string, i int, m map[string]bool) error {
	if !dnfIdxCheck(dnf, i, "dnfState3 check error") {
		return DnfFmtError
	}
	if i+2 <= len(*dnf) && string((*dnf)[i:i+2]) == "in" {
		return dnfState4(dnf, skipSpace(dnf, i+2), m)
	}
	DEBUG("state3 internal error")
	return DnfFmtError
}

/* state4: get leftDelimOfSet, default is '{' */
func dnfState4(dnf *string, i int, m map[string]bool) error {
	if !dnfIdxCheck(dnf, i, "dnfState4 check error") {
		return DnfFmtError
	}
	if (*dnf)[i] != leftDelimOfSet {
		DEBUG("state4 internal error")
		return DnfFmtError
	}
	return dnfState5(dnf, skipSpace(dnf, i+1), m)
}

/* state5: get elem of set */
func dnfState5(dnf *string, i int, m map[string]bool) error {
	if !dnfIdxCheck(dnf, i, "dnfState5 check error") {
		return DnfFmtError
	}

	j := skipString(dnf, i)
	if j >= len(*dnf) {
		DEBUG("state5 internal error")
		return DnfFmtError
	}

	// val := string((*dnf)[i:j])
	// _ = val

	return dnfState6(dnf, skipSpace(dnf, j), m)
}

/* state6: get next val(',') or get end of set('}')*/
func dnfState6(dnf *string, i int, m map[string]bool) error {
	if !dnfIdxCheck(dnf, i, "dnfState6 check error") {
		return DnfFmtError
	}
	if (*dnf)[i] == separatorOfSet {
		return dnfState7(dnf, skipSpace(dnf, i+1), m)
	}
	if (*dnf)[i] == rightDelimOfSet {
		return dnfState8(dnf, skipSpace(dnf, i+1), m)
	}
	DEBUG("state6 internal error")
	return DnfFmtError
}

/* state7: get next val */
func dnfState7(dnf *string, i int, m map[string]bool) error {
	if !dnfIdxCheck(dnf, i, "dnfState7 check error") {
		return DnfFmtError
	}

	j := skipString(dnf, i)
	if j >= len(*dnf) {
		DEBUG("state7 internal error")
		return DnfFmtError
	}

	// val := string((*dnf)[i:j])
	// _ = val

	return dnfState6(dnf, skipSpace(dnf, j), m)
}

/* state8: get 'and' or end of this conj(')') */
func dnfState8(dnf *string, i int, m map[string]bool) error {
	if !dnfIdxCheck(dnf, i, "dnfState8 check error") {
		return DnfFmtError
	}
	if i+3 < len(*dnf) && string((*dnf)[i:i+3]) == "and" {
		return dnfState1(dnf, skipSpace(dnf, i+3), m)
	}
	if (*dnf)[i] == rightDelimOfConj {
		return dnfState9(dnf, skipSpace(dnf, i+1), m)
	}
	DEBUG("state8 internal error")
	return DnfFmtError
}

/* state9: end of dnf or next conj(get "or") */
func dnfState9(dnf *string, i int, m map[string]bool) error {
	if i == len(*dnf) {
		/* accept */
		return nil
	}
	if i+2 < len(*dnf) && string((*dnf)[i:i+2]) == "or" {
		return dnfStart(dnf, skipSpace(dnf, i+2))
	}
	DEBUG("state9 internal error")
	return DnfFmtError
}
