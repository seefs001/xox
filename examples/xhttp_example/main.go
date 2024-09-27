package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/seefs001/xox/xhttp"
	"github.com/seefs001/xox/xlog"
	"github.com/seefs001/xox/xmw"
)

// UserHandler handles requests for user information
func UserHandler(c *xhttp.Context) {
	xlog.Info("Handling user request")
	userID := c.GetParam("id")
	if userID == "" {
		xlog.Warn("Missing user ID in request")
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing user ID"})
		return
	}

	// Simulate fetching user data
	user := map[string]string{
		"id":   userID,
		"name": fmt.Sprintf("User %s", userID),
	}

	xlog.Info("User data fetched", "user", user)
	c.JSON(http.StatusOK, user)
}

// HeaderEchoHandler echoes back the received headers
func HeaderEchoHandler(c *xhttp.Context) {
	xlog.Info("Handling header echo request")
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		headers[key] = values[0]
	}

	xlog.Info("Echoing headers", "headers", headers)
	c.JSON(http.StatusOK, headers)
}

// ContextValueHandler demonstrates using context values
func ContextValueHandler(c *xhttp.Context) {
	xlog.Info("Handling context value request")
	c.WithValue("key", "value")
	val := c.GetContext().Value("key")
	xlog.Info("Context value set and retrieved", "value", val)
	c.JSON(http.StatusOK, map[string]string{"contextValue": val.(string)})
}

// ParamHandler demonstrates getting parameters as different types
func ParamHandler(c *xhttp.Context) {
	xlog.Info("Handling param request")
	id, err := c.GetParamInt("id")
	if err != nil {
		xlog.Warn("Invalid ID parameter", "error", err)
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	score, err := c.GetParamFloat("score")
	if err != nil {
		xlog.Warn("Invalid score parameter", "error", err)
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid score"})
		return
	}

	xlog.Info("Parameters parsed successfully", "id", id, "score", score)
	c.JSON(http.StatusOK, map[string]interface{}{
		"id":    id,
		"score": score,
	})
}

// BodyHandler demonstrates decoding request body
func BodyHandler(c *xhttp.Context) {
	xlog.Info("Handling body request")
	var data map[string]interface{}
	if err := c.GetBody(&data); err != nil {
		xlog.Warn("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	xlog.Info("Request body decoded", "data", data)
	c.JSON(http.StatusOK, data)
}

// RedirectHandler demonstrates redirection
func RedirectHandler(c *xhttp.Context) {
	xlog.Info("Handling redirect request")
	c.Redirect(http.StatusFound, "/redirected")
}

// RedirectedHandler handles the redirected request
func RedirectedHandler(c *xhttp.Context) {
	xlog.Info("Handling redirected request")
	c.JSON(http.StatusOK, map[string]string{"message": "You've been redirected"})
}

// CookieHandler demonstrates setting and getting cookies
func CookieHandler(c *xhttp.Context) {
	xlog.Info("Handling cookie request")
	// Set a cookie
	cookie := &http.Cookie{
		Name:    "example",
		Value:   "test",
		Expires: time.Now().Add(24 * time.Hour),
	}
	c.SetCookie(cookie)
	xlog.Info("Cookie set", "cookie", cookie)

	// Get the cookie we just set
	retrievedCookie, err := c.GetCookie("example")
	if err != nil {
		xlog.Error("Failed to retrieve cookie", "error", err)
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cookie"})
		return
	}

	xlog.Info("Cookie retrieved", "cookie", retrievedCookie)
	c.JSON(http.StatusOK, map[string]string{"cookie_value": retrievedCookie.Value})
}

// NewHandler demonstrates using xmw middleware
func NewHandler(c *xhttp.Context) {
	xlog.Info("Handling new request with xmw middleware")
	c.JSON(http.StatusOK, map[string]string{"message": "This handler uses xmw middleware"})
}

func main() {
	xlog.Info("Starting xhttp example server")

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

	xlog.Info("Routes set up successfully")

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

	xlog.Info("Middleware stack defined")

	// Apply middleware to the mux
	handler := xmw.Use(mux, middlewareStack...)

	xlog.Info("Middleware applied to mux")

	// Start the server with the middleware-wrapped handler
	xlog.Info("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
