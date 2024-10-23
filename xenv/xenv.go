package xenv

import (
	"bufio"
	"encoding/json"
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

	envFiles := []string{filename, filename + ".local"}
	var errs []error

	currentDir, err := os.Getwd()
	if err != nil {
		return xerror.Wrap(err, "error getting current directory")
	}

	for {
		for _, envFile := range envFiles {
			filePath := filepath.Join(currentDir, envFile)
			if x.FileExists(filePath) {
				if err := loadEnvFile(filePath); err != nil {
					errs = append(errs, xerror.Wrapf(err, "error loading file: %s", filePath))
				}
			}
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func loadEnvFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return xerror.Wrap(err, "error opening file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)

		if err := os.Setenv(key, value); err != nil {
			return xerror.Wrapf(err, "error setting environment variable %s", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return xerror.Wrap(err, "error reading file")
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
	return os.Setenv(key, value)
}

// Unset removes an environment variable
func Unset(key string) error {
	return os.Unsetenv(key)
}

// GetBool retrieves the boolean value of an environment variable
func GetBool(key string) bool {
	return x.StringToBool(os.Getenv(key))
}

// GetBoolDefault retrieves the boolean value of an environment variable or returns a default value if not set
func GetBoolDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return x.StringToBool(value)
	}
	return defaultValue
}

// GetInt retrieves the integer value of an environment variable
func GetInt(key string) (int, error) {
	return x.StringToInt(os.Getenv(key))
}

// GetIntDefault retrieves the integer value of an environment variable or returns a default value if not set
func GetIntDefault(key string, defaultValue int) int {
	if value, err := GetInt(key); err == nil {
		return value
	}
	return defaultValue
}

// GetFloat64 retrieves the float64 value of an environment variable
func GetFloat64(key string) (float64, error) {
	return x.StringToFloat64(os.Getenv(key))
}

// GetFloat64Default retrieves the float64 value of an environment variable or returns a default value if not set
func GetFloat64Default(key string, defaultValue float64) float64 {
	if value, err := GetFloat64(key); err == nil {
		return value
	}
	return defaultValue
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

// GetDurationDefault retrieves the time.Duration value of an environment variable or returns a default value if not set
func GetDurationDefault(key string, defaultValue time.Duration) time.Duration {
	if value, err := GetDuration(key); err == nil {
		return value
	}
	return defaultValue
}

// GetMap retrieves a map of strings from an environment variable
func GetMap(key, pairSep, kvSep string) map[string]string {
	return x.StringToMap(os.Getenv(key), pairSep, kvSep)
}

// GetJSON retrieves and unmarshals a JSON-encoded environment variable
func GetJSON(key string, v interface{}) error {
	return json.Unmarshal([]byte(os.Getenv(key)), v)
}

// GetJSONDefault retrieves and unmarshals a JSON-encoded environment variable or returns a default value if not set
func GetJSONDefault(key string, defaultValue interface{}, v interface{}) error {
	if value := os.Getenv(key); value != "" {
		if err := json.Unmarshal([]byte(value), v); err == nil {
			return nil
		}
	}

	defaultJSON, err := json.Marshal(defaultValue)
	if err != nil {
		return err
	}
	return json.Unmarshal(defaultJSON, v)
}
