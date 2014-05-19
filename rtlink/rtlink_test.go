package rtlink

import (
	"net/http"
	"net/url"
	"testing"
)

func TestURL(t *testing.T) {
	for i, test := range []struct {
		pat, s string
		q      url.Values
		expect string
	}{
		{"/hello/", "world", nil, "/hello/world"},
		{"/hello", "world", nil, "/hello"},
		{"/hello/", "world", url.Values{"a": {"1"}}, "/hello/world?a=1"},
		{"/hello", "world", url.Values{"a": {"1"}}, "/hello?a=1"},
	} {
		u := URL(test.pat, test.s, test.q).String()
		if u != test.expect {
			t.Errorf("test %d: URL(%q, %q, %v) == %q (!= %q)", i, test.pat, test.s, test.q, u, test.expect)
		}
	}
}

func TestAbsURL(t *testing.T) {
	for i, test := range []struct {
		https        bool
		host, pat, s string
		q            url.Values
		expect       string
	}{
		{false, "example.com", "/hello/", "world", nil, "http://example.com/hello/world"},
		{true, "example.com", "/hello/", "world", nil, "https://example.com/hello/world"},
	} {
		u := AbsURL(test.https, test.host, test.pat, test.s, test.q).String()
		if u != test.expect {
			t.Errorf("test %d: URL(%b, %q, %q, %q, %v) == %q (!= %q)", i, test.https, test.host, test.pat, test.s, test.q, u, test.expect)
		}
	}
}

func TestLinker(t *testing.T) {
	for i, test := range []struct {
		ln     Linker
		pat, s string
		q      url.Values
		expect string
	}{
		{LinkerFunc(URL), "/hello/", "world", nil, "/hello/world"},
		{LinkerFunc(URL), "/hello", "world", nil, "/hello"},
		{LinkerFunc(URL), "/hello/", "world", url.Values{"a": {"1"}}, "/hello/world?a=1"},
		{LinkerFunc(URL), "/hello", "world", url.Values{"a": {"1"}}, "/hello?a=1"},
	} {
		u := test.ln.URL(test.pat, test.s, test.q).String()
		if u != test.expect {
			t.Errorf("test %d: URL(%q, %q, %v) == %q (!= %q)", i, test.pat, test.s, test.q, u, test.expect)
		}
	}
}

func TestHostLinker(t *testing.T) {
	for i, test := range []struct {
		https  bool
		req    *http.Request
		pat, s string
		q      url.Values
		expect string
	}{
		{false, &http.Request{Host: "example.com"}, "/hello/", "world", nil, "http://example.com/hello/world"},
		{true, &http.Request{Host: "example.com"}, "/hello/", "world", nil, "https://example.com/hello/world"},
	} {
		ln := NewHostLinker(test.req, test.https)
		u := ln.URL(test.pat, test.s, test.q).String()
		if u != test.expect {
			t.Errorf("test %d: ln.URL(%q, %q, %v) == %q (!= %q)", i, test.pat, test.s, test.q, u, test.expect)
		}
	}
}
