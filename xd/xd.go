package xd

import (
	"fmt"
	"reflect"
	"sync"

	"strings"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xerror"
)

func generateServiceName[T any]() string {
	var t T
	name := fmt.Sprintf("%T", t)
	if name != "<nil>" {
		return name
	}
	return fmt.Sprintf("%T", new(T))
}

type Container struct {
	mu              sync.RWMutex
	services        map[string]any
	invokedServices map[string]bool
	logf            func(format string, args ...any)
}

var defaultContainer = &Container{
	services:        make(map[string]any),
	invokedServices: make(map[string]bool),
	logf:            func(format string, args ...any) {},
}

func NewContainer() *Container {
	return &Container{
		services:        make(map[string]any),
		invokedServices: make(map[string]bool),
		logf:            func(format string, args ...any) {},
	}
}

func GetDefaultContainer() *Container {
	return defaultContainer
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
	if c == nil {
		return
	}
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

	if provider == nil {
		if c.logf != nil {
			c.logf("Provider function is nil for service %s", serviceName)
		}
		return
	}

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
	var zero T
	if c == nil {
		return zero, xerror.New("container is nil")
	}

	serviceName := generateServiceName[T]()
	if name != "" {
		serviceName = fmt.Sprintf("%s:%s", serviceName, name)
	}

	c.mu.RLock()
	service, exists := c.services[serviceName]
	c.mu.RUnlock()

	if !exists {
		return zero, xerror.Errorf("service %s not found", serviceName)
	}

	if service == nil {
		return zero, xerror.Errorf("service %s is nil", serviceName)
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

	serviceValue := reflect.ValueOf(service)
	expectedType := reflect.TypeOf((*T)(nil)).Elem()
	if !serviceValue.Type().AssignableTo(expectedType) {
		return zero, xerror.Errorf("service %s type mismatch: expected %v, got %v",
			serviceName, expectedType, serviceValue.Type())
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
		return xerror.Errorf("InjectStructNamed: Non-pointer value passed")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return xerror.Errorf("InjectStructNamed: Value is not a struct")
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

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("xd")
		if tag == "-" {
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
				// Skip if service not found, don't return error
				continue
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
					return xerror.Errorf("incompatible types for field %s: expected %v, got %v",
						fieldType.Name, field.Type(), serviceValue.Type())
				}
			}

			field.Set(serviceValue)
		}

		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
			if err := injectStructNamedRecursive(c, field, fieldNames); err != nil {
				return xerror.Wrap(err, "failed to inject nested struct")
			}
		}
	}

	return nil
}

// RemoveService removes a service of the specified type from the container.
func (c *Container) RemoveService(serviceType reflect.Type) {
	c.mu.Lock()
	defer c.mu.Unlock()

	serviceName := serviceType.String()
	for key := range c.services {
		if strings.HasPrefix(key, serviceName) {
			delete(c.services, key)
			delete(c.invokedServices, key)
			if c.logf != nil {
				c.logf("Service %s removed", key)
			}
		}
	}
}

// HasService checks if a service of the specified type is registered in the container.
func (c *Container) HasService(serviceType reflect.Type) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	serviceName := serviceType.String()
	for key := range c.services {
		if key == serviceName || strings.HasPrefix(key, serviceName+":") {
			return true
		}
	}
	return false
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
	newContainer.services = x.CopyMap(c.services)
	newContainer.invokedServices = x.CopyMap(c.invokedServices)
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
	if c == nil || service == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	serviceType := reflect.TypeOf(service)
	serviceName := serviceType.String()
	if name != "" {
		serviceName = fmt.Sprintf("%s:%s", serviceName, name)
	}
	c.services[serviceName] = service
	if c.logf != nil {
		c.logf("Service %s of type %v set directly", serviceName, serviceType)
	}
}

// RemoveNamedService removes a named service of the specified type from the container.
func (c *Container) RemoveNamedService(serviceType reflect.Type, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	serviceName := serviceType.String()
	if name != "" {
		serviceName = fmt.Sprintf("%s:%s", serviceName, name)
	}

	delete(c.services, serviceName)
	delete(c.invokedServices, serviceName)
	if c.logf != nil {
		c.logf("Named service %s removed", serviceName)
	}
}

// New utility methods
func (c *Container) GetServiceNames() []string {
	if c == nil {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.services))
	for name := range c.services {
		names = append(names, name)
	}
	return names
}

func (c *Container) ServiceCount() int {
	if c == nil {
		return 0
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.services)
}

func (c *Container) IsServiceInvoked(serviceName string) bool {
	if c == nil {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.invokedServices[serviceName]
}
