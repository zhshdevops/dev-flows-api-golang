/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-25  @author mengyuan
 */

package models

import (
	"fmt"
	"time"

	"api-server/models/common"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

const (
	VolumeCreate = 1
	VolumeDelete = 2
)
const (
	VolumeOpSuccess = 1
	VolumeOpFailed  = 2
)

// Volume is the structrue of table volume
type VolumeLog struct {
	ID              string    `orm:"pk;column(id)"`
	VolumeID        string    `orm:"column(volume_id)"`
	UserID          uint32    `orm:"column(user_id)"`
	Namespace       string    `orm:"size(255);column(namespace)"`
	Cluster         string    `orm:"size(45);column(cluster)"`
	Name            string    `orm:"size(45);column(name)"`
	Operation       uint8     `orm:"column(operation)"`
	OperationDetail string    `orm:"size(1000);column(operation_detail)"`
	Result          uint8     `orm:"column(result)"`
	CreateTime      time.Time `orm:"type(datetime);column(creation_time)"`
}

func NewVolumeLog() *VolumeLog {
	return &VolumeLog{}
}

func (t *VolumeLog) TableName() string {
	return "tenx_volume_log"
}

// Insert insert one log
func (t *VolumeLog) Insert(orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	_, err := o.Insert(t)
	if err != nil {
		glog.Errorln("Insert app table failed.", err)
	}
	return err
}

// InsertMulti insert multi logs
func (t *VolumeLog) InsertMulti(logs []VolumeLog, orms ...orm.Ormer) (int64, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	num, err := o.InsertMulti(len(logs), logs)
	if err != nil {
		glog.Errorln("Insert volume_log table failed.", err)
	} else if num != int64(len(logs)) {
		glog.Errorf("Insert volume_log table failed. insert %v, success %v\n", len(logs), num)
	}
	return num, err
}

func (t *VolumeLog) GetCountByType(namespaceList, clusterList []string, beginTime, endTime string) (map[string]int, error) {
	method := "GetCountByType"

	qb, err := orm.NewQueryBuilder(common.DType.String())
	if err != nil {
		glog.Errorln(method, "NewQueryBuilder failed.", err)
		return nil, err
	}
	qb.Select("operation, COUNT(*) as cnt").
		From(t.TableName()).
		Where("creation_time > ?").
		And("creation_time <= ?").
		And(fmt.Sprintf("result =  %v", VolumeOpSuccess))

	// copy slice to avoid modify original parameters
	// qb.In() doesnot add quote (sql is like namespace IN (space1, space2), it's wrong
	// so we add qoute ourself
	var spaceListCopy []string
	var clusterListCopy []string
	for _, space := range namespaceList {
		spaceListCopy = append(spaceListCopy, "'"+space+"'")
	}
	for _, cluster := range clusterList {
		clusterListCopy = append(clusterListCopy, "'"+cluster+"'")
	}
	if len(spaceListCopy) > 0 {
		qb.And("namespace").In(spaceListCopy...)
	}
	if len(clusterListCopy) > 0 {
		qb.And("cluster").In(clusterListCopy...)
	}
	qb.GroupBy("operation")
	sql := qb.String()

	o := orm.NewOrm()
	var result []struct {
		Operation int
		Cnt       int
	}
	o.Raw(sql, beginTime, endTime).QueryRows(&result)

	// slice to map
	countMap := make(map[string]int)
	for _, r := range result {
		countMap[NewVolumeLog().convertOpTypeToReadable(r.Operation)] = r.Cnt
	}
	return countMap, nil
}

func (t *VolumeLog) convertOpTypeToReadable(typ int) string {
	opMap := map[int]string{
		VolumeCreate: "volumeCreate",
		VolumeDelete: "volumeDelete",
	}
	op, ok := opMap[typ]
	if !ok {
		return "unknown"
	}
	return op
}
