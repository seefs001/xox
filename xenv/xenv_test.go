package xenv_test

import (
	"os"
	"testing"
	"time"

	"github.com/seefs001/xox/xenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Create a temporary .env file
	content := `
TEST_KEY=test_value
TEST_BOOL=true
TEST_INT=42
TEST_FLOAT=3.14
TEST_SLICE=a,b,c
TEST_JSON={"key": "value"}
`
	tmpfile, err := os.CreateTemp("", "test.env")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	tmpfile.Close()

	// Test Load function
	err = xenv.Load(xenv.LoadOptions{Filename: tmpfile.Name()})
	assert.NoError(t, err)

	// Test Get function
	assert.Equal(t, "test_value", xenv.Get("TEST_KEY"))

	// Test GetDefault function
	assert.Equal(t, "test_value", xenv.GetDefault("TEST_KEY", "default"))
	assert.Equal(t, "default", xenv.GetDefault("NON_EXISTENT", "default"))

	// Test GetBool function
	assert.True(t, xenv.GetBool("TEST_BOOL"))

	// Test GetInt function
	intVal, err := xenv.GetInt("TEST_INT")
	assert.NoError(t, err)
	assert.Equal(t, 42, intVal)

	// Test GetFloat64 function
	floatVal, err := xenv.GetFloat64("TEST_FLOAT")
	assert.NoError(t, err)
	assert.Equal(t, 3.14, floatVal)

	// Test GetSlice function
	assert.Equal(t, []string{"a", "b", "c"}, xenv.GetSlice("TEST_SLICE", ","))

	// Test GetJSON function
	var jsonData map[string]string
	err = xenv.GetJSON("TEST_JSON", &jsonData)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"key": "value"}, jsonData)
}

func TestSetUnset(t *testing.T) {
	// Test Set function
	err := xenv.Set("TEST_SET", "set_value")
	assert.NoError(t, err)
	assert.Equal(t, "set_value", xenv.Get("TEST_SET"))

	// Test Unset function
	err = xenv.Unset("TEST_SET")
	assert.NoError(t, err)
	assert.Empty(t, xenv.Get("TEST_SET"))

	// Test Set with empty value
	err = xenv.Set("TEST_EMPTY", "")
	assert.NoError(t, err)
	assert.Empty(t, xenv.Get("TEST_EMPTY"))

	// Test Unset non-existent variable
	err = xenv.Unset("NON_EXISTENT")
	assert.NoError(t, err)
}

func TestMustGet(t *testing.T) {
	// Set a test environment variable
	os.Setenv("TEST_MUST_GET", "must_get_value")

	// Test MustGet function
	assert.Equal(t, "must_get_value", xenv.MustGet("TEST_MUST_GET"))

	// Test MustGet function with non-existent key
	assert.Panics(t, func() { xenv.MustGet("NON_EXISTENT") })
}

func TestGetDuration(t *testing.T) {
	// Set a test environment variable
	os.Setenv("TEST_DURATION", "5s")

	// Test GetDuration function
	duration, err := xenv.GetDuration("TEST_DURATION")
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, duration)
}

func TestGetMap(t *testing.T) {
	// Set a test environment variable
	os.Setenv("TEST_MAP", "key1:value1,key2:value2")

	// Test GetMap function
	result := xenv.GetMap("TEST_MAP", ",", ":")
	expected := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	assert.Equal(t, expected, result)
}
