package rt

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCompose(t *testing.T) {
	for i, test := range []struct{ pat, param, path string }{
		{"/", "", "/"},
		{"/", "/", "//"},
		{"/", "hello", "/hello"},
		{"/hello", "world", "/hello"},
		{"/hello/", "world", "/hello/world"},
		{"hello/", "world", "hello/world"},
		// weird cases
		{"", "", ""},
		{"", "hello", ""},
	} {
		path := Compose(test.pat, test.param)
		if path != test.path {
			t.Errorf("test %d: Compose(%q, %q) == %q (!= %q)", i, test.pat, test.param, path, test.path)
		}
	}
}

func TestDecompose(t *testing.T) {
	for i, test := range []struct{ pat, path, param string }{
		{"/", "/", ""},
		{"/", "/hello", "hello"},
		{"/hello/", "/hello/world", "world"},
		{"/hello", "/hello/world", ""},
		// weird cases not generally seen
		{"", "", ""},
		{"", "hello", ""},
		{"/", "", ""},
	} {
		param := Decompose(test.pat, test.path)
		if param != test.param {
			t.Errorf("test %d: Deompose(%q, %q) == %q (!= %q)", i, test.pat, test.path, param, test.param)
		}
	}
}

func TestServeMuxParameter(t *testing.T) {
	mux := NewServeMux()
	mux.HandleFunc("/hello/", func(resp http.ResponseWriter, req *http.Request) {
		_, reqPat := mux.Handler(req)
		name := Decompose(reqPat, req.URL.Path)
		fmt.Fprint(resp, name)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	resp, err := http.Get(server.URL + "/hello/world")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	p, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(p) != "world" {
		t.Fatalf("unexpected response: %q", string(p))
	}
}

func TestServeMuxCheckReverse(t *testing.T) {
	const (
		root   = "/"
		hello  = "/hello/"
		ignore = "/ignore/"
	)
	nop := func(http.ResponseWriter, *http.Request) {}
	mux := NewServeMux()
	mux.HandleFunc(root, nop)
	mux.HandleFunc(hello, nop)
	mux.HandleFunc(ignore, nop)

	type goodReverse struct{ root, hello, ignore string }

	var struc interface{}
	var err error
	var expect error

	struc = struct{}{}
	expect = IrreversibleRoutesError{root, hello, ignore}
	err = mux.CheckReverse(struc)
	if !reflect.DeepEqual(err, expect) {
		t.Fatalf("unexpected error: %v", err)
	}

	struc = 1
	err = mux.CheckReverse(&struc)
	if err == nil {
		t.Fatalf("unexpected reserve: %#v", struc)
	}

	struc = goodReverse{root, hello, ignore}
	err = mux.CheckReverse(struc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	struc = struct{ root, hello, ignore string }{root, "/other/", ignore}
	expect = NonExistentRoutesError{"/other/"}
	err = mux.CheckReverse(struc)
	if !reflect.DeepEqual(err, expect) {
		t.Fatalf("unexpected error: %v", err)
	}

	_struc := (*struct{})(nil)
	struc = &_struc
	err = mux.CheckReverse(struc)
	if err == nil {
		t.Fatalf("unexpected reserve: %#v", struc)
	}

	struc = struct{ foo int }{}
	err = mux.CheckReverse(struc)
	if err == nil {
		t.Fatalf("unexpected reserve: %#v", struc)
	}
}

func TestDeref(t *testing.T) {
	_, err := deref(reflect.Value{})
	if err == nil {
		t.Fatalf("unexpected deference of invalid value.")
	}
}
