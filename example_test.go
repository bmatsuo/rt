package rt

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
)

// This shows how you can extract a path parameter using http.ServeMux.
func ExampleDecompose_parameter() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello/", func(resp http.ResponseWriter, req *http.Request) {
		_, pat := mux.Handler(req)
		name := Decompose(pat, req.URL.Path)
		fmt.Fprintln(resp, "hello", name)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	resp, err := http.Get(server.URL + "/hello/world")
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, resp.Body)
	resp.Body.Close()
	// Output: hello world
}

// This example demonstrates how to write reversible routes and how to ensure your routes are reversible at runtime.
func ExampleServeMux_CheckReverse() {
	type ExampleRoutes struct {
		Hello   string
		Goodbye string
	}
	routes := ExampleRoutes{
		"/hello/",
		"/goodbye/",
	}
	mux := NewServeMux()
	mux.HandleFunc(routes.Hello, func(resp http.ResponseWriter, req *http.Request) { fmt.Fprintln(resp, "hello world") })
	mux.HandleFunc(routes.Goodbye, func(resp http.ResponseWriter, req *http.Request) { fmt.Fprintln(resp, "goodbye") })
	fmt.Println(mux.CheckReverse(routes))

	mux.HandleFunc("/monitor", func(resp http.ResponseWriter, req *http.Request) { fmt.Fprintln(resp, "ok") })
	fmt.Println(mux.CheckReverse(routes).(IrreversibleRoutesError)) // some errors may be ignorable

	mux = NewServeMux()
	mux.HandleFunc(routes.Hello, func(resp http.ResponseWriter, req *http.Request) { fmt.Fprintln(resp, "hello world") })
	fmt.Println(mux.CheckReverse(routes))
	// Output:
	// <nil>
	// irreversible routes: ["/monitor"]
	// non-existent routes: ["/goodbye/"]
}
