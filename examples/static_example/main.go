package main

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/seefs001/xox/xmw"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	// Create a new mux for routing
	mux := http.NewServeMux()

	// API handlers
	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from API!")
	})

	// Create static file configuration
	staticConfig := xmw.StaticConfig{
		Root:      "static",
		Index:     "index.html",
		Browse:    false,
		MaxAge:    3600, // 1 hour cache
		Prefix:    "/",
		APIPrefix: "/api",
		EmbedFS:   staticFiles,
		Files:     make(map[string]xmw.File),
	}

	// Add a dynamic file
	dynamicContent := strings.NewReader("This is a dynamic file content.")
	staticConfig.AddFile("/dynamic.txt", dynamicContent, "text/plain", time.Now())

	// Create middleware stack
	middlewareStack := []xmw.Middleware{
		xmw.Logger(),
		xmw.Recover(),
		xmw.Static(staticConfig),
	}

	// Apply middleware to the mux
	handler := xmw.Use(mux, middlewareStack...)

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", handler)
}
