/*
Package rtlink defines helper functions for creating links with rt.

This API is experimental.
*/
package rtlink

import (
	"net/http"
	"net/url"

	"github.com/bmatsuo/rt"
)

type Link struct {
	Rel  string   `json:"rel"`
	HRef *url.URL `json:"href"`
}

// AbsURL is like URL but returns an absolute URL.  If https is false then the
// "http" scheme is used.
func AbsURL(https bool, host, pat, s string, q url.Values) *url.URL {
	u := URL(pat, s, q)
	u.Host = host
	if https {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}
	return u
}

// URL creates a relative URL from the path composition of pat and s with query
// q.
func URL(pat, s string, q url.Values) *url.URL {
	return &url.URL{
		Path:     rt.Compose(pat, s),
		RawQuery: q.Encode(),
	}
}

// Linker is a type that can be used for making links.
//	ln := NewHostLinker(req, false)
//	friendLink := Link{
//		"urn:myvocab:friend",
//		ln.URL("/friends/", friend.ID, nil),
//	}
//	friendsLink := Link{
//		"urn:myvocab:friend-list",
//		ln.URL("/friends/", "", url.Values{
//			{"maxValues": 10},
//			{"sortBy": "relevance"},
//		},
//	)}
type Linker interface {
	// URL returns a URL with a path composed of pat and s and with query q.
	// The returned URL may not be absolute.
	URL(pat, s string, q url.Values) *url.URL
}

// LinkerFunc is a function that implements Linker.
type LinkerFunc func(pat, s string, q url.Values) *url.URL

// URL returns the result of fn(pat, s, q).
func (fn LinkerFunc) URL(pat, s string, q url.Values) *url.URL {
	return fn(pat, s, q)
}

// Linker is a helper type for creating links.
type hostLinker struct {
	req      *http.Request
	useHTTPS bool
}

// NewHostLinker returns a Linker directing clients absolute URLs on req.Host.
// If useHTTPS is false the "http" scheme is used in URLs.
func NewHostLinker(req *http.Request, useHTTPS bool) Linker {
	return &hostLinker{req, useHTTPS}
}

// URL is like the function URL but uses ln and req.Host to determine the
// scheme and host to use.
func (ln *hostLinker) URL(pat, s string, q url.Values) *url.URL {
	return AbsURL(ln.useHTTPS, ln.req.Host, pat, s, q)
}
