/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-27  @author liuyang
 */

package models

import (
	"fmt"
	"sort"
	"time"

	"api-server/models/common"
	sqlstatus "api-server/models/sql/status"

	"github.com/astaxie/beego/orm"
)

// ServiceInfo service status
type ServiceInfo struct {
	Name          string                  `json:"name,omitempty"` // used in array
	CreateTime    time.Time               `json:"create_time"` // sort key
	Address       string                  `json:"address"`
	Quota         string                  `json:"quota"`
	Envs          []ServiceEnvInfo        `json:"envs,omitempty"`
	Ports         []ServicePortInfo       `json:"ports,omitempty"`
	Volumes       []ServiceVolumeInfo     `json:"volumes,omitempty"`
	AutoScale     *AutoScaleInfo          `json:"auto_scale,omitempty"`
	HA            *ServiceHAInfo          `json:"ha,omitempty"`
	Instances     map[string]InstanceInfo `json:"instances,omitempty"`
	InstanceCount uint32                  `json:"instance_count,omitempty"`
}

// ServiceEnvInfo service env info
type ServiceEnvInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ServicePortInfo service port info
type ServicePortInfo struct {
	InstancePort uint `json:"instance_port,omitempty"`
	ServicePort  uint `json:"svc_port,omitempty"`
}

// ServiceVolumeInfo service volume info
type ServiceVolumeInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
	Size uint   `json:"size"`
}

// ServiceHAInfo service ha info
type ServiceHAInfo struct {
	Open    bool              `json:"open"`
	Options []ServiceHAOption `json:"options"`
}

// ServiceHAOption service ha option
type ServiceHAOption struct {
	Type       string `json:"type"`
	Timeout    uint   `json:"timeout"`
	Interval   uint   `json:"interval"`
	FirstDelay uint   `json:"first_delay"`
	Port       uint   `json:"port"`
	Path       string `json:"path"`
}

//ServiceTimeScaleRecord service time scale record
type ServiceTimeScaleRecord struct {
	Id           int64     `json:"id" orm:"pk;column(id)"`
	Name         string    `json:"name",orm:"size(63),column(name)"`
	Namespace    string    `json:"namespace",orm:"size(255),column(namespace)"`
	Spec         string    `json:"spec",orm:"size(255),column(spec)"`
	Desired      int32     `json:"desired",orm:"size(1),column(desired)"`
	ClusterId    string    `json:"clusterID",orm:"size(255),column(cluster_id)"`
	Status       uint      `json:"status",orm:"size(1),column(status)"` //0 off 1 on
	CreationTime time.Time `json:"creationTime" orm:"column(creation_time)" `
}

func (record *ServiceTimeScaleRecord) TableName() string {
	return "tenx_service_time_scale_record"
}

// serviceByCreateTime sort service status by create time
type serviceByCreateTime []*ServiceInfo

// Len implement for sort.Interface
func (t serviceByCreateTime) Len() int {
	return len(t)
}

// Swap implement for sort.Interface
func (t serviceByCreateTime) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Less implement for sort.Interface
func (t serviceByCreateTime) Less(i, j int) bool {
	return t[i].CreateTime.Before(t[j].CreateTime)
}

// SortServiceInfoByCreateTime sort service status by create time
func SortServiceInfoByCreateTime(svcs []*ServiceInfo) {
	sort.Stable(serviceByCreateTime(svcs))
}

func (record *ServiceTimeScaleRecord) Insert() (int64, error) {
	o := orm.NewOrm()
	record.CreationTime = time.Now()
	record.Status = 1
	id, err := o.Insert(record)
	return id, err
}

func (record *ServiceTimeScaleRecord) Update() error {
	o := orm.NewOrm()
	switch common.DType {
	case common.DRMySQL:
		_, err := o.Update(record)
		return err
	case common.DRPostgres:
		sql := `update "tenx_service_time_scale_record" set name=?, namespace=?, spec=?,desired=?, status=?;`
		_, err := o.Raw(sql, record.Name, record.Namespace, record.Spec, record.Desired, record.Status).Exec()
		return err
	}

	return fmt.Errorf("Driver %s not supported", common.DType)
}

func (record *ServiceTimeScaleRecord) Delete() error {
	o := orm.NewOrm()
	// o.Update(md, ...)
	switch common.DType {
	case common.DRMySQL:
		sql := `delete from "tenx_service_time_scale_record" where name=?, namespace=?;`
		_, err := o.Raw(sql, record.Name, record.Namespace).Exec()
		return err
	case common.DRPostgres:
		sql := `delete from "tenx_service_time_scale_record" where name=?, namespace=?;`
		_, err := o.Raw(sql, record.Name, record.Namespace).Exec()
		return err
	}

	return fmt.Errorf("Driver %s not supported", common.DType)
}

func (record *ServiceTimeScaleRecord) Get() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where name=? and namespace=?;", record.TableName())
	err := o.Raw(sql, record.Name, record.Namespace).QueryRow(record)
	return sqlstatus.ParseErrorCode(err)
}

func (record *ServiceTimeScaleRecord) GetByID(id uint64) (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where id=?;", record.TableName())
	err := o.Raw(sql, id).QueryRow(record)
	return sqlstatus.ParseErrorCode(err)
}

func QueryServiceTimeScaleRecords() (records []ServiceTimeScaleRecord, err error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select name,namespace,spec,desired,cluster_id from %v where status=?;", new(ServiceTimeScaleRecord).TableName())
	_, err = o.Raw(sql, 1).QueryRows(&records)
	return records, err
}
