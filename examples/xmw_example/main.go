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
		sm := xmw.GetSessionManager(r)
		if sm != nil {
			if name, ok := sm.Get(r, "name"); ok {
				fmt.Fprintf(w, "Hello, %s!", name)
				return
			}
		}
		fmt.Fprintf(w, "Hello, World!")
	})

	timeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Current time: %s", time.Now().Format(time.RFC3339))
	})

	setNameHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sm := xmw.GetSessionManager(r)
		if sm != nil {
			name := r.URL.Query().Get("name")
			if name != "" {
				sm.Set(r, "name", name)
				fmt.Fprintf(w, "Name set to: %s", name)
			} else {
				fmt.Fprintf(w, "Please provide a name using the 'name' query parameter")
				return
			}
		}
	})

	clearSessionHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sm := xmw.GetSessionManager(r)
		if sm != nil {
			sm.Clear(r)
			fmt.Fprintf(w, "Session cleared")
		}
	})

	// Create a mux (multiplexer) to route requests
	mux := http.NewServeMux()
	mux.Handle("/", helloHandler)
	mux.Handle("/time", timeHandler)
	mux.Handle("/setname", setNameHandler)
	mux.Handle("/clear", clearSessionHandler)

	// Define middleware stack
	// middlewareStack := []xmw.Middleware{
	// 	xmw.Logger(),
	// 	xmw.Recover(),
	// 	xmw.Timeout(xmw.TimeoutConfig{
	// 		Timeout: 5 * time.Second,
	// 	}),
	// 	xmw.CORS(xmw.CORSConfig{
	// 		AllowOrigins: []string{"*"},
	// 		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	}),
	// 	xmw.Compress(),
	// 	xmw.Session(xmw.SessionConfig{
	// 		Store:       xmw.NewMemoryStore(),
	// 		CookieName:  "session_id",
	// 		MaxAge:      3600,             // 1 hour
	// 		SessionName: "custom_session", // You can specify a custom session name here
	// 	}),
	// }

	// Apply middleware to the mux
	// finalHandler := xmw.Use(mux, middlewareStack...)
	finalHandler := xmw.Use(mux, xmw.DefaultMiddlewareSet...)

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", finalHandler)
}
