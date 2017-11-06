/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-22  @author liuyang
 */

package integration

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

func init() {
}

const (
	RecordTypeVSphere string = "vsphere"
	RecordTypeCeph    string = "ceph"
)

// Record integration db record
type Record struct {
	ID         string    `orm:"pk;column(id)"`
	Name       string    `orm:"column(name)"`
	Type       string    `orm:"column(type)"`
	Username   string    `orm:"column(username)"`
	Namespace  string    `orm:"column(namespace)"`
	Config     string    `orm:"column(config)"`
	CreateTime time.Time `orm:"column(create_time)"`
}

// TableName return tenx_integration
func (t *Record) TableName() string {
	return "tenx_integration"
}

// List list integration items by namespace
func (t *Record) List() ([]Record, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(t.TableName())
	var records []Record
	_, err := qs.All(&records)
	return records, err
}

// Insert insert record to db
func (t *Record) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(t)
	return err
}

// Update update record to db
func (t *Record) Update() error {
	o := orm.NewOrm()
	_, err := o.Update(t)
	return err
}

// Delete by ID and other parameters
func (t *Record) Delete() error {
	o := orm.NewOrm()

	if t.ID == "" {
		return nil
	}

	_, err := o.Delete(t)
	return err
}

// Get get record by id
func (t *Record) Get() error {
	o := orm.NewOrm()

	err := o.QueryTable(t.TableName()).Filter("id", t.ID).One(t)
	return err
}

// Status status of currently intergrated items
type Status struct {
	VSphere []*VSphere `json:"vsphere"`
	Ceph    []*Ceph    `json:"ceph"`
}

// GetStatus get integration status
func GetStatus() (*Status, error) {
	method := "GetStatus"
	var table Record
	records, err := table.List()
	if err != nil {
		return nil, err
	}

	s := &Status{}

	for _, r := range records {
		if r.Type == RecordTypeVSphere {
			v, err := NewVSphere(r.Config)
			if err != nil {
				glog.Errorln(method, "failed to get vsphere config", err)
				continue
			}
			v.ID = r.ID
			// v.Password = ""
			s.VSphere = append(s.VSphere, v)
		} else if r.Type == RecordTypeCeph {
			// add parse ceph config
		}
	}

	return s, nil
}
