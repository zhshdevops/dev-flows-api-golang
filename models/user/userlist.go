/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang/glog"

	"dev-flows-api-golang/models/common"
	sqlstatus "dev-flows-api-golang/models/sql/status"
)

var (
	// SortByMapping is the mapping from user passed key to database field
	SortByMapping = map[string]string{
		"userName":     "user_name",
		"email":        "email",
		"creationTime": "creation_time",
	}
	// SortByAccountMapping is the mapping from user account passed key to database field
	SortByAccountMapping = map[string]string{
		"balance": "balance",
	}
	// FilterByMapping is the mapping from user passed key to database field
	FilterByMapping = map[string]string{
		"userName": "user_name",
		"email":    "email",
		"phone":    "phone",
		"role":     "role",
		"type":     "type",
	}
)

// User tenx_users record
type User struct {
	UserID         int32     `json:"userID"`
	UserName       string    `json:"userName"`
	Namespace      string    `json:"namespace"`
	DisplayName    string    `json:"displayName"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	CreationTime   time.Time `json:"creationTime"`
	APIToken       string    `json:"apiToken"`
	Role           int32     `json:"role"`
	Avatar         string    `json:"avatar"`
	Balance        int64     `json:"balance"`
	TeamCount      int       `json:"teamCount"`
	Active         int8      `json:"active"`
	LoginFrequency int       `json:"login_frequency"`
	EnvEdition     int       `json:"env_edition"`
	Migrated       int       `json:"migrated"`
	Password       string    `json:"password"`
	Type           int       `json:"type"`
}

// UserList represent a list of users
type UserList struct {
	ListMeta common.ListMeta `json:"listMeta"`
	Users    []User          `json:"users"`
}

// ToUser converts a UserModel to User
func ToUser(um *UserModel, ua *UserAccountModel) *User {
	if um == nil {
		return nil
	}
	u := &User{
		UserID:         um.UserID,
		UserName:       um.Username,
		Namespace:      um.Namespace,
		DisplayName:    um.Displayname,
		Email:          um.Email,
		Phone:          um.Phone,
		CreationTime:   um.CreationTime,
		APIToken:       um.APIToken,
		Role:           um.Role,
		Avatar:         um.Avatar,
		Active:         um.Active,
		LoginFrequency: um.LoginFrequency,
		EnvEdition:     um.EnvEdition,
		Migrated:       um.Migrated,
		//Password:       um.Password,
		Type: um.Type,
	}
	if ua != nil {
		u.Balance = ua.Balance
	}
	return u
}

// ToUsers converts []UserModel to []User
func ToUsers(ums []UserModel, useraccounts []UserAccountModel, sortby string) ([]User, error) {
	if len(ums) != len(useraccounts) {
		glog.Errorf("users number %d and user accounts number %d does not match\n", len(ums), len(useraccounts))
	}
	var users []User
	if _, ok := SortByMapping[sortby]; ok {
		userMap := make(map[int32]UserAccountModel)
		for _, ua := range useraccounts {
			userMap[ua.UserID] = ua
		}

		for _, um := range ums {
			ua := userMap[um.UserID]
			users = append(users, *ToUser(&um, &ua))
		}
	} else {
		userMap := make(map[int32]UserModel)
		for _, um := range ums {
			userMap[um.UserID] = um
		}
		for _, ua := range useraccounts {
			um := userMap[ua.UserID]
			users = append(users, *ToUser(&um, &ua))
		}
	}
	return users, nil
}

// ValidateUserDataSelect filter out invalid sort, filter options, and return valid ones
func ValidateUserDataSelect(old *common.DataSelectQuery) (*common.DataSelectQuery, error) {
	if old == nil {
		return common.NewDataSelectQuery(common.DefaultPagination, common.NewSortQuery([]string{"a", SortByMapping["userName"]}), common.NoFilter), nil
	}

	dataselect := common.NewDataSelectQuery(common.DefaultPagination, common.NoSort, common.NoFilter)
	if old.FilterQuery != nil {
		for _, f := range old.FilterQuery.FilterByList {
			prop, ok := FilterByMapping[f.Property]
			if !ok {
				glog.Errorf("user list, invalid filter by options: %s\n", f.Property)
				return nil, fmt.Errorf("Invalid filter option: %s", f.Property)
			}
			f.Property = prop
			dataselect.FilterQuery.FilterByList = append(dataselect.FilterQuery.FilterByList, f)
		}
	}

	if old.SortQuery != nil {
		for _, sq := range old.SortQuery.SortByList {
			prop, ok := SortByMapping[sq.Property]
			if !ok {
				prop, ok = SortByAccountMapping[sq.Property]
				if !ok {
					glog.Errorf("user list, invalid sort options: %s\n", sq.Property)
					return nil, fmt.Errorf("invalid sort option: %s\n", sq.Property)
				}
			}
			sq.Property = prop
			dataselect.SortQuery.SortByList = append(dataselect.SortQuery.SortByList, sq)

		}
	}

	if old.PaginationQuery != nil {
		dataselect.PaginationQuery = common.NewPaginationQuery(old.PaginationQuery.From, old.PaginationQuery.Size)
	}
	return dataselect, nil
}

// GetUserList returns a list of users, no team count, no space count
func GetUserList(dataselect *common.DataSelectQuery) (*UserList, uint32, error) {
	var sortby string
	if dataselect != nil && dataselect.SortQuery != nil && len(dataselect.SortQuery.SortByList) > 0 {
		sortby = dataselect.SortQuery.SortByList[0].Property
	}

	// validate data select options
	dataselect, err := ValidateUserDataSelect(dataselect)
	if err != nil {
		return nil, sqlstatus.SQLErrSyntax, err
	}

	var userlist = &UserList{}

	// get user count
	u := &UserModel{}
	count, errCode, err := u.Count(dataselect)
	if err != nil {
		glog.Errorf("get user number fails, error code:%d, error:%v\n", errCode, err)
		return nil, errCode, err
	}
	userlist.ListMeta.Total = count

	if _, ok := SortByMapping[sortby]; ok { // sort by user's fields
		// get user list
		userModels, errCode, err := u.List(dataselect)
		if err != nil {
			glog.Errorf("get user list fails, pagination: %s, sort:%s, code:%d, error:%v\n", dataselect.PaginationQuery, dataselect.SortQuery, errCode, err)
			return nil, errCode, err
		}

		var userIDs []int32
		for _, um := range userModels {
			userIDs = append(userIDs, um.UserID)
		}
		userAccountModel := &UserAccountModel{}
		userAccounts, errCode, err := userAccountModel.ListByIDs(userIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get user account list fails, code:%s, error:%v\n", errCode, err)
			return nil, errCode, err
		}

		userlist.Users, err = ToUsers(userModels, userAccounts, sortby)
		if err != nil {
			errcode, err := sqlstatus.ParseErrorCode(err)
			return nil, errcode, err
		}
	} else { // sort by account info
		// get user list
		userModels, errCode, err := u.List(common.NewDataSelectQuery(common.NoPagination, common.NoSort, dataselect.FilterQuery))
		if err != nil {
			glog.Errorf("get user list fails, pagination: %s, sort:%s, code:%d, error:%v\n", dataselect.PaginationQuery, dataselect.SortQuery, errCode, err)
			return nil, errCode, err
		}
		var userIDs []int32
		for _, um := range userModels {
			userIDs = append(userIDs, um.UserID)
		}

		userAccountModel := &UserAccountModel{}
		userAccounts, errCode, err := userAccountModel.ListByIDs(userIDs, common.NewDataSelectQuery(dataselect.PaginationQuery, dataselect.SortQuery, common.NoFilter))
		if err != nil {
			glog.Errorf("get user account list fails, code:%s, error:%v\n", errCode, err)
			return nil, errCode, err
		}

		userlist.Users, err = ToUsers(userModels, userAccounts, sortby)
		if err != nil {
			errcode, err := sqlstatus.ParseErrorCode(err)
			return nil, errcode, err
		}
	}

	return userlist, sqlstatus.SQLSuccess, nil
}

// GetUserListByIDs returns a list of users according to a list of ids
func GetUserListByIDs(userIDs []int32, dataselect *common.DataSelectQuery) (*UserList, uint32, error) {
	if len(userIDs) == 0 {
		return &UserList{Users: []User{}}, sqlstatus.SQLSuccess, nil
	}
	var sortby string
	if dataselect != nil && dataselect.SortQuery != nil && len(dataselect.SortQuery.SortByList) > 0 {
		sortby = dataselect.SortQuery.SortByList[0].Property
	}

	// validate data select options
	dataselect, err := ValidateUserDataSelect(dataselect)
	if err != nil {
		return nil, sqlstatus.SQLErrSyntax, err
	}

	var userlist = &UserList{ListMeta: common.ListMeta{Total: len(userIDs)}, Users: []User{}}

	u := &UserModel{}

	if _, ok := SortByMapping[sortby]; ok {
		// get user list
		userModels, errCode, err := u.ListByIDs(userIDs, dataselect)
		if err != nil {
			glog.Errorf("get user list fails, pagination: %s, sort:%s, code:%s, error:%v\n", dataselect.PaginationQuery, dataselect.SortQuery, errCode, err)
			return nil, errCode, err
		}

		userIDs = nil // clear
		for _, um := range userModels {
			userIDs = append(userIDs, um.UserID)
		}
		if userIDs == nil {
			glog.Error("the user information is incomplete in database\n")
			err = errors.New("the user information is incomplete in database")
			return userlist, sqlstatus.SQLErrNoRowFound, err
		}

		userAccountModel := &UserAccountModel{}
		userAccounts, errCode, err := userAccountModel.ListByIDs(userIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get user account list fails, code:%s, error:%v\n", errCode, err)
			return nil, errCode, err
		}

		userlist.Users, err = ToUsers(userModels, userAccounts, sortby)
		if err != nil {
			errcode, err := sqlstatus.ParseErrorCode(err)
			return nil, errcode, err
		}
	} else {
		userAccountModel := &UserAccountModel{}
		userAccounts, errCode, err := userAccountModel.ListByIDs(userIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get user account list fails, code:%s, error:%v\n", errCode, err)
			return nil, errCode, err
		}
		userIDs = nil // clear
		for _, ua := range userAccounts {
			userIDs = append(userIDs, ua.UserID)
		}
		if userIDs == nil {
			glog.Error("the user information is incomplete in database\n")
			err = errors.New("the user information is incomplete in database")
			return userlist, sqlstatus.SQLErrNoRowFound, err
		}

		// get user list
		userModels, errCode, err := u.ListByIDs(userIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get user list fails, pagination: %s, sort:%s, code:%s, error:%v\n", dataselect.PaginationQuery, dataselect.SortQuery, errCode, err)
			return nil, errCode, err
		}

		if len(userModels) == 0 {
			return userlist, sqlstatus.SQLSuccess, nil
		}

		userlist.Users, err = ToUsers(userModels, userAccounts, sortby)
		if err != nil {
			errcode, err := sqlstatus.ParseErrorCode(err)
			return nil, errcode, err
		}
	}

	return userlist, sqlstatus.SQLSuccess, nil
}

//UpdateUserRole update user role
func UpdateUserRole(userID int32, role int32) (int64, error) {
	userModel := UserModel{}
	n, err := userModel.UpdateUserRole(userID, role)
	if nil != err {
		glog.Errorf("update user role fails, userID: %s, role: %s, error: %v", userID, role, err)
		return 0, err
	}
	return n, err
}
