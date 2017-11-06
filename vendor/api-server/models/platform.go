/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2017-2-10  @author yangle
 */

package models

import (
	"fmt"

	"github.com/astaxie/beego/orm"

	sqlstatus "api-server/models/sql/status"
)

// Sec is the structure of table tenx_sec
type Platform struct {
	Name  string `orm:"pk;column(name)"`
	Value string `orm:"column(value)"`
}

func NewPlatform() *Platform {
	return &Platform{}
}

// TableName return table name
func (t *Platform) TableName() string {
	return "tenx_platform"
}

//get platform record
// GetByName get user by name
func (t *Platform) Get(name string) (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where name=?;", t.TableName())
	err := o.Raw(sql, name).QueryRow(t)
	return sqlstatus.ParseErrorCode(err)
}

//get platform list
func (t *Platform) List() ([]Platform, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(t.TableName())
	var records []Platform
	_, err := qs.All(&records)
	return records, err
}

//update platform
func (t *Platform) Update() (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Update(t)
	return sqlstatus.ParseErrorCode(err)
}

//insert license
func (t *Platform) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(t)
	return err
}
