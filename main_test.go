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
		{[]string{"--help"}, "help", true},
		{[]string{"-help"}, "help", true},
		{[]string{"--version"}, "help", false},
		{[]string{}, "version", false},
	}
	for _, tc := range cases {
		if got := hasFlag(tc.args, tc.name); got != tc.want {
			t.Fatalf("hasFlag(%q, %q) = %v, want %v", tc.args, tc.name, got, tc.want)
		}
	}
}
