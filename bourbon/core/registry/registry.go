package registry

import "sync"

type Registry struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]interface{}),
	}
}

func (r *Registry) Register(name string, service interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[name] = service
}

func (r *Registry) Get(name string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	service, ok := r.services[name]
	return service, ok
}

func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.services[name]
	return ok
}
