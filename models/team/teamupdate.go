/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team

import (
	"github.com/golang/glog"

	sqlstatus "dev-flows-api-golang/models/sql/status"
)

// UpdateTeam update the current team and returns its detail
func UpdateTeam(teamID string, spec Team) (*Team, uint32, error) {
	tm := &TeamModel{TeamID: teamID}
	errcode, err := tm.Get()
	if err != nil {
		glog.Errorf("update team %s fails, errocode: %d, error:%v\n", teamID, errcode, err)
		return nil, errcode, err
	}

	// check every fields
	if len(spec.TeamName) > 0 {
		tm.TeamName = spec.TeamName
	}
	if len(spec.Description) > 0 {
		tm.Description = spec.Description
	}

	if errcode, err := tm.Update(); err != nil {
		glog.Errorf("update team %s fails, error code:%d, error:%v\n", tm.TeamID, errcode, err)
		return nil, errcode, err
	}

	return ToTeam(tm, nil, nil, nil), sqlstatus.SQLSuccess, nil
}
