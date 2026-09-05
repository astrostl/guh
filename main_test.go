package main

import "testing"

func TestHasFlag(t *testing.T) {
	cases := []struct {
		args []string
		name string
		want bool
	}{
		{[]string{"--version"}, "version", true},
		{[]string{"-version"}, "version", true},
		{[]string{"--demo"}, "demo", true},
		{[]string{"-demo"}, "demo", true},
		{[]string{"--demo"}, "version", false},
		{[]string{}, "version", false},
	}
	for _, tc := range cases {
		if got := hasFlag(tc.args, tc.name); got != tc.want {
			t.Fatalf("hasFlag(%q, %q) = %v, want %v", tc.args, tc.name, got, tc.want)
		}
	}
}
