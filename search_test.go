package godnf_test

import (
	"fmt"

	dnf "github.com/brg-liuwei/godnf"
)

func ExampleCondToString() {
	conds := []dnf.Cond{
		{"platform", "iOS"},
		{"city", "ShangHai"},
		{"gender", "female"},
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
		{"platform", "iOS"},
		{"city", "ShangHai"},
		{"gender", "female"},
	}
	fmt.Println(dnf.ConditionsToString(conds))
	// Output:
	// { (platform: iOS), (city: ShangHai), (gender: female) }
}
