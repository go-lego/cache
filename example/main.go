package main

import (
	"fmt"

	"github.com/go-lego/cache"
	"github.com/go-lego/cache/driver"
)

func main() {
	dv := driver.NewDriver(driver.Port(6379))
	r := cache.NewCache(cache.Driver(dv))
	if err := r.Init(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r.Get("int"))
	fmt.Println(r.Incr("int", 5))
	// fmt.Println(r.MGet([]string{"test.1", "test.2"}))
	// fmt.Println(r.HGetAll("ttt"))
	// r.Get("test")
	// r.Set("test", "test")
	// r.Get("test")
	// r, err := redis.Dial("tcp", "127.0.0.1:6379")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// ss, err := redis.Strings(r.Do("MGET", "test.1", "ssssss", "test.2", "kkkkk"))
	// for _, s := range ss {
	// 	fmt.Printf("%s %T\n", s, s)
	// }
	// fmt.Println("Length:", len(ss))
	// fmt.Println(redis.String(r.Do("GET", "not.exist")))
	// fmt.Println(redis.String(r.Do("HGET", "not.existh", "tt")))
	// fmt.Println(redis.StringMap(r.Do("HGETALL", "not.existh")))

	// fmt.Println(redis.String(r.Do("GET", "int")))
	// fmt.Println(redis.String(r.Do("GET", "float")))

	// r.Do("SET", "incrtest", "10")
	// r.Do("INCRBY", "incrtest", "3")
	// fmt.Println(redis.String(r.Do("GET", "incrtest")))
}
