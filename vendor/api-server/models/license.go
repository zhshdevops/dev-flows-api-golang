/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2017-2-10  @author yangle
 */

package models

import (
	sqlstatus "api-server/models/sql/status"
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

// Sec is the structure of table tenx_sec
type License struct {
	ID           uint64    `orm:"pk;column(id)"`
	LicenseUID   string    `orm:"column(license_uid)"`
	SerialNumber string    `orm:"type(text);column(serial_number)"`
	CreateTime   time.Time `orm:"type(timestamp);column(create_time)"`
	Owner        int32     `orm:"type(int);column(owner_id)"`
}

func NewLicense() *License {
	return &License{}
}

// TableName return table name
func (t *License) TableName() string {
	return "tenx_licenses"
}

func (t *License) GetSerialNumber() (string, error) {
	method := "models.Sec.GetSerialNumber"
	o := orm.NewOrm()
	License := License{}
	err := o.QueryTable(t.TableName()).One(&License, "SerialNumber")
	if err != nil {
		glog.Errorln(method, "failed.", err)
	}
	return License.SerialNumber, err
}

//get license record
// GetByName get user by name
func (t *License) GetByUID(uid string) (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where license_uid=?;", t.TableName())
	err := o.Raw(sql, uid).QueryRow(t)
	return sqlstatus.ParseErrorCode(err)
}

//get license list
func (t *License) List() ([]License, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(t.TableName())
	var records []License
	_, err := qs.All(&records)
	return records, err
}

//insert license
func (t *License) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(t)
	return err
}
