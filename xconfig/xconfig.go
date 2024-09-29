package xconfig

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xcast"
	"github.com/seefs001/xox/xerror"
	"github.com/seefs001/xox/xlog"
)

// Config holds the configuration values.
type Config struct {
	data   map[string]any
	mu     sync.RWMutex
	prefix string
}

// defaultConfig is the default instance of Config.
var defaultConfig = NewConfig()

// NewConfig creates a new Config instance.
func NewConfig() *Config {
	return &Config{
		data: make(map[string]any),
	}
}

// LoadFromJSON loads configuration from a JSON string.
func (c *Config) LoadFromJSON(jsonStr string) error {
	var jsonData map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		return xerror.Wrap(err, "failed to unmarshal JSON")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FlattenMap(jsonData, "")
	return nil
}

// LoadFromEnv loads configuration from environment variables or a specified file.
func (c *Config) LoadFromEnv(options ...func(*Config)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Apply options first to set the prefix if any
	for _, option := range options {
		option(c)
	}

	// Load environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		// If prefix is set, only load variables with that prefix
		if c.prefix != "" {
			if strings.HasPrefix(key, c.prefix) {
				c.data[strings.TrimPrefix(key, c.prefix)] = value
			}
		} else {
			// If no prefix is set, load all variables
			c.data[key] = value
		}
	}

	return nil
}

// WithEnvPrefix sets the prefix for environment variables.
func WithEnvPrefix(prefix string) func(*Config) {
	return func(c *Config) {
		c.prefix = prefix
	}
}

// WithEnvFile loads configuration from a specified file.
func WithEnvFile(filePath string) func(*Config) {
	return func(c *Config) {
		file, err := os.Open(filePath)
		if err != nil {
			xlog.Error("failed to open env file: %v", err)
			return
		}
		defer file.Close()

		var jsonData map[string]any
		if err := json.NewDecoder(file).Decode(&jsonData); err != nil {
			xlog.Error("failed to decode env file: %v", err)
			return
		}

		c.FlattenMap(jsonData, "")
	}
}

// GetInt retrieves a value by key and converts it to an int.
func (c *Config) GetInt(key string) (int, error) {
	value, err := c.get(key)
	if err != nil {
		return 0, err
	}
	return xcast.ToInt(value)
}

// GetInt32 retrieves a value by key and converts it to an int32.
func (c *Config) GetInt32(key string) (int32, error) {
	value, err := c.get(key)
	if err != nil {
		return 0, err
	}
	return xcast.ToInt32(value)
}

// GetString retrieves a value by key and converts it to a string.
func (c *Config) GetString(key string) (string, error) {
	value, err := c.get(key)
	if err != nil {
		return "", err
	}
	return xcast.ToString(value)
}

// GetBool retrieves a value by key and converts it to a bool.
func (c *Config) GetBool(key string) (bool, error) {
	value, err := c.get(key)
	if err != nil {
		return false, err
	}
	return xcast.ToBool(value)
}

// get retrieves a value by key.
func (c *Config) get(key string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.data[key]
	if !exists {
		return nil, xerror.Errorf("key %s not found", key)
	}
	return value, nil
}

// Put sets a value by key.
func (c *Config) Put(key string, value any) {
	c.mu.Lock()
	c.data[key] = value
	c.mu.Unlock()
}

// LoadFromStruct loads configuration from a struct, using struct tags to map fields.
func (c *Config) LoadFromStruct(s any) error {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return xerror.New("input must be a struct")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loadStructFields(v, "")
}

func (c *Config) loadStructFields(v reflect.Value, prefix string) error {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			// Skip unexported fields
			continue
		}
		tag := field.Tag.Get("config")
		if tag == "" {
			tag = field.Name
		}
		if prefix != "" {
			tag = prefix + "." + tag
		}
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// Recursively handle embedded structs
			if err := c.loadStructFields(v.Field(i), prefix); err != nil {
				return err
			}
		} else {
			c.data[tag] = v.Field(i).Interface()
		}
	}
	return nil
}

// LoadFromMap loads configuration from a map[string]any.
func (c *Config) LoadFromMap(data map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, value := range data {
		c.data[key] = value
	}
}

// LoadFromStringMap loads configuration from a map[string]string.
func (c *Config) LoadFromStringMap(data map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, value := range data {
		c.data[key] = value
	}
}

// GetAll returns all configuration data.
func (c *Config) GetAll() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return x.CopyMap(c.data)
}

// FlattenMap flattens a nested map into a single-level map with dot notation keys.
func (c *Config) FlattenMap(data map[string]any, prefix string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := value.(type) {
		case map[string]any:
			c.FlattenMap(v, fullKey)
		case []any:
			for i, item := range v {
				indexStr, err := xcast.ToString(i)
				if err != nil {
					xlog.Error("Failed to convert index to string: %v", err)
					continue
				}
				itemKey := fullKey + "." + indexStr
				if subMap, ok := item.(map[string]any); ok {
					c.FlattenMap(subMap, itemKey)
				} else {
					c.data[itemKey] = item
				}
			}
		default:
			c.data[fullKey] = value
		}
	}
}

// ParseToStruct parses the configuration to a struct.
func (c *Config) ParseToStruct(result interface{}) error {
	jsonData, err := json.Marshal(c.GetAll())
	if err != nil {
		return xerror.Wrap(err, "failed to marshal config data")
	}
	if err := json.Unmarshal(jsonData, result); err != nil {
		return xerror.Wrap(err, "failed to unmarshal config data to struct")
	}
	return nil
}

// Default methods to interact with the defaultConfig instance

func LoadFromJSON(jsonStr string) error {
	return defaultConfig.LoadFromJSON(jsonStr)
}

func LoadFromEnv(options ...func(*Config)) error {
	return defaultConfig.LoadFromEnv(options...)
}

func GetInt(key string) (int, error) {
	return defaultConfig.GetInt(key)
}

func GetInt32(key string) (int32, error) {
	return defaultConfig.GetInt32(key)
}

func GetString(key string) (string, error) {
	return defaultConfig.GetString(key)
}

func GetBool(key string) (bool, error) {
	return defaultConfig.GetBool(key)
}

func Put(key string, value any) {
	defaultConfig.Put(key, value)
}

func LoadFromStruct(s any) error {
	return defaultConfig.LoadFromStruct(s)
}

func LoadFromMap(data map[string]any) {
	defaultConfig.LoadFromMap(data)
}

func LoadFromStringMap(data map[string]string) {
	defaultConfig.LoadFromStringMap(data)
}

func GetAll() map[string]any {
	return defaultConfig.GetAll()
}

// ParseToStruct parses the default configuration to a struct using generics.
func ParseToStruct[T any]() (T, error) {
	var result T
	err := defaultConfig.ParseToStruct(&result)
	return result, err
}

// LoadFromStructGeneric loads configuration from a struct using generics to the default config.
func LoadFromStructGeneric[T any](s T) error {
	return defaultConfig.LoadFromStruct(s)
}

// GetDefaultConfig returns the default configuration instance.
func GetDefaultConfig() *Config {
	return defaultConfig
}
