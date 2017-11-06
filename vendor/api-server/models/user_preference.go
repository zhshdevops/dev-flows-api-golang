/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-10-19  @author liuyang
 */

package models

import (
	"time"

	shortuid "api-server/modules/tenx/id"

	"github.com/astaxie/beego/orm"
)

// user_preference | CREATE TABLE `user_preference` (
//   `id` char(17) NOT NULL,
//   `owner_name` varchar(200) NOT NULL,
//   `type` varchar(45) NOT NULL COMMENT '3rdparty-registry',
//   `config_detail` text NOT NULL,
//   `description` text,
//   `create_time` datetime DEFAULT NULL,
//   PRIMARY KEY (`id`),
//   UNIQUE KEY `id_UNIQUE` (`id`),
//   KEY `user_config` (`type`,`owner_name`)
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8

// UserPreferenceTable record of user_preference table
type UserPreferenceTable struct {
	ID             string             `orm:"pk;column(id)"`
	OwnerName      string             `orm:"column(owner_name)"`
	OwnerNamespace string             `orm:"column(owner_namespace)"`
	Type           UserPreferenceType `orm:"column(type)"`
	ConfigDetail   string             `orm:"column(config_detail)"`
	Description    string             `orm:"column(description)"`
	CreateTime     time.Time          `orm:"column(create_time)"`
}

// UserPreferenceType user preference type
type UserPreferenceType string

func (upt *UserPreferenceType) String() string {
	return string(*upt)
}

const (
	// UserPreference_ThirdPartyRegistry user preference type third party registry
	UserPreference_ThirdPartyRegistry UserPreferenceType = "3rdparty-registry"
	UserPreference_TenxCloudHub       UserPreferenceType = "tenxcloud-hub"
)

func init() {
	orm.RegisterModel(new(UserPreferenceTable))
}

// TableName return user_preference
func (up *UserPreferenceTable) TableName() string {
	return "tenx_user_preference"
}

func NewUserPreferenceTableRecord(t UserPreferenceType, username, namespace, description, configDetail string) *UserPreferenceTable {
	return &UserPreferenceTable{
		ID:             shortuid.NewUserPreference(),
		OwnerName:      username,
		OwnerNamespace: namespace,
		Type:           t,
		ConfigDetail:   configDetail,
		Description:    description,
		CreateTime:     time.Now(),
	}
}

// Insert insert record
func (up *UserPreferenceTable) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(up)
	return err
}

// Delete delete record
func (up *UserPreferenceTable) Delete() error {
	o := orm.NewOrm()
	_, err := o.QueryTable(up.TableName()).Filter("id", up.ID).Filter("owner_namespace", up.OwnerNamespace).Filter("owner_name", up.OwnerName).Delete()
	return err
}

// DeleteByType delete record by type
func (up *UserPreferenceTable) DeleteByType() error {
	o := orm.NewOrm()
	_, err := o.QueryTable(up.TableName()).Filter("type", up.Type).Filter("owner_namespace", up.OwnerNamespace).Filter("owner_name", up.OwnerName).Delete()
	return err
}

// List list user configs by username and type
func (up *UserPreferenceTable) List(namespace string, uptype UserPreferenceType) ([]UserPreferenceTable, error) {
	o := orm.NewOrm()
	var configs []UserPreferenceTable
	_, err := o.QueryTable(up.TableName()).Filter("type", string(uptype)).Filter("owner_namespace", namespace).All(&configs)
	return configs, err
}
