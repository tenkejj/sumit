package main

import "testing"

func TestSanitizeNIP(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"679-310-80-59", "6793108059"},
		{"679 310 80 59", "6793108059"},
		{"PL679-310-80-59", "6793108059"},
		{"679.310.80.59", "6793108059"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := sanitizeNIP(tc.in); got != tc.want {
			t.Errorf("sanitizeNIP(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
