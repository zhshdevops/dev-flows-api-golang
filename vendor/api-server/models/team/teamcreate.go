/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"

	idfactory "api-server/modules/tenx/id"
	"fmt"
)

// CreateTeam creates a new user and returns its detail
func CreateTeam(spec Team, orms ...orm.Ormer) (*Team, uint32, error) {
	teamModel := &TeamModel{
		TeamID:       idfactory.NewTeam(),
		TeamName:     spec.TeamName,
		Description:  spec.Description,
		CreatorID:    spec.CreatorID,
		CreationTime: time.Now(),
	}
	errcode, err := teamModel.Insert(orms...)
	if err != nil {
		glog.Errorf("create team %s fails, error code:%d, error:%v\n", spec.TeamName, errcode, err)
		return nil, errcode, err
	}

	return ToTeam(teamModel, nil, nil, nil), errcode, err
}

//CreateOrUpdateTeam
func CreateOrUpdateTeam(spec Team, orms ...orm.Ormer) (*Team, error) {
	var o orm.Ormer
	if len(orms) == 0 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	spec.TeamID = idfactory.NewTeam()
	sql := "insert into %s (id, name, description, creator_id, creation_time) value(?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE creator_id = ?"
	sql = fmt.Sprintf(sql, "tenx_teams")
	_, err := o.Raw(sql, spec.TeamID, spec.TeamName, spec.Description, spec.CreatorID, time.Now(), spec.CreatorID).Exec()
	if nil != err {
		return nil, err
	}
	var result TeamModel
	err = o.QueryTable("tenx_teams").Filter("name", spec.TeamName).One(&result)
	if nil != err {
		return nil, err
	}
	spec.TeamID = result.TeamID
	return &spec, err
}

func CreateOrIgnoreTeam(spec Team, orms ...orm.Ormer) (*Team, error) {
	var o orm.Ormer
	if len(orms) == 0 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	spec.TeamID = idfactory.NewTeam()
	sql := "insert ignore %s (id, name, description, creator_id, creation_time) value(?, ?, ?, ?, ?)"
	sql = fmt.Sprintf(sql, "tenx_teams")
	_, err := o.Raw(sql, spec.TeamID, spec.TeamName, spec.Description, spec.CreatorID, time.Now()).Exec()
	if nil != err {
		return nil, err
	}
	var result TeamModel
	err = o.QueryTable("tenx_teams").Filter("name", spec.TeamName).One(&result)
	if nil != err {
		return nil, err
	}
	spec.TeamID = result.TeamID
	return &spec, err
}

//Exist check the team existence
func Exist(name string) bool {
	o := orm.NewOrm()
	return o.QueryTable("tenx_teams").Filter("name", name).Exist()
}

//ExistByID check the team exitence by ID
func ExistByID(id string) bool {
	o := orm.NewOrm()
	return o.QueryTable("tenx_teams").Filter("id", id).Exist()
}
