// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	icache "github.com/cn-joyconn/gocache/icache"
)

var (
	// Timer for how often to recycle the expired cache items in memory (in seconds)
	DefaultEvery = 60 // 1 minute
)

// MemoryItem stores memory cache item.
type MemoryItem struct {
	val         []byte
	createdTime time.Time
	lifespan    time.Duration
}

func (mi *MemoryItem) isExpire() bool {
	// 0 means forever
	if mi.lifespan == 0 {
		return false
	}
	return time.Now().Sub(mi.createdTime) > mi.lifespan
}

// MemoryCache is a memory cache adapter.
// Contains a RW locker for safe map storage.
type MemoryCache struct {
	sync.RWMutex
	dur   time.Duration
	items map[string]*MemoryItem
	Every int // run an expiration check Every clock time
}

// NewMemoryCache returns a new MemoryCache.
func NewMemoryCache() icache.ICache {
	cache := MemoryCache{items: make(map[string]*MemoryItem)}
	return &cache
}

// Get returns cache from memory.
// If non-existent or expired, return nil.
func (bc *MemoryCache) Get( key string, value interface{}) (err error) {
	if(value==nil){
        return errors.New("val is unkown interface!!!")
	}
	var val []byte
	bc.RLock()
	defer func(){
		bc.RUnlock()
		switch value.(type) {
			case []byte:
				value = val
			default:
				err = jsonDecode(val, value)
		}
	}()
	if itm, ok := bc.items[key]; ok {
		if itm.isExpire() {
			return errors.New("the key is expired")
		}
		val = itm.val
		
		
		// enc,_ := json.Marshal(value)
		// fmt.Println("get")
		// fmt.Println(string(enc))
		return  err
	}
	return  errors.New("the key isn't exist")
}


// GetMulti gets caches from memory.
// If non-existent or expired, return nil.
func (bc *MemoryCache) GetMulti( keys []string,values []interface{}) (err error) {
	if(values==nil){
        return errors.New("val is unkown interface!!!")
    }
	keysErr := make([]string, 0)
	var value interface{}
	for _, ki := range keys {
		err = bc.Get( ki,value)
		if err != nil {
			keysErr = append(keysErr, fmt.Sprintf("key [%s] error: %s", ki, err.Error()))
			continue
		}
		values = append(values, value)
	}

	if len(keysErr) == 0 {
		return  nil
	}
	return  errors.New(strings.Join(keysErr, "; "))
}

// Put puts cache into memory.
// If lifespan is 0, it will never overwrite this value unless restarted
func (bc *MemoryCache) Put( key string, val interface{}, timeout time.Duration) (err error) {
	var cache []byte = make([]byte, 0)
    // 如果value是[]byte就不需要转化了
    switch val.(type) {
		case []byte:
			cache = val.([]byte)
		default:
			cache, err = jsonEncode(val)
	}
	
	bc.Lock()
	defer bc.Unlock()
	cacheItem := MemoryItem{
		val:         cache,
		createdTime: time.Now(),
		lifespan:    timeout,
	}
	bc.items[key] = &cacheItem
	return nil
}

// Delete cache in memory.
func (bc *MemoryCache) Delete( key string) error {
	bc.Lock()
	defer bc.Unlock()
	if _, ok := bc.items[key]; !ok {
		return errors.New("key not exist")
	}
	delete(bc.items, key)
	if _, ok := bc.items[key]; ok {
		return errors.New("delete key error")
	}
	return nil
}
// Incr increases cache counter in memory.
// Supports int,int32,int64,uint,uint32,uint64.
func (bc *MemoryCache) incr( key string,val int64)(err error) {
	bc.Lock()
	defer bc.Unlock()
	var value int64
	var lifespan time.Duration
	if itm, ok := bc.items[key]; ok {
		if itm.isExpire() {
			return errors.New("the key is expired")
		}	
		lifespan = itm.lifespan		
	}
    err = bc.Get(key, &value)
    if err != nil {
        return errors.New("value is not an integer or out of range")
    }
	err = bc.Put(key, value + val, lifespan)
    return err
	
}
// Incr increases cache counter in memory.
// Supports int,int32,int64,uint,uint32,uint64.
func (bc *MemoryCache) Incr( key string)(err error) {
	err  = bc.incr(key,1)
	return
}

// Decr decreases counter in memory.
func (bc *MemoryCache) Decr( key string) (err error) {
	err  = bc.incr(key,-1)
	return
}

// IsExist checks if cache exists in memory.
func (bc *MemoryCache) IsExist( key string) (bool, error) {
	bc.RLock()
	defer bc.RUnlock()
	if v, ok := bc.items[key]; ok {
		return !v.isExpire(), nil
	}
	return false, nil
}

// ClearAll deletes all cache in memory.
func (bc *MemoryCache) ClearAll() error {
	bc.Lock()
	defer bc.Unlock()
	bc.items = make(map[string]*MemoryItem)
	return nil
}

// StartAndGC starts memory cache. Checks expiration in every clock time.
func (bc *MemoryCache) StartAndGC(config string) error {
	var cf map[string]int
	json.Unmarshal([]byte(config), &cf)
	if _, ok := cf["interval"]; !ok {
		cf = make(map[string]int)
		cf["interval"] = DefaultEvery
	}
	dur := time.Duration(cf["interval"]) * time.Second
	bc.Every = cf["interval"]
	bc.dur = dur
	go bc.vacuum()
	return nil
}

// check expiration.
func (bc *MemoryCache) vacuum() {
	bc.RLock()
	every := bc.Every
	bc.RUnlock()

	if every < 1 {
		return
	}
	for {
		<-time.After(bc.dur)
		bc.RLock()
		if bc.items == nil {
			bc.RUnlock()
			return
		}
		bc.RUnlock()
		if keys := bc.expiredKeys(); len(keys) != 0 {
			bc.clearItems(keys)
		}
	}
}

// expiredKeys returns keys list which are expired.
func (bc *MemoryCache) expiredKeys() (keys []string) {
	bc.RLock()
	defer bc.RUnlock()
	for key, itm := range bc.items {
		if itm.isExpire() {
			keys = append(keys, key)
		}
	}
	return
}

// ClearItems removes all items who's key is in keys
func (bc *MemoryCache) clearItems(keys []string) {
	bc.Lock()
	defer bc.Unlock()
	for _, key := range keys {
		delete(bc.items, key)
	}
}

func init() {
 	icache.Register("memory", NewMemoryCache)
}

// json序列化
func jsonEncode(data interface{}) ([]byte, error) {
    enc,err := json.Marshal(data)
    if err != nil {
        return nil, err
	}
    return enc, nil
}

// json反序列化
func jsonDecode(data []byte, to interface{}) error {

	err := json.Unmarshal(data, to)
    return err
}