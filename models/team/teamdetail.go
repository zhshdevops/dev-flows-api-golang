/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team

import (
	"github.com/golang/glog"

	sqlstatus "dev-flows-api-golang/models/sql/status"
	"dev-flows-api-golang/models/team2resource"
)

// GetTeamDetail returns a team's details from team id
func GetTeamDetail(teamID string) (*Team, uint32, error) {
	t := &TeamModel{TeamID: teamID}
	if errcode, err := t.Get(); err != nil {
		glog.Errorf("get team %s fails, error code:%d, error:%v\n", teamID, errcode, err)
		return nil, errcode, err
	}

	teamUserInfoModel := &TeamUserInfoModel{}
	if errcode, err := teamUserInfoModel.Get(teamID); err != nil {
		glog.Errorf("get team %s's summary info fails, error code:%d, error:%v\n", teamID, errcode, err)
		return nil, errcode, err
	}

	teamSpaceInfoModel := &TeamSpaceInfoModel{}
	if errcode, err := teamSpaceInfoModel.Get(teamID); err != nil {
		glog.Errorf("get team %s's space summary info fails, error code:%d, error:%v\n", teamID, errcode, err)
		return nil, errcode, err
	}

	resourceInfoModel := &team2resource.TeamResourceInfoModel{}
	if errcode, err := resourceInfoModel.GetByTeamID(teamID); err != nil {
		glog.Errorf("get team resource summary of team %s fails, error code:%d, error:%v\n", teamID, errcode, err)
		return nil, errcode, err
	}

	return ToTeam(t, teamUserInfoModel, teamSpaceInfoModel, resourceInfoModel), sqlstatus.SQLSuccess, nil
}
