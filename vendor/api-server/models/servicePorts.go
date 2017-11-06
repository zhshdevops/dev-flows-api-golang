/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-27  @author mengyuan
 */

package models

import (
	"time"

	"fmt"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

const SvcAddressTableName = "tenx_service_ports"

type SvcAddress struct {
	Protocol string
	Host     string
	Port     int
	PortName string
	GroupID  string
}

// ServicePorts is the structrue of table service_ports
type ServicePorts struct {
	ID          uint64    `orm:"pk;column(id)"`
	ClusterID   string    `orm:"size(45);column(cluster_id)"`
	Namespace   string    `orm:"size(255);column(namespace)"`
	AppName     string    `orm:"size(63);column(app_name)"`
	ServiceName string    `orm:"size(253);column(service_name)"`
	Host        string    `orm:"size(50);column(host)"`
	Port        int       `orm:"size(5);column(port)"`
	LBGroup     string    `orm:"size(20);column(lbgroup)"`
	CreateTime  time.Time `orm:"type(datetime);column(creation_time)"`
}

// MaximumPortNumber maximum port number for random
var MaximumPortNumber = 65536

// MinimumPortNumber minimum port number for random
var MinimumPortNumber = 10000

// TableName return table name
func (t *ServicePorts) TableName() string {
	return SvcAddressTableName
}

func NewServicePorts() *ServicePorts {
	return &ServicePorts{}
}

// InsertMultiSvcPort insert multi port info
func InsertMultiSvcPort(clusterID, namespace, appName string, addressMap map[string][]SvcAddress) error {
	method := "InsertMultiSvcPort"
	if len(addressMap) == 0 {
		glog.Infoln(method, "addressMap len is 0. no need to record address")
		return nil
	}
	now := time.Now()
	o := orm.NewOrm()
	portInfos := make([]ServicePorts, 0, 1)
	info := ServicePorts{
		ClusterID:  clusterID,
		Namespace:  namespace,
		AppName:    appName,
		CreateTime: now,
	}
	for svcName, addrList := range addressMap {
		for _, addr := range addrList {
			info.ServiceName = svcName
			info.Host = addr.Host
			info.Port = addr.Port
			info.LBGroup = addr.GroupID
			portInfos = append(portInfos, info)
		}
	}
	if _, err := o.InsertMulti(len(portInfos), portInfos); err != nil {
		return err
	}
	return nil
}

//update tenx_service_ports set lbgroup="ddd"

func ChangeMultiSvcPort(namespace, oldlb string, addressMap map[string][]SvcAddress) error {
	method := "InsertMultiSvcPort"
	if len(addressMap) == 0 {
		glog.Infoln(method, "addressMap len is 0. no need to record address")
		return nil
	}

	for svcName, addrList := range addressMap {
		for _, addr := range addrList {
			sql := fmt.Sprintf("update %s set lbgroup=%q where namespace=%q and lbgroup=%q and service_name=%q and port=%d  and host=%q",
				SvcAddressTableName,addr.GroupID, namespace,oldlb,svcName,addr.Port,addr.Host)
			//glog.Errorf("sql : %q", sql)
			_, err := orm.NewOrm().Raw(sql).Exec()
			if err != nil {
				glog.Errorf("delete failed: %v", err)
				return err
			}

		}
	}

	return nil
}

func DeleteSvcAddressByAddr(namespace string, addrs []SvcAddress) error {
	if len(addrs) < 1 {
		return nil
	}
	rawSQL := fmt.Sprintf("delete from %s where namespace=%q and ", SvcAddressTableName, namespace)
	addrSQL := func() string {
		var result string
		conditions := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			cond := fmt.Sprintf("(host=%q and port=%d)", addr.Host, addr.Port)
			conditions = append(conditions, cond)
		}
		result = strings.Join(conditions, " or ")
		if len(conditions) > 1 {
			result = "(" + result + ")"
		}
		return result
	}()
	sql := rawSQL + addrSQL + ";"
	_, err := orm.NewOrm().Raw(sql).Exec()
	if err != nil {
		glog.Errorf("delete failed: %v", err)
		return err
	}
	return nil
}

// DeleteSvcAddressByAppName delete address by app name
func DeleteSvcAddressByAppName(clusterID, namespace string, appNames []string) error {
	method := "DeleteByAppName"
	o := orm.NewOrm()
	_, err := o.QueryTable(SvcAddressTableName).
		Filter("cluster_id", clusterID).
		Filter("namespace", namespace).
		Filter("app_name__in", appNames).
		Delete()
	if err != nil {
		glog.Errorln(method, "delete failed.", err)
	}
	return err
}

// DeleteSvcAddressByServiceName delete address by service name
func DeleteSvcAddressByServiceName(clusterID, namespace string, svcNames []string) error {
	method := "DeleteByServiceName"
	o := orm.NewOrm()
	_, err := o.QueryTable(SvcAddressTableName).
		Filter("cluster_id", clusterID).
		Filter("namespace", namespace).
		Filter("service_name__in", svcNames).
		Delete()
	if err != nil {
		glog.Errorln(method, "delete failed.", err)
	}
	return err
}

// DeleteSvcAddressByLBGroup delete ports by cluster and lbgroup info
func DeleteSvcAddressByLBGroup(clusterID, lbGroup string) error {
	method := "DeleteSvcAddressByLBGroup"
	o := orm.NewOrm()
	_, err := o.QueryTable(SvcAddressTableName).
		Filter("cluster_id", clusterID).
		Filter("lbgroup", lbGroup).
		Delete()
	if err != nil {
		glog.Errorln(method, "delete failed.", err)
	}
	return err
}

// GetAllPortsByIP get all ports by externalIP
func GetAllPortsByIP(ip, lbGroup string) ([]int, error) {
	method := "GetAllPortsByIp"

	var addressList []ServicePorts
	o := orm.NewOrm()
	ports := make([]int, 0, 1)
	_, err := o.QueryTable(SvcAddressTableName).
		Filter("host", ip).
		Filter("lbgroup", lbGroup).
		Limit(-1).
		All(&addressList, "Port")
	if err != nil {
		glog.Errorln(method, "failed.", err)
		return ports, err
	}
	for _, addr := range addressList {
		ports = append(ports, addr.Port)
	}
	return ports, nil
}

// GetConflictPorts return used ports except used by service itself
func GetConflictPorts(clusterID, externalIP, lbGroup, namespace, serviceName string) ([]int32, error) {
	o := orm.NewOrm()
	usedPorts := make([]int32, 0, 1)
	_, err := o.Raw("SELECT port from "+SvcAddressTableName+" where cluster_id = ? and host = ? and lbgroup = ? and (namespace != ? or service_name != ?)", clusterID, externalIP, lbGroup, namespace, serviceName).QueryRows(&usedPorts)
	return usedPorts, err
}

func (t *ServicePorts) DeleteAllByNamespace(namespaceList []string, orms ...orm.Ormer) (int64, error) {
	if len(namespaceList) == 0 {
		return 0, nil
	}
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	num, err := o.QueryTable(t.TableName()).
		Filter("namespace__in", namespaceList).
		Delete()
	return num, err
}
