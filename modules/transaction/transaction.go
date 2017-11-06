package transaction

import (
"github.com/astaxie/beego/orm"
"github.com/golang/glog"
)

type transaction struct {
	fail     bool
	rollback bool
	finish   bool
	o        orm.Ormer
}

// New create a transaction and start
func New() transaction {
	t := transaction{
		o: orm.NewOrm(),
	}
	err := t.o.Begin()
	if err != nil && err != orm.ErrTxHasBegan {
		glog.Errorln("trasaction start failed.", err)
		t.fail = true
	}
	return t
}

// Do call function f if not rollback and not finish
func (t *transaction) Do(f func()) *transaction {
	if t.fail || t.rollback || t.finish {
		return t
	}
	f()
	return t
}

// Done rollback or commit this transaction
func (t *transaction) Done() {
	// if transaction start failed, no need to rollback or commit
	if t.fail {
		return
	}
	if t.rollback {
		if err := t.O().Rollback(); err != nil {
			glog.Errorln("trasaction rollback failed.", err)
			t.fail = true
		}
	} else {
		if err := t.O().Commit(); err != nil {
			glog.Errorln("trasaction commit failed.", err)
			t.fail = true
		}
	}
}

// O return the inner Ormer
func (t *transaction) O() orm.Ormer {
	return t.o
}

// Rollback set rollback flag.
// If set true subsequent Do function not execute any more
// and in Done function this transaction will be rollback
func (t *transaction) Rollback(logs ...interface{}) {
	t.rollback = true
	if len(logs) > 0 {
		glog.Errorln(logs...)
	}
}

// IsCommit check if this transaction have been commit successful
func (t *transaction) IsCommit() bool {
	return !t.rollback && !t.fail
}

// Finish set finish flag.
// If set true subsequent Do function not execute any more
// and in Done function this transaction will be commit
func (t *transaction) Finish(logs ...interface{}) {
	t.finish = true
	if len(logs) > 0 {
		glog.Infoln(logs...)
	}
}
