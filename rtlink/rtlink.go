// Package rtlink provides linking for web servers.
package rtlink

import (
	"net/http"
	"net/url"

	"github.com/bmatsuo/rt"
)

// AbsURL is like URL but returns an absolute URL.  If https is false then the
// "http" scheme is used.
func AbsURL(https bool, host, pat, s string, q url.Values) url.URL {
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
func URL(pat, s string, q url.Values) url.URL {
	return url.URL{
		Path:     rt.Compose(pat, s),
		RawQuery: q.Encode(),
	}
}

type Linker struct {
	UseHTTPS bool
}

func (ln *Linker) URL(req *http.Request, pat, s string, q url.Values) url.URL {
	return AbsURL(ln.UseHTTPS, req.Host, pat, s, q)
}

func (ln *Linker) Link(req *http.Request, rel, pat, s string, q url.Values) Link {
	u := ln.URL(req, pat, s, q)
	return Link{u, rel}
}

type Link struct {
	HRef url.URL `json:"href"`
	Rel  string  `json:"rel"`
}
