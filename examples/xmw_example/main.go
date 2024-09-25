package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/seefs001/xox/xmw"
)

func main() {
	// Create multiple handler functions
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	timeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Current time: %s", time.Now().Format(time.RFC3339))
	})

	// Create a mux (multiplexer) to route requests
	mux := http.NewServeMux()
	mux.Handle("/", helloHandler)
	mux.Handle("/time", timeHandler)

	// Define middleware stack
	middlewareStack := []xmw.Middleware{
		xmw.Logger(),
		xmw.Recover(),
		xmw.Timeout(xmw.TimeoutConfig{
			Timeout: 5 * time.Second,
		}),
		xmw.CORS(xmw.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		}),
		xmw.Compress(),
		xmw.BasicAuth(xmw.BasicAuthConfig{
			Users: map[string]string{"user": "password"},
			Realm: "Restricted",
		}),
	}

	// Apply middleware to the mux
	finalHandler := xmw.Use(mux, middlewareStack...)

	// Start the server
	http.ListenAndServe(":8080", finalHandler)
}
