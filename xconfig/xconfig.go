package xconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/seefs001/xox/xcast"
)

// Config holds the configuration values.
type Config struct {
	data map[string]any
	mu   sync.RWMutex
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
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FlattenMap(jsonData, "")
	return nil
}

// LoadFromEnv loads configuration from environment variables or a specified file.
func (c *Config) LoadFromEnv(options ...func(*Config)) error {
	for _, option := range options {
		option(c)
	}
	return nil
}

// WithEnvPrefix sets the prefix for environment variables.
func WithEnvPrefix(prefix string) func(*Config) {
	return func(c *Config) {
		c.mu.Lock()
		defer c.mu.Unlock()
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key, value := parts[0], parts[1]
			if strings.HasPrefix(key, prefix) {
				c.data[strings.TrimPrefix(key, prefix)] = value
			}
		}
	}
}

// WithEnvFile loads configuration from a specified file.
func WithEnvFile(filePath string) func(*Config) {
	return func(c *Config) {
		c.mu.Lock()
		defer c.mu.Unlock()
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("failed to open env file: %v\n", err)
			return
		}
		defer file.Close()

		var jsonData map[string]any
		if err := json.NewDecoder(file).Decode(&jsonData); err != nil {
			fmt.Printf("failed to decode env file: %v\n", err)
			return
		}

		c.FlattenMap(jsonData, "")
	}
}

// GetInt retrieves a value by key and converts it to an int.
func (c *Config) GetInt(key string) (int, error) {
	c.mu.RLock()
	value, exists := c.data[key]
	c.mu.RUnlock()
	if !exists {
		return 0, fmt.Errorf("key %s not found", key)
	}

	result, err := xcast.ToInt(value)
	if err != nil {
		return 0, fmt.Errorf("failed to convert value to int: %w", err)
	}

	return result, nil
}

// GetInt32 retrieves a value by key and converts it to an int32.
func (c *Config) GetInt32(key string) (int32, error) {
	c.mu.RLock()
	value, exists := c.data[key]
	c.mu.RUnlock()
	if !exists {
		return 0, fmt.Errorf("key %s not found", key)
	}

	result, err := xcast.ToInt32(value)
	if err != nil {
		return 0, fmt.Errorf("failed to convert value to int32: %w", err)
	}

	return result, nil
}

// GetString retrieves a value by key and converts it to a string.
func (c *Config) GetString(key string) (string, error) {
	c.mu.RLock()
	value, exists := c.data[key]
	c.mu.RUnlock()
	if !exists {
		return "", fmt.Errorf("key %s not found", key)
	}

	result, err := xcast.ToString(value)
	if err != nil {
		return "", fmt.Errorf("failed to convert value to string: %w", err)
	}

	return result, nil
}

// GetBool retrieves a value by key and converts it to a bool.
func (c *Config) GetBool(key string) (bool, error) {
	c.mu.RLock()
	value, exists := c.data[key]
	c.mu.RUnlock()
	if !exists {
		return false, fmt.Errorf("key %s not found", key)
	}

	result, err := xcast.ToBool(value)
	if err != nil {
		return false, fmt.Errorf("failed to convert value to bool: %w", err)
	}

	return result, nil
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
		return fmt.Errorf("input must be a struct")
	}

	t := v.Type()
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("config")
		if tag == "" {
			tag = field.Name
		}
		c.data[tag] = v.Field(i).Interface()
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
	return c.data
}

// FlattenMap flattens a nested map into a single-level map with dot notation keys.
func (c *Config) FlattenMap(data map[string]any, prefix string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		if subMap, ok := value.(map[string]any); ok {
			c.FlattenMap(subMap, fullKey)
		} else {
			c.data[fullKey] = value
		}
	}
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

// GetDefaultConfig returns the default configuration instance.
func GetDefaultConfig() *Config {
	return defaultConfig
}
