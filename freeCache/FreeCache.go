package freeCache

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	icache "github.com/cn-joyconn/joyconn-gocache/icache"
	freecache "github.com/coocood/freecache"
)

//缺点:
//     1.当需要缓存的数据占用超过提前分配缓存的 1/1024 则不能缓存
//     2.当分配内存过多则会导致内存占用高 最好不要超过100MB的内存分配
//     3.key的长度需要小于65535字节

var (
	DefaultSize                  = 64 //64M
	free        *freecache.Cache = nil
)

type FreeCache struct {
	Free *freecache.Cache
}

// 如果是需要集成到beego，则init函数必须打开，反之可以注释掉
func init() {
	
	gob.Register(map[string]interface{}{})
	gob.Register(map[string][]int{})
    gob.Register(map[string][]int64{})
    icache.Register("freecache", NewFreeCache)
}

func NewFreeCache() icache.ICache {
	cache := FreeCache{}
	return &cache
}

// 该函数是可以在任意地方使用，初始化Free进程缓存
func NewFree(m int) *FreeCache {
	if free == nil {
		cacheSize := m * 1024 * 1024
		free = freecache.NewCache(cacheSize)
	}
	//beeFree := Cache{}
	beeFree := new(FreeCache)
	beeFree.Free = free
	return beeFree
}

// 推荐使用,
func (free *FreeCache) Get(key string, value interface{}) error {
	if value == nil {
		return errors.New("val is unkown interface!!!")
	}
	cache, err := free.Free.Get([]byte(key))
	if len(cache) > 0 && err == nil {
		switch value.(type) {
			case []byte:
				value = cache
			default:
				err = jsonDecode(cache, value)
		}
		return nil
	} else {
		return err
	}
}

//批量获取keys
func (free *FreeCache) GetMulti(keys []string, values []interface{}) error {
	if values == nil {
		return errors.New("val is unkown interface!!!")
	}
	var value interface{}
	var err error
	for i := 0; i < len(keys); i++ {
		err = free.Get(keys[i], value)
		if err == nil {
			values = append(values, value)
		}
	}
	return nil
}

//设置缓存
func (free *FreeCache) Put(key string, val interface{}, timeout time.Duration) error {
	var cache []byte = make([]byte, 0)
	var err error
	// 如果value是[]byte就不需要转化了
	switch val.(type) {
		case []byte:
			cache = val.([]byte)
		default:
			cache, err = jsonEncode(val)
	}
	if err == nil {
		err = free.Free.Set([]byte(key), cache, int(timeout.Seconds()))
	}
	return err
}

// 删除key
func (free *FreeCache) Delete(key string) error {
	b := free.Free.Del([]byte(key))
	if b {
		return nil
	} else {
		return errors.New("del" + key + " error!!!")
	}
}

//对key值为int64的加1
func (free *FreeCache) Incr(key string) error {
	var value int64
	err := free.Get(key, &value)
	if err != nil {
		return errors.New("value is not an integer or out of range")
	}
	t, err := free.Free.TTL([]byte(key))
	free.Put(key, value+1, time.Duration(t)*time.Second)
	return err
}

//对key值为int64的减1
func (free *FreeCache) Decr(key string) error {
	var value int64
	err := free.Get(key, &value)
	if err != nil {
		return errors.New("value is not an integer or out of range")
	}
	t, err := free.Free.TTL([]byte(key))
	free.Put(key, value-1, time.Duration(t)*time.Second)
	return err
}

//判断指定key是否存在
func (free *FreeCache) IsExist(key string) (bool, error) {
	var buf []byte
	err := free.Get(key, buf)
	var exist = false
	if err == nil {
		if buf == nil {

		} else {
			exist = true
		}
	}
	if buf == nil || err == nil {

	}
	return exist, err
}

//清出所有缓存
func (free *FreeCache) ClearAll() error {
	free.Free.Clear()
	return nil
}

//在beego框架中，注册时会自动执行该函数初始化
func (cache *FreeCache) StartAndGC(config string) error {
	var cf map[string]int
	json.Unmarshal([]byte(config), &cf)
	if _, ok := cf["size"]; !ok {
		cf = make(map[string]int)
		cf["size"] = DefaultSize
	}
	//NewFree(int(cf["size"]))
	if free == nil {
		cacheSize := int(cf["size"]) * 1024 * 1024
		free = freecache.NewCache(cacheSize)
	}

	cache.Free = free
	cache.Free.ResetStatistics()
	// free = nil
	return nil
}

//输出cache状态
func (free *FreeCache) String() string {
	info := fmt.Sprintf("EntryCount is %d,ExpiredCount is %d,HitCount is %d,HitRate is %f,EvacuateCount Is %d,AverageAccessTime is %d,LookupCount is %d .", free.Free.EntryCount(), free.Free.ExpiredCount(),
		free.Free.HitCount(), free.Free.HitRate(), free.Free.EvacuateCount(), free.Free.AverageAccessTime(), free.Free.LookupCount())
	return info
}

//获取cache状态
func (free *FreeCache) CacheStatus() map[string]interface{} {

	infoMap := map[string]interface{}{
		"HitCount":          free.Free.HitCount(),
		"HitRate":           free.Free.HitRate(),
		"EvacuateCount":     free.Free.EvacuateCount(),
		"AverageAccessTime": free.Free.AverageAccessTime(),
		"LookupCount":       free.Free.LookupCount(),
		"EntryCount":        free.Free.EntryCount(),
		"ExpiredCount":      free.Free.ExpiredCount(),
	}
	return infoMap
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
