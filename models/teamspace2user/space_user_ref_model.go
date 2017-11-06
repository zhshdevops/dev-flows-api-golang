/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package teamspace2user

import (
	"fmt"

	sqlstatus "dev-flows-api-golang/models/sql/status"

	"github.com/astaxie/beego/orm"

	"dev-flows-api-golang/models/common"
)

// SpaceUserModel tenx_space_user_ref, teamspace_id and user_id together as the primary key
type SpaceUserModel struct {
	SpaceID string `orm:"column(teamspace_id);pk" json:"spaceID"`
	UserID  int32  `orm:"column(user_id)" json:"userID"`
	Role    int32  `orm:"column(role)" json:"role"`
}

func NewSpaceUsrModel() *SpaceUserModel {
	return &SpaceUserModel{}
}

func (s SpaceUserModel) TableName() string {
	return "tenx_space_user_ref"
}

// Insert add a new ref record
func (s *SpaceUserModel) Insert() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("INSERT INTO %s (teamspace_id, user_id, role) VALUES (?, ?, ?);", s.TableName())
	_, err := o.Raw(sql, s.SpaceID, s.UserID, s.Role).Exec()
	return sqlstatus.ParseErrorCode(err)
}

// Delete deletes a record
func (s *SpaceUserModel) Delete() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("delete from %s where teamspace_id=? and user_id=?;", s.TableName())
	res, err := o.Raw(sql, s.SpaceID, s.UserID).Exec()
	if err != nil {
		return sqlstatus.ParseErrorCode(err)
	}
	if rowNumber, err := res.RowsAffected(); err != nil {
		return sqlstatus.ParseErrorCode(err)
	} else if rowNumber == 0 {
		return sqlstatus.SQLErrNoRowFound, fmt.Errorf("user %d not in the space", s.UserID)
	}
	return sqlstatus.SQLSuccess, nil
}

// ListByUser lists all teamspace user ref by userID
func (s *SpaceUserModel) ListByUser(userID int32, dataselect *common.DataSelectQuery) ([]SpaceUserModel, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT  * FROM %s WHERE user_id = ? %s %s;`, s.TableName(), dataselect.SortQuery, dataselect.PaginationQuery)
	var refs []SpaceUserModel
	_, err := o.Raw(sql, userID).QueryRows(&refs)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return refs, errCode, err
}

// CountUser return the number of users in certain namespace
func (s SpaceUserModel) CountUser(spaceID string) (int64, uint32, error) {
	o := orm.NewOrm()
	number, err := o.QueryTable(s.TableName()).Filter("teamspace_id", spaceID).Count()
	errcode, err := sqlstatus.ParseErrorCode(err)
	return number, errcode, err
}

// ListBySpace returns a list of team spaces according to space ids
func (s *SpaceUserModel) ListBySpace(spaceID string) ([]SpaceUserModel, uint32, error) {
	o := orm.NewOrm()

	// space query ids
	sql := fmt.Sprintf(`SELECT 
    *
FROM
    %s
WHERE
    teamspace_id=?;`, s.TableName())

	var spaceModels []SpaceUserModel
	_, err := o.Raw(sql, spaceID).QueryRows(&spaceModels)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return spaceModels, errCode, err
}

// DeleteBySpace deletes refs by space ID
func (s *SpaceUserModel) DeleteBySpace(spaceID string) (uint32, error) {
	o := orm.NewOrm()
	_, err := o.QueryTable(s.TableName()).Filter("teamspace_id", spaceID).Delete()
	return sqlstatus.ParseErrorCode(err)
}

// DeleteByUser deletes refs by userID
func (s *SpaceUserModel) DeleteByUser(userID int32) (uint32, error) {
	o := orm.NewOrm()
	_, err := o.QueryTable(s.TableName()).Filter("userID", userID).Delete()
	return sqlstatus.ParseErrorCode(err)
}
