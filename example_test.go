package rt

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
)

func curl(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, resp.Body)
	resp.Body.Close()
}

// This shows how you can extract a path parameter using http.ServeMux.
func ExampleDecompose_parameter() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello/", func(resp http.ResponseWriter, req *http.Request) {
		_, pat := mux.Handler(req)
		name := Decompose(pat, req.URL.Path)
		fmt.Fprintf(resp, "hello %s!", name)
	})

	// this is where you would call http.ListenAndServe() and be done.
	server := httptest.NewServer(mux)
	defer server.Close()
	curl(server.URL + "/hello/world")
	// Output: hello world!
}

// This is the most trivial example of rt.  In fact it works with plain old
// http.ServeMux.
func Example_hello() {
	mux := NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world!")
	})

	// this is where you would call http.ListenAndServe() and be done.
	server := httptest.NewServer(mux)
	defer server.Close()
	curl(server.URL + "/hello")
	// Output: hello world!
}

// This complete example demonstrates how to implement and use safe, reversible
// routes with rt.  The actual server logic is bizarre so don't pay too much
// attention.
func Example_reverse() {
	type Server struct {
		rts struct {
			Sessions string `rt:"/v1/sessions/"`
			Users    string `rt:"/v1/users/"`
		}
	}
	newServer := func() (*Server, error) {
		s := new(Server)
		err := Struct(&s.rts)
		return s, err
	}
	// in practice this is a method on the Server type.
	httpRoutes := func(server *Server) (mux *ServeMux, err error) {
		mux = NewServeMux()
		defer func() {
			err = mux.CheckReverse(server.rts)
		}()
		mux.HandleFunc(server.rts.Sessions, func(w http.ResponseWriter, r *http.Request) {
			// locate the user associated with the session id path parameter.
			_, pat := mux.Handler(r)
			sessionID := Decompose(pat, r.URL.Path)
			userID, err := base64.URLEncoding.DecodeString(sessionID)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			// redirect the client to the associated user resource.
			u := "http://" + r.Host
			u += Compose(server.rts.Users, string(userID))
			http.Redirect(w, r, u, http.StatusSeeOther)
		})
		mux.HandleFunc(server.rts.Users, func(w http.ResponseWriter, r *http.Request) {
			_, pat := mux.Handler(r)
			userID := Decompose(pat, r.URL.Path)
			fmt.Fprintf(w, "hello %s!", userID)
		})

		return // naked return is required for deferred error handling
	}

	s, err := newServer()
	if err != nil {
		log.Fatal(err)
	}
	mux, err := httpRoutes(s)
	if err != nil {
		log.Fatal(err)
	}

	// this is where you would call http.ListenAndServe() and be done.
	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()
	sessionID := base64.URLEncoding.EncodeToString([]byte("world"))
	curl(httpServer.URL + Compose(s.rts.Sessions, sessionID))
	// Output: hello world!
}

// This example shows how Struct() populates struct field values.
func ExampleStruct() {
	server := new(struct {
		rts struct {
			Users      string `rt:"/v1/users/"`
			Pets       string `rt:"/v1/pets/"`
			Deductions string `rt:"/v1/deductions/"`
		}
	})
	server.rts.Deductions = "/v1/tax_deductions/"
	err := Struct(&server.rts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(server.rts)
	// Output:
	// {/v1/users/ /v1/pets/ /v1/tax_deductions/}
}
