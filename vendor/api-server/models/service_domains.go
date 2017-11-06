/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-10-13  @author liuyang
 */

package models

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
	"github.com/pborman/uuid"

	sqlstatus "api-server/models/sql/status"
)

// CREATE TABLE `service_domains` (
//   `domain_id` varchar(36) NOT NULL COMMENT 'service 绑定域名',
//   `domain_name` varchar(253) NOT NULL,
//   `domain_port` int(11) NOT NULL,
//   `service_name` varchar(256) NOT NULL,
//   `cluster_id` varchar(36) NOT NULL,
//   PRIMARY KEY (`domain_id`),
//   UNIQUE KEY `domain_id_UNIQUE` (`domain_id`),
//   UNIQUE KEY `unique_domain_name` (`domain_name`,`cluster_id`)
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8

// ServiceDomainTable service_domains
type ServiceDomainTable struct {
	DomainID    string `orm:"pk;column(domain_id)"`
	DomainName  string `orm:"column(domain_name)"`
	DomainPort  int    `orm:"column(domain_port)"`
	ServiceName string `orm:"column(service_name)"`
	LBGroup     string `orm:"column(lbgroup)"`
	ClusterID   string `orm:"column(cluster_id)"`
}

const (
	BindDomainKey = "binding_domains"
	BindPortKey   = "binding_port"
)

const serviceDomains = "tenx_service_domains"

// TableName return service_domains
func (t *ServiceDomainTable) TableName() string {
	return serviceDomains
}

func init() {
	orm.RegisterModel(new(ServiceDomainTable))
}

// BindDomain add domain record
func (t *ServiceDomainTable) BindDomain(clusterID, lbGroup, serviceName, domainName string, domainPort int) error {
	t.DomainID = uuid.New()
	t.DomainName = domainName
	t.DomainPort = domainPort
	t.ServiceName = serviceName
	t.LBGroup = lbGroup
	t.ClusterID = clusterID

	o := orm.NewOrm()
	_, err := o.Insert(t)
	code, _ := sqlstatus.ParseErrorCode(err)
	if code == sqlstatus.SQLErrDuplicateEntry {
		return fmt.Errorf("domain is already bound")
	}
	return err
}

// UnbindDomain add domain record
func (t *ServiceDomainTable) UnbindDomain(clusterID, serviceName, domainName string, domainPort int) error {
	o := orm.NewOrm()
	_, err := o.QueryTable(t.TableName()).Filter("domain_name", domainName).Filter("domain_port", domainPort).Filter("service_name", serviceName).Filter("cluster_id", clusterID).Delete()
	return err
}

// DeleteBindingDomains delete all binding domains of the specified service list
func DeleteBindingDomains(clusterID string, serviceList []string) error {
	method := "DeleteBindingDomains"
	if len(serviceList) == 0 {
		return nil
	}
	o := orm.NewOrm()
	_, err := o.QueryTable(serviceDomains).
		Filter("cluster_id", clusterID).
		Filter("service_name__in", serviceList).
		Delete()
	if err != nil {
		glog.Errorln(method, "failed to delete binding domains", err)
	}
	return err
}
