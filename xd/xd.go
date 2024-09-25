package xd

import (
	"fmt"
	"reflect"
	"sync"
)

// Container is the main structure for the dependency injection container
type Container struct {
	mu       sync.RWMutex
	services map[reflect.Type]any
}

// New creates a new dependency injection container
func New() *Container {
	return &Container{
		services: make(map[reflect.Type]any),
	}
}

// Provide registers a service in the container
func (c *Container) Provide(service any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[reflect.TypeOf(service)] = service
}

// Resolve retrieves a service from the container
func (c *Container) Resolve(t reflect.Type) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if service, ok := c.services[t]; ok {
		return service, nil
	}

	return nil, fmt.Errorf("service of type %v not found", t)
}

// MustResolve retrieves a service from the container, panicking if not found
func (c *Container) MustResolve(t reflect.Type) any {
	result, err := c.Resolve(t)
	if err != nil {
		panic(err)
	}
	return result
}

// ProvideFactory registers a factory function to create a service
func (c *Container) ProvideFactory(t reflect.Type, factory func() any) {
	c.Provide(factory)
}

// Reset clears all services from the container
func (c *Container) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services = make(map[reflect.Type]any)
}
