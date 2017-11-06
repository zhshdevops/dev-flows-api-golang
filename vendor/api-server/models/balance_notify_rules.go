/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-12-05  @author mengyuan
 */

package models

import (
	"time"

	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

type BalanceNotifyRules struct {
	ID            int       `orm:"pk;column(id)"`
	Namespace     string    `orm:"column(namespace)"`
	NamespaceType int       `orm:"column(namespace_type)"`
	Threshold     int       `orm:"column(threshold)"`
	ReceiverID    int32     `orm:"column(receiver_id)"`
	ReceiverName  string    `orm:"column(receiver_name)"`
	NotifyWay     int       `orm:"column(notify_way)"`
	CreateTime    time.Time `orm:"column(create_time)"`
	ModifyTime    time.Time `orm:"column(modify_time)"`
}

func NewBalanceNofifyRules() *BalanceNotifyRules {
	return &BalanceNotifyRules{}
}

func (t *BalanceNotifyRules) TableName() string {
	return "tenx_balance_notify_rules"
}

// SetRule add or update one rule
func (t *BalanceNotifyRules) SetRule() error {
	method := "SetRule"
	o := orm.NewOrm()
	// insert ignore, if
	sql := `INSERT IGNORE INTO %s(namespace,namespace_type,threshold,receiver_id,receiver_name,notify_way,create_time,modify_time)
	VALUES(?,?,?,?,?,?,?,?)`
	sql = fmt.Sprintf(sql, t.TableName())
	ret, err := o.Raw(sql, t.Namespace, t.NamespaceType, t.Threshold, t.ReceiverID, t.ReceiverName, t.NotifyWay, t.CreateTime, t.ModifyTime).Exec()
	if err != nil {
		glog.Errorln(method, "insert ignore failed.", err)
		return err
	}
	affected, err := ret.RowsAffected()
	if err != nil {
		glog.Errorln(method, "get affected rows number failed.", err)
		return err
	}
	// insert success
	if affected == 1 {
		return nil
	}
	// already have record, update it

	values := orm.Params{
		"threshold":   t.Threshold,
		"notify_way":  t.NotifyWay,
		"modify_time": t.ModifyTime,
	}
	_, err = o.QueryTable(t.TableName()).
		Filter("namespace", t.Namespace).
		Filter("receiver_id", t.ReceiverID).
		Filter("namespace_type", t.NamespaceType).
		Update(values)
	if err != nil {
		glog.Errorln(method, "update rules failed.", err)
		return err
	}

	return nil
}
func (t *BalanceNotifyRules) GetRule(space string, receiverID int32, spaceType int) error {
	return orm.NewOrm().QueryTable(t.TableName()).
		Filter("namespace", space).
		Filter("receiver_id", receiverID).
		Filter("namespace_type", spaceType).
		One(t)
}
