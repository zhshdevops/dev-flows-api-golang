/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package teamspace2user

import (
	"github.com/golang/glog"

	"api-server/models/common"
	"api-server/models/teamspace"
)

// GetSpaceListByUser fetches spaces according to user id
func GetSpaceListByUser(userID int32, dataselect *common.DataSelectQuery) (*teamspace.SpaceList, uint32, error) {
	if dataselect == nil {
		dataselect = common.DefaultDataSelect
	}

	tuf := &SpaceUserModel{UserID: userID}
	tufs, errcode, err := tuf.ListByUser(userID, common.NewDataSelectQuery(common.NoPagination, common.NoSort, common.NoFilter))
	if err != nil {
		glog.Errorf("get space list by user id %d fails, error code:%d, error:%v\n", userID, errcode, err)
		return nil, errcode, err
	}

	var spaceIDs []string
	roleMap := make(map[string]int32)
	for _, item := range tufs {
		spaceIDs = append(spaceIDs, item.SpaceID)
		roleMap[item.SpaceID] = item.Role
	}

	spacelist, errcode, err := teamspace.GetSpaceListByIDs(spaceIDs, dataselect)
	if err != nil {
		return spacelist, errcode, err
	}
	for idx := range spacelist.Spaces {
		spacelist.Spaces[idx].Role = roleMap[spacelist.Spaces[idx].SpaceID]
	}
	return spacelist, errcode, err
}
