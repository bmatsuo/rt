/*
Package rt expands parameter and reverse-route support for http.ServeMux.

This API is experimental.

Route Parameters

The Decompose function allows access to "suffix parameters" of HTTP request
paths.  The syntax for this function is compatible with http.ServeMux and the
two are meant to be used together..

Reverse Routes

The Compose function provides path construction for routes.  An abstraction for
string concatenation following the rules of http.ServeMux. Compose is flexible
and can easily be used with http.ServeMux.

For safer reverse routing rt defines the ServeMux type with an API identical to
http.ServeMux except for an additional method, CheckReverse, to help ensure
runtime references to routes are valid.

The final tool rt provides for building reverse routes is the Struct function.
It provides a convenient way to define reversible routes.
*/
package rt

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// HostPath returns pat's host and path parts.
func HostPath(pat string) (host, path string) {
	if pat == "" {
		return "", ""
	}
	if pat[0] == '/' {
		return "", pat
	}
	i := strings.Index(pat, "/")
	if i < 0 {
		return pat, "" // this is weird. but consistent with stdlib.
	}
	return pat[:i], pat[i:]
}

// Compose creates a path from pat with parameter s.
// Compose returns pat if pat does not take a route parameter.
func Compose(pat, s string) string {
	_, pat = HostPath(pat)
	if !strings.HasSuffix(pat, "/") {
		return pat
	}
	return pat + s
}

// Decompose returns the parameter of pat that matches path.
// Decompose returns the empty string if pat doesn't match path.
// If pat is a host-specific pattern the host is ignored.
func Decompose(pat, path string) string {
	_, pat = HostPath(pat)
	if !strings.HasSuffix(pat, "/") {
		return ""
	}
	suf := strings.TrimPrefix(path, pat)
	if suf == path {
		return ""
	}
	return suf
}

// ServeMux is a drop-in replacement for http.ServeMux that can be checked for
// reversable routes.
type ServeMux struct {
	mut  sync.Mutex
	pats []string
	mux  *http.ServeMux
}

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{mux: http.NewServeMux()}
}

// Handle behaves the same as the corresponding http.ServeMux method.
func (mux *ServeMux) Handle(pat string, h http.Handler) {
	mux.mut.Lock()
	mux.pats = append(mux.pats, pat)
	mux.mux.Handle(pat, h)
	mux.mut.Unlock()
}

// HandleFunc behaves the same as the corresponding http.ServeMux method.
func (mux *ServeMux) HandleFunc(pat string, h http.HandlerFunc) {
	mux.Handle(pat, h)
}

// Handler behaves the same as the corresponding http.ServeMux method.
func (mux *ServeMux) Handler(req *http.Request) (http.Handler, string) {
	return mux.mux.Handler(req)
}

// ServeHTTP behaves the same as the corresponding http.ServeMux method.
func (mux *ServeMux) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	mux.mux.ServeHTTP(resp, req)
}

// IrreversibleRoutesError contains routes which cannot be produced with a
// potential reverse mapping.
type IrreversibleRoutesError []string

func (err IrreversibleRoutesError) Error() string {
	return fmt.Sprintf("irreversible routes: %q", []string(err))
}

// NonExistentRoutesError contains routes from a potential reverse mapping that
// are not present in a ServeMux.
type NonExistentRoutesError []string

func (err NonExistentRoutesError) Error() string {
	return fmt.Sprintf("non-existent routes: %q", []string(err))
}

// CheckReverse ensures structure is a reverse map for routes handled by mux.
// Structure must be a struct containing only string fields.
func (mux *ServeMux) CheckReverse(structure interface{}) error {
	val, err := deref(reflect.ValueOf(structure))
	if err != nil {
		return err
	}
	kind := val.Kind()
	if kind != reflect.Struct {
		return fmt.Errorf("non-struct value")
	}
	var all []string
	numField := val.NumField()
	for i := 0; i < numField; i++ {
		field := val.Field(i)
		fkind := field.Kind()
		if fkind != reflect.String {
			return fmt.Errorf("non-string field %q", val.Type().Field(i).Name)
		}
		fstring := field.String()
		all = append(all, fstring)
	}
	sort.Strings(all)

	mux.mut.Lock()
	defer mux.mut.Unlock()
	sort.Strings(mux.pats)

	var irrev IrreversibleRoutesError
	var notexist NonExistentRoutesError
	i, j := 0, 0
	n, m := len(mux.pats), len(all)
	for i < n && j < m {
		switch {
		case mux.pats[i] < all[j]:
			irrev = append(irrev, mux.pats[i])
			i++
		case mux.pats[i] > all[j]:
			notexist = append(notexist, all[j])
			j++
		default:
			i++
			j++
		}
	}
	if i < n {
		irrev = append(irrev, mux.pats[i:]...)
	}
	if j < m {
		notexist = append(notexist, all[j:]...)
	}
	if len(notexist) > 0 {
		return notexist
	}
	if len(irrev) > 0 {
		return irrev
	}
	return nil
}

// deref repeatedly dereferences val until the result is not a pointer.
func deref(val reflect.Value) (reflect.Value, error) {
	if !val.IsValid() {
		return val, fmt.Errorf("invalid value")
	}
	kind := val.Kind()
	if kind != reflect.Ptr {
		return val, nil
	}
	if val.IsNil() {
		return reflect.Value{}, fmt.Errorf("nil")
	}
	return deref(reflect.Indirect(val))
}

// Struct fills the fields of the struct at structptr with values from the
// field tags.
func Struct(structptr interface{}) error {
	val := reflect.ValueOf(structptr)
	kind := val.Kind()
	if kind != reflect.Ptr {
		return fmt.Errorf("not a struct pointer")
	}
	val, err := deref(val)
	if err != nil {
		return err
	}
	kind = val.Kind()
	if kind != reflect.Struct {
		return fmt.Errorf("not a struct pointer")
	}
	typ := val.Type()
	numField := val.NumField()
	for i := 0; i < numField; i++ {
		fieldv := val.Field(i)
		fieldt := typ.Field(i)
		fieldk := fieldt.Type.Kind()
		if fieldk != reflect.String {
			return fmt.Errorf("non-string field %q (%s)", fieldt.Name, fieldt.Type.Name())
		}
		if fieldv.Len() != 0 {
			continue
		}
		tag := fieldt.Tag.Get("rt")
		if tag == "" {
			return fmt.Errorf("empty field with no tag value %q", fieldt.Name)
		}
		fieldv.Set(reflect.ValueOf(tag))
	}
	return nil
}
