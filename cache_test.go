package cache

import (
	"errors"
	"testing"

	dmock "github.com/go-lego/cache/driver/mock"
	"github.com/golang/mock/gomock"
)

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().Get("test").Return("test", nil)
	d.EXPECT().Get("testno").Return("", errors.New("test"))

	if len(c.keys) > 0 {
		t.Error("Initial memory was expected to be empty, but: ", c.keys)
	}
	v, err := c.Get("test")
	if err != nil {
		t.Error("No error was expected for get, but: ", err)
	}
	if v != "test" {
		t.Error("Get key 'test' was expected to value 'test', but: ", v)
	}

	v, ok := c.keys["test"]
	if !ok || v != "test" {
		t.Error("Memory incorrect after get: ", v, ok)
	}
	// get from memory
	v, err = c.Get("test")
	if err != nil {
		t.Error("No error was expected for get, but: ", err)
	}
	if v != "test" {
		t.Error("Get key 'test' was expected to value 'test', but: ", v)
	}

	_, err = c.Get("testno")
	if err == nil {
		t.Error("Error 'test' was expected, but: ", err)
	}
}

func TestSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().Set("test", "test").Return(nil)

	if len(c.keys) > 0 {
		t.Error("Initial memory was expected to be empty, but: ", c.keys)
	}
	err := c.Set("test", "test")
	if err != nil {
		t.Error("No error was expected for set, but: ", err)
	}
	v, ok := c.keys["test"]
	if !ok || v != "test" {
		t.Error("Memory incorrect after set: ", v, ok)
	}
}

func TestTransSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	if len(c.keys) > 0 {
		t.Error("Initial memory was expected to be empty, but: ", c.keys)
	}
	if c.getCurrentTransaction() != nil {
		t.Error("Initial transaction was expected to nil")
	}
	tx := c.BeginTransaction()
	if c.getCurrentTransaction() != tx {
		t.Error("Current transaction is not the one begin")
	}
	err := c.Set("test", "test")
	if err != nil {
		t.Error("No error was expected for set, but: ", err)
	}
	v, ok := c.keys["test"]
	if !ok || v != "test" {
		t.Error("Memory incorrect after set: ", v, ok)
	}
	if len(c.tx.cmds) != 1 {
		t.Error("Transaction commands size was expected to 1")
	}
	if c.tx.cmds[0].t != typeSet {
		t.Error("Transaction first command type was expected to typeSet")
	}
	if c.tx.cmds[0].args[0] != "test" || c.tx.cmds[0].args[1] != "test" {
		t.Error("Transaction first command arguments incorrect")
	}

	d.EXPECT().Set("test", "test").Return(nil)
	err = tx.Commit()
	if err != nil {
		t.Error("No error was expected for transaction commit, but: ", err)
	}

	if c.tx.active {
		t.Error("Transaction status should be inactive after commit")
	}
}

func TestMGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	// init memory
	c.keys["test1"] = "1"
	c.keys["test2"] = flagValueNil
	c.delKeys["test5"] = ""

	d.EXPECT().MGet([]string{"test3", "test4"}).Return(map[string]string{"test3": "t3", "test4": ""}, nil)

	ret, err := c.MGet([]string{"test1", "test2", "test3", "test4", "test5"})

	// t.Error(ret)
	if err != nil {
		t.Error("No error was expectected for MGet, but: ", err)
	}
	if ret["test1"] != "1" || ret["test2"] != "" || ret["test3"] != "t3" || ret["test4"] != "" || ret["test5"] != "" {
		t.Error("MGet result was incorrect")
	}
}

func TestMSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	// init memory
	c.delKeys["test2"] = ""

	d.EXPECT().MSet(map[string]interface{}{"test1": 1, "test2": "good"}).Return(nil)

	err := c.MSet(map[string]interface{}{"test1": 1, "test2": "good"})

	if err != nil {
		t.Error("No error was expectected for MSet, but: ", err)
	}
	if len(c.delKeys) > 0 {
		t.Error("MSet should delete the delKeys memory")
	}
	if c.keys["test1"] != "1" || c.keys["test2"] != "good" {
		t.Error("MSet memory incorrect")
	}
}

func TestTransMSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	// init memory
	c.delKeys["test2"] = ""

	// d.EXPECT().MSet(map[string]interface{}{"test1": 1, "test2": "good"}).Return(nil)
	tx := c.BeginTransaction()
	args := map[string]interface{}{"test1": 1, "test2": "good"}
	err := c.MSet(args)

	if err != nil {
		t.Error("No error was expectected for MSet, but: ", err)
	}
	if len(c.delKeys) > 0 {
		t.Error("MSet should delete the delKeys memory")
	}
	if c.keys["test1"] != "1" || c.keys["test2"] != "good" {
		t.Error("MSet memory incorrect")
	}
	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeMSet {
		t.Error("Transaction first command type was expected to typeMSet")
	}
	d.EXPECT().MSet(args).Return(nil)
	err = tx.Commit()
	if err != nil {
		t.Error("No error was expected for transaction commit, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after commit")
	}
}

func TestDel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.keys["test"] = "ok"

	d.EXPECT().Del("test").Return(nil)

	err := c.Del("test")

	if err != nil {
		t.Error("No error was expected for del, but: ", err)
	}
	if len(c.keys) > 0 || len(c.delKeys) == 0 {
		t.Error("Del should update memory keys and delKeys")
	}
	_, err = c.Get("test")
	if err != ErrValueNil {
		t.Error("Deleted value should return error")
	}
}

func TestTransDel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.keys["test"] = "ok"

	tx := c.BeginTransaction()
	err := c.Del("test")

	if err != nil {
		t.Error("No error was expected for del, but: ", err)
	}
	if len(c.keys) > 0 || len(c.delKeys) == 0 {
		t.Error("Del should update memory keys and delKeys")
	}
	_, err = c.Get("test")
	if err != ErrValueNil {
		t.Error("Deleted value should return error")
	}
	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeDel {
		t.Error("Transaction first command type was expected to typeDel")
	}

	d.EXPECT().Del("test").Return(nil)
	err = tx.Commit()
	if err != nil {
		t.Error("No error was expected for transaction commit, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after commit")
	}
}

func TestExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.delKeys["test"] = ""
	c.keys["test1"] = "ok"
	c.keys["test2"] = flagValueNil

	d.EXPECT().Exists("test3").Return(true, nil)

	b, err := c.Exists("test")

	if err != nil {
		t.Error("No error was expected for exists, but: ", err)
	}
	if b {
		t.Error("Key 'test' should exist")
	}

	b, err = c.Exists("test1")

	if err != nil {
		t.Error("No error was expected for exists, but: ", err)
	}
	if !b {
		t.Error("Key 'test1' should exist")
	}

	b, err = c.Exists("test2")

	if err != nil {
		t.Error("No error was expected for exists, but: ", err)
	}
	if b {
		t.Error("Key 'test2' should not exist")
	}

	b, err = c.Exists("test3")

	if err != nil {
		t.Error("No error was expected for exists, but: ", err)
	}
	if !b {
		t.Error("Key 'test3' should exist")
	}
}

func TestExpire(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().Expire("test", int64(30)).Return(nil)

	err := c.Expire("test", 30)

	if err != nil {
		t.Error("No error was expected for del, but: ", err)
	}
}

func TestTransExpire(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	tx := c.BeginTransaction()
	err := c.Expire("test", 30)

	if err != nil {
		t.Error("No error was expected for del, but: ", err)
	}

	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeExpire {
		t.Error("Transaction first command type was expected to typeExpire")
	}

	d.EXPECT().Expire("test", int64(30)).Return(nil)
	err = tx.Commit()
	if err != nil {
		t.Error("No error was expected for transaction commit, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after commit")
	}
}

func TestIncr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().Incr("test", 10).Return("", errors.New("test"))
	d.EXPECT().Incr("test1", 11).Return("22", nil)

	_, err := c.Incr("test", 10)

	if err == nil || err.Error() != "test" {
		t.Error("Error was expected")
	}

	v, err := c.Incr("test1", 11)
	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Increased value was incorrect")
	}
	if c.keys["test1"] != "22" {
		t.Error("Increase memory was not updated")
	}
}

func TestTransIncrCommit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().Incr("test1", 11).Return("22", nil)

	tx := c.BeginTransaction()

	v, err := c.Incr("test1", 11)

	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Increased value was incorrect")
	}
	if c.keys["test1"] != "22" {
		t.Error("Increase memory was not updated")
	}

	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeIncr {
		t.Error("Transaction first command type was expected to typeIncr")
	}

	err = tx.Commit()

	if err != nil {
		t.Error("No error was expected for transaction commit, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after commit")
	}
}

func TestTransIncrRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().Incr("test1", 11).Return("22", nil)
	d.EXPECT().Decr("test1", 11).Return("11", nil)

	tx := c.BeginTransaction()

	v, err := c.Incr("test1", 11)

	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Increased value was incorrect")
	}
	if c.keys["test1"] != "22" {
		t.Error("Increase memory was not updated")
	}
	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeIncr {
		t.Error("Transaction first command type was expected to typeIncr")
	}

	err = tx.Rollback()

	if err != nil {
		t.Error("No error was expected for transaction rollback, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after rollback")
	}
}

func TestDecr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().Decr("test", 10).Return("", errors.New("test"))
	d.EXPECT().Decr("test1", 11).Return("22", nil)

	_, err := c.Decr("test", 10)

	if err == nil || err.Error() != "test" {
		t.Error("Error was expected")
	}

	v, err := c.Decr("test1", 11)
	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Decreased value was incorrect")
	}
	if c.keys["test1"] != "22" {
		t.Error("Decreased memory was not updated")
	}
}

func TestTransDecrCommit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().Decr("test1", 11).Return("22", nil)

	tx := c.BeginTransaction()

	v, err := c.Decr("test1", 11)

	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Decreased value was incorrect")
	}
	if c.keys["test1"] != "22" {
		t.Error("Decreased memory was not updated")
	}

	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeDecr {
		t.Error("Transaction first command type was expected to typeDecr")
	}

	err = tx.Commit()

	if err != nil {
		t.Error("No error was expected for transaction commit, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after commit")
	}
}

func TestTransDecrRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().Decr("test1", 11).Return("22", nil)
	d.EXPECT().Incr("test1", 11).Return("11", nil)

	tx := c.BeginTransaction()

	v, err := c.Decr("test1", 11)

	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Decreased value was incorrect")
	}
	if c.keys["test1"] != "22" {
		t.Error("Decreased memory was not updated")
	}
	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeDecr {
		t.Error("Transaction first command type was expected to typeDecr")
	}

	err = tx.Rollback()

	if err != nil {
		t.Error("No error was expected for transaction commit, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after commit")
	}
}

func TestHGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.delKeys["test1"] = ""
	c.hsets = map[string]map[string]string{
		"hash": map[string]string{
			"test1": "1",
			"test2": "good",
		},
	}

	d.EXPECT().HGet("hash", "test3").Return("OK", nil)

	_, err := c.HGet("test1", "a")
	if err != ErrValueNil {
		t.Error("ErrValueNil was expected, but: ", err)
	}
	v, err := c.HGet("hash", "test1")
	if err != nil {
		t.Error("No error was expected, but: ", err)
	}
	if v != "1" {
		t.Error("HGet value was expected to '1', but: ", v)
	}
	v, err = c.HGet("hash", "test3")
	if err != nil {
		t.Error("No error was expected, but: ", err)
	}
	if v != "OK" {
		t.Error("HGet value was expected to 'OK', but: ", v)
	}
}

func TestHSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.delKeys["hash"] = ""

	d.EXPECT().HSet("hash", "test1", 1).Return(nil)

	err := c.HSet("hash", "test1", 1)

	if err != nil {
		t.Error("No error was expected, but: ", err)
	}
	m, ok := c.hsets["hash"]
	if !ok {
		t.Error("HSet memory is not set")
	}
	if m["test1"] != "1" {
		t.Error("HSet memory is not set correctly")
	}
	if _, ok := c.delKeys["hash"]; ok {
		t.Error("HSet delete keys is not cleaned")
	}
}

func TestTransHSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	tx := c.BeginTransaction()
	c.delKeys["hash"] = ""

	err := c.HSet("hash", "test1", 1)

	if err != nil {
		t.Error("No error was expected, but: ", err)
	}
	m, ok := c.hsets["hash"]
	if !ok {
		t.Error("HSet memory is not set")
	}
	if m["test1"] != "1" {
		t.Error("HSet memory is not set correctly")
	}
	if _, ok := c.delKeys["hash"]; ok {
		t.Error("HSet delete keys is not cleaned")
	}
	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeHSet {
		t.Error("Transaction first command type was expected to typeHSet")
	}

	d.EXPECT().HSet("hash", "test1", 1).Return(nil)
	tx.Commit()
	if c.tx.active {
		t.Error("Transaction should be inactive after commit")
	}
	if len(c.tx.cmds) != 0 {
		t.Error("Transaction commands should be empty after commit")
	}
}

func TestHMGetAllHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.hsets["hash"] = map[string]string{
		"test1": "1",
		"test2": "good",
		"test3": "ok",
		"test4": flagValueNil,
	}

	res, err := c.HMGet("hash", []string{"test1", "test3", "test4"})
	if err != nil {
		t.Error("No error was expected, but: ", err)
	}
	if len(res) != 3 {
		t.Error("Result size was expected to 3")
	}
	if res["test1"] != "1" || res["test3"] != "ok" || res["test4"] != "" {
		t.Error("Result was incorrect")
	}
}

func TestHMGetPartHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.hsets["hash"] = map[string]string{
		"test1": "1",
		"test2": "good",
		"test3": "ok",
		"test4": flagValueNil,
	}

	d.EXPECT().HMGet("hash", []string{"test5", "test6"}).Return(map[string]string{"test5": "test5"}, nil)

	res, err := c.HMGet("hash", []string{"test1", "test3", "test4", "test5", "test6"})
	if err != nil {
		t.Error("No error was expected, but: ", err)
	}
	if len(res) != 4 {
		t.Error("Result size was expected to 4")
	}
	if res["test1"] != "1" || res["test3"] != "ok" || res["test4"] != "" || res["test5"] != "test5" || res["test6"] != "" {
		t.Error("Result was incorrect")
	}
	if c.hsets["hash"]["test5"] != "test5" {
		t.Error("Memory was not updated")
	}
}

func TestHMGetDeleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.delKeys["hash"] = ""

	res, err := c.HMGet("hash", []string{"test1", "test3", "test4", "test5", "test6"})
	if err != nil {
		t.Error("No error was expected, but: ", err)
	}
	if len(res) != 0 {
		t.Error("Result size was expected to 0")
	}
}

func TestHMSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.delKeys["hash"] = ""
	c.hsets["hash"] = map[string]string{
		"test1": "1",
		"test2": "good",
		"test3": "ok",
		"test4": flagValueNil,
	}

	d.EXPECT().HMSet("hash", map[string]interface{}{"test1": 10, "test4": "tt", "test5": "test5"}).Return(nil)

	err := c.HMSet("hash", map[string]interface{}{"test1": 10, "test4": "tt", "test5": "test5"})

	if err != nil {
		t.Error("No error was expected")
	}
	if _, ok := c.delKeys["hash"]; ok {
		t.Error("Memory deleted key should be empty")
	}
	m := c.hsets["hash"]
	if m["test1"] != "10" || m["test2"] != "good" || m["test3"] != "ok" || m["test4"] != "tt" || m["test5"] != "test5" {
		t.Error("Memory hset was not updated")
	}
}

func TestTransHMSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.delKeys["hash"] = ""
	c.hsets["hash"] = map[string]string{
		"test1": "1",
		"test2": "good",
		"test3": "ok",
		"test4": flagValueNil,
	}

	tx := c.BeginTransaction()

	err := c.HMSet("hash", map[string]interface{}{"test1": 10, "test4": "tt", "test5": "test5"})

	if err != nil {
		t.Error("No error was expected")
	}
	if _, ok := c.delKeys["hash"]; ok {
		t.Error("Memory deleted key should be empty")
	}
	m := c.hsets["hash"]
	if m["test1"] != "10" || m["test2"] != "good" || m["test3"] != "ok" || m["test4"] != "tt" || m["test5"] != "test5" {
		t.Error("Memory hset was not updated")
	}
	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeHMSet {
		t.Error("Transaction first command type was expected to typeHMSet")
	}

	d.EXPECT().HMSet("hash", map[string]interface{}{"test1": 10, "test4": "tt", "test5": "test5"}).Return(nil)
	tx.Commit()
	if c.tx.active {
		t.Error("Transaction should be inactive after commit")
	}
	if len(c.tx.cmds) != 0 {
		t.Error("Transaction commands should be empty after commit")
	}
}

func TestHIncr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().HIncr("hash", "test", 10).Return("", errors.New("test"))
	d.EXPECT().HIncr("hash", "test1", 11).Return("22", nil)

	_, err := c.HIncr("hash", "test", 10)

	if err == nil || err.Error() != "test" {
		t.Error("Error was expected")
	}

	v, err := c.HIncr("hash", "test1", 11)
	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Increased value was incorrect")
	}
	if c.hsets["hash"]["test1"] != "22" {
		t.Error("Increase memory was not updated")
	}
}

