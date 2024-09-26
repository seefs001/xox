package xd

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xlog"
)

func generateServiceName[T any]() string {
	var t T

	// struct
	name := fmt.Sprintf("%T", t)
	if name != "<nil>" {
		return name
	}

	// interface
	return fmt.Sprintf("%T", new(T))
}

type Container struct {
	mu       sync.RWMutex
	services map[string]any

	invokedServices map[string]bool

	logf func(format string, args ...any)
}

var defaultContainer = &Container{
	services:        make(map[string]any),
	invokedServices: make(map[string]bool),
	logf:            func(format string, args ...any) {}, // Default no-op logger
}

func NewContainer() *Container {
	return &Container{
		services:        make(map[string]any),
		invokedServices: make(map[string]bool),
		logf:            defaultContainer.logf,
	}
}

func SetLogger(logf func(format string, args ...any)) {
	defaultContainer.logf = logf
}

// Provide registers a service provider for type T in the container.
func Provide[T any](c *Container, provider func(c *Container) (T, error)) {
	ProvideNamed[T](c, "", provider)
}

// ProvideNamed registers a named service provider for type T in the container.
func ProvideNamed[T any](c *Container, name string, provider func(c *Container) (T, error)) {
	serviceName := generateServiceName[T]()
	if name != "" {
		serviceName = fmt.Sprintf("%s:%s", serviceName, name)
	}

	c.mu.Lock()
	if _, exists := c.services[serviceName]; exists {
		c.mu.Unlock()
		if c.logf != nil {
			c.logf("Service %s already provided", serviceName)
		}
		return
	}
	c.mu.Unlock()

	service, err := provider(c)
	if err != nil {
		if c.logf != nil {
			c.logf("Error providing service %s: %v", serviceName, err)
		}
		return
	}

	c.mu.Lock()
	c.services[serviceName] = service
	c.mu.Unlock()

	if c.logf != nil {
		c.logf("Service %s provided successfully", serviceName)
	}
}

// Invoke retrieves a service of type T from the container.
func Invoke[T any](c *Container) (T, error) {
	return InvokeNamed[T](c, "")
}

// InvokeNamed retrieves a named service of type T from the container.
func InvokeNamed[T any](c *Container, name string) (T, error) {
	serviceName := generateServiceName[T]()
	if name != "" {
		serviceName = fmt.Sprintf("%s:%s", serviceName, name)
	}

	c.mu.RLock()
	service, exists := c.services[serviceName]
	c.mu.RUnlock()

	if !exists {
		var zero T
		return zero, fmt.Errorf("service %s not found", serviceName)
	}

	c.mu.Lock()
	if !c.invokedServices[serviceName] {
		c.invokedServices[serviceName] = true
		c.mu.Unlock()
		if c.logf != nil {
			c.logf("Service %s invoked for the first time", serviceName)
		}
	} else {
		c.mu.Unlock()
	}

	// Check if the service type matches the requested type
	if reflect.TypeOf(service) != reflect.TypeOf((*T)(nil)).Elem() {
		if c.logf != nil {
			c.logf("Warning: Service %s type mismatch. Expected %T, got %T", serviceName, *new(T), service)
		}
	}

	return service.(T), nil
}

func MustInvoke[T any](c *Container) T {
	return x.Must1(Invoke[T](c))
}

func MustInvokeNamed[T any](c *Container, name string) T {
	return x.Must1(InvokeNamed[T](c, name))
}

func ListServices(c *Container) []ServiceInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	services := make([]ServiceInfo, 0, len(c.services))
	for serviceName, service := range c.services {
		info := ServiceInfo{
			Name:      serviceName,
			Type:      reflect.TypeOf(service).String(),
			IsInvoked: c.invokedServices[serviceName],
		}
		services = append(services, info)
	}

	if c.logf != nil {
		c.logf("Listed %d services with detailed information", len(services))
	}

	return services
}

type ServiceInfo struct {
	Name      string
	Type      string
	IsInvoked bool
}

// InjectStruct injects services into the fields of a struct.
func InjectStruct[T any](c *Container, s T) error {
	return InjectStructNamed(c, s, nil)
}

