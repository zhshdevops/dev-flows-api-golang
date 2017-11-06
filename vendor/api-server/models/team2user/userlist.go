/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */
package team2user

import (
	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"

	"api-server/models/common"
	sqlstatus "api-server/models/sql/status"
	"api-server/models/user"
)

var (
	// SortByTeamInfoMapping is the mapping from team count passed key to database field
	SortByTeamInfoMapping = map[string]string{
		"teamCount": "team_count",
	}
	// FilterByMapping is the mapping from user passed key to database field
	FilterByMapping = map[string]string{
		"userName": "tenx_users.user_name",
		"email":    "tenx_users.email",
		"phone":    "tenx_users.phone",
		"role":     "tenx_users.role",
		"type":     "tenx_users.type",
	}
)

// GetUserList fetches users
func GetUserList(dataselect *common.DataSelectQuery) (*user.UserList, uint32, error) {
	var sortby string
	if dataselect != nil && dataselect.SortQuery != nil && len(dataselect.SortQuery.SortByList) > 0 {
		sortby = dataselect.SortQuery.SortByList[0].Property
	}

	userInfoModel := &UserInfoModel{}

	var userlist *user.UserList
	var userInfoModels []UserInfoModel
	var errcode uint32
	var err error

	if _, ok := SortByTeamInfoMapping[sortby]; ok {
		dataselect.SortQuery.SortByList[0].Property = SortByTeamInfoMapping[sortby]
		userInfoModels, errcode, err = userInfoModel.ListUserInfos(dataselect)
		if err != nil {
			glog.Errorf("get user's team info fails, error code:%d, error:%v\n", errcode, err)
			return nil, errcode, err
		}
		var userIDs []int32
		for _, ui := range userInfoModels {
			userIDs = append(userIDs, ui.UserID)
		}
		userlist, errcode, err = user.GetUserListByIDs(userIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get user info of %v fails, error code:%d, error:%v\n", userIDs, errcode, err)
			return nil, errcode, err
		}
		userMap := make(map[int32]user.User)
		for _, u := range userlist.Users {
			userMap[u.UserID] = u
		}
		userlist.Users = nil
		for _, ui := range userInfoModels {
			u := userMap[ui.UserID]
			u.TeamCount = ui.TeamCount
			userlist.Users = append(userlist.Users, u)
		}

	} else {
		userlist, errcode, err = user.GetUserList(dataselect)
		if err != nil {
			glog.Errorf("get user list fails, error code:%d, error:%v\n", errcode, err)
			return userlist, errcode, err
		}

		var userIDs []int32
		for _, item := range userlist.Users {
			userIDs = append(userIDs, item.UserID)
		}

		userInfoModels, errcode, err = userInfoModel.ListUserInfoByIDs(userIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get user info of %v fails, error code:%d, error:%v\n", userIDs, errcode, err)
			return userlist, errcode, err
		}
		infoMap := make(map[int32]UserInfoModel)
		for _, info := range userInfoModels {
			infoMap[info.UserID] = info
		}
		for idx, u := range userlist.Users {
			userlist.Users[idx].TeamCount = infoMap[u.UserID].TeamCount
			// userlist.Users[idx].SpaceCount = infoMap[u.UserID].SpaceCount
		}
	}

	// get user count
	u := &user.UserModel{}
	if dataselect.FilterQuery != nil {
		for idx, f := range dataselect.FilterQuery.FilterByList {
			val, ok := user.FilterByMapping[f.Property]
			if ok {
				dataselect.FilterQuery.FilterByList[idx].Property = val
			}
		}
	}

	count, errCode, err := u.Count(dataselect)
	if err != nil {
		glog.Errorf("get user number fails, error code:%d, error:%v\n", errCode, err)
		return nil, errCode, err
	}
	userlist.ListMeta.Total = count

	return userlist, sqlstatus.SQLSuccess, nil
}

