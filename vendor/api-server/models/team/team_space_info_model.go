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

// TeamSpaceInfoModel calculate team summary information
type TeamSpaceInfoModel struct {
	TeamID     string `orm:"column(team_id);pk"`
	SpaceCount int    `orm:"column(space_count)"`
}

// List lists TeamInfoDetails by certain order
func (ti *TeamSpaceInfoModel) List(dataselect *common.DataSelectQuery) ([]TeamSpaceInfoModel, uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf(`SELECT 
    tenx_teams.id AS team_id,
    COUNT(tenx_team_space.id) AS space_count
FROM
    tenx_teams
        LEFT JOIN
    tenx_team_space ON tenx_team_space.team_id = tenx_teams.id
WHERE %s
GROUP BY tenx_teams.id
%s %s;`, dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)

	var teaminfolist []TeamSpaceInfoModel
	_, err := o.Raw(sql).QueryRows(&teaminfolist)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return teaminfolist, errcode, err
}

// ListByTeamIDs lists TeamInfoDetails by certain order
func (ti *TeamSpaceInfoModel) ListByTeamIDs(teamIDs []string, dataselect *common.DataSelectQuery) ([]TeamSpaceInfoModel, uint32, error) {
	o := orm.NewOrm()

	teamIDStr := fmt.Sprintf("'%s'", strings.Join(teamIDs, "','"))
	sql := fmt.Sprintf(`SELECT 
    tenx_teams.id AS team_id,
    COUNT(tenx_team_space.id) AS space_count
FROM
    tenx_teams
        LEFT JOIN
    tenx_team_space ON tenx_team_space.team_id = tenx_teams.id
    where tenx_teams.id in (%s)
GROUP BY tenx_teams.id
%s %s;`, teamIDStr, dataselect.SortQuery, dataselect.PaginationQuery)

	var teaminfolist []TeamSpaceInfoModel
	_, err := o.Raw(sql).QueryRows(&teaminfolist)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return teaminfolist, errcode, err
}

// GetTeamInfoByNamespace get team info by specified space id
func (ti *TeamSpaceInfoModel) GetTeamInfoByNamespace(namespace string) (string, string, string, error) {
	o := orm.NewOrm()

	sql := `SELECT 
    tenx_teams.id AS team_id, tenx_teams.name as team_name, tenx_team_space.id as space_id
FROM
    tenx_teams
        LEFT JOIN
    tenx_team_space ON tenx_team_space.team_id = tenx_teams.id
    where tenx_team_space.namespace = ?;`
	var teamID, teamName, spaceID string
	err := o.Raw(sql, namespace).QueryRow(&teamID, &teamName, &spaceID)
	return teamID, teamName, spaceID, err
}

// Get returns team summery info
func (ti *TeamSpaceInfoModel) Get(teamID string) (uint32, error) {
	o := orm.NewOrm()

	ti.TeamID = teamID
	count, err := o.QueryTable("tenx_team_space").Filter("team_id", teamID).Count()
	if err != nil {
		return sqlstatus.ParseErrorCode(err)
	}
	ti.SpaceCount = int(count)

	return sqlstatus.ParseErrorCode(err)
}
