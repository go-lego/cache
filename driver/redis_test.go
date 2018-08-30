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

func TestRedisMSet(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("MSET", "test1", "ok", "test2", "good").Expect("OK")

	err := r.MSet(map[string]string{"test1": "ok", "test2": "good"})
	if err != nil {
		t.Error("No error was expected to MSet, but: ", err)
	}
}

func TestRedisDel(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("DEL", "test1").Expect("OK")

	err := r.Del("test1")
	if err != nil {
		t.Error("No error was expected to del, but: ", err)
	}
}

func TestRedisExists(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("EXISTS", "test1").Expect(int64(1))
	c.Command("EXISTS", "test2").Expect(int64(0))

	b, err := r.Exists("test1")
	if err != nil {
		t.Error("No error was expected to exists, but: ", err)
	}
	if !b {
		t.Error("Key 'test1' should exists")
	}
	b, err = r.Exists("test2")
	if err != nil {
		t.Error("No error was expected to exist, but: ", err)
	}
	if b {
		t.Error("Key 'test2' should not exist")
	}
}

func TestRedisExpire(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("EXPIRE", "test1", int64(123456)).Expect("OK")

	err := r.Expire("test1", int64(123456))
	if err != nil {
		t.Error("No error was expected to expire, but: ", err)
	}
}

func TestRedisIncr(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("INCRBY", "test1", "13").Expect("100")

	nv, err := r.Incr("test1", "13")
	if err != nil {
		t.Error("No error was expected to incr, but: ", err)
	}
	if nv != "100" {
		t.Error("Incr return value incorrect")
	}
}

func TestRedisDecr(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("DECRBY", "test1", "13").Expect("100")

	nv, err := r.Decr("test1", "13")
	if err != nil {
		t.Error("No error was expected to decr, but: ", err)
	}
	if nv != "100" {
		t.Error("Decr return value incorrect")
	}
}

func TestRedisHGet(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("HGET", "test1", "k1").Expect("100")

	v, err := r.HGet("test1", "k1")
	if err != nil {
		t.Error("No error was expected to HGet, but: ", err)
	}
	if v != "100" {
		t.Error("HGet return value incorrect")
	}
}

func TestRedisHSet(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("HSET", "test1", "k1", "100").Expect("OK")

	err := r.HSet("test1", "k1", "100")
	if err != nil {
		t.Error("No error was expected to HSet, but: ", err)
	}
}

func TestRedisHMGet(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("HMGET", "test1", "k1", "k2", "k3").Expect([]interface{}{[]byte("k1"), []byte("ok"), []byte("k2"), []byte("good"), []byte("k3"), []byte("1")})

	m, err := r.HMGet("test1", []string{"k1", "k2", "k3"})
	if err != nil {
		t.Error("No error was expected to HMGet, but: ", err)
	}
	if m["k1"] != "ok" || m["k2"] != "good" || m["k3"] != "1" {
		t.Error("HMGet return value incorrect")
	}
}

func TestRedisHMSet(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("HMSET", "test1", "k1", "ok", "k2", "good", "k3", "1").Expect("ok")

	err := r.HMSet("test1", map[string]string{"k1": "ok", "k2": "good", "k3": "1"})
	if err != nil {
		t.Error("No error was expected to HMSet, but: ", err)
	}
}

func TestRedisHGetAll(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("HGETALL", "test1").Expect([]interface{}{[]byte("k1"), []byte("ok"), []byte("k2"), []byte("good"), []byte("k3"), []byte("1")})

	m, err := r.HGetAll("test1")
	if err != nil {
		t.Error("No error was expected to HMGet, but: ", err)
	}
	if m["k1"] != "ok" || m["k2"] != "good" || m["k3"] != "1" {
		t.Error("HMGet return value incorrect")
	}
}

func TestRedisHDel(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("HDEL", "test1", "k1").Expect("OK")

	err := r.HDel("test1", "k1")
	if err != nil {
		t.Error("No error was expected to HDel, but: ", err)
	}
}

func TestRedisHExists(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("HEXISTS", "test1", "k1").Expect(int64(1))
	c.Command("HEXISTS", "test1", "k2").Expect(int64(0))

	b, err := r.HExists("test1", "k1")
	if err != nil {
		t.Error("No error was expected to HExists, but: ", err)
	}
	if !b {
		t.Error("Key 'test1'.'k1' should exists")
	}
	b, err = r.HExists("test1", "k2")
	if err != nil {
		t.Error("No error was expected to HExists, but: ", err)
	}
	if b {
		t.Error("Key 'test1'.'k2' should not exist")
	}
}

func TestRedisHIncr(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("HINCRBY", "test1", "k1", "13").Expect("100")

	nv, err := r.HIncr("test1", "k1", "13")
	if err != nil {
		t.Error("No error was expected to HIncr, but: ", err)
	}
	if nv != "100" {
		t.Error("HIncr return value incorrect")
	}
}

func TestRedisHDecr(t *testing.T) {
	c := redigomock.NewConn()
	r := &redisDriver{
		pool: &testRedisPool{conn: c},
	}

	c.Command("HINCRBY", "test1", "k1", "-13").Expect("100")
	c.Command("HINCRBY", "test1", "k1", "13").Expect("100")

	nv, err := r.HDecr("test1", "k1", "13")
	if err != nil {
		t.Error("No error was expected to HDecr, but: ", err)
	}
	if nv != "100" {
		t.Error("HDecr return value incorrect")
	}

	nv, err = r.HDecr("test1", "k1", "-13")
	if err != nil {
		t.Error("No error was expected to HDecr, but: ", err)
	}
	if nv != "100" {
		t.Error("HDecr return value incorrect")
	}

	_, err = r.HDecr("test1", "k1", "1-3")
	if err == nil {
		t.Error("'Invalid delta' error was expected to HDecr, but: ", err)
	}
}
