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
