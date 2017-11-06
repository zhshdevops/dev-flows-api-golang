/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package user

import (
	"time"

	"github.com/astaxie/beego/orm"

	sqlstatus "api-server/models/sql/status"
	"api-server/modules/transaction"
	"api-server/util/misc"
	"api-server/util/rand"
	"api-server/util/secure"
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultUserBalance = 100000
)

// UserSpec defines the fields needed to create a new user
type UserSpec struct {
	UserName           string `json:"userName"`
	Password           string `json:"password"`
	OldPassword        string `json:"oldPassword"`
	Namespace          string `json:"namespace"`
	DisplayName        string `json:"displayName"`
	Email              string `json:"email"`
	Phone              string `json:"phone"`
	APIToken           string `json:"apiToken"`
	Role               int32  `json:"role"`
	Avatar             string `json:"avatar"`
	InviteCode         string `json:"inviteCode"`
	AccountType        string `json:"accountType"`
	AccountID          string `json:"accountID"`
	AccountDetail      string `json:"accountDetail"`
	DefaultUserBalance *int64 `json:"defaultUserBalance"`
}

// SpaceModel tenx_team_space record
type SpaceModel struct {
	SpaceID      string    `orm:"column(id);pk"`
	SpaceName    string    `orm:"column(name)"`
	Namespace    string    `orm:"column(namespace)"`
	Description  string    `orm:"column(description)"`
	TeamID       string    `orm:"column(team_id)"`
	CreationTime time.Time `orm:"column(creation_time)"`
}

const (
	UserStatusInactive     int8 = 0
	UserStatusActive       int8 = 1
	UserStatusNeedResetPWD int8 = 2 // not used yet
)

// CreateUser creates a new user and returns its detail
func CreateUser(spec UserSpec, isStandard bool, activated bool) (*User, uint32, error, []string) {
	method := "CreateUser"
	encodedPass := secure.EncodeMD5(spec.Password)
	// if user does not pass api Token, generate one
	apiToken := spec.APIToken
	if len(apiToken) == 0 {
		apiToken = string(rand.Krand(48, 1))
	}

	activeStatus := UserStatusInactive
	if activated {
		activeStatus = UserStatusActive
	}
	um := &UserModel{
		Username:     spec.UserName,
		Password:     encodedPass,
		Namespace:    spec.Namespace,
		Displayname:  spec.DisplayName,
		Email:        spec.Email,
		Phone:        spec.Phone,
		APIToken:     apiToken,
		Role:         spec.Role,
		CreationTime: time.Now(),
		Active:       activeStatus,
		Type:         1, // database user by default
	}

	userAccModel := &UserAccountModel{
		Namespace:           um.Namespace,
		Balance:             0.0,
		LastCost:            0,
		LastChargeAdminID:   0,
		LastChargeAdminName: "",
		LastChargeAmount:    0,
		StorageLimit:        0,
	}
	var errcode uint32
	var err error
	dupItems := make([]string, 0, 1)
	trans := transaction.New()
	trans.Do(func() {
		// lock
		sql := `SELECT user_name, email, phone FROM tenx_users WHERE user_name = ? OR email = ? `
		userModel := make([]UserModel, 0, 1)
		if um.Phone != "" {
			sql += `OR phone = ? FOR UPDATE`
			_, err = trans.O().Raw(sql, um.Username, um.Email, um.Phone).QueryRows(&userModel)
		} else {
			sql += `FOR UPDATE`
			_, err = trans.O().Raw(sql, um.Username, um.Email).QueryRows(&userModel)
		}
		if err != orm.ErrNoRows && err != nil {
			// db error
			trans.Rollback(method, "select for update failed.", err)
		} else if err == nil && len(userModel) > 0 {
			// duplicate
			// get duplicate item
			for _, user := range userModel {
				if user.Username == um.Username {
					dupItems = append(dupItems, "username")
				}
				if user.Email == um.Email {
					dupItems = append(dupItems, "email")
				}
				if (user.Phone != "") && (user.Phone == um.Phone) {
					dupItems = append(dupItems, "phone")
				}
			}
			dupItems = misc.RemoveDuplication(dupItems)
			err = errors.New("duplicate items")
			trans.Rollback(method, "duplicate items", strings.Join(dupItems, ","))
		}
	}).Do(func() {
		sql := `SELECT namespace FROM tenx_team_space WHERE namespace = ? FOR UPDATE `
		spaceModels := make([]SpaceModel, 0, 1)
		_, err = trans.O().Raw(sql, um.Namespace).QueryRows(&spaceModels)
		if err != orm.ErrNoRows && err != nil {
			//db error
			trans.Rollback(method, "select teamspace for update failed.", err)
		} else if err == nil && len(spaceModels) > 0 {
			//the namespace already exists
			err = fmt.Errorf("the namespace %s already exists for teamspace", um.Namespace)
			trans.Rollback(method, fmt.Sprintf("the namespace %s already exists for teamspace", um.Namespace))
		}
	}).Do(func() {
		errcode, err = um.Insert(trans.O())
		if err != nil {
			trans.Rollback(fmt.Sprintf("insert %#v fails, error code:%d, error:%v\n", um, errcode, err))
		}
	}).Do(func() {
		userAccModel.UserID = um.UserID
		if !isStandard {
			var defaultUserBalance int64 = DefaultUserBalance
			if nil != spec.DefaultUserBalance {
				defaultUserBalance = *spec.DefaultUserBalance
			}
			// For pro/lite edition, set default balance
			userAccModel.Balance = defaultUserBalance
		}
		errcode, err = userAccModel.Insert(trans.O())
		if err != nil {
			trans.Rollback(fmt.Sprintf("create user account %d %s fails, error code:%d, error:%v\n", userAccModel.UserID, userAccModel.Namespace, errcode, err))
		}
	}).Done()
	if !trans.IsCommit() {
		return nil, errcode, err, dupItems
	}

	return ToUser(um, userAccModel), sqlstatus.SQLSuccess, nil, nil
}

// UsernameExist check the user existence
func UsernameExist(name string) bool {
	if !isValidUserName(name) {
		return true
	}

	return orm.NewOrm().QueryTable("tenx_users").Filter("user_name", name).Exist() ||
		orm.NewOrm().QueryTable("tenx_team_space").Filter("namespace", name).Exist()
}

func EmailExist(email string) bool {
	return orm.NewOrm().QueryTable("tenx_users").Filter("email", email).Exist()
}

func PhoneExist(phone string) bool {
	return orm.NewOrm().QueryTable("tenx_users").Filter("phone", phone).Exist()
}

func isValidUserName(userName string) bool {
	switch userName {
	case
		"admin",
		"library",
		"tenxcloud",
		"service",
		"master",
		"administrator",
		"kube-system",
		"kubelet",
		"user-foo":
		return false
	}
	return true
}
