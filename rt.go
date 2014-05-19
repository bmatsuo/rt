/*
Package rt provides route composition.

This API is experimental.
*/
package rt

import "strings"

// Compose creates a path from pat with parameter s.
// Compose returns pat if pat does not take a route parameter.
func Compose(pat, s string) string {
	if !strings.HasSuffix(pat, "/") {
		return pat
	}
	return pat + s
}

// Decompose returns the parameter of pat that matches path.
// Decompose returns the empty string if pat doesn't match path.
func Decompose(pat, path string) string {
	if !strings.HasSuffix(pat, "/") {
		return ""
	}
	suf := strings.TrimPrefix(path, pat)
	if suf == path {
		return ""
	}
	return suf
}
