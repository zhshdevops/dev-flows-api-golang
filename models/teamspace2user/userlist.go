/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package teamspace2user

import (
	"github.com/golang/glog"

	"dev-flows-api-golang/models/common"
	sqlstatus "dev-flows-api-golang/models/sql/status"
	"dev-flows-api-golang/models/user"
)

// GetUserListBySpace fetches users  according to space id
func GetUserListBySpace(spaceID string, dataselect *common.DataSelectQuery) (*user.UserList, uint32, error) {
	if dataselect == nil {
		dataselect = common.DefaultDataSelect
	}

	tuf := &SpaceUserModel{SpaceID: spaceID}
	tufs, errcode, err := tuf.ListBySpace(spaceID)
	if err != nil {
		glog.Errorf("get user space ref list by space id %s fails, error code:%d, error:%v\n", spaceID, errcode, err)
		return nil, errcode, err
	}
	if len(tufs) == 0 {
		return &user.UserList{Users: []user.User{}}, sqlstatus.SQLSuccess, nil
	}

	var userIDs []int32
	for _, item := range tufs {
		userIDs = append(userIDs, item.UserID)
	}

	if len(userIDs) == 0 {
		return &user.UserList{Users: []user.User{}}, sqlstatus.SQLSuccess, nil
	}

	return user.GetUserListByIDs(userIDs, dataselect)
}
