package monitor

import (
	"sync"
)

// Tags is the interface for managing tags
type Tags interface {
	Add(key, value string)
	Delete(key string)
	Get() map[string]string
}

// tagsImpl is the implementation of the Tags interface
type tagsImpl struct {
	mu   sync.RWMutex
	tags map[string]string
}

// NewTags creates a new instance of tagsImpl
func NewTags() Tags {
	return &tagsImpl{
		tags: make(map[string]string),
	}
}

// Add adds a new tag to the Tags map
func (t *tagsImpl) Add(key, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tags[key] = value
}

// Delete deletes a tag from the Tags map
func (t *tagsImpl) Delete(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tags, key)
}

// Get returns a copy of the Tags map
func (t *tagsImpl) Get() map[string]string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	copy := make(map[string]string, len(t.tags))
	for k, v := range t.tags {
		copy[k] = v
	}
	return copy
}
