/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team2user

import (
	"fmt"

	sqlstatus "api-server/models/sql/status"
	teamspaceModal "api-server/models/teamspace"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"

	"api-server/models/common"
)

const (
	TeamNormalUser = 0
	TeamAdmin      = 1
	SuperAdminUser = 2
)

// TeamUserModel tenx_team_user_ref, team_id and user_id together as the primary key
type TeamUserModel struct {
	TeamID string `orm:"column(team_id);pk"`
	UserID int32  `orm:"column(user_id)"`
	Role   int32  `orm:"column(role)"` // currently not used, because team creator is leader
}

func (s TeamUserModel) TableName() string {
	return "tenx_team_user_ref"
}

func NewTeamUserModel() *TeamUserModel {
	return &TeamUserModel{}
}

// ListByUser lists all teams user ref by userID
func (s *TeamUserModel) ListByUser(userID int32, dataselect *common.DataSelectQuery) ([]TeamUserModel, uint32, error) {
	o := orm.NewOrm()

	//Get user role, for admin user, return all team spaces
	type RoleModel struct {
		Role int32 `orm:"column(role)"`
	}
	sql := `SELECT role FROM tenx_users WHERE user_id = ?;`

	roleModel := RoleModel{}
	var refs []TeamUserModel
	err := o.Raw(sql, userID).QueryRow(&roleModel)
	if err == nil {
		if roleModel.Role == SuperAdminUser {
			sql = fmt.Sprintf(`SELECT * FROM %s %s %s;`, s.TableName(), dataselect.SortQuery, dataselect.PaginationQuery)
			_, err = o.Raw(sql).QueryRows(&refs)
		} else {
			sql = fmt.Sprintf(`SELECT * FROM %s WHERE user_id = ? %s %s;`, s.TableName(), dataselect.SortQuery, dataselect.PaginationQuery)
			_, err = o.Raw(sql, userID).QueryRows(&refs)
		}
	}

	errCode, _ := sqlstatus.ParseErrorCode(err)
	if err != nil {
		return nil, errCode, err
	}
	return refs, errCode, err
}

