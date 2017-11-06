/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package teamspace2user

import (
	"github.com/golang/glog"

	sqlstatus "api-server/models/sql/status"
	"api-server/models/teamspace"
	"api-server/models/teamspace2app"
)

// GetSpaceDetail fetches space details according to space id
func GetSpaceDetail(spaceID string) (*teamspace.Space, uint32, error) {
	spaceModel := &teamspace.SpaceModel{SpaceID: spaceID}
	errcode, err := spaceModel.Get()
	if err != nil {
		glog.Errorf("get team space %s fails, error code:%d, error:%v\n", spaceID, errcode, err)
		return nil, errcode, err
	}
	spaceAccModel := &teamspace.SpaceAccountModel{SpaceID: spaceID}
	errcode, err = spaceAccModel.Get()
	if err != nil {
		glog.Errorf("team space %s exist, but lack account information. error code:%d, error:%v\n", spaceID, errcode, err)
		return nil, errcode, err
	}

	spaceAppInfoModel := &teamspace2app.SpaceAppInfoModel{}
	errcode, err = spaceAppInfoModel.Get(spaceID)
	if err != nil {
		glog.Errorf("get team space %s's app informaion fails, error code:%d, error:%v\n", spaceID, errcode, err)
		return nil, errcode, err
	}

	spaceDetail := teamspace.ToSpace(spaceModel, spaceAccModel, spaceAppInfoModel)

	// get user number
	spaceUserModel := SpaceUserModel{}
	count, errcode, err := spaceUserModel.CountUser(spaceID)
	if err != nil {
		glog.Errorf("get user count in space %s fails, error code:%d, error:%v\n", spaceID, errcode, err)
		return nil, errcode, err
	}

	spaceDetail.UserCount = count

	return spaceDetail, sqlstatus.SQLSuccess, nil
}
