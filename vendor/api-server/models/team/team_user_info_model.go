/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team

import (
	"fmt"
	"strings"

	sqlstatus "api-server/models/sql/status"

	"github.com/astaxie/beego/orm"

	"api-server/models/common"
)

// TeamUserInfoModel calculate team summary information
type TeamUserInfoModel struct {
	TeamID    string `orm:"column(team_id);pk"`
	UserCount int    `orm:"column(user_count)"`
}

// List lists TeamInfoDetails by certain order
func (ti *TeamUserInfoModel) List(dataselect *common.DataSelectQuery) ([]TeamUserInfoModel, uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf(`SELECT 
    tenx_teams.id AS team_id,
    COUNT(tenx_team_user_ref.user_id) AS user_count
FROM
    tenx_teams
        LEFT JOIN
    tenx_team_user_ref ON tenx_team_user_ref.team_id = tenx_teams.id
WHERE %s
GROUP BY tenx_teams.id
%s %s;`, dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)

	var teaminfolist []TeamUserInfoModel
	_, err := o.Raw(sql).QueryRows(&teaminfolist)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return teaminfolist, errcode, err
}

// ListByTeamIDs lists TeamInfoDetails by certain order
func (ti *TeamUserInfoModel) ListByTeamIDs(teamIDs []string, dataselect *common.DataSelectQuery) ([]TeamUserInfoModel, uint32, error) {
	o := orm.NewOrm()

	teamIDStr := fmt.Sprintf("'%s'", strings.Join(teamIDs, "','"))
	sql := fmt.Sprintf(`SELECT 
    tenx_teams.id AS team_id,
    COUNT(tenx_team_user_ref.user_id) AS user_count
FROM
    tenx_teams
        LEFT JOIN
    tenx_team_user_ref ON tenx_team_user_ref.team_id = tenx_teams.id
    where tenx_teams.id in (%s)
GROUP BY tenx_teams.id
%s %s;`, teamIDStr, dataselect.SortQuery, dataselect.PaginationQuery)

	var teaminfolist []TeamUserInfoModel
	_, err := o.Raw(sql).QueryRows(&teaminfolist)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return teaminfolist, errcode, err
}

// Get returns team summery info
func (ti *TeamUserInfoModel) Get(teamID string) (uint32, error) {
	o := orm.NewOrm()

	ti.TeamID = teamID
	count, err := o.QueryTable("tenx_team_user_ref").Filter("team_id", teamID).Count()
	ti.UserCount = int(count)

	return sqlstatus.ParseErrorCode(err)
}
