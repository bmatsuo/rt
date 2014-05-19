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
func ExampleServeMux_reverse() {
	type ExampleRoutes struct {
		Sessions string
		Users    string
	}
	routes := ExampleRoutes{
		"/sessions/",
		"/users/",
	}
	mux := NewServeMux()
	mux.HandleFunc(routes.Sessions, func(resp http.ResponseWriter, req *http.Request) {
		userID := "123"
		path := Compose(routes.Users, userID)
		u := "http" + req.Host + path
		http.Redirect(resp, req, u, http.StatusSeeOther)
	})
	mux.HandleFunc(routes.Users, func(resp http.ResponseWriter, req *http.Request) { http.Error(resp, "forbidden", 403) })
	fmt.Println(mux.CheckReverse(routes))
	// Output:
	// <nil>
}
