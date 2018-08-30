package driver

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

type testRedisPool struct {
	conn redis.Conn
}

func (p *testRedisPool) Get() redis.Conn {
	return p.conn
}

func TestRedisGet(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("GET", "test").Expect("test")
	c.Command("GET", "testno").ExpectError(redis.ErrNil)

	v, err := r.Get("test")
	if err != nil {
		t.Error("No error was expected to get 'test', but: ", err)
	}
	if v != "test" {
		t.Error("'test' value was expected to 'test', but: ", v)
	}

	_, err = r.Get("testno")
	if err != redis.ErrNil {
		t.Error("Expected error: ", redis.ErrNil, " but: ", err)
	}
}

func TestRedisSet(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("SET", "test", "test").Expect("OK")

	err := r.Set("test", "test")
	if err != nil {
		t.Error("No error was expected to get 'test', but: ", err)
	}
}

func TestRedisMGet(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("MGET", "test1", "test2", "testno", "test3").Expect([]interface{}{"1", "2", nil, "test3"})
	ret, err := r.MGet([]string{"test1", "test2", "testno", "test3"})
	if err != nil {
		t.Error("No error was expected to MGet, but: ", err)
	}
	if len(ret) != 4 {
		t.Error("MGet result length was expected to 4, but: ", len(ret))
	}
	if ret["test1"] != "1" || ret["test2"] != "2" || ret["test3"] != "test3" || ret["testno"] != "" {
		t.Error("MGet result was incorrect: ", ret)
	}
}
