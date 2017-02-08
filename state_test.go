package godnf_test

import (
	"testing"

	dnf "github.com/brg-liuwei/godnf"
)

func setDelim() {
	dnf.SetDelimOfConj(rune('('), rune(')'))
	dnf.SetDelimOfSet(rune('{'), rune('}'))
	dnf.SetSeparatorOfSet(rune(','))
}

func TestDnfCheck(t *testing.T) {
	dnf.SetDebug(false)
	setDelim()
	checkDnf := func(dnfStr string, noErr bool) {
		err := dnf.DnfCheck(dnfStr)
		if (err == nil) != noErr {
			t.Error("test ", dnfStr, " fail")
		}
	}

	checkDnf("", false)   // test dnfStart
	checkDnf("  ", false) // test dnfStart

	checkDnf("(", false)                         // test State1
	checkDnf("(   ", false)                      // test State1
	checkDnf(" [ city in { Beijing }]  ", false) // test State1

	checkDnf(" (city in{Beijing})  ", true) // test skipString

	checkDnf("  ( city in { Beijing }) ", true)    // test State2
	checkDnf(" ( city not in { Beijing })", true)  // test State2
	checkDnf(" ( city not on { Beijing })", false) // test State2
	checkDnf(" ( city at { Beijing })", false)     // test State2

	checkDnf("(city in [ Beijing })", false) // test State4

	checkDnf("( city in { ShangHai ShenZheng })", false) // test State5

	checkDnf("( city in { ShangHai ])", false)            // test State6
	checkDnf("( city in { ShangHai, ShenZheng ))", false) // test State7

	checkDnf("(city in {SH} and gender not in { female}) or (age in {3, 5})", true)
	checkDnf("(city in {SH} and gender not in { female}) or (age in {3, 5} and city in {HZ})", true)
	checkDnf("(city in {SH} and city not in { BJ }) or (age in {3, 5} and city in {HZ})", false)
	checkDnf("(city in {SH}) or (age in {3, 5} and city in {HZ}", false)
}
