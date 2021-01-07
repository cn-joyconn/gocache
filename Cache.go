package cache

import (
	"encoding/json"
	"strconv"
	"time"

	_ "github.com/cn-joyconn/joyconn-gocache/freeCache"
	icache "github.com/cn-joyconn/joyconn-gocache/icache"
	_ "github.com/cn-joyconn/joyconn-gocache/memory"
	_ "github.com/cn-joyconn/joyconn-gocache/redis"
	log "github.com/cn-joyconn/joyconn-gologs"
	filetool "github.com/cn-joyconn/joyconn-goutils/filetool"
	strtool "github.com/cn-joyconn/joyconn-goutils/strtool"
	yaml "gopkg.in/yaml.v2"
)
var globlaCache icache.ICache
var logger = log.GetLogger("")
func init() {
	cachecfg := make(map[string]interface{})
	var err error
	selfDir := filetool.SelfDir()
	configPath := selfDir + "/conf/cache.yml"
	var cacheconfs cacheConf
	if filetool.IsExist(configPath) {
		configBytes, err := filetool.ReadFileToBytes(configPath)
		if err != nil {
			logger.Error(err.Error())
		}
		err = yaml.Unmarshal(configBytes, &cacheconfs)
		if err != nil {
			logger.Error("解析log.yml文件失败")
		}
	} else {
		logger.Error("未找到log.yml")
	}
	if cacheconfs.Cachedriver == "freecache" {
		if cacheconfs.Freecache.Size<1{
			cachecfg["size"] = 10
		}else {
			cachecfg["size"] = strconv.Itoa(cacheconfs.Freecache.Size)
		}
		 //config.String("cache::joyconn.cache.free.cache.size")
		cacheConfJson,_ := json.Marshal(cachecfg)
		globlaCache, err = icache.NewCache("freecache", string(cacheConfJson))
	} else if cacheconfs.Cachedriver == "redis" {
		cachecfg["key"] = cacheconfs.Redis.Key
		cachecfg["conn"] = cacheconfs.Redis.Conn
		cachecfg["dbNum"]= strconv.Itoa(cacheconfs.Redis.Dbnum)
		cachecfg["password"] = cacheconfs.Redis.Password
		cacheConfJson, _ := json.Marshal(cachecfg)
		globlaCache, err = icache.NewCache("redis", string(cacheConfJson))
	}else{
		cachecfg["interval"]=60
		cacheConfJson, _ := json.Marshal(cachecfg)
		globlaCache, err = icache.NewCache("memory",string(cacheConfJson))
	}
	if err != nil {
		panic(err)
	}
}
//cacheConf 缓存配置 
type cacheConf struct{ 
	Cachedriver string
	Redis struct {
		  Key string
		  Conn string
		  Dbnum int
		  Password string
	}
	Freecache struct {
		  Size int
	}
}
type  Cache struct {
	Catalog      string
	CacheName string 
}
func (Cache  *Cache)getkey(key string) (result string){
	result = Cache.Catalog + ":" + Cache.CacheName + ":" + key
	return 
}


// Get a cached value by key.
func (Cache  *Cache)Get( key string,val interface{}) ( error){
	skey := Cache.getkey(key)
	return globlaCache.Get(skey,&val)
}
// GetMulti is a batch version of Get.
func (Cache  *Cache)GetMulti(keys []string,values []interface{}) ( error){
	skeys :=make( []string,len(keys))
	var skey string
	for i := 0; i < len(keys); i++ {
		skey = Cache.getkey(keys[i])
		skeys[i]=skey
	}
	return globlaCache.GetMulti(skeys,values)
}
// Set a cached value with key and expire time.
func (Cache  *Cache)Put( key string, val interface{}, timeout time.Duration) error{
	skey := Cache.getkey(key)
	return globlaCache.Put(skey,&val, timeout * time.Second)
}
// Delete cached value by key.
func (Cache  *Cache)Delete( key string) error{
	skey := Cache.getkey(key)
	return globlaCache.Delete(skey)
}
// Increment a cached int value by key, as a counter.
func (Cache  *Cache)Incr( key string) error{
	skey := Cache.getkey(key)
	return globlaCache.Incr(skey);

}
// Decrement a cached int value by key, as a counter.
func (Cache  *Cache)Decr( key string) error{
	skey := Cache.getkey(key)
	return globlaCache.Decr(skey);

}
// Check if a cached value exists or not.
func (Cache  *Cache)IsExist( key string) (bool, error){
	skey := Cache.getkey(key)
	return globlaCache.IsExist(skey);
}
// Clear all cache.
func ClearAll() error{
	return globlaCache.ClearAll();
}

