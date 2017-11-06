/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team2user

import (
	"github.com/golang/glog"

	"api-server/models/common"
	"api-server/models/team"
)

// GetTeamListByUser fetches teams according to user id
func GetTeamListByUser(userID int32, dataselect *common.DataSelectQuery) (*team.TeamList, uint32, error) {
	if dataselect == nil {
		dataselect = common.DefaultDataSelect
	}

	tuf := &TeamUserModel{UserID: userID}
	tufs, errcode, err := tuf.ListByUser(userID, common.NoDataSelect)
	if err != nil {
		glog.Errorf("get team list by user id %d fails, error code:%d, error:%v\n", userID, errcode, err)
		return nil, errcode, err
	}

	var teamIDs []string
	for _, item := range tufs {
		teamIDs = append(teamIDs, item.TeamID)
	}

	return team.GetTeamListByIDs(false, teamIDs, dataselect)
}
