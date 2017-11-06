/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team2user

import (
	"fmt"
	"strconv"
	"strings"

	sqlstatus "dev-flows-api-golang/models/sql/status"

	"github.com/astaxie/beego/orm"

	"dev-flows-api-golang/models/common"
)

const (
	SortByTeamCount  = "team_count"
	SortBySpaceCount = "space_count"
)

// UserInfoModel defines the summery information of resources in a team
type UserInfoModel struct {
	UserID     int32 `orm:"column(user_id);pk"`
	TeamCount  int   `orm:"column(team_count)"`
	SpaceCount int   `orm:"column(space_count)"`
}

// ListUserInfos lists all user infos
func (u *UserInfoModel) ListUserInfos(dataselect *common.DataSelectQuery) ([]UserInfoModel, uint32, error) {
	o := orm.NewOrm()
	if dataselect == nil {
		dataselect = common.DefaultDataSelect
	}
	sql := fmt.Sprintf(`SELECT 
    tenx_users.user_id AS user_id,
    COUNT(tenx_team_user_ref.team_id) AS team_count,
    COUNT(tenx_space_user_ref.teamspace_id) AS space_count
FROM
    tenx_users
        LEFT JOIN
    tenx_team_user_ref ON tenx_team_user_ref.user_id = tenx_users.user_id
        LEFT JOIN
    tenx_space_user_ref ON tenx_space_user_ref.user_id = tenx_users.user_id
WHERE %s
GROUP BY user_id %s %s;`, strings.Replace(dataselect.FilterQuery.String(), "role", "tenx_users.role", -1), dataselect.SortQuery, dataselect.PaginationQuery)
	var userInfoModels []UserInfoModel
	_, err := o.Raw(sql).QueryRows(&userInfoModels)
	if err != nil {
		errcode, err := sqlstatus.ParseErrorCode(err)
		return nil, errcode, err
	}
	return userInfoModels, sqlstatus.SQLSuccess, nil
}

// ListUserInfoByIDs lists all user's infos in the userIDs list
func (u *UserInfoModel) ListUserInfoByIDs(userIDs []int32, dataselect *common.DataSelectQuery) ([]UserInfoModel, uint32, error) {
	if len(userIDs) == 0 {
		return []UserInfoModel{}, sqlstatus.SQLSuccess, nil
	}

	var userIDStrs []string
	for _, userID := range userIDs {
		userIDStrs = append(userIDStrs, strconv.Itoa(int(userID)))
	}

	o := orm.NewOrm()
	if dataselect == nil {
		dataselect = common.NoDataSelect
	}

	sql := fmt.Sprintf(`SELECT 
    tenx_users.user_id AS user_id,
    COUNT(tenx_team_user_ref.team_id) AS team_count,
    COUNT(tenx_space_user_ref.teamspace_id) AS space_count
FROM
    tenx_users
        LEFT JOIN
    tenx_team_user_ref ON tenx_team_user_ref.user_id = tenx_users.user_id
        LEFT JOIN
    tenx_space_user_ref ON tenx_space_user_ref.user_id = tenx_users.user_id
WHERE
    tenx_users.user_id IN (%s)
GROUP BY user_id %s %s;`, strings.Join(userIDStrs, ","), dataselect.SortQuery, dataselect.SortQuery)
	var userInfoModels []UserInfoModel
	_, err := o.Raw(sql).QueryRows(&userInfoModels)
	if err != nil {
		errcode, err := sqlstatus.ParseErrorCode(err)
		return nil, errcode, err
	}
	return userInfoModels, sqlstatus.SQLSuccess, nil
}