func TestTransHIncrCommit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().HIncr("hash", "test1", 11).Return("22", nil)

	tx := c.BeginTransaction()

	v, err := c.HIncr("hash", "test1", 11)

	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Increased value was incorrect")
	}
	if c.hsets["hash"]["test1"] != "22" {
		t.Error("Increase memory was not updated")
	}

	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeHIncr {
		t.Error("Transaction first command type was expected to typeHIncr")
	}

	err = tx.Commit()

	if err != nil {
		t.Error("No error was expected for transaction commit, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after commit")
	}
	if len(c.tx.cmds) != 0 {
		t.Error("Transaction command should be cleaned after commit")
	}
}

func TestTransHIncrRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().HIncr("hash", "test1", 11).Return("22", nil)
	d.EXPECT().HDecr("hash", "test1", 11).Return("11", nil)

	tx := c.BeginTransaction()

	v, err := c.HIncr("hash", "test1", 11)

	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Increased value was incorrect")
	}
	if c.hsets["hash"]["test1"] != "22" {
		t.Error("Increase memory was not updated")
	}
	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeHIncr {
		t.Error("Transaction first command type was expected to typeHIncr")
	}

	err = tx.Rollback()

	if err != nil {
		t.Error("No error was expected for transaction rollback, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after rollback")
	}
	if len(c.tx.cmds) != 0 {
		t.Error("Transaction command should be cleaned after rollback")
	}
}

func TestHDecr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().HDecr("hash", "test", 10).Return("", errors.New("test"))
	d.EXPECT().HDecr("hash", "test1", 11).Return("22", nil)

	_, err := c.HDecr("hash", "test", 10)

	if err == nil || err.Error() != "test" {
		t.Error("Error was expected")
	}

	v, err := c.HDecr("hash", "test1", 11)
	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Decreased value was incorrect")
	}
	if c.hsets["hash"]["test1"] != "22" {
		t.Error("Decrease memory was not updated")
	}
}

func TestTransHDecrCommit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().HDecr("hash", "test1", 11).Return("22", nil)

	tx := c.BeginTransaction()

	v, err := c.HDecr("hash", "test1", 11)

	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Decreased value was incorrect")
	}
	if c.hsets["hash"]["test1"] != "22" {
		t.Error("Decrease memory was not updated")
	}

	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeHDecr {
		t.Error("Transaction first command type was expected to typeHDecr")
	}

	err = tx.Commit()

	if err != nil {
		t.Error("No error was expected for transaction commit, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after commit")
	}
	if len(c.tx.cmds) != 0 {
		t.Error("Transaction command should be cleaned after commit")
	}
}

func TestTransHDecrRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	d.EXPECT().HDecr("hash", "test1", 11).Return("22", nil)
	d.EXPECT().HIncr("hash", "test1", 11).Return("11", nil)

	tx := c.BeginTransaction()

	v, err := c.HDecr("hash", "test1", 11)

	if err != nil {
		t.Error("No error was expected")
	}
	if v != "22" {
		t.Error("Decreased value was incorrect")
	}
	if c.hsets["hash"]["test1"] != "22" {
		t.Error("Decrease memory was not updated")
	}
	if len(c.tx.cmds) != 1 || c.tx.cmds[0].t != typeHDecr {
		t.Error("Transaction first command type was expected to typeHDecr")
	}

	err = tx.Rollback()

	if err != nil {
		t.Error("No error was expected for transaction rollback, but: ", err)
	}
	if c.tx.active {
		t.Error("Transaction status should be inactive after rollback")
	}
	if len(c.tx.cmds) != 0 {
		t.Error("Transaction command should be cleaned after rollback")
	}
}

func TestHGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := dmock.NewMockDriver(ctrl)
	c := newCacheImpl(Driver(d))

	c.hsets["hash"] = map[string]string{
		"test1": "tt",
		"test2": flagValueNil,
	}

	d.EXPECT().HGetAll("hash").Return(map[string]string{"test1": "1", "test2": "good", "test3": "OK"}, nil)

	res, err := c.HGetAll("hash")

	if err != nil {
		t.Error("No error was expected, but: ", err)
	}
	if res["test1"] != "tt" || res["test2"] != "" || res["test3"] != "OK" {
		t.Error("Result was incorrect")
	}
}
