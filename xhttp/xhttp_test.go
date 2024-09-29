package xhttp_test

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/seefs001/xox/xhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseWriter(t *testing.T) {
	t.Run("WriteJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		rw := &xhttp.ResponseWriter{ResponseWriter: w}
		data := map[string]string{"message": "hello"}

		err := rw.WriteJSON(data)
		assert.NoError(t, err)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var result map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, data, result)
	})

	t.Run("WriteXML", func(t *testing.T) {
		w := httptest.NewRecorder()
		rw := &xhttp.ResponseWriter{ResponseWriter: w}
		data := struct {
			XMLName xml.Name `xml:"response"`
			Message string   `xml:"message"`
		}{Message: "hello"}

		err := rw.WriteXML(&data)
		assert.NoError(t, err)
		assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))

		var result struct {
			XMLName xml.Name `xml:"response"`
			Message string   `xml:"message"`
		}
		err = xml.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, data.Message, result.Message)
	})
}

func TestContext(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		c := xhttp.NewContext(w, r)

		data := map[string]string{"message": "hello"}
		err := c.JSON(http.StatusOK, data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var result map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, data, result)
	})

	t.Run("XML", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		c := xhttp.NewContext(w, r)

		data := struct {
			XMLName xml.Name `xml:"response"`
			Message string   `xml:"message"`
		}{Message: "hello"}
		err := c.XML(http.StatusOK, &data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))

		var result struct {
			XMLName xml.Name `xml:"response"`
			Message string   `xml:"message"`
		}
		err = xml.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, data.Message, result.Message)
	})

	t.Run("String", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		c := xhttp.NewContext(w, r)

		err := c.String(http.StatusOK, "Hello, %s!", "world")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		assert.Equal(t, "Hello, world!", w.Body.String())
	})

	t.Run("GetParam", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?key=value", nil)
		c := xhttp.NewContext(httptest.NewRecorder(), r)

		assert.Equal(t, "value", c.GetParam("key"))
	})

	t.Run("GetHeader", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Test", "test-value")
		c := xhttp.NewContext(httptest.NewRecorder(), r)

		assert.Equal(t, "test-value", c.GetHeader("X-Test"))
	})

	t.Run("SetHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		c := xhttp.NewContext(w, httptest.NewRequest("GET", "/", nil))

		c.SetHeader("X-Test", "test-value")
		assert.Equal(t, "test-value", w.Header().Get("X-Test"))
	})

	t.Run("GetBody", func(t *testing.T) {
		body := strings.NewReader(`{"message": "hello"}`)
		r := httptest.NewRequest("POST", "/", body)
		r.Header.Set("Content-Type", "application/json")
		c := xhttp.NewContext(httptest.NewRecorder(), r)

		var data map[string]string
		err := c.GetBody(&data)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"message": "hello"}, data)
	})
}

func TestRouter(t *testing.T) {
	t.Run("HandleFunc", func(t *testing.T) {
		router := xhttp.NewRouter()
		router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, "OK", w.Body.String())
	})

	t.Run("Group", func(t *testing.T) {
		router := xhttp.NewRouter()
		group := router.Group("/api")
		group.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("API OK"))
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/test", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, "API OK", w.Body.String())
	})
}

func TestWrap(t *testing.T) {
	handler := xhttp.Wrap(func(c *xhttp.Context) {
		c.String(http.StatusOK, "Hello, %s!", "world")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello, world!", w.Body.String())
}

func TestBind(t *testing.T) {
	type TestStruct struct {
		Name  string `form:"name" json:"name"`
		Age   int    `form:"age" json:"age"`
		Email string `form:"email" json:"email"`
	}

	t.Run("BindQuery", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?name=John&age=30&email=john@example.com", nil)
		c := xhttp.NewContext(httptest.NewRecorder(), r)

		var data TestStruct
		err := c.Bind(&data)
		require.NoError(t, err)
		assert.Equal(t, "John", data.Name)
		assert.Equal(t, 30, data.Age)
		assert.Equal(t, "john@example.com", data.Email)
	})

	t.Run("BindJSON", func(t *testing.T) {
		body := strings.NewReader(`{"name":"Jane","age":25,"email":"jane@example.com"}`)
		r := httptest.NewRequest("POST", "/", body)
		r.Header.Set("Content-Type", "application/json")
		c := xhttp.NewContext(httptest.NewRecorder(), r)

		var data TestStruct
		err := c.Bind(&data)
		require.NoError(t, err)
		assert.Equal(t, "Jane", data.Name)
		assert.Equal(t, 25, data.Age)
		assert.Equal(t, "jane@example.com", data.Email)
	})
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
	}{
		{
			name:     "X-Forwarded-For",
			headers:  map[string]string{"X-Forwarded-For": "203.0.113.195"},
			expected: "203.0.113.195",
		},
		{
			name:     "X-Real-IP",
			headers:  map[string]string{"X-Real-IP": "203.0.113.195"},
			expected: "203.0.113.195",
		},
		{
			name:     "RemoteAddr",
			headers:  map[string]string{},
			expected: "192.0.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}
			r.RemoteAddr = "192.0.2.1:1234"
			c := xhttp.NewContext(httptest.NewRecorder(), r)

			assert.Equal(t, tt.expected, c.ClientIP())
		})
	}
}
