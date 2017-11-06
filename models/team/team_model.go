/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team

import (
	"fmt"
	"strings"
	"time"

	sqlstatus "dev-flows-api-golang/models/sql/status"

	"github.com/astaxie/beego/orm"

	"dev-flows-api-golang/models/common"

	"dev-flows-api-golang/util/misc"

	"github.com/golang/glog"
)

// TeamModel tenx_users record
type TeamModel struct {
	TeamID       string    `orm:"column(id);pk"`
	TeamName     string    `orm:"column(name);size(200)"`
	Description  string    `orm:"column(description)"`
	CreatorID    int32     `orm:"column(creator_id)"`
	CreationTime time.Time `orm:"column(creation_time)"`
}

// TableName return tenx_users
func (t *TeamModel) TableName() string {
	return "tenx_teams"
}

func NewTeamModel() *TeamModel {
	return &TeamModel{}
}

// Insert add a new team record to db
func (t *TeamModel) Insert(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	_, err := o.Insert(t)
	return sqlstatus.ParseErrorCode(err)
}

// Get fetch a user record by id
func (t *TeamModel) Get() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where id=?;", t.TableName())
	err := o.Raw(sql, t.TeamID).QueryRow(t)
	return sqlstatus.ParseErrorCode(err)
}

// GetByName get user by name
func (t *TeamModel) GetByName(name string) (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where name=?;", t.TableName())
	err := o.Raw(sql, name).QueryRow(t)
	return sqlstatus.ParseErrorCode(err)
}

// Update updates all fields of the team
func (t *TeamModel) Update() (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Update(t)
	return sqlstatus.ParseErrorCode(err)
}

// Delete fetch a user record by id
func (t *TeamModel) Delete(teamID string, orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	_, err := o.QueryTable(t.TableName()).Filter("id", teamID).Delete()
	return sqlstatus.ParseErrorCode(err)
}

// ListIDs list all team ids, according to filter
func (t *TeamModel) ListIDs(creatorId int32, dataselect *common.DataSelectQuery) ([]string, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT 
    id
FROM
    %s where creator_id = %d and %s;`, t.TableName(), creatorId, dataselect.FilterQuery)
	var teamModels []TeamModel
	_, err := o.Raw(sql).QueryRows(&teamModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}

	if len(teamModels) == 0 {
		return []string{}, sqlstatus.SQLSuccess, nil
	}

	// get the user IDs
	var teamIDStr []string
	for _, um := range teamModels {
		teamIDStr = append(teamIDStr, um.TeamID)
	}
	return teamIDStr, sqlstatus.SQLSuccess, nil
}

// ListIDs list all team ids, according to filter
func (t *TeamModel) ListAllIDs() ([]string, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT 
    id
FROM
    %s;`, t.TableName())
	var teamModels []TeamModel
	_, err := o.Raw(sql).QueryRows(&teamModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}

	if len(teamModels) == 0 {
		return []string{}, sqlstatus.SQLSuccess, nil
	}

	// get the user IDs
	var teamIDStr []string
	for _, um := range teamModels {
		teamIDStr = append(teamIDStr, um.TeamID)
	}
	return teamIDStr, sqlstatus.SQLSuccess, nil
}

// List lists all teams with pagination
func (t *TeamModel) List(dataselect *common.DataSelectQuery) ([]TeamModel, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT 
    id
FROM
    %s where %s
%s %s;`, t.TableName(), dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var teamModels []TeamModel
	_, err := o.Raw(sql).QueryRows(&teamModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}

	if len(teamModels) == 0 {
		return []TeamModel{}, sqlstatus.SQLSuccess, nil
	}

	// get the user IDs
	var teamIDStr []string
	for _, um := range teamModels {
		teamIDStr = append(teamIDStr, fmt.Sprintf("'%s'", um.TeamID))
	}

	sql = fmt.Sprintf(`SELECT 
    *
FROM
    %s
WHERE
    id IN (%s) %s;`, t.TableName(), strings.Join(teamIDStr, ","), dataselect.SortQuery)
	_, err = o.Raw(sql).QueryRows(&teamModels)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return teamModels, errCode, err
}

// Count counts all teams
func (t *TeamModel) Count(dataselect *common.DataSelectQuery) (int, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT 
    count(*) as total
FROM
    %s WHERE %s;`, t.TableName(), dataselect.FilterQuery)
	meta := &common.ListMeta{}
	err := o.Raw(sql).QueryRow(meta)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return 0, errCode, err
	}
	return meta.Total, sqlstatus.SQLSuccess, nil
}

