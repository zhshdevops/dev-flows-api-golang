/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-14  @author mengyuan
 */
package alert

import (
	"fmt"
	"time"

	sqlutil "dev-flows-api-golang/models/sql/util"

	"dev-flows-api-golang/modules/transaction"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

type Receivers struct {
	Email []struct {
		Addr   string `json:"addr"`
		Desc   string `json:"desc"`
		Status *int   `json:"status,omitempty"`
	} `json:"email"`
}
type NotifyReceiverGroup struct {
	ID         int64     `orm:"pk;column(id)"`
	GroupID    string    `orm:"column(group_id)"`
	Creator    string    `orm:"column(creator)"`
	Name       string    `orm:"column(name)"`
	NameSpace  string    `orm:"column(namespace)"`
	Desc       string    `orm:"column(desc)"`
	Receivers  string    `orm:"column(receivers)"`
	CreateTime time.Time `orm:"column(create_time)"`
	ModifyTime time.Time `orm:"column(modify_time)"`
}

func NewNotifyReceiverGroup() *NotifyReceiverGroup {
	return new(NotifyReceiverGroup)
}
func (t *NotifyReceiverGroup) TableName() string {
	return "tenx_alert_notify_receiver_group"
}

func (t *NotifyReceiverGroup) Get() error {
	return orm.NewOrm().QueryTable(t.TableName()).Filter("group_id", t.GroupID).One(t)
}

func (t *NotifyReceiverGroup) ListByUser(creator, namespace, name string) ([]NotifyReceiverGroup, error) {
	var items []NotifyReceiverGroup
	o := orm.NewOrm()
	qs := o.QueryTable(t.TableName()).Filter("namespace", namespace).Filter("creator", creator)
	if name != "" {
		qs = qs.Filter("name__contains", sqlutil.EscapeUnderlineInLikeStatement(name))
	}
	_, err := qs.All(&items)
	return items, err
}

func (t *NotifyReceiverGroup) BatchDelete(namespace, user string, ids []string) (int64, error) {
	return orm.NewOrm().QueryTable(t.TableName()).
		Filter("group_id__in", ids).
		Filter("namespace", namespace).
		Filter("creator", user).
		Delete()
}

func (t *NotifyReceiverGroup) Create() (exist bool, err error) {
	trans := transaction.New()
	trans.Do(func() {
		exist, err = t.IsAlreadyExist()
		if err != nil {
			glog.Errorf("check record existence failed: %v", err)
			trans.Finish()
		}
		if exist {
			trans.Finish()
		}
	}).Do(func() {
		_, err = orm.NewOrm().Insert(t)
		if err != nil {
			glog.Errorf("create receiver group failed: %v", err)
		}
	}).Done()
	return
}

func (t *NotifyReceiverGroup) Modify() (exist bool, err error) {
	n, err := orm.NewOrm().QueryTable(t.TableName()).
		Filter("group_id", t.GroupID).
		Update(orm.Params{
			"desc":        t.Desc,
			"modify_time": t.ModifyTime,
			"receivers":   t.Receivers,
		})
	return n == 1, err
}

func (t *NotifyReceiverGroup) IsAlreadyExist() (bool, error) {
	var count int64
	sql := fmt.Sprintf("select count(id) from %s where namespace=%q and name=%q;", t.TableName(), t.NameSpace, t.Name)
	err := orm.NewOrm().Raw(sql).QueryRow(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
