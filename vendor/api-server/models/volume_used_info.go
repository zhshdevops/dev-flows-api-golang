/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-10-25  @author mengyuan
 */

package models

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

// VolumeUsedInfo is the structrue of table volume_used_info
type VolumeUsedInfo struct {
	ID          uint64    `orm:"pk;column(id)"`
	Namespace   string    `orm:"size(255);column(namespace)"`
	Cluster     string    `orm:"size(45);column(cluster)"`
	VolumeName  string    `orm:"size(100);column(volume_name)"`
	AppName     string    `orm:"size(63);column(app_name)"`
	MountPath   string    `orm:"column(mount_path)"`
	ServiceName string    `orm:"size(255);column(service_name)"`
	CreateTime  time.Time `orm:"type(datetime);column(creation_time)"`
}

// TableName return table name
func (t *VolumeUsedInfo) TableName() string {
	return "tenx_volume_used_info"
}

func (t *VolumeUsedInfo) RecordVolumeUsedInfo(clusterID, namespace, appName string, volumeNames map[string][]string, mountInfo map[string]string) error {
	method := "RecordVolumeUsedInfo"
	if len(volumeNames) < 1 {
		return nil
	}
	infos := make([]VolumeUsedInfo, 0, 1)
	info := VolumeUsedInfo{
		Namespace:  namespace,
		Cluster:    clusterID,
		CreateTime: time.Now(),
	}
	for deploymentName, volumeNameList := range volumeNames {
		for _, volumeName := range volumeNameList {
			info.VolumeName = volumeName
			info.MountPath = mountInfo[volumeName]
			info.AppName = appName
			info.ServiceName = deploymentName
			infos = append(infos, info)
		}
	}

	o := orm.NewOrm()

	num, err := o.InsertMulti(len(infos), infos)
	if err != nil {
		glog.Errorln(method, "Insert failed.", err)
	} else if num != int64(len(infos)) {
		glog.Errorf("%s insert failed. insert %v, success %v\n", method, len(infos), num)
	}
	return err
}

func (t *VolumeUsedInfo) DeleteRecord(clusterID, namespace string, appNames, serviceNames []string) error {
	method := "DeleteRecord"
	o := orm.NewOrm()
	qs := o.QueryTable(t.TableName()).
		Filter("cluster", clusterID).
		Filter("namespace", namespace)
	if appNames != nil {
		qs = qs.Filter("app_name__in", appNames)
	}
	if serviceNames != nil {
		qs = qs.Filter("service_name__in", serviceNames)
	}
	_, err := qs.Delete()
	if err != nil {
		glog.Errorln(method, "failed.", err)
	}
	return err
}

func (t *VolumeUsedInfo) Get(clusterID, namespace, volumeName string) (VolumeUsedInfo, error) {
	o := orm.NewOrm()
	volume := VolumeUsedInfo{
		Namespace:  namespace,
		Cluster:    clusterID,
		VolumeName: volumeName,
	}
	err := o.Read(&volume, "namespace", "cluster", "volume_name")
	return volume, err
}

func (t *VolumeUsedInfo) List(clusterID, namespace string, fields ...string) ([]VolumeUsedInfo, error) {
	o := orm.NewOrm()
	volumes := make([]VolumeUsedInfo, 0, 1)
	_, err := o.QueryTable(t.TableName()).
		Filter("cluster", clusterID).
		Filter("namespace", namespace).
		All(&volumes, fields...)
	return volumes, err
}

func (t *VolumeUsedInfo) IsUsing(clusterID, namespace string, volumeNames []string) (bool, error) {
	o := orm.NewOrm()

	cnt, err := o.QueryTable(t.TableName()).
		Filter("cluster", clusterID).
		Filter("namespace", namespace).
		Filter("volume_name__in", volumeNames).
		Count()
	return cnt > 0, err
}
