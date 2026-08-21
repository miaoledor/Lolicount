// Package utils provides shared helpers used across drawer packages:
// SVG string escaping and display-size geometry. It has no business
// dependencies so every drawer can import it without coupling.
package utils

import "strings"

// EscapeXML escapes the five special XML characters in a text node.
func EscapeXML(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}
