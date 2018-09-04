package cache

// Transaction cache transaction interface
type Transaction interface {
	// Commit the transaction
	Commit() error

	// Rollback the transaction
	Rollback() error
}

// TransSupport interface to support transaction
type TransSupport interface {
	// BeforeCreate called before transaction creation
	BeforeCreate() error

	// AfterCreate called after transaction creation
	AfterCreate() error

	// BeforeCommit called before transaction commit
	BeforeCommit() error

	// AfterCommit called after transaction commit
	AfterCommit() error

	// BeforeRollback called before transaction rollback
	BeforeRollback() error

	// AfterRollback called after transaction rollback
	AfterRollback() error
}

const (
	typeSet    = 1
	typeDel    = 2
	typeExpire = 3
	typeIncr   = 4
	typeDecr   = 5
	typeMSet   = 6
	typeHSet   = 7
	typeHMSet  = 8
	typeHDel   = 9
	typeHIncr  = 10
	typeHDecr  = 11
)

type command struct {
	t    int
	args []interface{}
}

type transImpl struct {
	active bool

	c *cacheImpl

	cmds []*command
}

func newTransImpl(c *cacheImpl) *transImpl {
	ts, ok := c.options.Driver.(TransSupport)
	if ok {
		ts.BeforeCreate()
	}
	tx := &transImpl{
		active: true,
		c:      c,
		cmds:   []*command{},
	}

	if ok {
		ts.AfterCreate()
	}

	return tx
}

// Commit transaction
func (t *transImpl) Commit() error {
	d := t.c.options.Driver
	ts, ok := d.(TransSupport)
	if ok {
		ts.BeforeCommit()
	}
	if t.cmds != nil {

		for _, cmd := range t.cmds {
			var err error
			switch cmd.t {
			case typeSet:
				err = d.Set(cmd.args[0].(string), cmd.args[1])
			case typeDel:
				err = d.Del(cmd.args[0].(string))
			case typeExpire:
				err = d.Expire(cmd.args[0].(string), cmd.args[1].(int64))
			case typeMSet:
				err = d.MSet(cmd.args[0].(map[string]interface{}))
			case typeHSet:
				err = d.HSet(cmd.args[0].(string), cmd.args[1].(string), cmd.args[2])
			case typeHMSet:
				err = d.HMSet(cmd.args[0].(string), cmd.args[1].(map[string]interface{}))
			case typeHDel:
				err = d.HDel(cmd.args[0].(string), cmd.args[1].(string))
			}
			if err != nil {
				// TODO
			}
		}
	}
	if ok {
		ts.AfterCommit()
	}
	t.active = false
	t.cmds = []*command{}
	return nil
}

// Rollback transaction
func (t *transImpl) Rollback() error {
	d := t.c.options.Driver
	ts, ok := d.(TransSupport)
	if ok {
		ts.BeforeRollback()
	}
	if t.cmds != nil {
		l := len(t.cmds)
		for i := l - 1; i >= 0; i-- {
			var err error
			cmd := t.cmds[i]
			switch cmd.t {
			case typeIncr:
				_, err = d.Decr(cmd.args[0].(string), cmd.args[1])
			case typeDecr:
				_, err = d.Incr(cmd.args[0].(string), cmd.args[1])
			case typeHIncr:
				_, err = d.HDecr(cmd.args[0].(string), cmd.args[1].(string), cmd.args[2])
			case typeHDecr:
				_, err = d.HIncr(cmd.args[0].(string), cmd.args[1].(string), cmd.args[2])
			}
			if err != nil {
				// TODO
			}
		}
	}
	if ok {
		ts.AfterRollback()
	}
	t.active = false
	t.cmds = []*command{}
	return nil
}

func (t *transImpl) onSet(key string, value interface{}) {
	t.cmds = append(t.cmds, &command{
		t:    typeSet,
		args: []interface{}{key, value},
	})
}

func (t *transImpl) onDel(key string) {
	t.cmds = append(t.cmds, &command{
		t:    typeDel,
		args: []interface{}{key},
	})
}

func (t *transImpl) onExpire(key string, ex int64) {
	t.cmds = append(t.cmds, &command{
		t:    typeExpire,
		args: []interface{}{key, ex},
	})
}

func (t *transImpl) onIncr(key string, delta interface{}) {
	t.cmds = append(t.cmds, &command{
		t:    typeIncr,
		args: []interface{}{key, delta},
	})
}

func (t *transImpl) onDecr(key string, delta interface{}) {
	t.cmds = append(t.cmds, &command{
		t:    typeDecr,
		args: []interface{}{key, delta},
	})
}

func (t *transImpl) onMSet(kvs map[string]interface{}) {
	t.cmds = append(t.cmds, &command{
		t:    typeMSet,
		args: []interface{}{kvs},
	})
}

func (t *transImpl) onHSet(key string, hk string, value interface{}) {
	t.cmds = append(t.cmds, &command{
		t:    typeHSet,
		args: []interface{}{key, hk, value},
	})
}

func (t *transImpl) onHMSet(key string, kvs map[string]interface{}) {
	t.cmds = append(t.cmds, &command{
		t:    typeHMSet,
		args: []interface{}{key, kvs},
	})
}

func (t *transImpl) onHDel(key string, hk string) {
	t.cmds = append(t.cmds, &command{
		t:    typeHDel,
		args: []interface{}{key, hk},
	})
}

func (t *transImpl) onHIncr(key string, hk string, delta interface{}) {
	t.cmds = append(t.cmds, &command{
		t:    typeHIncr,
		args: []interface{}{key, hk, delta},
	})
}

func (t *transImpl) onHDecr(key string, hk string, delta interface{}) {
	t.cmds = append(t.cmds, &command{
		t:    typeHDecr,
		args: []interface{}{key, hk, delta},
	})
}
