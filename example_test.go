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

// This complete example demonstrates how to write reversible routes with rt
// with maximum runtime safety.  The actual server logic is bizarre.  But it
// demonstrates all the features of rt.
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
	// in practice this is method on the Server type.
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
			path := Compose(server.rts.Users, string(userID))
			u := "http://" + r.Host + path
			http.Redirect(w, r, u, http.StatusSeeOther)
		})
		mux.HandleFunc(server.rts.Users, func(w http.ResponseWriter, r *http.Request) {
			_, pat := mux.Handler(r)
			userID := Decompose(pat, r.URL.Path)
			json.NewEncoder(w).Encode(map[string]string{
				"id": userID,
			})
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

	// here you would call http.ListenAndServe() and be done.
	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	// make a request to the new server.
	sessionID := base64.URLEncoding.EncodeToString([]byte("123"))
	resp, err := http.Get(httpServer.URL + Compose(s.rts.Sessions, sessionID))
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, resp.Body)
	resp.Body.Close()
	// Output:
	// {"id":"123"}
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
