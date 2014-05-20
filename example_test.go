package rt

import (
	"encoding/base64"
	"encoding/json"
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
	rts := struct {
		Sessions string
		Users    string
	}{
		"/sessions/",
		"/users/",
	}
	mux := NewServeMux()
	mux.HandleFunc(rts.Sessions, func(w http.ResponseWriter, r *http.Request) {
		// locate the user associated with the session id path parameter.
		_, pat := mux.Handler(r)
		sessionID := Decompose(pat, r.URL.Path)
		userID, err := base64.URLEncoding.DecodeString(sessionID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// redirect the client to the associated user resource.
		path := Compose(rts.Users, string(userID))
		u := "http://" + r.Host + path
		http.Redirect(w, r, u, http.StatusSeeOther)
	})
	mux.HandleFunc(rts.Users, func(w http.ResponseWriter, r *http.Request) {
		_, pat := mux.Handler(r)
		userID := Decompose(pat, r.URL.Path)
		json.NewEncoder(w).Encode(map[string]string{
			"id": userID,
		})
	})
	err := mux.CheckReverse(rts) // ensure that routes are reversible
	if err != nil {
		log.Fatal(err)
	}

	server := httptest.NewServer(mux)
	defer server.Close()
	sessionID := base64.URLEncoding.EncodeToString([]byte("123"))
	resp, err := http.Get(server.URL + Compose(rts.Sessions, sessionID))
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, resp.Body)
	resp.Body.Close()
	// Output:
	// {"id":"123"}
}

// Basic usage of Struct() to implement reverse routes.  Deferring with
// mux.CheckReverse(rts) is important as route structure becomes more complex.
func ExampleStruct() {
	// in practice, server would be a type and httpRoutes should be a method
	// on that type.
	server := new(struct {
		rts struct {
			Users      string `rt:"/v1/users/"`
			Pets       string `rt:"/v1/pets/"`
			Deductions string `rt:"/v1/deductions/"`
		}
	})
	httpRoutes := func() (mux *ServeMux, err error) {
		err = Struct(&server.rts)
		if err != nil {
			return nil, err
		}
		fmt.Println(server.rts)
		mux = NewServeMux()
		defer func() { err = mux.CheckReverse(server.rts) }()

		ok := func(w http.ResponseWriter, r *http.Request) { fmt.Println("ok") }
		mux.HandleFunc(server.rts.Users, ok)
		mux.HandleFunc(server.rts.Pets, ok)
		mux.HandleFunc(server.rts.Deductions, ok)

		return // naked return is required here.
	}

	mux, err := httpRoutes()
	if err != nil {
		fmt.Println(err)
	}
	_ = mux
	// ... wrap mux up with middleware or call http.ListenAndServe()

	// Output:
	// {/v1/users/ /v1/pets/ /v1/deductions/}
}
