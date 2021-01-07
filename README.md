## 缓存使用说明
全局缓存库,支持 memory 、redis 、freecache。  
调用方式 
```
  import (jcache "github.com/cn-joyconn/gocache")
  
  func test(){
    cache := &Cache{Catalog:"joyconn",CacheName:"admin"}
    //put cache
    err := cache.Put("supermanager","asbsss",1000 )
    if err!=nil{
      fmt.Println("put cache error")
    }
    //get cache
    var value string
    err=cache.Get("supermanager", &value)
    if err!=nil{
      fmt.Println("get cache error")
    }
    fmt.Println(value)

    //delete cache
    cache.Delete("supermanager")
  }
```
## 缓存配置文件说明
配置文件位置 `` ./conf/cache.yml ``
### 示例
```
cache :   # cache 配置
  cachedriver : memory # cache 引擎 支持redis 、 memory 、freecache
  redis :   # redis 配置
    key : world # redis 中的key前缀
    conn : 127.0.0.1:6379 # redis 连接
    dbnum : 0  # redis 数据库序号
    password : 123456   # redis 连接密码
  freecache : # freecache 配置
    size : 256 # 预分配内存大小（需要一次性申请所有缓存空间，容量就是固定不变的） size为M
```
### cachedriver
 全局缓存采样的存储引擎,支持redis 、 memory 、freecache (默认采用memory)

### redis
采用redis作为缓存引擎时,需要设置redis连接配置


### freecache
采用freecache作为缓存引擎时,需要设置freecache预分配内存大小,size为M (默认10M)


