package monitor

//go:generate mockgen -source=$GOFILE -destination=../mocks/mock_$GOPACKAGE/$GOFILE -package=mock_$GOPACKAGE

import (
	"context"
	"fmt"
	"sync"

	"github.com/vd09/trading-algorithm-backtesting-system/constraint"
)

// Tags is the interface for managing tags
type Tags interface {
	Add(key string, value interface{})
	Delete(key string)
	Get() map[string]string
	With(key string, value interface{}) Tags
	AddTagsFromCtx(ctx context.Context)
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

// NewTags creates a new instance of tagsImpl
func NewTagsKV(key string, value interface{}) Tags {
	tags := NewTags()
	tags.Add(key, value)
	return tags
}

// Add adds a new tag to the Tags map
func (t *tagsImpl) Add(key string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if value == nil {
		t.tags[key] = "NA"
	} else {
		t.tags[key] = fmt.Sprintf("%v", value)
	}
}

// With adds a new tag to the Tags map
func (t *tagsImpl) With(key string, value interface{}) Tags {
	t.Add(key, value)
	return t
}

func (t *tagsImpl) AddTagsFromCtx(ctx context.Context) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if slice, ok := ctx.Value(constraint.COMMON_LABELS_CTX).(Labels); ok {
		for _, labelKey := range slice {
			if labelValue, ok := ctx.Value(labelKey).(string); ok {
				t.Add(labelKey, labelValue)
			}
		}
	}
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
