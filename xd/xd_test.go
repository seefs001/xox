package xd_test

import (
	"fmt"
	"reflect"
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

// TestProvideAndInvokeNamed tests the ProvideNamed and InvokeNamed functions
func TestProvideAndInvokeNamed(t *testing.T) {
	c := xd.NewContainer()

	type Config struct {
		Environment string
	}

	// Provide named Config services
	xd.ProvideNamed(c, "dev", func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "development"}, nil
	})
	xd.ProvideNamed(c, "prod", func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "production"}, nil
	})

	// Invoke named Config services
	devConfig := xd.MustInvokeNamed[*Config](c, "dev")
	prodConfig := xd.MustInvokeNamed[*Config](c, "prod")

	assert.Equal(t, "development", devConfig.Environment, "Unexpected dev environment")
	assert.Equal(t, "production", prodConfig.Environment, "Unexpected prod environment")

	// Test invoking non-existent named service
	_, err := xd.InvokeNamed[*Config](c, "staging")
	assert.Error(t, err, "Expected error when invoking non-existent named service")
}

// TestInjectStructNamed tests the InjectStructNamed function
func TestInjectStructNamed(t *testing.T) {
	c := xd.NewContainer()

	type Database struct {
		ConnectionString string
	}

	type Service struct {
		DevDB  *Database `xd:"-"`
		ProdDB *Database `xd:"-"`
	}

	xd.ProvideNamed(c, "dev", func(c *xd.Container) (*Database, error) {
		return &Database{ConnectionString: "dev-db-connection"}, nil
	})
	xd.ProvideNamed(c, "prod", func(c *xd.Container) (*Database, error) {
		return &Database{ConnectionString: "prod-db-connection"}, nil
	})

	s := &Service{}
	err := xd.InjectStructNamed(c, s, map[string]string{
		"DevDB":  "dev",
		"ProdDB": "prod",
	})

	assert.NoError(t, err, "InjectStructNamed should not return an error")
	assert.Equal(t, "dev-db-connection", s.DevDB.ConnectionString, "Unexpected dev database connection string")
	assert.Equal(t, "prod-db-connection", s.ProdDB.ConnectionString, "Unexpected prod database connection string")
}

// TestSetNamedService tests the SetNamedService function
func TestSetNamedService(t *testing.T) {
	c := xd.NewContainer()

	type Config struct {
		Environment string
	}

	c.SetNamedService("dev", &Config{Environment: "development"})
	c.SetNamedService("prod", &Config{Environment: "production"})

	devConfig := xd.MustInvokeNamed[*Config](c, "dev")
	prodConfig := xd.MustInvokeNamed[*Config](c, "prod")

	assert.Equal(t, "development", devConfig.Environment, "Unexpected dev environment")
	assert.Equal(t, "production", prodConfig.Environment, "Unexpected prod environment")
}

// TestListServicesWithNamed tests the ListServices function with named services
func TestListServicesWithNamed(t *testing.T) {
	c := xd.NewContainer()

	type Config struct {
		Environment string
	}

	xd.ProvideNamed(c, "dev", func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "development"}, nil
	})
	xd.ProvideNamed(c, "prod", func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "production"}, nil
	})

	// Invoke one of the named services
	xd.MustInvokeNamed[*Config](c, "dev")

	services := xd.ListServices(c)

	assert.Equal(t, 2, len(services), "Expected 2 services")

	for _, service := range services {
		assert.True(t, service.Name == "*xd_test.Config:dev" || service.Name == "*xd_test.Config:prod", "Unexpected service name")
		if service.Name == "*xd_test.Config:dev" {
			assert.True(t, service.IsInvoked, "Expected dev service to be invoked")
		} else {
			assert.False(t, service.IsInvoked, "Expected prod service to not be invoked")
		}
	}
}

// TestRemoveNamedService tests removing a named service
func TestRemoveNamedService(t *testing.T) {
	c := xd.NewContainer()

	type Config struct {
		Environment string
	}

	xd.ProvideNamed(c, "dev", func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "development"}, nil
	})

	// Check if the service exists
	devConfig, err := xd.InvokeNamed[*Config](c, "dev")
	assert.NoError(t, err, "Expected service to exist")
	assert.NotNil(t, devConfig, "Expected service to exist")

	// Remove the named service
	c.RemoveNamedService(reflect.TypeOf((*Config)(nil)).Elem(), "dev")

	// Try to invoke the removed service
	_, err = xd.InvokeNamed[*Config](c, "dev")
	assert.Error(t, err, "Expected error when invoking removed service")
}

