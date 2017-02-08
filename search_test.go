package godnf_test

import (
	"fmt"

	dnf "github.com/brg-liuwei/godnf"
)

func ExampleConditionToString() {
	conds := []dnf.Cond{
		{Key: "platform", Val: "iOS"},
		{Key: "city", Val: "ShangHai"},
		{Key: "gender", Val: "female"},
	}

	for _, cond := range conds {
		fmt.Println(cond.ToString())
	}
	// Unordered output:
	// (city: ShangHai)
	// (gender: female)
	// (platform: iOS)
}

func ExampleConditionsToString() {
	conds := []dnf.Cond{
		{Key: "platform", Val: "iOS"},
		{Key: "city", Val: "ShangHai"},
		{Key: "gender", Val: "female"},
	}
	fmt.Println(dnf.ConditionsToString(conds))
	// Output:
	// { (platform: iOS), (city: ShangHai), (gender: female) }
}
