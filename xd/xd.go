package xd

import (
	"fmt"
	"reflect"
	"sync"

	"strings"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xerror"
)

// ErrCircularDependency is returned when a circular dependency is detected
var ErrCircularDependency = xerror.New("circular dependency detected")

// Improved service name generation with type caching
var (
	serviceNameCache     = make(map[reflect.Type]string)
	serviceNameCacheLock sync.RWMutex
)

func generateServiceName[T any]() string {
	// Get type information
	t := reflect.TypeOf((*T)(nil)).Elem()

	// Check cache first
	serviceNameCacheLock.RLock()
	if name, ok := serviceNameCache[t]; ok {
		serviceNameCacheLock.RUnlock()
		return name
	}
	serviceNameCacheLock.RUnlock()

	// Generate name and cache it
	name := t.String()
	serviceNameCacheLock.Lock()
	serviceNameCache[t] = name
	serviceNameCacheLock.Unlock()

	return name
}

// ServiceProvider represents a function that can create a service
type ServiceProvider[T any] func(c *Container) (T, error)

// LazyServiceProvider wraps a provider to create services on-demand
type LazyServiceProvider[T any] struct {
	provider  ServiceProvider[T]
	instance  *T
	container *Container
	mu        sync.Mutex
	created   bool
	creating  bool // Used for circular dependency detection
}

// Get returns the service instance, creating it if needed
func (l *LazyServiceProvider[T]) Get() (T, error) {
	var zero T

	l.mu.Lock()
	// Check for circular dependencies
	if l.creating {
		l.mu.Unlock()
		return zero, ErrCircularDependency
	}

	if l.created {
		result := *l.instance
		l.mu.Unlock()
		return result, nil
	}

	if l.provider == nil {
		l.mu.Unlock()
		return zero, xerror.New("nil provider function")
	}

	// Mark as creating to detect circular dependencies
	l.creating = true
	l.mu.Unlock()

	// Create the instance
	instance, err := l.provider(l.container)

	l.mu.Lock()
	// Reset creating flag regardless of outcome
	l.creating = false

	if err != nil {
		l.mu.Unlock()
		return zero, err
	}

	l.instance = &instance
	l.created = true
	l.mu.Unlock()

	return instance, nil
}

type Container struct {
	mu              sync.RWMutex
	services        map[string]any
	invokedServices map[string]bool
	logf            func(format string, args ...any)
	constructing    map[string]bool // Track services being constructed for cycle detection
}

var defaultContainer = &Container{
	services:        make(map[string]any),
	invokedServices: make(map[string]bool),
	logf:            func(format string, args ...any) {},
	constructing:    make(map[string]bool),
}

