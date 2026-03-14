// Package serial provides FRX serialization and deserialization utilities for go-fastreport.
package serial

import (
	"fmt"
	"sort"
	"sync"

	"github.com/andrewloable/go-fastreport/report"
)

// Factory is a function that creates a new zero-value report object.
type Factory func() report.Base

// Registry maps type name strings to factory functions for FRX deserialization.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
	}
}

// Register registers a factory for the given type name.
// Returns an error if the name is already registered.
func (r *Registry) Register(name string, f Factory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("serial: type %q is already registered", name)
	}
	r.factories[name] = f
	return nil
}

// MustRegister registers a factory, panicking if the name is already registered.
func (r *Registry) MustRegister(name string, f Factory) {
	if err := r.Register(name, f); err != nil {
		panic(err)
	}
}

// Create creates a new object by type name.
// Returns (obj, nil) on success, (nil, error) if name is unknown.
func (r *Registry) Create(name string) (report.Base, error) {
	r.mu.RLock()
	f, ok := r.factories[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("serial: unknown type %q", name)
	}
	return f(), nil
}

// Has returns true if name is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	_, ok := r.factories[name]
	r.mu.RUnlock()
	return ok
}

// Names returns a sorted list of all registered type names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	r.mu.RUnlock()
	sort.Strings(names)
	return names
}

// DefaultRegistry is the package-level global registry.
var DefaultRegistry = NewRegistry()
