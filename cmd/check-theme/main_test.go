package main

import "testing"

func TestValidThemeName(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		// Lowercase, digits, hyphens.
		{"ao", true},
		{"umi-1", true},
		{"lian-ren", true},
		{"kabedon9", true},
		{"a", true},
		{"0", true},
		{"-", true},

		// Uppercase letters are now accepted.
		{"Kyouko", true},
		{"Miki", true},
		{"Ryouichi", true},
		{"Tenzen", true},
		{"ABCxyz", true},
		{"MixCase-9", true},

		// Reserved words are rejected.
		{"demo", false},
		{"random", false},
		// Reserved check must not be case-folded: only the exact
		// lowercase reserved tokens are forbidden.
		{"Demo", true},
		{"RANDOM", true},

		// Empty and invalid characters.
		{"", false},
		{"under_score", false},
		{"hello world", false},
		{"café", false},
		{"主题", false},
		{".dot", false},
		{"a/b", false},
	}
	for _, c := range cases {
		got := validThemeName(c.name)
		if got != c.want {
			t.Errorf("validThemeName(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}
