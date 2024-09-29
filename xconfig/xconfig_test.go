package xconfig_test

import (
	"os"
	"testing"

	"github.com/seefs001/xox/xconfig"
	"github.com/stretchr/testify/assert"
)

func TestLoadFromJSON(t *testing.T) {
	config := xconfig.NewConfig()
	xconfig.LoadFromEnv()
	jsonStr := `{"key1": "value1", "key2": 2, "nested": {"key3": "value3"}}`
	err := config.LoadFromJSON(jsonStr)
	assert.NoError(t, err)
	value1, err := config.GetString("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value1)
	value2, err := config.GetInt("key2")
	assert.NoError(t, err)
	assert.Equal(t, 2, value2)
	value3, err := config.GetString("nested.key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", value3)
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PREFIX_KEY1", "value1")
	os.Setenv("PREFIX_KEY2", "2")
	config := xconfig.NewConfig()
	err := config.LoadFromEnv(xconfig.WithEnvPrefix("PREFIX_"))
	assert.NoError(t, err)
	value1, err := config.GetString("KEY1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value1)
	value2, err := config.GetInt("KEY2")
	assert.NoError(t, err)
	assert.Equal(t, 2, value2)
}

func TestWithEnvFile(t *testing.T) {
	filePath := "test_env.json"
	fileContent := `{"key1": "value1", "key2": 2, "nested": {"key3": "value3"}}`
	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	assert.NoError(t, err)
	defer os.Remove(filePath)

	config := xconfig.NewConfig()
	err = config.LoadFromEnv(xconfig.WithEnvFile(filePath))
	assert.NoError(t, err)
	value1, err := config.GetString("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value1)
	value2, err := config.GetInt("key2")
	assert.NoError(t, err)
	assert.Equal(t, 2, value2)
	value3, err := config.GetString("nested.key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", value3)
}

func TestGetInt(t *testing.T) {
	config := xconfig.NewConfig()
	config.Put("key1", 1)
	value, err := config.GetInt("key1")
	assert.NoError(t, err)
	assert.Equal(t, 1, value)
}

func TestGetInt32(t *testing.T) {
	config := xconfig.NewConfig()
	config.Put("key1", int32(1))
	value, err := config.GetInt32("key1")
	assert.NoError(t, err)
	assert.Equal(t, int32(1), value)
}

func TestGetString(t *testing.T) {
	config := xconfig.NewConfig()
	config.Put("key1", "value1")
	value, err := config.GetString("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)
}

func TestGetBool(t *testing.T) {
	config := xconfig.NewConfig()
	config.Put("key1", true)
	value, err := config.GetBool("key1")
	assert.NoError(t, err)
	assert.Equal(t, true, value)
}

func TestPut(t *testing.T) {
	config := xconfig.NewConfig()
	config.Put("key1", "value1")
	value, err := config.GetString("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)
}

func TestLoadFromStruct(t *testing.T) {
	type ConfigStruct struct {
		Key1 string `config:"key1"`
		Key2 int    `config:"key2"`
	}
	config := xconfig.NewConfig()
	err := config.LoadFromStruct(ConfigStruct{Key1: "value1", Key2: 2})
	assert.NoError(t, err)
	value1, err := config.GetString("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value1)
	value2, err := config.GetInt("key2")
	assert.NoError(t, err)
	assert.Equal(t, 2, value2)
}

func TestGetAll(t *testing.T) {
	config := xconfig.NewConfig()
	config.Put("key1", "value1")
	config.Put("key2", 2)
	allData := config.GetAll()
	assert.Equal(t, "value1", allData["key1"])
	assert.Equal(t, 2, allData["key2"])
}

func TestFlattenMap(t *testing.T) {
	config := xconfig.NewConfig()
	nestedMap := map[string]any{
		"key1": "value1",
		"nested": map[string]any{
			"key2": "value2",
		},
	}
	config.FlattenMap(nestedMap, "")
	value1, err := config.GetString("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value1)
	value2, err := config.GetString("nested.key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", value2)
}

func TestDefaultConfig(t *testing.T) {
	xconfig.Put("key1", "value1")
	value1, err := xconfig.GetString("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value1)
	value2, err := xconfig.GetDefaultConfig().GetString("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value2)
}

func TestLoadFromJSONWithNestedStructures(t *testing.T) {
	config := xconfig.NewConfig()
	jsonStr := `{
		"database": {
			"host": "localhost",
			"port": 5432,
			"credentials": {
				"username": "admin",
				"password": "secret"
			}
		},
		"features": ["auth", "logging", "caching"],
		"limits": {
			"maxConnections": 100,
			"timeout": 30
		}
	}`
	err := config.LoadFromJSON(jsonStr)
	assert.NoError(t, err)

	// Test nested string
	dbHost, err := config.GetString("database.host")
	assert.NoError(t, err)
	assert.Equal(t, "localhost", dbHost)

	// Test nested int
	dbPort, err := config.GetInt("database.port")
	assert.NoError(t, err)
	assert.Equal(t, 5432, dbPort)

	// Test deeply nested string
	username, err := config.GetString("database.credentials.username")
	assert.NoError(t, err)
	assert.Equal(t, "admin", username)

	// Test non-existent key
	_, err = config.GetString("nonexistent.key")
	assert.Error(t, err)

	// Test invalid type conversion
	_, err = config.GetInt("database.host")
	assert.Error(t, err)
}

func TestLoadFromEnvWithComplexValues(t *testing.T) {
	os.Setenv("APP_NAME", "MyApp")
	os.Setenv("APP_VERSION", "1.2.3")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("APP_PORT", "8080")
	os.Setenv("APP_ADMINS", "admin1,admin2,admin3")
	defer func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("APP_VERSION")
		os.Unsetenv("APP_DEBUG")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_ADMINS")
	}()

	config := xconfig.NewConfig()
	err := config.LoadFromEnv(xconfig.WithEnvPrefix("APP_"))
	assert.NoError(t, err)

	// Test string value
	appName, err := config.GetString("NAME")
	assert.NoError(t, err)
	assert.Equal(t, "MyApp", appName)

	// Test bool value
	debug, err := config.GetBool("DEBUG")
	assert.NoError(t, err)
	assert.True(t, debug)

	// Test int value
	port, err := config.GetInt("PORT")
	assert.NoError(t, err)
	assert.Equal(t, 8080, port)

	// Test non-existent key
	_, err = config.GetString("NONEXISTENT")
	assert.Error(t, err)

	// Test invalid type conversion
	_, err = config.GetInt("NAME")
	assert.Error(t, err)
}

func TestLoadFromStructWithEmbeddedAndUnexportedFields(t *testing.T) {
	type BaseConfig struct {
		Environment string `config:"env"`
	}

	type AppConfig struct {
		BaseConfig
		Name    string `config:"name"`
		Version string `config:"version"`
		port    int    // Unexported field
		Debug   bool   `config:"debug"`
	}

	config := xconfig.NewConfig()
	err := config.LoadFromStruct(AppConfig{
		BaseConfig: BaseConfig{Environment: "production"},
		Name:       "MyApp",
		Version:    "1.0.0",
		port:       8080,
		Debug:      true,
	})
	assert.NoError(t, err)

	// Test embedded struct field
	env, err := config.GetString("env")
	assert.NoError(t, err)
	assert.Equal(t, "production", env)

	// Test normal struct field
	name, err := config.GetString("name")
	assert.NoError(t, err)
	assert.Equal(t, "MyApp", name)

	// Test unexported field (should not be included)
	_, err = config.GetInt("port")
	assert.Error(t, err)

	// Test bool field
	debug, err := config.GetBool("debug")
	assert.NoError(t, err)
	assert.True(t, debug)

	// Verify that only exported fields are loaded
	allData := config.GetAll()
	assert.Contains(t, allData, "env")
	assert.Contains(t, allData, "name")
	assert.Contains(t, allData, "version")
	assert.Contains(t, allData, "debug")
	assert.NotContains(t, allData, "port")
}

func TestFlattenMapWithComplexNestedStructures(t *testing.T) {
	config := xconfig.NewConfig()
	nestedMap := map[string]any{
		"database": map[string]any{
			"master": map[string]any{
				"host": "localhost",
				"port": 5432,
			},
			"slave": []any{
				map[string]any{"host": "slave1", "port": 5433},
				map[string]any{"host": "slave2", "port": 5434},
			},
		},
		"cache": map[string]any{
			"redis": map[string]any{
				"host": "localhost",
				"port": 6379,
			},
		},
	}
	config.FlattenMap(nestedMap, "")

	// Test deeply nested values
	masterHost, err := config.GetString("database.master.host")
	assert.NoError(t, err)
	assert.Equal(t, "localhost", masterHost)

	masterPort, err := config.GetInt("database.master.port")
	assert.NoError(t, err)
	assert.Equal(t, 5432, masterPort)

	// Test array values
	slave1Host, err := config.GetString("database.slave.0.host")
	assert.NoError(t, err)
	assert.Equal(t, "slave1", slave1Host)

	// Test cache values
	redisPort, err := config.GetInt("cache.redis.port")
	assert.NoError(t, err)
	assert.Equal(t, 6379, redisPort)
}

func TestParseToStruct(t *testing.T) {
	config := xconfig.NewConfig()
	config.Put("name", "MyApp")
	config.Put("version", "1.0.0")
	config.Put("debug", true)
	config.Put("port", 8080)

	type AppConfig struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Debug   bool   `json:"debug"`
		Port    int    `json:"port"`
	}

	var appConfig AppConfig
	err := config.ParseToStruct(&appConfig)
	assert.NoError(t, err)

	assert.Equal(t, "MyApp", appConfig.Name)
	assert.Equal(t, "1.0.0", appConfig.Version)
	assert.True(t, appConfig.Debug)
	assert.Equal(t, 8080, appConfig.Port)
}

func TestDefaultConfigMethods(t *testing.T) {
	// Test default config methods
	xconfig.Put("key1", "value1")
	xconfig.Put("key2", 42)
	xconfig.Put("key3", true)

	value1, err := xconfig.GetString("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value1)

	value2, err := xconfig.GetInt("key2")
	assert.NoError(t, err)
	assert.Equal(t, 42, value2)

	value3, err := xconfig.GetBool("key3")
	assert.NoError(t, err)
	assert.True(t, value3)

	// Test non-existent key
	_, err = xconfig.GetString("nonexistent")
	assert.Error(t, err)

	// Test invalid type conversion
	_, err = xconfig.GetInt("key1")
	assert.Error(t, err)
}
