/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-09  @author mengyuan
 */

package models

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

// Sec is the structure of table tenx_sec
type Sec struct {
	ID           uint64    `orm:"pk;column(id)"`
	SerialNumber string    `orm:"type(text);column(serial_number)"`
	CreateTime   time.Time `orm:"type(datetime);column(create_time)"`
}

func NewSec() *Sec {
	return &Sec{}
}

// TableName return table name
func (t *Sec) TableName() string {
	return "tenx_sec"
}

func (t *Sec) GetSerialNumber() (string, error) {
	method := "models.Sec.GetSerialNumber"
	o := orm.NewOrm()
	sec := Sec{}
	err := o.QueryTable(t.TableName()).One(&sec, "SerialNumber")
	if err != nil {
		glog.Errorln(method, "failed.", err)
	}
	return sec.SerialNumber, err
}

//get license list
func (t *Sec) List() ([]Sec, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(t.TableName())
	var records []Sec
	_, err := qs.All(&records)
	return records, err
}
