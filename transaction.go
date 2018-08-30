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

// Commit transaction
func (t *transImpl) Commit() error {
	ts, ok := t.c.options.Driver.(TransSupport)
	if ok {
		ts.BeforeCommit()
	}
	if t.cmds != nil {
		for _, cmd := range t.cmds {
			switch cmd.t {
			case typeSet:
				t.c.options.Driver.Set(cmd.args[0].(string), cmd.args[1].(string))
			case typeDel:
				t.c.options.Driver.Del(cmd.args[0].(string))
			case typeExpire:
				t.c.options.Driver.Expire(cmd.args[0].(string), cmd.args[1].(int64))
			case typeIncr:
				t.c.options.Driver.Incr(cmd.args[0].(string), cmd.args[1].(string))
			}
		}
	}
	if ok {
		ts.AfterCommit()
	}
	t.cmds = nil
	return nil
}

// Rollback transaction
func (t *transImpl) Rollback() error {
	ts, ok := t.c.options.Driver.(TransSupport)
	if ok {
		ts.BeforeRollback()
	}
	if t.cmds != nil {
		l := len(t.cmds)
		for i := l - 1; i >= 0; i-- {
			cmd := t.cmds[i]
			switch cmd.t {
			case typeIncr:
				t.c.options.Driver.Decr(cmd.args[0].(string), cmd.args[1].(string))
			case typeDecr:
				t.c.options.Driver.Incr(cmd.args[0].(string), cmd.args[1].(string))
			}
		}
	}
	if ok {
		ts.AfterRollback()
	}
	t.cmds = nil
	return nil
}

func (t *transImpl) onSet(key string, value string) {
	if t.cmds == nil {
		t.cmds = []*command{}
	}
	t.cmds = append(t.cmds, &command{
		t:    typeSet,
		args: []interface{}{key, value},
	})
}

func (t *transImpl) onDel(key string) {
	if t.cmds == nil {
		t.cmds = []*command{}
	}
	t.cmds = append(t.cmds, &command{
		t:    typeDel,
		args: []interface{}{key},
	})
}

func (t *transImpl) onExpire(key string, ex int64) {
	if t.cmds == nil {
		t.cmds = []*command{}
	}
	t.cmds = append(t.cmds, &command{
		t:    typeExpire,
		args: []interface{}{key, ex},
	})
}

func (t *transImpl) onIncr(key string, delta string) {
	if t.cmds == nil {
		t.cmds = []*command{}
	}
	t.cmds = append(t.cmds, &command{
		t:    typeIncr,
		args: []interface{}{key, delta},
	})
}

func (t *transImpl) onDecr(key string, delta string) {
	if t.cmds == nil {
		t.cmds = []*command{}
	}
	t.cmds = append(t.cmds, &command{
		t:    typeDecr,
		args: []interface{}{key, delta},
	})
}
