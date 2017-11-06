/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-21  @author shouhong.zhang
 */

package team2resource

import (
	"fmt"
	"time"

	"api-server/models/common"
	sqlstatus "api-server/models/sql/status"

	"github.com/astaxie/beego/orm"
)

//RequestStatus request status
type RequestStatus int

const (
	//Pending the request status is pending
	Pending RequestStatus = iota
	//Approved the request is approved
	Approved
)

// TeamResourceRequestModel tenx_team_resource_request, team_id and resource_id together as the primary key
type TeamResourceRequestModel struct {
	TeamID        string    `orm:"column(team_id);"`
	ResourceID    string    `orm:"column(resource_id)"`
	ResourceType  int       `orm:"column(resource_type)"`
	RequestUserID int32     `orm:"column(request_user_id)"`
	RequestTime   time.Time `orm:"column(request_time);type(datetime)"`
	ApproveUserID int32     `orm:"column(approve_user_id)"`
	ApproveTime   time.Time `orm:"column(approve_time);type(datetime)"`
	Status        int       `orm:"column(status)"`
}

//TableName the table name of tenx_team_resource_request
func (tc *TeamResourceRequestModel) TableName() string {
	return "tenx_team_resource_request"
}

// Insert add a new ref record
func (tc *TeamResourceRequestModel) Insert() (uint32, error) {
	o := orm.NewOrm()
	sql := "INSERT INTO tenx_team_resource_request (team_id, resource_id, resource_type, request_user_id, request_time, status) VALUES (?, ?, ?, ?, ?, ?);"
	_, err := o.Raw(sql, tc.TeamID, tc.ResourceID, tc.ResourceType,
		tc.RequestUserID, tc.RequestTime, tc.Status).Exec()
	return sqlstatus.ParseErrorCode(err)
}

//ListByTeamResourceRequestStatus returns a list of team cluster requests according to team id, no pagination, no sort
func (tc *TeamResourceRequestModel) ListByTeamResourceRequestStatus(teamID string, requestType RequestStatus, dataselect *common.DataSelectQuery) ([]TeamResourceRequestModel, uint32, error) {
	o := orm.NewOrm()

	// space query ids
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE team_id=?;`, tc.TableName())

	var refs []TeamResourceRequestModel
	_, err := o.Raw(sql, teamID).QueryRows(&refs)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return refs, errCode, err
}