func NewContainer() *Container {
	return &Container{
		services:        make(map[string]any),
		invokedServices: make(map[string]bool),
		logf:            func(format string, args ...any) {},
		constructing:    make(map[string]bool),
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

	// Detect circular dependencies during service construction
	if c.constructing[serviceName] {
		c.mu.Unlock()
		if c.logf != nil {
			c.logf("Circular dependency detected for service %s", serviceName)
		}
		return
	}

	c.constructing[serviceName] = true
	c.mu.Unlock()

	if provider == nil {
		c.mu.Lock()
		delete(c.constructing, serviceName)
		c.mu.Unlock()
		if c.logf != nil {
			c.logf("Provider function is nil for service %s", serviceName)
		}
		return
	}

	service, err := provider(c)

	c.mu.Lock()
	delete(c.constructing, serviceName)

	if err != nil {
		c.mu.Unlock()
		if c.logf != nil {
			c.logf("Error providing service %s: %v", serviceName, err)
		}
		return
	}

	c.services[serviceName] = service
	c.mu.Unlock()

	if c.logf != nil {
		c.logf("Service %s provided successfully", serviceName)
	}
}

// ProvideLazy registers a lazy service provider that creates the service only when first requested
func ProvideLazy[T any](c *Container, provider func(c *Container) (T, error)) {
	ProvideLazyNamed[T](c, "", provider)
}

// ProvideLazyNamed registers a named lazy service provider
func ProvideLazyNamed[T any](c *Container, name string, provider func(c *Container) (T, error)) {
	if c == nil || provider == nil {
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

	lazyProvider := &LazyServiceProvider[T]{
		provider:  provider,
		container: c,
	}

	c.services[serviceName] = lazyProvider
	c.mu.Unlock()

	if c.logf != nil {
		c.logf("Lazy service %s provided successfully", serviceName)
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

	// Check for lazy provider
	if lazyProvider, ok := service.(*LazyServiceProvider[T]); ok {
		instance, err := lazyProvider.Get()
		if err != nil {
			if err == ErrCircularDependency {
				return zero, xerror.Errorf("circular dependency detected when resolving service %s", serviceName)
			}
			return zero, xerror.Wrap(err, fmt.Sprintf("failed to initialize lazy service %s", serviceName))
		}

		// Update invoked status
		c.mu.Lock()
		if !c.invokedServices[serviceName] {
			c.invokedServices[serviceName] = true
			c.mu.Unlock()
			if c.logf != nil {
				c.logf("Lazy service %s initialized and invoked for the first time", serviceName)
			}
		} else {
			c.mu.Unlock()
		}

		return instance, nil
	}

	// Update invoked status for non-lazy services
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
	if c == nil {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	services := make([]ServiceInfo, 0, len(c.services))
	for serviceName, service := range c.services {
		info := ServiceInfo{
			Name:      serviceName,
			Type:      reflect.TypeOf(service).String(),
			IsInvoked: c.invokedServices[serviceName],
			IsLazy:    strings.Contains(reflect.TypeOf(service).String(), "LazyServiceProvider"),
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
	IsLazy    bool
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
		if v.IsNil() {
			// Skip nil pointers
			return nil
		}
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
		// Check if this field should be injected
		if tag != "-" {
			// Skip this field if it doesn't have the injection tag
			continue
		}

		// Field with tag "-" should be injected based on type
		// Use type name as service name
		serviceName := fieldType.Type.String()
		// Apply custom name from fieldNames if available
		if fieldNames != nil {
			if name, ok := fieldNames[fieldType.Name]; ok {
				serviceName = fmt.Sprintf("%s:%s", serviceName, name)
			}
		}

		// Try to get the service
		c.mu.RLock()
		service, exists := c.services[serviceName]

		// If exact match doesn't exist, try the base type without named qualifier
		if !exists && strings.Contains(serviceName, ":") {
			baseServiceName := strings.Split(serviceName, ":")[0]
			service, exists = c.services[baseServiceName]
		}
		c.mu.RUnlock()

		if exists {
			// Handle lazy providers
			lazyProvider := reflect.ValueOf(service)
			if lazyProvider.Type().String() == fmt.Sprintf("*xd.LazyServiceProvider[%s]", fieldType.Type.String()) {
				// Call Get() method to retrieve the actual service
				getLazyInstance := lazyProvider.MethodByName("Get")
				if getLazyInstance.IsValid() {
					results := getLazyInstance.Call(nil)
					if len(results) == 2 && results[1].IsNil() {
						service = results[0].Interface()

						// Mark as invoked
						c.mu.Lock()
						if !c.invokedServices[serviceName] {
							c.invokedServices[serviceName] = true
							c.mu.Unlock()
							if c.logf != nil {
								c.logf("Lazy service %s initialized and injected into field %s",
									serviceName, fieldType.Name)
							}
						} else {
							c.mu.Unlock()
						}
					} else if len(results) == 2 && !results[1].IsNil() {
						// Error initializing lazy service
						continue
					}
				}
			}

			serviceValue := reflect.ValueOf(service)

			// Handle type compatibility
			if !serviceValue.Type().AssignableTo(field.Type()) {
				if field.Kind() == reflect.Ptr && serviceValue.Kind() != reflect.Ptr {
					// Field is pointer but service is value, create a pointer to the service
					ptr := reflect.New(serviceValue.Type())
					ptr.Elem().Set(serviceValue)
					serviceValue = ptr
				} else if field.Kind() != reflect.Ptr && serviceValue.Kind() == reflect.Ptr {
					// Field is value but service is pointer, dereference the pointer
					if serviceValue.IsNil() {
						continue // Skip nil pointers
					}
					serviceValue = serviceValue.Elem()
				} else {
					return xerror.Errorf("incompatible types for field %s: expected %v, got %v",
						fieldType.Name, field.Type(), serviceValue.Type())
				}
			}

			// Only set if types are compatible
			if serviceValue.Type().AssignableTo(field.Type()) {
				field.Set(serviceValue)

				// Mark service as invoked for non-lazy providers
				if lazyProvider.Type().String() != fmt.Sprintf("*xd.LazyServiceProvider[%s]", fieldType.Type.String()) {
					c.mu.Lock()
					if !c.invokedServices[serviceName] {
						c.invokedServices[serviceName] = true
						c.mu.Unlock()
						if c.logf != nil {
							c.logf("Service %s injected into field %s for the first time",
								serviceName, fieldType.Name)
						}
					} else {
						c.mu.Unlock()
					}
				}
			}
		}

		// Recursively inject nested structs
		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct) {
			if err := injectStructNamedRecursive(c, field, fieldNames); err != nil {
				return xerror.Wrap(err, "failed to inject nested struct")
			}
		}
	}

	return nil
}

// RemoveService removes a service of the specified type from the container.
func (c *Container) RemoveService(serviceType reflect.Type) {
	if c == nil {
		return
	}

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
	if c == nil {
		return false
	}

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

// HasServiceNamed checks if a named service is registered in the container
func (c *Container) HasServiceNamed(serviceType reflect.Type, name string) bool {
	if c == nil {
		return false
	}

	serviceName := serviceType.String()
	if name != "" {
		serviceName = fmt.Sprintf("%s:%s", serviceName, name)
	}

	c.mu.RLock()
	_, exists := c.services[serviceName]
	c.mu.RUnlock()

	return exists
}

// Clear removes all services from the container.
func (c *Container) Clear() {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[string]any)
	c.invokedServices = make(map[string]bool)
	c.constructing = make(map[string]bool)

	if c.logf != nil {
		c.logf("All services cleared from container")
	}
}

// Clone creates a new Container with the same services as the current one.
func (c *Container) Clone() *Container {
	if c == nil {
		return NewContainer()
	}

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
	if c == nil {
		return
	}

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

// ValidateAll attempts to resolve all registered services to check for errors
func (c *Container) ValidateAll() []error {
	if c == nil {
		return []error{xerror.New("container is nil")}
	}

	var errors []error

	c.mu.RLock()
	serviceNames := make([]string, 0, len(c.services))
	for name := range c.services {
		serviceNames = append(serviceNames, name)
	}
	c.mu.RUnlock()

	for _, name := range serviceNames {
		if err := c.validateService(name); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateService attempts to resolve a single service by name to check for errors
func (c *Container) validateService(serviceName string) error {
	c.mu.RLock()
	service, exists := c.services[serviceName]
	c.mu.RUnlock()

	if !exists {
		return xerror.Errorf("service %s not found", serviceName)
	}

	if service == nil {
		return xerror.Errorf("service %s is nil", serviceName)
	}

	serviceType := reflect.TypeOf(service)

	// Check if it's a lazy service
	if strings.Contains(serviceType.String(), "LazyServiceProvider") {
		getLazyInstance := reflect.ValueOf(service).MethodByName("Get")
		if getLazyInstance.IsValid() {
			results := getLazyInstance.Call(nil)
			if len(results) == 2 && !results[1].IsNil() {
				// Extract error
				err := results[1].Interface().(error)
				return xerror.Wrap(err, fmt.Sprintf("validation failed for lazy service %s", serviceName))
			}
		}
	}

	return nil
}

// GetServiceNames returns all service names registered in the container
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

// SetServiceFactory adds a factory method that will be called each time the service is requested
func SetServiceFactory[T any](c *Container, factory func() T) {
	SetNamedServiceFactory[T](c, "", factory)
}

// SetNamedServiceFactory adds a named factory method that will be called each time the service is requested
func SetNamedServiceFactory[T any](c *Container, name string, factory func() T) {
	if c == nil || factory == nil {
		return
	}

	wrapper := func(c *Container) (T, error) {
		return factory(), nil
	}

	ProvideLazyNamed[T](c, name, wrapper)
}

// Get is a convenience method for getting a service of type T
func Get[T any](c *Container) (T, error) {
	return Invoke[T](c)
}

// GetNamed is a convenience method for getting a named service of type T
func GetNamed[T any](c *Container, name string) (T, error) {
	return InvokeNamed[T](c, name)
}

// MustGet is a convenience method that panics if the service cannot be retrieved
func MustGet[T any](c *Container) T {
	return MustInvoke[T](c)
}

// MustGetNamed is a convenience method that panics if the named service cannot be retrieved
func MustGetNamed[T any](c *Container, name string) T {
	return MustInvokeNamed[T](c, name)
}
