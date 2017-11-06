/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package teamspace

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"

	sqlstatus "dev-flows-api-golang/models/sql/status"
	"dev-flows-api-golang/models/user"
	idfactory "dev-flows-api-golang/util/uuid"
	"dev-flows-api-golang/modules/transaction"
)

const (
	DefaultUserBalance = 100000
)

// SpaceSpec defines the fields needed to create a new space
type SpaceSpec struct {
	SpaceName   string `json:"spaceName"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

// generate a valid namespace
func generateNamespace(strlen int) (string, uint32, error) {
	spaceModel := SpaceModel{}
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	for i := 0; i < 5; i++ {
		rand.Seed(time.Now().UTC().UnixNano())
		result := make([]byte, strlen)
		for i := 0; i < strlen; i++ {
			result[i] = chars[rand.Intn(len(chars))]
		}
		namespace := "team-" + string(result)
		alreadyExist, errcode, err := spaceModel.CheckNamespace(namespace)
		if err != nil {
			glog.Errorf("check namespace %s fails, error code:%d, error:%v\n", namespace, errcode, err)
			return "", errcode, err
		}
		if alreadyExist == false {
			return namespace, sqlstatus.SQLSuccess, nil
		}
	}
	glog.Errorf("generate namespace fails, already tried 5 times\n")
	return "", sqlstatus.SQLErrUnCategoried, fmt.Errorf("generate namespace fails, have tried too many times")
}

// CreateSpace creates a new space and returns its detail
func CreateSpace(teamID string, isStandard bool, spec SpaceSpec) (*Space, uint32, error) {
	method := "CreateSpace"

	spaceModel := &SpaceModel{
		SpaceID:      idfactory.NewTeamspace(),
		TeamID:       teamID,
		SpaceName:    spec.SpaceName,
		Namespace:    spec.SpaceName,
		Description:  spec.Description,
		CreationTime: time.Now(),
	}

	spaceAccModel := &SpaceAccountModel{
		SpaceID:             spaceModel.SpaceID,
		Namespace:           spaceModel.Namespace,
		Balance:             0.0,
		LastCost:            0,
		LastChargeAdminID:   0,
		LastChargeAdminName: "",
		LastChargeAmount:    0,
		StorageLimit:        0,
	}

	var errcode uint32
	var err error
	trans := transaction.New()
	trans.Do(func() {
		sql := `SELECT namespace FROM tenx_users WHERE namespace = ? FOR UPDATE `
		userModels := make([]user.UserModel, 0, 1)
		_, err = trans.O().Raw(sql, spaceModel.Namespace).QueryRows(&userModels)
		if err != orm.ErrNoRows && err != nil {
			//db error
			trans.Rollback(method, "select user for update failed.", err)
		} else if err == nil && len(userModels) > 0 {
			//the namespace already exists
			err = fmt.Errorf("the namespace %s already exists for user", spaceModel.Namespace)
			trans.Rollback(method, fmt.Sprintf("the namespace %s already exists for user", spaceModel.Namespace))
		}
	}).Do(func() {
		sql := `SELECT namespace FROM tenx_team_space WHERE namespace = ? FOR UPDATE `
		spaceModels := make([]SpaceModel, 0, 1)
		_, err = trans.O().Raw(sql, spaceModel.Namespace).QueryRows(&spaceModels)
		if err != orm.ErrNoRows && err != nil {
			//db error
			trans.Rollback(method, "select teamspace for update failed.", err)
		} else if err == nil && len(spaceModels) > 0 {
			//the namespace already exists
			err = fmt.Errorf("the namespace %s already exists for teamspace", spaceModel.Namespace)
			trans.Rollback(method, fmt.Sprintf("the namespace %s already exists for teamspace", spaceModel.Namespace))
		}
	}).Do(func() {
		if errcode, err = spaceModel.Insert(trans.O()); err != nil {
			trans.Rollback(method, fmt.Sprintf("insert %#v fails, error code:%d, error:%v\n", spaceModel, errcode, err))
		}
	}).Do(func() {
		if !isStandard {
			// For pro/lite edition, set default balance
			spaceAccModel.Balance = DefaultUserBalance
		}
		if errcode, err = spaceAccModel.Insert(trans.O()); err != nil {
			trans.Rollback(method, fmt.Sprintf("create space account %d %s fails, error code:%d, error:%v\n", spaceAccModel.SpaceID, spaceAccModel.Namespace, errcode, err))
		}
	}).Done()

	if !trans.IsCommit() {
		return nil, errcode, err
	}

	return ToSpace(spaceModel, spaceAccModel, nil), sqlstatus.SQLSuccess, nil
}

//Exist check teamspace existence
func Exist(name string) bool {
	o := orm.NewOrm()
	return o.QueryTable("tenx_team_space").Filter("name", name).Exist()
}
