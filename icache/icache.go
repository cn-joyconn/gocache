
package icache

import (
	"fmt"
	"time"
)


type ICache interface {
	// Get a cached value by key.
	Get( key string,val interface{}) ( error)
	// GetMulti is a batch version of Get.
	GetMulti(keys []string,val []interface{}) ( error)
	// Set a cached value with key and expire time.
	Put( key string, val interface{}, timeout time.Duration) error
	// Delete cached value by key.
	Delete( key string) error
	// Increment a cached int value by key, as a counter.
	Incr( key string) error
	// Decrement a cached int value by key, as a counter.
	Decr( key string) error
	// Check if a cached value exists or not.
	IsExist( key string) (bool, error)
	// Clear all cache.
	ClearAll() error
	// Start gc routine based on config string settings.
	StartAndGC(config string) error
}

// Instance is a function create a new Cache Instance
type Instance func() ICache

var adapters = make(map[string]Instance)

// Register makes a cache adapter available by the adapter name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, adapter Instance) {
	if adapter == nil {
		panic("cache: Register adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("cache: Register called twice for adapter " + name)
	}
	adapters[name] = adapter
}

// NewCache creates a new cache driver by adapter name and config string.
// config: must be in JSON format such as {"interval":360}.
// Starts gc automatically.
func NewCache(adapterName, config string) (adapter ICache, err error) {
	instanceFunc, ok := adapters[adapterName]
	if !ok {
		err = fmt.Errorf("cache: unknown adapter name %q (forgot to import?)", adapterName)
		return
	}
	adapter = instanceFunc()
	err = adapter.StartAndGC(config)
	if err != nil {
		adapter = nil
	}
	return
}
