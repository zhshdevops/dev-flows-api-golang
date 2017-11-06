/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package user

import (
	"github.com/golang/glog"

	sqlstatus "api-server/models/sql/status"
	"api-server/util/secure"

	"github.com/astaxie/beego/orm"
)

// UpdateUser update the current user and returns its detail
func UpdateUser(userID int32, spec UserSpec) (*User, uint32, error) {
	um := &UserModel{UserID: userID}
	errcode, err := um.Get()
	if err != nil {
		glog.Errorf("update user %d fails, errocode: %d, error:%v\n", userID, errcode, err)
		return nil, errcode, err
	}

	// check every fields
	if len(spec.Password) > 0 {
		um.Password = secure.EncodeMD5(spec.Password)
	}
	if len(spec.DisplayName) > 0 {
		um.Displayname = spec.DisplayName
	}
	if len(spec.Email) > 0 {
		um.Email = spec.Email
	}
	if len(spec.Phone) > 0 {
		um.Phone = spec.Phone
	}
	if len(spec.APIToken) > 0 {
		um.APIToken = spec.APIToken
	}
	if len(spec.Avatar) > 0 {
		um.Avatar = spec.Avatar
	}
	// Update role
	// spec.Role should be actual_role + 1
	if spec.Role > 0 {
		um.Role = spec.Role - 1
	}
	if errcode, err := um.Update(); err != nil {
		glog.Errorf("update %#v fails, error code:%d, error:%v\n", um, errcode, err)
		return nil, errcode, err
	}

	userAccModel := &UserAccountModel{UserID: userID}
	if errcode, err := userAccModel.Get(); err != nil {
		glog.Errorf("get user %d %s's account information fails, error code:%d, error:%v\n", um.UserID, um.Username, errcode, err)
		return nil, errcode, err
	}

	return ToUser(um, userAccModel), sqlstatus.SQLSuccess, nil
}

// UpdatePasswordByEmail update the user password by its email
func UpdatePasswordByEmail(email string, password string) error {
	sql := `UPDATE tenx_users SET password=? WHERE email=? `
	_, err := orm.NewOrm().Raw(sql, secure.EncodeMD5(password), email).Exec()
	return err
}

// SetAdminPassword set the admin user password
func SetAdminPassword(password string) error {
	sql := `UPDATE tenx_users SET password=?, active=1 WHERE user_name=? `
	_, err := orm.NewOrm().Raw(sql, secure.EncodeMD5(password), SuperUserName).Exec()
	return err
}