// GetUserListByIDs fetches users
func GetUserListByIDs(userIDs []int32, dataselect *common.DataSelectQuery) (*user.UserList, uint32, error) {
	var sortby string
	if dataselect != nil && dataselect.SortQuery != nil && len(dataselect.SortQuery.SortByList) > 0 {
		sortby = dataselect.SortQuery.SortByList[0].Property
	}

	userInfoModel := &UserInfoModel{}

	var userlist *user.UserList
	var userInfoModels []UserInfoModel
	var errcode uint32
	var err error

	// get user count
	userNumber := len(userIDs)
	if _, ok := SortByTeamInfoMapping[sortby]; ok {
		dataselect.SortQuery.SortByList[0].Property = SortByTeamInfoMapping[sortby]
		userInfoModels, errcode, err = userInfoModel.ListUserInfoByIDs(userIDs, dataselect)
		if err != nil {
			glog.Errorf("get user's team info fails, error code:%d, error:%v\n", errcode, err)
			return nil, errcode, err
		}
		userIDs = nil
		for _, ui := range userInfoModels {
			userIDs = append(userIDs, ui.UserID)
		}
		userlist, errcode, err = user.GetUserListByIDs(userIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get user info of %v fails, error code:%d, error:%v\n", userIDs, errcode, err)
			return nil, errcode, err
		}
		userMap := make(map[int32]user.User)
		for _, u := range userlist.Users {
			userMap[u.UserID] = u
		}
		userlist.Users = nil
		for _, ui := range userInfoModels {
			u := userMap[ui.UserID]
			u.TeamCount = ui.TeamCount
			userlist.Users = append(userlist.Users, u)
		}
	} else {
		userlist, errcode, err = user.GetUserListByIDs(userIDs, dataselect)
		if err != nil {
			glog.Errorf("get user list fails, error code:%d, error:%v\n", errcode, err)
			return userlist, errcode, err
		}

		userIDs = nil
		for _, item := range userlist.Users {
			userIDs = append(userIDs, item.UserID)
		}

		userInfoModels, errcode, err = userInfoModel.ListUserInfoByIDs(userIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get user info of %v fails, error code:%d, error:%v\n", userIDs, errcode, err)
			return userlist, errcode, err
		}
		infoMap := make(map[int32]UserInfoModel)
		for _, info := range userInfoModels {
			infoMap[info.UserID] = info
		}
		for idx, u := range userlist.Users {
			userlist.Users[idx].TeamCount = infoMap[u.UserID].TeamCount
			// userlist.Users[idx].SpaceCount = infoMap[u.UserID].SpaceCount
		}
	}

	userlist.ListMeta.Total = userNumber

	return userlist, sqlstatus.SQLSuccess, nil
}

// GetUserIDsByTeam fetches user IDs according to team id
func GetUserIDsByTeam(teamID string, dataselect *common.DataSelectQuery) ([]int32, uint32, error) {
	tuf := &TeamUserModel{TeamID: teamID}
	tufs, errcode, err := tuf.ListByTeam(teamID, common.NoDataSelect)
	if err != nil {
		glog.Errorf("get team list by team id %s fails, error code:%d, error:%v\n", teamID, errcode, err)
		return nil, errcode, err
	}

	if len(tufs) == 0 {
		return []int32{}, sqlstatus.SQLSuccess, nil
	}

	var userIDs []int32
	for _, item := range tufs {
		userIDs = append(userIDs, item.UserID)
	}

	userModel := &user.UserModel{}
	dataselect, err = user.ValidateUserDataSelect(dataselect)
	if err != nil {
		glog.Errorf("invalid dataselect, error:%v\n", err)
		return nil, sqlstatus.SQLErrSyntax, err
	}
	userIDs, errcode, err = userModel.ListIDsByIDs(userIDs, dataselect)
	if err != nil {
		glog.Errorf("get filter user id fails, error code:%d, error:%v\n", errcode, err)
		return nil, errcode, err
	}

	return userIDs, sqlstatus.SQLSuccess, nil
}

// AddTeamUserRef add team - user ref record
func AddTeamUserRef(teamID string, userIDs []int32, isTeamAdmin bool, orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) == 0 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	var role int32
	if isTeamAdmin {
		role = 1
	}
	for _, userID := range userIDs {
		teamUserModel := &TeamUserModel{TeamID: teamID, UserID: userID, Role: role}
		if errcode, err := teamUserModel.Insert(o); err != nil {
			return errcode, err
		}
	}
	return sqlstatus.SQLSuccess, nil
}

//AddOrIgnoreUserRef add or ignore add user ref
func AddOrIgnoreUserRef(teamID string, userID int32, isTeamAdmin bool, orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) == 0 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	var role int
	if isTeamAdmin {
		role = 1
	}
	teamUserModel := &TeamUserModel{}
	return teamUserModel.InsertIgnore(teamID, userID, role, o)
}

// DeleteTeamUsers deletes team - user ref record
func DeleteTeamUsers(teamID string, userIDs []int32) (uint32, error) {
	for _, userID := range userIDs {
		teamUserModel := &TeamUserModel{TeamID: teamID, UserID: userID}
		if errcode, err := teamUserModel.Delete(); err != nil {
			return errcode, err
		}
	}
	return sqlstatus.SQLSuccess, nil
}
