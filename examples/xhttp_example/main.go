package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/seefs001/xox/xhttp"
	"github.com/seefs001/xox/xmw"
)

// UserHandler handles requests for user information
func UserHandler(c *xhttp.Context) {
	userID := c.GetParam("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing user ID"})
		return
	}

	// Simulate fetching user data
	user := map[string]string{
		"id":   userID,
		"name": fmt.Sprintf("User %s", userID),
	}

	c.JSON(http.StatusOK, user)
}

// HeaderEchoHandler echoes back the received headers
func HeaderEchoHandler(c *xhttp.Context) {
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		headers[key] = values[0]
	}

	c.JSON(http.StatusOK, headers)
}

// ContextValueHandler demonstrates using context values
func ContextValueHandler(c *xhttp.Context) {
	c.WithValue("key", "value")
	val := c.GetContext().Value("key")
	c.JSON(http.StatusOK, map[string]string{"contextValue": val.(string)})
}

// ParamHandler demonstrates getting parameters as different types
func ParamHandler(c *xhttp.Context) {
	id, err := c.GetParamInt("id")
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	score, err := c.GetParamFloat("score")
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid score"})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id":    id,
		"score": score,
	})
}

// BodyHandler demonstrates decoding request body
func BodyHandler(c *xhttp.Context) {
	var data map[string]interface{}
	if err := c.GetBody(&data); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	c.JSON(http.StatusOK, data)
}

// RedirectHandler demonstrates redirection
func RedirectHandler(c *xhttp.Context) {
	c.Redirect(http.StatusFound, "/redirected")
}

// RedirectedHandler handles the redirected request
func RedirectedHandler(c *xhttp.Context) {
	c.JSON(http.StatusOK, map[string]string{"message": "You've been redirected"})
}

// CookieHandler demonstrates setting and getting cookies
func CookieHandler(c *xhttp.Context) {
	// Set a cookie
	cookie := &http.Cookie{
		Name:    "example",
		Value:   "test",
		Expires: time.Now().Add(24 * time.Hour),
	}
	c.SetCookie(cookie)

	// Get the cookie we just set
	retrievedCookie, err := c.GetCookie("example")
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cookie"})
		return
	}

	c.JSON(http.StatusOK, map[string]string{"cookie_value": retrievedCookie.Value})
}

// NewHandler demonstrates using xmw middleware
func NewHandler(c *xhttp.Context) {
	c.JSON(http.StatusOK, map[string]string{"message": "This handler uses xmw middleware"})
}

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Set up routes
	mux.HandleFunc("/user", xhttp.Wrap(UserHandler))
	mux.HandleFunc("/echo-headers", xhttp.Wrap(HeaderEchoHandler))
	mux.HandleFunc("/context-value", xhttp.Wrap(ContextValueHandler))
	mux.HandleFunc("/params", xhttp.Wrap(ParamHandler))
	mux.HandleFunc("/body", xhttp.Wrap(BodyHandler))
	mux.HandleFunc("/redirect", xhttp.Wrap(RedirectHandler))
	mux.HandleFunc("/redirected", xhttp.Wrap(RedirectedHandler))
	mux.HandleFunc("/cookie", xhttp.Wrap(CookieHandler))

	// Add a new route that uses xmw middleware
	mux.HandleFunc("/new", xhttp.Wrap(NewHandler))

	// Define middleware stack
	middlewareStack := []xmw.Middleware{
		xmw.Logger(xmw.LoggerConfig{
			Output:   log.Writer(),
			UseColor: true,
		}),
		xmw.Recover(xmw.RecoverConfig{
			EnableStackTrace: true,
		}),
		xmw.Timeout(xmw.TimeoutConfig{
			Timeout: 5 * time.Second,
		}),
		xmw.CORS(xmw.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		}),
		xmw.Compress(xmw.CompressConfig{
			Level: 5,
		}),
	}

	// Apply middleware to the mux
	handler := xmw.Use(mux, middlewareStack...)

	// Start the server with the middleware-wrapped handler
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
