package main

import (
	"math"
	"testing"
)

func TestValidateRules(t *testing.T) {
	tests := []struct {
		rule Rule
		exp  *Rule
	}{
		{Rule{}, nil},
		{Rule{App: "a"}, nil},
		{Rule{Space: "s"}, nil},
		{Rule{Org: "o"}, nil},
		{Rule{App: "a", Org: "o"}, nil},
		{Rule{Space: "s", Org: "o"}, nil},
		{Rule{App: "a", Space: "s"}, nil},
		{Rule{App: "a", Space: "s", Org: "o"}, nil},

		{Rule{MinInstances: 3}, nil},
		{Rule{MaxInstances: 5}, nil},
		{Rule{MinInstances: 3, MaxInstances: 5}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 3}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5}, nil},

		{
			Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: 40, MaxCpu: 60},
			&Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: 40, MaxCpu: 60, MinMem: math.MaxInt32, MaxMem: math.MaxInt32},
		},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: 60, MaxCpu: 60}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: 80, MaxCpu: 60}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 0, MaxInstances: 5, MinCpu: 40, MaxCpu: 60}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 2, MaxInstances: 5, MinCpu: 40, MaxCpu: 60}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 5, MaxInstances: 5, MinCpu: 40, MaxCpu: 60}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 6, MaxInstances: 5, MinCpu: 40, MaxCpu: 60}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: 40, MaxCpu: 160}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: 40, MaxCpu: -60}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: -40, MaxCpu: 60}, nil},
		{Rule{App: "a", Space: "s", Org: "", MinInstances: 3, MaxInstances: 5, MinCpu: 40, MaxCpu: 60}, nil},
		{Rule{App: "a", Space: "", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: 40, MaxCpu: 60}, nil},
		{Rule{App: "", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: 40, MaxCpu: 60}, nil},

		{
			Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinMem: 50, MaxMem: 70},
			&Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: math.MaxInt32, MaxCpu: math.MaxInt32, MinMem: 50, MaxMem: 70},
		},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinMem: 70, MaxMem: 70}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinMem: 90, MaxMem: 70}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinMem: 50, MaxMem: -70}, nil},
		{Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinMem: -50, MaxMem: 70}, nil},

		{
			Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: 40, MaxCpu: 60, MinMem: 50, MaxMem: 70},
			&Rule{App: "a", Space: "s", Org: "o", MinInstances: 3, MaxInstances: 5, MinCpu: 40, MaxCpu: 60, MinMem: 50, MaxMem: 70},
		},
	}

	for idx, test := range tests {
		rules := []Rule{test.rule}
		err := validateRules(rules)
		if err != nil && test.exp != nil {
			t.Fatalf("test %d: failed: %s", idx, err)
		} else if err == nil && (test.exp == nil || rules[0] != *test.exp) {
			t.Fatalf("test %d: succeeded: %s", idx)
		}
	}
}
