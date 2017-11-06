/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2017-04-27  @author houxinzhu
 */

package models

import (
	"time"

	"encoding/json"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

type Snapshot struct {
	ID           string    `orm:"pk;column(id)",json:"id"`
	Cluster      string    `orm:"size(45);column(cluster)",json:"cluster"`
	Namespace    string    `orm:"size(255);column(namespace)",json:"namespace"`
	Name         string    `orm:"size(100);column(name)",json:"name"`
	UserID       uint32    `orm:"column(user_id)",json:"userId"`
	Driver       string    `orm:"size(255);column(driver)",json:"driver"`
	DriverDetail string    `orm:"column(driver_detail)",json:"driverDetail"`
	Volume       string    `orm:"size(45);column(volume)",json:"volume"`
	CreateTime   time.Time `orm:"type(datetime);column(creation_time)",json:"createTime"`
}

type SnapDriverDetail struct {
	Size   int    `json:"size"` // MB
	FsType string `json:"fsType"`
}

type RespSnapshot struct {
	FsType string `json:"fstype"`
	Size   int    `json:"size"`
	Snapshot
}

func (s *Snapshot) TableName() string {
	return "tenx_snapshot"
}

func (s *Snapshot) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(s)
	if err != nil {
		glog.Errorln("Insert snapshot table failed.", err)
	}
	return err
}

func (s *Snapshot) GetSnapInfoByName(clusterID, namespace, volumeName, SnapshotName string) error {
	o := orm.NewOrm()
	if err := o.QueryTable(s.TableName()).
		Filter("cluster", clusterID).
		Filter("namespace", namespace).
		Filter("volume", volumeName).
		Filter("name", SnapshotName).
		One(s); err != nil {
		return err
	}
	return nil

}

func (s *Snapshot) GetSnapDetailByName(clusterID, namespace, volumeName, SnapshotName string) (detail *SnapDriverDetail, err error) {
	if err := s.GetSnapInfoByName(clusterID, namespace, volumeName, SnapshotName); err != nil {
		return nil, err
	}
	detail = &SnapDriverDetail{}
	if err := json.Unmarshal([]byte(s.DriverDetail), detail); err != nil {
		return nil, err
	}
	return detail, nil

}
func (t *Snapshot) Delete(o orm.Ormer, clusterID, namespace string, volumeName string, snapshot string) error {
	_, err := o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("cluster", clusterID).
		Filter("volume", volumeName).
		Filter("name", snapshot).
		Delete()
	return err
}

func (t *Snapshot) DeleteMulti(clusterID, namespace string, volumeName string, snapshots []string) error {
	o := orm.NewOrm()
	method := "model.snapshot.DeleteMulti"
	num, err := o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("cluster", clusterID).
		Filter("volume", volumeName).
		Filter("name__in", snapshots).
		Delete()
	if num != int64(len(snapshots)) {
		glog.Errorf("%s volume count:%v delete count %v\n", method, len(snapshots), num)
	}
	return err
}

func (t *Snapshot) List(clusterID, namespace string, fields ...string) ([]Snapshot, error) {
	o := orm.NewOrm()
	snapshot := make([]Snapshot, 0, 1)
	_, err := o.QueryTable(t.TableName()).
		Filter("cluster", clusterID).
		Filter("namespace", namespace).OrderBy("-creation_time").
		All(&snapshot, fields...)
	return snapshot, err
}

func (t *Snapshot) CheckSnapExists(clusterID, namespace, VolumeName string, SnapshotName []string) bool {
	o := orm.NewOrm()
	return o.QueryTable(t.TableName()).
		Filter("cluster", clusterID).
		Filter("volume", VolumeName).
		Filter("namespace", namespace).
		Filter("name__in", SnapshotName).
		Exist()
}

func (t *Snapshot) DeleteMultiByVolume(clusterID, namespace, volume string) error {
	method := "DeleteMultiByVolume"
	o := orm.NewOrm()
	snapshots, err := t.ListSnapshotByVolume(clusterID, namespace, volume) //list all snapshots
	if err != nil {
		return err
	}
	num, err := o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("cluster", clusterID).
		Filter("volume__exact", volume).
		Delete()
	if num != int64(len(snapshots)) {
		glog.Errorf("%s snapshot count:%v delete count %v\n", method, len(snapshots), num)
	}
	return err
}

func (t *Snapshot) ListSnapshotByVolume(clusterID, namespace string, volume string) ([]Snapshot, error) {
	snapshot := []Snapshot{}
	o := orm.NewOrm()
	_, err := o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("cluster", clusterID).
		Filter("volume__exact", volume).
		All(&snapshot, "name")
	return snapshot, err
}

// LockRow lock snapshot
func (t *Snapshot) LockRow(o orm.Ormer, clusterID, namespace, volume string, snapshot string) (string, error) {
	name := ""
	err := o.Raw("select name from "+t.TableName()+" where cluster = ? and namespace = ? and volume = ? and name = ? for update", clusterID, namespace, volume, snapshot).
		QueryRow(&name)
	if err == orm.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return name, nil
}
