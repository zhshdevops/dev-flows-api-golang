/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-16  @author mengyuan
 */

package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

// CiFlows is the structure of table tenx_ci_flows
type CiFlows struct {
	FlowID             string    `orm:"pk;column(flow_id)"`
	Name               string    `orm:"column(name)"`
	Owner              string    `orm:"column(owner)"`
	Namespace          string    `orm:"column(namespace)"`
	InitType           int       `orm:"column(init_type)"`
	NotificationConfig string    `orm:"column(notification_config)"`
	CreateTime         time.Time `orm:"type(datetime);column(create_time)"`
	UpdateTime         time.Time `orm:"type(datetime);column(update_time)"`
}

func (t *CiFlows) TableName() string {
	return "tenx_ci_flows"
}

func NewCiFlows() *CiFlows {
	return &CiFlows{}
}

func (t *CiFlows) GetCountByNamespace(namespaceList []string) (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable(t.TableName()).
		Filter("namespace__in", namespaceList).
		Count()
}