// CountByCreator by creator_id
func (t *TeamModel) CountByCreator() (int, error) {
	o := orm.NewOrm()

	meta := &common.ListMeta{}
	err := o.Raw(fmt.Sprintf(`SELECT count(*) as total FROM %s WHERE creator_id = ?`, t.TableName()), t.CreatorID).QueryRow(meta)
	if err != nil {
		return 0, err
	}
	return meta.Total, nil
}

// ListByIDs lists all teams by team IDs
func (t *TeamModel) ListByIDs(teamIDs []string, dataselect *common.DataSelectQuery) ([]TeamModel, uint32, error) {
	if len(teamIDs) == 0 {
		return []TeamModel{}, sqlstatus.SQLSuccess, nil
	}
	o := orm.NewOrm()

	teamIDStr := fmt.Sprintf("'%s'", strings.Join(teamIDs, "','"))
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE id IN (%s) and %s %s %s;`, t.TableName(), teamIDStr,
		dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var teamModels []TeamModel
	_, err := o.Raw(sql).QueryRows(&teamModels)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return teamModels, errCode, err
}

type TeamInfo struct {
	TeamID       string    `orm:"column(id)" json:"id"`
	Name         string    `json:"name"`
	CreationTime time.Time `orm:"column(creation_time)" json:"creationTime"`
	Role         int       `json:"-"`
	CreatorID    int32     `orm:"column(creator_id)" json:"-"`
	Balance      int64     `json:"balance"`
	MemberCount  int       `orm:"column(membercount)" json:"memberCount"`
	IsCreator    bool      `json:"isCreator"`
	IsAdmin      bool      `json:"isAdmin"`
}

func (t *TeamModel) GetUserAllTeam(userID int32) ([]TeamInfo, error) {
	method := "GetUserAllTeam"
	sql := `SELECT t1.id, t1.name, t1.creation_time, t1.creator_id, t2.role
	FROM tenx_teams AS t1 INNER JOIN tenx_team_user_ref AS t2
	ON t1.id = t2.team_id
	WHERE user_id = ?
	ORDER BY t2.role DESC`
	o := orm.NewOrm()
	baseInfo := make([]TeamInfo, 0, 1)
	if _, err := o.Raw(sql, userID).QueryRows(&baseInfo); err != nil {
		glog.Errorln(method, "get team base info failed.", err)
		return nil, err
	}
	if len(baseInfo) == 0 {
		return baseInfo, nil
	}

	teamIDList := make([]string, 0, len(baseInfo))
	for i := range baseInfo {
		teamIDList = append(teamIDList, baseInfo[i].TeamID)
	}

	teamIDStr := misc.SliceToString(teamIDList, "'", ",")

	// get team users number
	sql = `SELECT COUNT(*) AS membercount,team_id as id
	FROM tenx_team_user_ref
	WHERE team_id IN (` + teamIDStr + `)
	GROUP BY team_id`
	memberCountInfo := make([]TeamInfo, 0, 1)
	if _, err := o.Raw(sql).QueryRows(&memberCountInfo); err != nil {
		glog.Errorln(method, "get member count info failed.", err)
		return nil, err
	}

	// get balance info
	sql = `SELECT SUM(balance) as balance,t1.team_id as id
	FROM tenx_team_space AS t1 INNER JOIN tenx_team_space_account AS t2
	ON t1.id = t2.teamspace_id
	WHERE t1.team_id IN (` + teamIDStr + `)
	GROUP BY t1.team_id`
	balanceInfo := make([]TeamInfo, 0, 1)
	if _, err := o.Raw(sql).QueryRows(&balanceInfo); err != nil {
		glog.Errorln(method, "get balance info failed.", err)
		return nil, err
	}

	for i := range baseInfo {
		for _, member := range memberCountInfo {
			if member.TeamID == baseInfo[i].TeamID {
				baseInfo[i].MemberCount = member.MemberCount
				break
			}
		}
		for _, balance := range balanceInfo {
			if balance.TeamID == baseInfo[i].TeamID {
				baseInfo[i].Balance = balance.Balance
				break
			}
		}
		baseInfo[i].IsCreator = (baseInfo[i].CreatorID == userID)
		baseInfo[i].IsAdmin = (baseInfo[i].Role == 1 || baseInfo[i].Role == 2)
	}
	return baseInfo, nil
}
