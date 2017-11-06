/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2017-05-04  @author lizhen
 */

package labels

import (
	"errors"
	"github.com/astaxie/beego/orm"
	"time"
)

//const createTable = `CREATE TABLE IF NOT EXISTS tenx_labels (
//  id INT AUTO_INCREMENT PRIMARY KEY,
//  label VARCHAR(64) NOT NULL COMMENT 'key of label',
//  value VARCHAR(64) NOT NULL COMMENT 'value of label',
//  target VARCHAR(32) NOT NULL COMMENT 'target type, eg. node, pod, service...',
//  cluster_id VARCHAR(48) NOT NULL COMMENT 'tenx_clusters.id',
//  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
//  create_by INT NOT NULL COMMENT 'tenx_users.user_id',
//  CONSTRAINT UNIQUE INDEX USING HASH (label, value, target, cluster_id)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8;`

const tableName = "tenx_labels"

type LabelModel struct {
	ID        int       `orm:"pk;column(id)"`
	Key       string    `orm:"size(64);column(label)"`
	Value     string    `orm:"size(64);column(value)"`
	Target    string    `orm:"size(32);column(target)"`
	ClusterID string    `orm:"size(48);column(cluster_id)"`
	CreateAt  time.Time `orm:"column(create_at)"`
	CreateBy  int       `orm:"column(create_by)"`
}

func (LabelModel) TableName() string {
	return tableName
}

func (LabelModel) TableEngine() string {
	return "InnoDB"
}

type LabelDTO struct {
	ID        *int       `json:"id,omitempty"`
	Key       string     `json:"key"`
	Value     string     `json:"value"`
	Target    string     `json:"target,omitempty"`
	ClusterID string     `json:"clusterID,omitempty"`
	CreateAt  *time.Time `json:"createAt,omitempty"`
	CreateBy  *int       `json:"createBy,omitempty"`
}

func (dto LabelDTO) ToORM() *LabelModel {
	return &LabelModel{
		ID:        *dto.ID,
		Key:       dto.Key,
		Value:     dto.Value,
		Target:    dto.Target,
		ClusterID: dto.ClusterID,
		CreateAt:  *dto.CreateAt,
		CreateBy:  *dto.CreateBy,
	}
}

func (orm LabelModel) ToDTO() *LabelDTO {
	return &LabelDTO{
		ID:        &orm.ID,
		Key:       orm.Key,
		Value:     orm.Value,
		Target:    orm.Target,
		ClusterID: orm.ClusterID,
		CreateAt:  &orm.CreateAt,
		CreateBy:  &orm.CreateBy,
	}
}

func NewLabel(key, value, target, clusterID string, by int) (id int, err error) {
	var id64 int64
	id64, err = orm.NewOrm().Insert(&LabelModel{
		Key:       key,
		Value:     value,
		Target:    target,
		ClusterID: clusterID,
		CreateBy:  by,
	})
	id = int(id64)
	return
}

func NewLabels(withoutIDs []*LabelDTO, by int) (withIDs []*LabelDTO, err error) {
	withIDs = withoutIDs
	for _, withoutID := range withoutIDs {
		var id int
		if id, err = NewLabel(withoutID.Key, withoutID.Value, withoutID.Target, withoutID.ClusterID, by); err != nil {
			return
		}
		withoutID.ID = &id
		now := time.Now()
		withoutID.CreateAt = &now
		withoutID.Target = ""
	}
	return
}

func ModifyByID(id int, key, value string) (err error) {
	_, err = orm.NewOrm().Update(&LabelModel{
		ID:    id,
		Key:   key,
		Value: value,
	}, "label", "value")
	return
}

func DeleteByID(id int) (err error) {
	_, err = orm.NewOrm().Delete(&LabelModel{ID: id})
	return
}

func FindByID(id int) (label *LabelModel, err error) {
	var labels []*LabelModel
	if labels, err = find(map[string]interface{}{"id": id}); err != nil {
		return
	}
	label = labels[0]
	return
}

func FindByUserID(id int) (labels []*LabelModel, err error) {
	labels, err = find(map[string]interface{}{"create_by": id})
	return
}

func FindByTargetAndClusterID(target, clusterID string) (labels []*LabelDTO, err error) {
	var orms []*LabelModel
	orms, err = find(map[string]interface{}{"target": target, "cluster_id": clusterID, "order_by": "-create_at"})
	labels = toDTO(orms)
	return
}

var ErrColumnNameNotString = errors.New("order by should be specified by column name in string")

func find(constraint map[string]interface{}) (labels []*LabelModel, err error) {
	selector := orm.NewOrm().QueryTable(tableName)
	for key, value := range constraint {
		if key == "order_by" {
			if column, ok := value.(string); !ok {
				err = ErrColumnNameNotString
				return
			} else {
				selector = selector.OrderBy(column)
			}
		} else {
			selector = selector.Filter(key, value)
		}
	}
	_, err = selector.All(&labels)
	return
}

func toDTO(orms []*LabelModel) (dtos []*LabelDTO) {
	dtos = make([]*LabelDTO, len(orms))
	for index, orm := range orms {
		dtos[index] = orm.ToDTO()
	}
	return
}
