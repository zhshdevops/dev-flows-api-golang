/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package teamspace

import (
	"fmt"
	"strings"
	"time"

	sqlstatus "api-server/models/sql/status"

	"github.com/astaxie/beego/orm"

	"api-server/models/common"
)

// SpaceModel tenx_users record
type SpaceModel struct {
	SpaceID      string    `orm:"column(id);pk"`
	SpaceName    string    `orm:"column(name)"`
	Namespace    string    `orm:"column(namespace)"`
	Description  string    `orm:"column(description)"`
	TeamID       string    `orm:"column(team_id)"`
	CreationTime time.Time `orm:"column(creation_time)"`
}

func (s SpaceModel) TableName() string {
	return "tenx_team_space"
}
func NewSpaceModel() *SpaceModel {
	return &SpaceModel{}
}
func (s *SpaceModel) ListAll() ([]SpaceModel, error) {
	o := orm.NewOrm()
	var results []SpaceModel

	_, err := o.QueryTable(s.TableName()).All(&results)
	return results, err
}
// ListSpaceIDsByTeams returns a list of team spaces according to team ids
func (s *SpaceModel) ListSpaceIDsByTeams(teamIDs []string, dataselect *common.DataSelectQuery) ([]string, uint32, error) {
	if len(teamIDs) == 0 {
		return []string{}, sqlstatus.SQLSuccess, nil
	}
	o := orm.NewOrm()

	var teamIDStrs []string
	for _, teamID := range teamIDs {
		teamIDStrs = append(teamIDStrs, fmt.Sprintf("'%s'", teamID))
	}
	sql := fmt.Sprintf(`SELECT id FROM %s WHERE team_id in (%s) and %s;`, s.TableName(), strings.Join(teamIDStrs, ","), dataselect.FilterQuery)
	var spaceModels []SpaceModel
	_, err := o.Raw(sql).QueryRows(&spaceModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}

	if len(spaceModels) == 0 {
		return []string{}, sqlstatus.SQLSuccess, nil
	}

	// get the user IDs
	var spaceIDs []string
	for _, um := range spaceModels {
		spaceIDs = append(spaceIDs, um.SpaceID)
	}
	return spaceIDs, sqlstatus.SQLSuccess, nil
}

// ListByTeams returns a list of team spaces according to team ids
func (s *SpaceModel) ListByTeams(teamIDs []string, dataselect *common.DataSelectQuery) ([]SpaceModel, uint32, error) {
	if len(teamIDs) == 0 {
		return []SpaceModel{}, sqlstatus.SQLSuccess, nil
	}

	o := orm.NewOrm()

	var teamIDStrs []string
	for _, teamID := range teamIDs {
		teamIDStrs = append(teamIDStrs, fmt.Sprintf("'%s'", teamID))
	}
	sql := fmt.Sprintf(`SELECT id FROM %s WHERE team_id in (%s) and %s %s %s;`, s.TableName(), strings.Join(teamIDStrs, ","), dataselect.FilterQuery,
		dataselect.SortQuery, dataselect.PaginationQuery)
	var spaceModels []SpaceModel
	_, err := o.Raw(sql).QueryRows(&spaceModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}

	if len(spaceModels) == 0 {
		return []SpaceModel{}, sqlstatus.SQLSuccess, nil
	}

	// get the user IDs
	var spaceIDStr []string
	for _, um := range spaceModels {
		spaceIDStr = append(spaceIDStr, fmt.Sprintf("'%s'", um.SpaceID))
	}

	sql = fmt.Sprintf(`SELECT * FROM %s WHERE id IN (%s);`, s.TableName(), strings.Join(spaceIDStr, ","))
	_, err = o.Raw(sql).QueryRows(&spaceModels)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return spaceModels, errCode, err
}

// CountByTeams counts space number by a list of team ids
func (s *SpaceModel) CountByTeams(teamIDs []string, dataselect *common.DataSelectQuery) (int, uint32, error) {
	o := orm.NewOrm()

	var teamIDStrs []string
	for _, teamID := range teamIDs {
		teamIDStrs = append(teamIDStrs, fmt.Sprintf("'%s'", teamID))
	}
	sql := fmt.Sprintf(`SELECT
    count(*) as total
FROM
    %s
WHERE team_id in (%s) and %s;`, s.TableName(), strings.Join(teamIDStrs, ","), dataselect.FilterQuery)
	meta := &common.ListMeta{}
	err := o.Raw(sql).QueryRow(meta)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return 0, errCode, err
	}
	return meta.Total, sqlstatus.SQLSuccess, nil
}

