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