// InjectStructNamed injects named services into the fields of a struct.
func InjectStructNamed[T any](c *Container, s T, fieldNames map[string]string) error {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr {
		xlog.Warn("InjectStructNamed: Non-pointer value passed, creating a new pointer", "type", reflect.TypeOf(s))
		ptr := reflect.New(reflect.TypeOf(s))
		ptr.Elem().Set(v)
		v = ptr
	}
	return injectStructNamedRecursive(c, v, fieldNames)
}

func injectStructNamedRecursive(c *Container, v reflect.Value, fieldNames map[string]string) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		tag := fieldType.Tag.Get("xd")
		if tag != "-" {
			// If the field is a struct or pointer to struct, recursively inject
			if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
				if err := injectStructNamedRecursive(c, field, fieldNames); err != nil {
					return err
				}
			}
			continue
		}

		serviceName := fieldType.Type.String()
		if fieldNames != nil {
			if name, ok := fieldNames[fieldType.Name]; ok {
				serviceName = fmt.Sprintf("%s:%s", serviceName, name)
			}
		}

		c.mu.RLock()
		service, exists := c.services[serviceName]
		c.mu.RUnlock()

		if !exists {
			return fmt.Errorf("service %s not found for field %s", serviceName, fieldType.Name)
		}

		serviceValue := reflect.ValueOf(service)
		if field.Kind() != serviceValue.Kind() {
			if field.Kind() == reflect.Ptr && serviceValue.Kind() != reflect.Ptr {
				ptr := reflect.New(serviceValue.Type())
				ptr.Elem().Set(serviceValue)
				serviceValue = ptr
			} else if field.Kind() != reflect.Ptr && serviceValue.Kind() == reflect.Ptr {
				serviceValue = serviceValue.Elem()
			} else {
				return fmt.Errorf("incompatible types for field %s: expected %v, got %v",
					fieldType.Name, field.Type(), serviceValue.Type())
			}
		}

		if !field.CanSet() {
			return fmt.Errorf("cannot set field %s", fieldType.Name)
		}

		field.Set(serviceValue)

		// Recursively inject for struct fields
		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
			if err := injectStructNamedRecursive(c, field, fieldNames); err != nil {
				return err
			}
		}
	}

	return nil
}

// HasService checks if a service of the specified type is registered in the container.
func (c *Container) HasService(serviceType reflect.Type) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	serviceName := serviceType.String()
	_, exists := c.services[serviceName]
	return exists
}

// RemoveService removes a service of the specified type from the container.
func (c *Container) RemoveService(serviceType reflect.Type) {
	c.mu.Lock()
	defer c.mu.Unlock()

	serviceName := serviceType.String()
	delete(c.services, serviceName)
	delete(c.invokedServices, serviceName)
	if c.logf != nil {
		c.logf("Service %s removed", serviceName)
	}
}

// Clear removes all services from the container.
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[string]any)
	c.invokedServices = make(map[string]bool)
	if c.logf != nil {
		c.logf("All services cleared from container")
	}
}

// Clone creates a new Container with the same services as the current one.
func (c *Container) Clone() *Container {
	c.mu.RLock()
	defer c.mu.RUnlock()

	newContainer := NewContainer()
	for serviceName, service := range c.services {
		newContainer.services[serviceName] = service
	}
	for serviceName, invoked := range c.invokedServices {
		newContainer.invokedServices[serviceName] = invoked
	}
	newContainer.logf = c.logf

	if c.logf != nil {
		c.logf("Container cloned with %d services", len(c.services))
	}

	return newContainer
}

// SetService directly sets a service in the container without using a provider function.
func (c *Container) SetService(service any) {
	c.SetNamedService("", service)
}

// SetNamedService directly sets a named service in the container without using a provider function.
func (c *Container) SetNamedService(name string, service any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	serviceName := reflect.TypeOf(service).String()
	if name != "" {
		serviceName = fmt.Sprintf("%s:%s", serviceName, name)
	}
	c.services[serviceName] = service
	if c.logf != nil {
		c.logf("Service %s set directly", serviceName)
	}
}
