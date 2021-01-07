package cache

import (
	"fmt"
	"testing"
)



func TestCache(t *testing.T) {
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