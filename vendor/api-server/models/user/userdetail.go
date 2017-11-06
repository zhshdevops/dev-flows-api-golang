/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package user

import (
	"github.com/golang/glog"

	sqlstatus "api-server/models/sql/status"
)

// GetUserDetail returns a user's details from user id
func GetUserDetail(userID int32) (*User, uint32, error) {
	um := &UserModel{UserID: userID}
	// TODO: use one SQL to get all information
	if errcode, err := um.Get(); err != nil {
		glog.Errorf("get user %d fails, error code:%d, error:%v\n", userID, errcode, err)
		return nil, errcode, err
	}
	ua := &UserAccountModel{UserID: userID}
	if errcode, err := ua.Get(); err != nil {
		glog.Errorf("get user account %d fails, error code:%d, error:%v\n", userID, errcode, err)
		return ToUser(um, nil), errcode, err
	}

	return ToUser(um, ua), sqlstatus.SQLSuccess, nil
}

// GetUserDetailByUsernameAndPassword returns a user's details from user name
func GetUserDetailByUsernameAndPassword(username string, password string) (*User, uint32, error) {
	um := &UserModel{Username: username}
	// TODO: use one SQL to get all information
	if errcode, err := um.GetByUsernameAndPassword(password); err != nil {
		glog.Errorf("get user %s fails, error code:%d, error:%v\n", username, errcode, err)
		return nil, errcode, err
	}
	ua := &UserAccountModel{UserID: um.UserID}
	if errcode, err := ua.Get(); err != nil {
		glog.Errorf("get user account %d fails, error code:%d, error:%v\n", um.UserID, errcode, err)
		return ToUser(um, nil), errcode, err
	}
	return ToUser(um, ua), sqlstatus.SQLSuccess, nil
}

// GetUserDetailByEmailAndPassword returns a user's details from user email
func GetUserDetailByEmailAndPassword(email string, password string) (*User, uint32, error) {
	um := &UserModel{Email: email}
	if errcode, err := um.GetByEmailAndPassword(password); err != nil {
		glog.Errorf("get user %s fails, error code:%d, error:%v\n", email, errcode, err)
		return nil, errcode, err
	}
	ua := &UserAccountModel{UserID: um.UserID}
	if errcode, err := ua.Get(); err != nil {
		glog.Errorf("get user account %d fails, error code:%d, error:%v\n", um.UserID, errcode, err)
		return ToUser(um, nil), errcode, err
	}
	return ToUser(um, ua), sqlstatus.SQLSuccess, nil
}

// GetUserDetailByThirdPartyAccount returns a user's details by third party account info
func GetUserDetailByThirdPartyAccount(accountType, accountID string) (*User, uint32, error) {
	tpa := &ThirdPartyAccountModel{AccountType: accountType, AccountID: accountID}
	if errcode, err := tpa.GetByAccountTypeAndID(); err != nil {
		glog.Errorf("get third party account fails, account type: %s, account id: %s, error code: %d, error:%v\n",
			tpa.AccountType, tpa.AccountID, errcode, err)
		return nil, errcode, err
	}

	um := &UserModel{}
	if errcode, err := um.GetByNamespace(tpa.Namespace); err != nil {
		glog.Errorf("get user %s fails, error code:%d, error:%v\n", tpa.Namespace, errcode, err)
		return nil, errcode, err
	}

	ua := &UserAccountModel{UserID: um.UserID}
	if errcode, err := ua.Get(); err != nil {
		glog.Errorf("get user account %d fails, error code:%d, error:%v\n", um.UserID, errcode, err)
		return ToUser(um, nil), errcode, err
	}
	return ToUser(um, ua), sqlstatus.SQLSuccess, nil
}
