package app

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/ishubhamsingh2e/bourbon/core"
)

var (
	// ErrTargetNotPointer is returned when the target passed to Resolve is not a non-nil pointer.
	ErrTargetNotPointer = errors.New("container: resolve target must be a non-nil pointer")
	// ErrBindingNotFound is returned when no matching binding exists for the requested type.
	ErrBindingNotFound = errors.New("container: no binding found for type")
	// ErrCircularDependency is returned when a dependency cycle is detected during resolution.
	ErrCircularDependency = errors.New("container: circular dependency detected")
	// ErrInvalidConstructor is returned when the constructor signature is invalid.
	ErrInvalidConstructor = errors.New("container: constructor must be a function")
)

type bindingKind int

const (
	kindInstance bindingKind = iota
	kindSingletonFactory
	kindTransientFactory
)

type binding struct {
	kind        bindingKind
	instance    any
	constructor reflect.Value
}

// Container provides a thread-safe dependency injection registry.
type Container struct {
	mu       sync.RWMutex
	bindings map[reflect.Type]*binding
}

// NewContainer initializes a new empty Container.
func NewContainer() *Container {
	return &Container{
		bindings: make(map[reflect.Type]*binding),
	}
}

// Singleton binds an instantiated service or a constructor that evaluates once.
func (c *Container) Singleton(service any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val := reflect.ValueOf(service)
	typ := reflect.TypeOf(service)

	if typ == nil {
		panic("container: cannot bind nil service as singleton")
	}

	if val.Kind() == reflect.Func {
		// Constructor that should only be evaluated once upon first resolve
		outTyp := typ.Out(0)
		c.bindings[outTyp] = &binding{
			kind:        kindSingletonFactory,
			constructor: val,
		}
		return
	}

	c.bindings[typ] = &binding{
		kind:     kindInstance,
		instance: service,
	}
}

// Bind registers a constructor function for transient instantiation on each Resolve.
func (c *Container) Bind(constructor any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val := reflect.ValueOf(constructor)
	typ := reflect.TypeOf(constructor)

	if typ == nil || val.Kind() != reflect.Func {
		panic(ErrInvalidConstructor)
	}

	if typ.NumOut() < 1 || typ.NumOut() > 2 {
		panic("container: constructor must return 1 value or (value, error)")
	}

	outTyp := typ.Out(0)
	c.bindings[outTyp] = &binding{
		kind:        kindTransientFactory,
		constructor: val,
	}
}

// Transient registers a constructor function for transient instantiation on each Resolve.
func (c *Container) Transient(service any) {
	c.Bind(service)
}

// Resolve populates the target pointer with an instance of the requested type.
func (c *Container) Resolve(target any) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return ErrTargetNotPointer
	}

	targetElem := val.Elem()
	targetType := targetElem.Type()

	resolved, err := c.resolveType(targetType, make(map[reflect.Type]bool))
	if err != nil {
		return err
	}

	targetElem.Set(resolved)
	return nil
}

func (c *Container) resolveType(t reflect.Type, resolving map[reflect.Type]bool) (reflect.Value, error) {
	c.mu.Lock()
	b, exists := c.bindings[t]
	if !exists {
		c.mu.Unlock()
		return reflect.Value{}, fmt.Errorf("%w: %v", ErrBindingNotFound, t)
	}

	if b.kind == kindInstance {
		inst := reflect.ValueOf(b.instance)
		c.mu.Unlock()
		return inst, nil
	}
	c.mu.Unlock()

	if resolving[t] {
		return reflect.Value{}, fmt.Errorf("%w: %v", ErrCircularDependency, t)
	}
	resolving[t] = true
	defer delete(resolving, t)

	fn := b.constructor
	fnType := fn.Type()
	numIn := fnType.NumIn()
	inArgs := make([]reflect.Value, numIn)

	for i := 0; i < numIn; i++ {
		argType := fnType.In(i)
		argVal, err := c.resolveType(argType, resolving)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("container: failed resolving argument %d (%v) for %v: %w", i, argType, t, err)
		}
		inArgs[i] = argVal
	}

	results := fn.Call(inArgs)
	if len(results) == 2 && !results[1].IsNil() {
		errVal, ok := results[1].Interface().(error)
		if ok && errVal != nil {
			return reflect.Value{}, fmt.Errorf("container: constructor error for %v: %w", t, errVal)
		}
	}

	resultVal := results[0]

	if b.kind == kindSingletonFactory {
		c.mu.Lock()
		b.kind = kindInstance
		b.instance = resultVal.Interface()
		c.mu.Unlock()
	}

	return resultVal, nil
}

var _ core.Container = (*Container)(nil)
