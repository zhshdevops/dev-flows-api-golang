/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team2resource

import (
	"fmt"

	sqlstatus "api-server/models/sql/status"

	"github.com/astaxie/beego/orm"

	"api-server/models/common"
)

type ResourceType int

const (
	ResourceCluster ResourceType = iota
	ResourceStorage
)

// TeamResourceModel tenx_team_resource_ref, team_id and resource_id together as the primary key
type TeamResourceModel struct {
	TeamID       string `orm:"column(team_id);pk"`
	ResourceID   string `orm:"column(resource_id)"`
	ResourceType int    `orm:"column(resource_type)"`
}

func (tc TeamResourceModel) TableName() string {
	return "tenx_team_resource_ref"
}

// Insert add a new ref record
func (tc *TeamResourceModel) Insert(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	sql := "INSERT INTO tenx_team_resource_ref (team_id, resource_id, resource_type) VALUES (?, ?, ?);"
	_, err := o.Raw(sql, tc.TeamID, tc.ResourceID, tc.ResourceType).Exec()
	return sqlstatus.ParseErrorCode(err)
}

// Delete deletes a record
func (tc *TeamResourceModel) Delete() (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Raw("delete from tenx_team_resource_ref where team_id=? and resource_id=? and resource_type=?;", tc.TeamID, tc.ResourceID, tc.ResourceType).Exec()
	return sqlstatus.ParseErrorCode(err)
}

// ListByResource lists all team cluster ref by resourceID, no pagination, no sort
func (tc *TeamResourceModel) ListByResource(resourceID string, dataselect *common.DataSelectQuery) ([]TeamResourceModel, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT * FROM %s WHERE resource_id = ?;`, tc.TableName())
	var refs []TeamResourceModel
	_, err := o.Raw(sql, resourceID).QueryRows(&refs)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}
	errCode, err := sqlstatus.ParseErrorCode(err)
	return refs, errCode, err
}

// ListByTeam returns a list of team cluster ref according to team id, no pagination, no sort
func (tc *TeamResourceModel) ListByTeam(teamID string, dataselect *common.DataSelectQuery) ([]TeamResourceModel, uint32, error) {
	o := orm.NewOrm()

	// space query ids
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE team_id=?;`, tc.TableName())

	var refs []TeamResourceModel
	_, err := o.Raw(sql, teamID).QueryRows(&refs)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return refs, errCode, err
}

// ListByTeamResourceType returns a list of team cluster ref according to team id, no pagination, no sort
func (tc *TeamResourceModel) ListByTeamResourceType(teamID string, resourceType ResourceType, dataselect *common.DataSelectQuery) ([]TeamResourceModel, uint32, error) {
	o := orm.NewOrm()

	// space query ids
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE team_id=?;`, tc.TableName())

	var refs []TeamResourceModel
	_, err := o.Raw(sql, teamID).QueryRows(&refs)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return refs, errCode, err
}

// DeleteByTeamID deletes all records with certain team_id
func (tc *TeamResourceModel) DeleteByTeamID(teamID string) (uint32, error) {
	o := orm.NewOrm()

	_, err := o.QueryTable(tc.TableName()).Filter("team_id", teamID).Delete()
	return sqlstatus.ParseErrorCode(err)
}
