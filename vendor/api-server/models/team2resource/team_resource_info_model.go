/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team2resource

import (
	"fmt"
	"strings"

	sqlstatus "api-server/models/sql/status"

	"github.com/astaxie/beego/orm"

	"api-server/models/common"
)

// TeamResourceInfoModel defines the summery information of resources in a team
type TeamResourceInfoModel struct {
	TeamID       string `orm:"column(team_id);pk"`
	ClusterCount int    `orm:"column(cluster_count)"`
	// StorageCount int    `orm:"column(storage_count)"`  // a way to extend resource kind in this struct (advice)
}

// ListByTeams consumes a list of team IDs, produces resource summery information
// SortBy: cluster_count
func (tc *TeamResourceInfoModel) ListByTeams(teamIDs []string, dataselect *common.DataSelectQuery) ([]TeamResourceInfoModel, uint32, error) {
	if len(teamIDs) == 0 {
		return []TeamResourceInfoModel{}, sqlstatus.SQLSuccess, nil
	}

	o := orm.NewOrm()

	teamIDStr := fmt.Sprintf("'%s'", strings.Join(teamIDs, "','"))
	// cluster number
	sql := fmt.Sprintf(`SELECT 
    tenx_teams.id as team_id, COUNT(resource_id) AS cluster_count
FROM
	tenx_teams
LEFT JOIN
    tenx_team_resource_ref on tenx_team_resource_ref.team_id = tenx_teams.id and tenx_team_resource_ref.resource_type = ?
WHERE
    tenx_teams.id IN (%s) 
GROUP BY tenx_teams.id %s %s;`, teamIDStr, dataselect.SortQuery, dataselect.PaginationQuery)
	var clusterInfos []TeamResourceInfoModel
	_, err := o.Raw(sql, int(ResourceCluster)).QueryRows(&clusterInfos)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}

	// if you want to support StorageNumber, use a new sql statement, like the following
	// 	sql = fmt.Sprintf(`SELECT
	//     team_id, COUNT(resource_id) AS storage_count
	// FROM
	//     tenx_team_resource_ref
	// WHERE
	//     team_id IN (%s)
	//         AND resource_type = ?;`, teamIDStr)
	// 	var storageInfos []TeamResourceInfoModel
	// 	_, err = o.Raw(sql, int(ResourceStorage)).QueryRows(&storageInfos)
	// 	if err != nil {
	// 		errCode, err := sqlstatus.ParseErrorCode(err)
	// 		return nil, errCode, err
	// 	}
	//  merge the results with the clusterInfos

	errCode, err := sqlstatus.ParseErrorCode(err)
	return clusterInfos, errCode, err
}

// GetByTeamID consumes a team ID, produces resource summery information
func (tc *TeamResourceInfoModel) GetByTeamID(teamID string) (uint32, error) {
	o := orm.NewOrm()

	tc.TeamID = teamID

	clusterInfo := TeamResourceInfoModel{}
	// cluster number
	sql := `SELECT 
    team_id, COUNT(resource_id) AS cluster_count
FROM
    tenx_team_resource_ref
WHERE
    team_id = ?
        AND resource_type = ?;`
	err := o.Raw(sql, teamID, int(ResourceCluster)).QueryRow(&clusterInfo)
	if err != nil {
		return sqlstatus.ParseErrorCode(err)
	}

	// if you want to support StorageNumber, use a new sql statement, like the following
	// 	sql = fmt.Sprintf(`SELECT
	//     team_id, COUNT(resource_id) AS storage_number
	// FROM
	//     tenx_team_resource_ref
	// WHERE
	//     team_id = ?
	//         AND resource_type = ?;`, teamID)
	// 	storageInfos := TeamResourceInfoModel{}
	// 	_, err = o.Raw(sql, teamID,  int(ResourceStorage)).QueryRows(&storageInfo)
	// 	if err != nil {
	// 		return sqlstatus.ParseErrorCode(err)
	// 	}
	//  merge the results with the clusterInfo
	tc.ClusterCount = clusterInfo.ClusterCount
	return sqlstatus.SQLSuccess, nil
}
