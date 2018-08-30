package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/garyburd/redigo/redis"
)

func main() {
	if true {
		runtime.GOMAXPROCS(1)
		// ct := make(chan int, 100)
		ct := make(chan int)
		go func() { //producer
			time.Sleep(500 * time.Millisecond)
			for i := 0; i < 30; i++ {
				ct <- i
			}
			close(ct)
		}()

		go func() { //consumer
			for {
				select {
				case i, ok := <-ct:
					if !ok {
						return
					}
					fmt.Println("consumer A: ", i)
					// time.Sleep(1 * time.Millisecond)
				}
			}
		}()
		go func() { //consumer
			for {
				select {
				case i, ok := <-ct:
					if !ok {
						return
					}
					fmt.Println("consumer B: ", i)
					// time.Sleep(1 * time.Millisecond)
				}
			}
		}()
		time.Sleep(1 * time.Second)
		return
	}

	if true {
		runtime.GOMAXPROCS(1)
		for i := 0; i < 10; i++ {
			go func() {
				fmt.Printf("A : %d\n", i)
			}()
		}
		for j := 0; j < 10; j++ {
			go func(i int) {
				fmt.Printf("B : %d\n", i)
			}(j)
		}
		time.Sleep(1 * time.Second)
		return
	}
	if true {
		type student struct {
			Name string
			Age  int
		}
		stus := []student{
			{Name: "Henry", Age: 10},
			{Name: "Toto", Age: 17},
			{Name: "Marry", Age: 15},
		}
		m := map[string]*student{}
		for _, s := range stus {
			m[s.Name] = &s
		}
		for k, v := range m {
			fmt.Println(k, v.Name, v.Age)
		}
		return
	}
	// dv := driver.NewDriver(driver.Port(6379))
	// r := cache.NewCache(cache.Driver(dv))
	// if err := r.Init(); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// r.Get("test")
	// r.Set("test", "test")
	// r.Get("test")
	r, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	ss, err := redis.Strings(r.Do("MGET", "test.1", "ssssss", "test.2", "kkkkk"))
	for _, s := range ss {
		fmt.Printf("%s %T\n", s, s)
	}
	fmt.Println("Length:", len(ss))
	// fmt.Println(redis.String(r.Do("GET", "not.exist")))
	// fmt.Println(redis.String(r.Do("HGET", "not.existh", "tt")))
	// fmt.Println(redis.StringMap(r.Do("HGETALL", "not.existh")))
}