// TestCloneContainer tests the Clone method of the Container
func TestCloneContainer(t *testing.T) {
	c := xd.NewContainer()

	type Config struct {
		Environment string
	}

	xd.Provide(c, func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "production"}, nil
	})

	// Invoke the service to mark it as invoked
	xd.MustInvoke[*Config](c)

	// Clone the container
	clonedContainer := c.Clone()

	// Check if the cloned container has the same service
	clonedConfig := xd.MustInvoke[*Config](clonedContainer)
	assert.Equal(t, "production", clonedConfig.Environment, "Cloned container should have the same service")

	// Check if the service is marked as invoked in the cloned container
	services := xd.ListServices(clonedContainer)
	assert.Equal(t, 1, len(services), "Cloned container should have 1 service")
	assert.True(t, services[0].IsInvoked, "Service should be marked as invoked in cloned container")

	// Modify the original container
	xd.Provide(c, func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "development"}, nil
	})

	// Check that the cloned container is not affected
	clonedConfigAfterModification := xd.MustInvoke[*Config](clonedContainer)
	assert.Equal(t, "production", clonedConfigAfterModification.Environment, "Cloned container should not be affected by changes to original")
}

// TestClearContainer tests the Clear method of the Container
func TestClearContainer(t *testing.T) {
	c := xd.NewContainer()

	type Config struct {
		Environment string
	}

	xd.Provide(c, func(c *xd.Container) (*Config, error) {
		return &Config{Environment: "production"}, nil
	})

	// Invoke the service
	xd.MustInvoke[*Config](c)

	// Clear the container
	c.Clear()

	// Check if the container is empty
	services := xd.ListServices(c)
	assert.Equal(t, 0, len(services), "Container should be empty after Clear")

	// Try to invoke the service again
	_, err := xd.Invoke[*Config](c)
	assert.Error(t, err, "Invoking cleared service should return an error")
}

// TestSetServiceAndSetNamedService tests the SetService and SetNamedService methods
func TestSetServiceAndSetNamedService(t *testing.T) {
	c := xd.NewContainer()

	type Config struct {
		Environment string
	}

	// Test SetService
	c.SetService(&Config{Environment: "production"})
	config := xd.MustInvoke[*Config](c)
	assert.Equal(t, "production", config.Environment, "SetService should set the service correctly")

	// Test SetNamedService
	c.SetNamedService("dev", &Config{Environment: "development"})
	devConfig := xd.MustInvokeNamed[*Config](c, "dev")
	assert.Equal(t, "development", devConfig.Environment, "SetNamedService should set the named service correctly")

	// Check that both services exist
	services := xd.ListServices(c)
	assert.Equal(t, 2, len(services), "Container should have 2 services")
}

// TestErrorHandlingAndEdgeCases tests various error handling scenarios and edge cases
func TestErrorHandlingAndEdgeCases(t *testing.T) {
	c := xd.NewContainer()

	// Test invoking non-existent service
	_, err := xd.Invoke[*struct{}](c)
	assert.Error(t, err, "Invoking non-existent service should return an error")

	// Test providing service with error
	xd.Provide(c, func(c *xd.Container) (*struct{}, error) {
		return nil, fmt.Errorf("service creation failed")
	})
	_, err = xd.Invoke[*struct{}](c)
	assert.Error(t, err, "Invoking service that failed to provide should return an error")

	// Test removing non-existent service
	c.RemoveService(reflect.TypeOf((*struct{})(nil)))
	assert.False(t, c.HasService(reflect.TypeOf((*struct{})(nil))), "Removing non-existent service should not cause issues")

	// Test injecting into non-struct type
	err = xd.InjectStruct(c, "not a struct")
	assert.Error(t, err, "Injecting into non-struct type should return an error")

	// Test injecting into struct with unexported fields
	type privateStruct struct {
		config *struct{} `xd:"-"`
		Public *struct{} `xd:"-"`
	}
	err = xd.InjectStruct(c, &privateStruct{})
	assert.NoError(t, err, "Injecting into struct with unexported fields should not return an error")

	// Provide a service for the public field
	xd.Provide(c, func(c *xd.Container) (*struct{}, error) {
		return &struct{}{}, nil
	})

	// Try injecting again
	ps := &privateStruct{}
	err = xd.InjectStruct(c, ps)
	assert.NoError(t, err, "Injecting into struct with unexported fields should not return an error")
	assert.NotNil(t, ps.Public, "Public field should be injected")
}