// ListByIDs returns a list of team spaces according to space ids
func (s *SpaceModel) ListByIDs(spaceIDs []string, dataselect *common.DataSelectQuery) ([]SpaceModel, uint32, error) {
	o := orm.NewOrm()

	// space query ids
	spaceStr := fmt.Sprintf("'%s'", strings.Join(spaceIDs, "','"))
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE id IN (%s) and %s %s %s;`, s.TableName(), spaceStr,
		dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)

	var spaceModels []SpaceModel
	_, err := o.Raw(sql).QueryRows(&spaceModels)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return spaceModels, errCode, err
}

// Get fetch one record from db
func (s *SpaceModel) Get(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	err := o.QueryTable(s.TableName()).Filter("id", s.SpaceID).One(s)
	return sqlstatus.ParseErrorCode(err)
}

// Insert add a new team space
func (s *SpaceModel) Insert(o orm.Ormer) (uint32, error) {
	sql := fmt.Sprintf("INSERT INTO %s (id, name, namespace, description, team_id, creation_time) VALUES (?,?,?,?,?,?);", s.TableName())
	_, err := o.Raw(sql, s.SpaceID, s.SpaceName, s.Namespace, s.Description, s.TeamID, s.CreationTime).Exec()
	return sqlstatus.ParseErrorCode(err)
}

// Delete deletes a record
func (s *SpaceModel) Delete() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("delete from %s where id=?;", s.TableName())
	res, err := o.Raw(sql, s.SpaceID).Exec()
	if err != nil {
		return sqlstatus.ParseErrorCode(err)
	}
	if rowNumber, err := res.RowsAffected(); err != nil {
		return sqlstatus.ParseErrorCode(err)
	} else if rowNumber == 0 {
		return sqlstatus.SQLErrNoRowFound, fmt.Errorf("space %s does not exist", s.SpaceID)
	}
	return sqlstatus.SQLSuccess, nil
}

// CheckNamespace checks if a namespace exists
func (s SpaceModel) CheckNamespace(namespace string) (bool, uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select count(namespace) as total from %s where namespace=?;", s.TableName())
	var res common.ListMeta
	if err := o.Raw(sql, namespace).QueryRow(&res); err != nil {
		errcode, err := sqlstatus.ParseErrorCode(err)
		return false, errcode, err
	}
	return res.Total != 0, sqlstatus.SQLSuccess, nil
}

func (s *SpaceModel) GetSpaceListByTeamID(teamID string) ([]string, error) {
	o := orm.NewOrm()
	var list orm.ParamsList
	if _, err := o.QueryTable(s.TableName()).
		Filter("team_id", teamID).
		ValuesFlat(&list, "namespace"); err != nil {
		return nil, err
	}
	var spaceList []string
	for _, item := range list {
		spaceList = append(spaceList, item.(string))
	}
	return spaceList, nil
}

func (s *SpaceModel) GetTeamIDByTeamspace(teamspace string) (string, error) {
	o := orm.NewOrm()
	result := SpaceModel{}
	err := o.QueryTable(s.TableName()).Filter("namespace", teamspace).One(&result, "TeamID")
	return result.TeamID, err
}

func (s *SpaceModel) DeleteByNamespace(namespaceList []string, orms ...orm.Ormer) (int64, error) {
	if len(namespaceList) == 0 {
		return 0, nil
	}
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	num, err := o.QueryTable(s.TableName()).
		Filter("namespace__in", namespaceList).
		Delete()
	return num, err
}

// IsSpaceExistInTeam check if space exists in the team
func (s *SpaceModel) IsSpaceExistInTeam(teamID string, spaceID string) (bool, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable(s.TableName()).
		Filter("id", spaceID).
		Filter("team_id", teamID).Count()
	if count > 0 {
		return true, nil
	}
	return false, err
}
