package xconfig_test

import (
	"os"
	"testing"

	"github.com/seefs001/xox/xconfig"
	"github.com/stretchr/testify/assert"
)

func TestLoadFromJSON(t *testing.T) {
	config := xconfig.NewConfig()
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