// ListByTeam returns a list of team spaces according to team ID
func (s *TeamUserModel) ListByTeam(teamID string, dataselect *common.DataSelectQuery) ([]TeamUserModel, uint32, error) {
	o := orm.NewOrm()

	// space query ids
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE team_id=?;`, s.TableName())

	var teamUserModels []TeamUserModel
	_, err := o.Raw(sql, teamID).QueryRows(&teamUserModels)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return teamUserModels, errCode, err
}

// CountByTeamID returns the number of members according to team ID
func (s *TeamUserModel) CountByTeamID(teamID string) (int, uint32, error) {
	o := orm.NewOrm()

	// space query ids
	sql := fmt.Sprintf(`SELECT count(user_id) AS count FROM %s WHERE team_id=?;`, s.TableName())

	var count int
	err := o.Raw(sql, teamID).QueryRow(&count)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return count, errCode, err
}

// DeleteByTeam returns a list of team spaces according to team ID
func (s *TeamUserModel) DeleteByTeam(teamID string, orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	_, err := o.QueryTable(s.TableName()).Filter("team_id", teamID).Delete()
	return sqlstatus.ParseErrorCode(err)
}

// DeleteByUser returns a list of team spaces according to team ID
func (s *TeamUserModel) DeleteByUser(userID int32) (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Raw("delete from tenx_team_user_ref where user_id=?;", userID).Exec()
	return sqlstatus.ParseErrorCode(err)
}

// Insert add a new ref record
func (s *TeamUserModel) Insert(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 0 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	sql := fmt.Sprintf("INSERT INTO %s (team_id, user_id, role) VALUES (?, ?, ?);", s.TableName())
	_, err := o.Raw(sql, s.TeamID, s.UserID, s.Role).Exec()
	return sqlstatus.ParseErrorCode(err)
}
func (s *TeamUserModel) InsertIgnore(teamID string, userID int32, role int, orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	sql := fmt.Sprintf("INSERT IGNORE INTO %s (team_id, user_id, role) VALUES (?, ?, ?);", s.TableName())
	_, err := o.Raw(sql, teamID, userID, role).Exec()
	return err
}

// Delete deletes a record
func (s *TeamUserModel) Delete() (uint32, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from tenx_team_user_ref where team_id=? and user_id=?;", s.TeamID, s.UserID).Exec()
	if err != nil {
		return sqlstatus.ParseErrorCode(err)
	}
	if rowNumber, err := res.RowsAffected(); err != nil {
		return sqlstatus.ParseErrorCode(err)
	} else if rowNumber == 0 {
		return sqlstatus.SQLErrNoRowFound, fmt.Errorf("user %d does not exist", s.UserID)
	}
	return sqlstatus.SQLSuccess, nil
}

func (t *TeamUserModel) GetRole(teamID string, userID int32) (int32, error) {
	method := "models.TeamUserRef.GetRole"
	tt := TeamUserModel{}
	o := orm.NewOrm()
	if err := o.QueryTable(t.TableName()).
		Filter("team_id", teamID).
		Filter("user_id", userID).
		One(&tt); err != nil {
		glog.Infof("%s team id %s, user id %d. %s\n", method, teamID, userID, err)
		return 0, err
	}
	return tt.Role, nil
}

func (t *TeamUserModel) CheckUserTeamspace(userID int32, teamspace string) bool {
	method := "CheckUserTeamspace"
	o := orm.NewOrm()
	result := struct {
		Cnt int
	}{}
	sql := `SELECT COUNT(*) AS cnt
	FROM tenx_team_user_ref AS t1 INNER JOIN tenx_team_space AS t2
	ON t1.team_id = t2.team_id
	WHERE t1.user_id = ?
	AND t2.namespace = ?`
	if err := o.Raw(sql, userID, teamspace).QueryRow(&result); err != nil {
		glog.Errorln(method, "failed.", err)
		return false
	}
	return result.Cnt > 0
}

func (t *TeamUserModel) IsUserInTeam(userID int32, teamID string) bool {
	o := orm.NewOrm()
	return o.QueryTable(t.TableName()).
		Filter("user_id", userID).
		Filter("team_id", teamID).
		Exist()
}

func (t *TeamUserModel) GetTeamUserEmailList(teamID string) ([]string, []string) {
	method := "GetTeamUserEmailList"

	adminEmails := make([]string, 0, 1)
	normalUserEmails := make([]string, 0, 1)

	sql := `SELECT t2.email,t1.role FROM tenx_team_user_ref AS t1
	INNER JOIN tenx_users AS t2
	ON t1.user_id=t2.user_id
	WHERE t1.team_id=?`
	result := make([]struct {
		Email string
		Role  int
	}, 0, 1)
	o := orm.NewOrm()
	_, err := o.Raw(sql, teamID).QueryRows(&result)
	if err != nil {
		glog.Errorln(method, "get user email failed.", err)
		return adminEmails, normalUserEmails
	}

	for _, item := range result {
		if item.Role == TeamNormalUser {
			normalUserEmails = append(normalUserEmails, item.Email)
		} else {
			adminEmails = append(adminEmails, item.Email)
		}
	}
	return adminEmails, normalUserEmails
}

//AuthUserCanUseTeamspace auth user can use this namespace
func (t *TeamUserModel) AuthUserCanUseTeamspace(user_id int32, teamspace string) (int, error) {
	teams := teamspaceModal.NewSpaceModel()
	o := orm.NewOrm()
	var result int
	sql := fmt.Sprintf("SELECT count(*) FROM %s t1 inner join %s t2 on t1.team_id = t2.team_id where t1.namespace = ? and t2.user_id = ?", teams.TableName(), t.TableName())
	err := o.Raw(sql, teamspace, user_id).QueryRow(&result)
	return result, err
}

func IsTeammate(one, another int32) (isTeammate bool, err error) {
	var sameTeams []string
	sameTeams, err = SameTeams(one, another)
	isTeammate = len(sameTeams) > 0
	return
}

const findSameTeams = `SELECT one.team_id FROM (
    (SELECT team_id FROM tenx_team_user_ref WHERE user_id = ?)
      AS one JOIN
    (SELECT team_id FROM tenx_team_user_ref WHERE user_id = ?)
      AS another ON one.team_id = another.team_id);`

func SameTeams(one, another int32) (sameTeams []string, err error) {
	_, err = orm.NewOrm().Raw(findSameTeams, one, another).QueryRows(&sameTeams)
	return
}
