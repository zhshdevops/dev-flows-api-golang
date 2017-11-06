/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-10-20  @author mengyuan
 */

package models

import (
	"encoding/json"
	"errors"
	"time"

	sqlutil "api-server/models/sql/util"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

// Volume is the structrue of table volume
type Volume struct {
	ID           string    `orm:"pk;column(id)"`
	Cluster      string    `orm:"size(45);column(cluster)"`
	Namespace    string    `orm:"size(255);column(namespace)"`
	Name         string    `orm:"size(100);column(name)"`
	UserID       uint32    `orm:"column(user_id)"`
	Driver       string    `orm:"size(255);column(driver)"`
	DriverDetail string    `orm:"column(driver_detail)"`
	CreateTime   time.Time `orm:"type(datetime);column(creation_time)"`
}

// TableName return table name
func (t *Volume) TableName() string {
	return "tenx_volume"
}

func NewVolume() *Volume {
	return &Volume{}
}

// LockRow lock volume
func (t *Volume) LockRow(o orm.Ormer, clusterID, namespace, volumeName string) (string, error) {
	name := ""
	err := o.Raw("select name from "+t.TableName()+" where cluster = ? and namespace = ? and name = ? for update", clusterID, namespace, volumeName).
		QueryRow(&name)
	if err == orm.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return name, nil
}

func (t *Volume) Insert(orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	_, err := o.Insert(t)
	if err != nil {
		glog.Errorln("Insert volume table failed.", err)
	}
	return err
}

func (t *Volume) DeleteMulti(clusterID, namespace string, volumes []string) error {
	method := "DeleteMulti"
	o := orm.NewOrm()
	num, err := o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("cluster", clusterID).
		Filter("name__in", volumes).
		Delete()
	if num != int64(len(volumes)) {
		glog.Errorf("%s volume count:%v delete count %v\n", method, len(volumes), num)
	}
	return err
}

func (t *Volume) UpdateRBDDetail(clusterID, namespace, volumeName, newType string, newSize int) error {
	method := "UpdateRBDDetail"

	detail, oldType, oldSize, err := t.GetRBDDriverDetail(clusterID, namespace, volumeName)
	if err != nil {
		return nil
	}

	needUpdate := false
	// fs type changed, update it
	if newType != "" && newType != oldType {
		glog.V(2).Infoln(method, "fsType changed")
		detail["fsType"] = newType
		needUpdate = true
	}
	// size changed, update it
	if newSize > 0 && newSize != oldSize {
		glog.V(2).Infoln(method, "size changed")
		detail["size"] = newSize
		needUpdate = true
	}
	if !needUpdate {
		return nil
	}
	detailStr, err := json.Marshal(detail)
	if err != nil {
		glog.Errorln(method, "marshal driver detail failed.", err)
		return err
	}

	o := orm.NewOrm()
	_, err = o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("cluster", clusterID).
		Filter("name", volumeName).
		Filter("driver", "rbd").
		Update(orm.Params{"driver_detail": detailStr})
	if err != nil {
		glog.Errorln(method, "update driver detail failed.", err)
		return err
	}
	return nil
}

func (t *Volume) GetRBDDriverDetail(clusterID, namespace, volumeName string) (detail map[string]interface{}, fsType string, size int, err error) {
	method := "GetRBDDriverDetail"
	volume := Volume{}
	o := orm.NewOrm()
	err = o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("cluster", clusterID).
		Filter("name", volumeName).
		Filter("driver", "rbd").
		One(&volume, "DriverDetail")
	if err != nil {
		glog.Errorln(method, "query volume failed.", err)
		return
	}
	detail = make(map[string]interface{})
	if err = json.Unmarshal([]byte(volume.DriverDetail), &detail); err != nil {
		glog.Errorln(method, "unmarshal driver_detail failed.", err)
		return
	}

	typeTmp, ok := detail["fsType"]
	if !ok {
		glog.Errorln(method, "driver_detail no fsType field")
		return nil, "", 0, errors.New("driver_detail no fsType field")
	}
	if fsType, ok = typeTmp.(string); !ok {
		glog.Errorln(method, "fsType invalid")
		return nil, "", 0, errors.New("fsType invalid")
	}
	sizeTmp, ok := detail["size"]
	if !ok {
		glog.Errorln(method, "driver_detail no size field")
		return nil, "", 0, errors.New("driver_detail no size field")
	}
	if _, ok = sizeTmp.(float64); !ok {
		glog.Errorln(method, "size invalid")
		return nil, "", 0, errors.New("size invalid")
	}
	size = int(sizeTmp.(float64))
	return
}

func (t *Volume) ListWithUseingInfo(clusterID, namespace, appName, volumeName string) ([]orm.Params, error) {
	method := "models.Volume.List"

	o := orm.NewOrm()
	sql := `SELECT t1.cluster AS cluster,
		t1.name AS name,
		t1.driver AS diskType,
		t1.driver_detail AS detail,
		t1.creation_time AS createTime,
		t2.id AS usedInfoID,
		t2.app_name AS appName,
		t2.service_name as serviceName,
		t2.mount_path AS mountPoint
		FROM tenx_volume AS t1 LEFT JOIN tenx_volume_used_info AS t2
		ON t1.cluster = t2.cluster AND t1.namespace = t2.namespace AND t1.name = t2.volume_name
		WHERE t1.cluster = ? AND t1.namespace = ?`
	var r orm.RawSeter
	// if appName is not empty, add it to WHERE clause
	if appName != "" && volumeName != "" {
		sql += " AND t2.app_name = ? AND t1.name = ?"
		r = o.Raw(sql, clusterID, namespace, appName, volumeName)
	} else if appName != "" && volumeName == "" {
		sql += " AND t2.app_name = ?"
		r = o.Raw(sql, clusterID, namespace, appName)
	} else if appName == "" && volumeName != "" {
		sql += " AND t1.name like ?"
		r = o.Raw(sql, clusterID, namespace, "%"+sqlutil.EscapeUnderlineInLikeStatement(volumeName)+"%")
	} else {
		r = o.Raw(sql, clusterID, namespace)
	}
	results := make([]orm.Params, 0, 1)
	_, err := r.Values(&results)
	if err != nil {
		glog.Errorln(method, "query db failed.", err)
	}
	// if no data selected, r.Values(&results) will change results to nil
	if results == nil {
		results = make([]orm.Params, 0, 1)
	}
	return results, err
}

func (t *Volume) List(clusterID, namespace string, fields ...string) ([]Volume, error) {
	o := orm.NewOrm()
	volumes := make([]Volume, 0, 1)
	_, err := o.QueryTable(t.TableName()).
		Filter("cluster", clusterID).
		Filter("namespace", namespace).
		All(&volumes, fields...)
	return volumes, err
}

// GetCountByNamespace get the count of template belonging namespaceList
func (t *Volume) GetCountByNamespace(namespaceList []string) (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable(t.TableName()).
		Filter("namespace__in", namespaceList).
		Count()
}

//ListVolume list cluster all volume
func (t *Volume) ListVolume(clusterID string, fields ...string) ([]Volume, error) {
	o := orm.NewOrm()
	volumes := make([]Volume, 0)
	_, err := o.QueryTable(t.TableName()).Filter("cluster", clusterID).All(&volumes, fields...)
	return volumes, err
}
func (t *Volume) CheckVolumeExists(clusterID, namespace, volumeName string) bool {
	o := orm.NewOrm()
	return o.QueryTable(t.TableName()).
		Filter("cluster", clusterID).
		Filter("namespace", namespace).
		Filter("name", volumeName).
		Exist()
}

func (t *Volume) GetOverviewInfo(cluster, namespace string) (usedSize, totalCnt, usedCnt int, err error) {
	method := "GetOverviewInfo"
	result, err := t.ListWithUseingInfo(cluster, namespace, "", "")
	if err != nil {
		return
	}

	for _, item := range result {
		totalCnt++
		if item["usedInfoID"] != nil {
			usedCnt++
		}
		detail := make(map[string]interface{})
		if err = json.Unmarshal([]byte(item["detail"].(string)), &detail); err != nil {
			glog.Errorln(method, "unmarshal driver_detail failed.", err)
			return
		}
		sizeTmp, ok := detail["size"]
		if !ok {
			err = errors.New("driver_detail no size field")
			glog.Errorln(method, err.Error())
			return
		}
		if _, ok = sizeTmp.(float64); !ok {
			err = errors.New("size invalid")
			glog.Errorln(method, err.Error())
			return
		}
		usedSize += int(sizeTmp.(float64))
	}
	return
}
