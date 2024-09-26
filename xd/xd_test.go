package xd_test

import (
	"testing"

	"github.com/seefs001/xox/xd"
	"github.com/stretchr/testify/assert"
)

// TestInjectAndExtractStruct tests injecting a struct and then extracting it
func TestInjectAndExtractStruct(t *testing.T) {
	c := xd.NewContainer()

	type Config struct {
		Environment string
	}

	type Service struct {
		Config *Config `xd:"-"`
		Name   string
	}

	// Provide the Config service
	xd.Provide(c, func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "production"}, nil
	})

	// Provide the Service
	xd.Provide(c, func(c *xd.Container) (*Service, error) {
		service := &Service{Name: "TestService"}
		err := xd.InjectStruct(c, service)
		return service, err
	})

	// Invoke the Service
	service := xd.MustInvoke[*Service](c)

	// Verify the injection
	assert.NotNil(t, service.Config, "Expected Config to be injected")
	assert.Equal(t, "production", service.Config.Environment, "Unexpected environment")
	assert.Equal(t, "TestService", service.Name, "Name should not be modified by injection")

	// Extract the Service struct again
	extractedService := xd.MustInvoke[*Service](c)

	// Verify the extracted struct
	assert.NotNil(t, extractedService, "Extracted Service should not be nil")
	assert.Equal(t, service, extractedService, "Extracted Service should be the same as the injected one")
	assert.Equal(t, "production", extractedService.Config.Environment, "Unexpected environment in extracted Service")
	assert.Equal(t, "TestService", extractedService.Name, "Unexpected name in extracted Service")
}

// TestProvideAndInvoke tests the basic functionality of Provide and Invoke
func TestProvideAndInvoke(t *testing.T) {
	c := xd.NewContainer()

	type Database struct {
		ConnectionString string
	}

	// Provide a Database service
	xd.Provide(c, func(c *xd.Container) (*Database, error) {
		return &Database{ConnectionString: "postgres://localhost:5432/mydb"}, nil
	})

	// Invoke the Database service
	result := xd.MustInvoke[*Database](c)
	assert.Equal(t, "postgres://localhost:5432/mydb", result.ConnectionString, "Unexpected connection string")
}

// TestProvideMultipleServices tests providing and invoking multiple services
func TestProvideMultipleServices(t *testing.T) {
	c := xd.NewContainer()

	type Logger struct {
		Level string
	}

	type Config struct {
		Environment string
	}

	xd.Provide(c, func(c *xd.Container) (*Logger, error) {
		return &Logger{Level: "info"}, nil
	})

	xd.Provide(c, func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "production"}, nil
	})

	loggerResult := xd.MustInvoke[*Logger](c)
	assert.Equal(t, "info", loggerResult.Level, "Unexpected logger level")

	configResult := xd.MustInvoke[*Config](c)
	assert.Equal(t, "production", configResult.Environment, "Unexpected config environment")
}

// TestInjectStruct tests the InjectStruct function with realistic structures
func TestInjectStruct(t *testing.T) {
	c := xd.NewContainer()

	type Database struct {
		ConnectionString string
	}

	type Logger struct {
		Level string
	}

	type Service struct {
		DB     *Database `xd:"-"`
		Log    *Logger   `xd:"-"`
		APIKey string
	}

	xd.Provide(c, func(c *xd.Container) (*Database, error) {
		return &Database{ConnectionString: "postgres://localhost:5432/mydb"}, nil
	})

	xd.Provide(c, func(c *xd.Container) (*Logger, error) {
		return &Logger{Level: "debug"}, nil
	})

	s := &Service{APIKey: "secret-key"}
	err := xd.InjectStruct(c, s)
	assert.NoError(t, err, "InjectStruct should not return an error")

	assert.NotNil(t, s.DB, "Expected Database to be injected")
	assert.Equal(t, "postgres://localhost:5432/mydb", s.DB.ConnectionString, "Unexpected database connection string")

	assert.NotNil(t, s.Log, "Expected Logger to be injected")
	assert.Equal(t, "debug", s.Log.Level, "Unexpected logger level")

	assert.Equal(t, "secret-key", s.APIKey, "APIKey should not be modified by injection")
}

// TestInjectStructWithNestedDependencies tests injecting structs with nested dependencies
func TestInjectStructWithNestedDependencies(t *testing.T) {
	c := xd.NewContainer()

	type Config struct {
		Environment string
	}

	type Database struct {
		ConnectionString string
		Config           *Config `xd:"-"`
	}

	type Service struct {
		DB  *Database `xd:"-"`
		Env string
	}

	xd.Provide(c, func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "staging"}, nil
	})

	xd.Provide(c, func(c *xd.Container) (*Database, error) {
		db := &Database{ConnectionString: "postgres://localhost:5432/stagingdb"}
		err := xd.InjectStruct(c, db)
		return db, err
	})

	s := &Service{Env: "local"}
	err := xd.InjectStruct(c, s)
	assert.NoError(t, err, "InjectStruct should not return an error")

	assert.NotNil(t, s.DB, "Expected Database to be injected")
	assert.Equal(t, "postgres://localhost:5432/stagingdb", s.DB.ConnectionString, "Unexpected database connection string")
	assert.NotNil(t, s.DB.Config, "Expected Config to be injected into Database")
	assert.Equal(t, "staging", s.DB.Config.Environment, "Unexpected environment in nested Config")
	assert.Equal(t, "local", s.Env, "Env should not be modified by injection")
}
