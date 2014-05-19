/*
Package rt provides route composition.

This API is experimental.
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
	mux.mut.Unlock()
	mux.mux.Handle(pat, h)
}

// HandlerFunc behaves the same as the corresponding http.ServeMux method.
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

// NonReversableRoutes are routes which cannot be produced with a reverse
// mapping.
type IrreversibleRoutesError []string

func (err IrreversibleRoutesError) Error() string {
	return fmt.Sprintf("irreversible routes: %q", []string(err))
}

// NonExistentRoutes are routes which are present in a ServeMux.
type NonExistentRoutesError []string

func (err NonExistentRoutesError) Error() string {
	return fmt.Sprintf("non-existent routes: %q", []string(err))
}

// Reversible checks if structure is reverse map for patterns handled by mux.
// Structure must be a struct with only string fields.
func (mux *ServeMux) Reversible(structure interface{}) error {
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
