package utils

import "testing"

func TestEscapeXML(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"abc", "abc"},
		{"a&b", "a&amp;b"},
		{"a<b>c", "a&lt;b&gt;c"},
		{`"q"`, "&quot;q&quot;"},
		{"'apos'", "&apos;apos&apos;"},
	}
	for _, tt := range tests {
		if got := EscapeXML(tt.in); got != tt.want {
			t.Errorf("EscapeXML(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
