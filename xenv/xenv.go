package xenv

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xerror"
)

// LoadOptions represents options for loading environment variables
type LoadOptions struct {
	Filename string
}

// Load reads the .env file and loads the environment variables
func Load(options ...LoadOptions) error {
	filename := ".env"
	if len(options) > 0 && options[0].Filename != "" {
		filename = options[0].Filename
	}

	// Try to find the .env file in the current directory and parent directories
	currentDir, err := os.Getwd()
	if err != nil {
		return xerror.Wrap(err, "error getting current directory")
	}

	for {
		filePath := filepath.Join(currentDir, filename)
		if x.FileExists(filePath) {
			return loadEnvFile(filePath)
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break // We've reached the root directory
		}
		currentDir = parentDir
	}

	return xerror.New(".env file not found")
}

func loadEnvFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return xerror.Wrap(err, "error opening .env file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return xerror.Newf("invalid line in .env file: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		value = strings.Trim(value, `"'`)

		if err := os.Setenv(key, value); err != nil {
			return xerror.Wrapf(err, "error setting environment variable %s", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return xerror.Wrap(err, "error reading .env file")
	}

	return nil
}

// Get retrieves the value of an environment variable
func Get(key string) string {
	return os.Getenv(key)
}

// GetDefault retrieves the value of an environment variable or returns a default value if not set
func GetDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Set sets the value of an environment variable
func Set(key, value string) error {
	return xerror.Wrap(os.Setenv(key, value), "error setting environment variable")
}

// Unset removes an environment variable
func Unset(key string) error {
	return xerror.Wrap(os.Unsetenv(key), "error unsetting environment variable")
}

// GetBool retrieves the boolean value of an environment variable
func GetBool(key string) bool {
	return x.StringToBool(os.Getenv(key))
}

// GetInt retrieves the integer value of an environment variable
func GetInt(key string) (int, error) {
	return x.StringToInt(os.Getenv(key))
}

// GetFloat64 retrieves the float64 value of an environment variable
func GetFloat64(key string) (float64, error) {
	return x.StringToFloat64(os.Getenv(key))
}

// GetSlice retrieves a slice of strings from an environment variable
func GetSlice(key, sep string) []string {
	return strings.Split(os.Getenv(key), sep)
}

// MustGet retrieves the value of an environment variable, panics if not set
func MustGet(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(xerror.Newf("environment variable %s is not set", key))
	}
	return value
}

// GetInt64 retrieves the int64 value of an environment variable
func GetInt64(key string) (int64, error) {
	return x.StringToInt64(os.Getenv(key))
}

// GetUint retrieves the uint value of an environment variable
func GetUint(key string) (uint, error) {
	return x.StringToUint(os.Getenv(key))
}

// GetUint64 retrieves the uint64 value of an environment variable
func GetUint64(key string) (uint64, error) {
	return x.StringToUint64(os.Getenv(key))
}

// GetDuration retrieves the time.Duration value of an environment variable
func GetDuration(key string) (time.Duration, error) {
	return x.StringToDuration(os.Getenv(key))
}

// GetMap retrieves a map of strings from an environment variable
func GetMap(key, pairSep, kvSep string) map[string]string {
	return x.StringToMap(os.Getenv(key), pairSep, kvSep)
}

// GetJSON retrieves and unmarshals a JSON-encoded environment variable
func GetJSON(key string, v interface{}) error {
	return x.UnmarshalJSON([]byte(os.Getenv(key)), v)
}
