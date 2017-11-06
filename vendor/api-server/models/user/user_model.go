/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package user

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"

	"api-server/models/common"
	sqlstatus "api-server/models/sql/status"
	"api-server/util/rand"
)

// UserModel tenx_users record
type UserModel struct {
	UserID         int32     `orm:"column(user_id);pk"`
	Username       string    `orm:"column(user_name);size(200)"`
	Namespace      string    `orm:"column(namespace);size(200)"`
	Displayname    string    `orm:"column(displayname);size(200)"`
	Password       string    `orm:"column(password);size(500)"`
	Email          string    `orm:"column(email);size(100)"`
	Phone          string    `orm:"column(phone);size(20)"`
	CreationTime   time.Time `orm:"column(creation_time);auto_now_add;type(datetime)"`
	LastLoginTime  time.Time `orm:"column(last_login_time);auto_now;type(datetime)"`
	LoginFrequency int       `orm:"column(login_frequency);default(1)"`
	ConfirmCode    string    `orm:"column(confirm_code);size(45)"`
	Active         int8      `orm:"column(active);default(0)"`
	APIToken       string    `orm:"column(api_token);size(48)"`
	Role           int32     `orm:"column(role)"`
	Avatar         string    `orm:"column(avatar);size(45)"`
	EnvEdition     int       `orm:"column(env_edition)"`
	Migrated       int       `orm:"column(migrated)"`
	Type           int       `orm:"column(type);default(1)"`
}

// DirectorySummary count of each user directory
type DirectorySummary struct {
	Type  int `orm:"column(type)";json:"type"`
	Count int `orm:"column(count)";json:"count"`
}

const (
	MigratedNo      = 0
	MigratedFromOld = 1
)

const (
	EnvEditionStanard      = 0
	EnvEditionProfessional = 1
)

const (
	DatabaseUser = 1
	LDAPUser     = 2
)

// TableName return tenx_users
func (u *UserModel) TableName() string {
	return "tenx_users"
}
func NewUserModel() *UserModel {
	return &UserModel{}
}

// Insert add a new user record to db
func (u *UserModel) Insert(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	switch common.DType {
	case common.DRMySQL:
		userID, err := o.Insert(u)
		if err != nil {
			return sqlstatus.ParseErrorCode(err)
		}
		u.UserID = int32(userID)
	case common.DRPostgres:
		sql := `INSERT INTO "tenx_users" ( "user_name", "namespace", "displayname", "password", "email", "phone", "creation_time", "login_frequency", "active", "api_token", "role") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING user_id;`
		err := o.Raw(sql, u.Username, u.Namespace, u.Displayname, u.Password, u.Email, u.Phone, time.Now(), u.LoginFrequency, u.Active, u.APIToken, u.Role).QueryRow(u)
		if err != nil {
			return sqlstatus.ParseErrorCode(err)
		}
	default:
		return sqlstatus.SQLErrUnCategoried, fmt.Errorf("driver %s not supported", common.DType)
	}

	glog.Infof("user %s added to table %s\n", u.Username, u.TableName())
	return u.Get(o)
}

// Get fetch a user record by id
func (u *UserModel) Get(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	err := o.QueryTable(u.TableName()).Filter("user_id", u.UserID).One(u)
	return sqlstatus.ParseErrorCode(err)
}

func (u *UserModel) GetByEamil(email string, orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	err := o.QueryTable(u.TableName()).Filter("email", email).One(u)
	return sqlstatus.ParseErrorCode(err)
}

// GetByUsernameAndPassword fetch a user record by user_name and password
func (u *UserModel) GetByUsernameAndPassword(password string) (uint32, error) {
	o := orm.NewOrm()
	err := o.QueryTable(u.TableName()).Filter("user_name", u.Username).Filter("password", password).One(u)
	return sqlstatus.ParseErrorCode(err)
}

// GetByEmailAndPassword fetch a user record by email and password
func (u *UserModel) GetByEmailAndPassword(password string) (uint32, error) {
	o := orm.NewOrm()
	err := o.QueryTable(u.TableName()).Filter("email", u.Email).Filter("password", password).One(u)
	return sqlstatus.ParseErrorCode(err)
}

// Update update a user record by id
func (u *UserModel) Update(cols ...string) (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Update(u, cols...)
	return sqlstatus.ParseErrorCode(err)
}

// UpdateLastLoginTime update last login time by user id
func (u *UserModel) UpdateLastLoginTime() (uint32, error) {
	o := orm.NewOrm()
	newInfo := orm.Params{"last_login_time": u.LastLoginTime, "login_frequency": u.LoginFrequency}
	_, err := o.QueryTable(u.TableName()).
		Filter("user_id", u.UserID).Update(newInfo)
	return sqlstatus.ParseErrorCode(err)
}

// GetByName get user by name
func (u *UserModel) GetByName(name string) (uint32, error) {
	o := orm.NewOrm()
	err := o.QueryTable(u.TableName()).Filter("user_name", name).One(u)
	return sqlstatus.ParseErrorCode(err)
}

func (u *UserModel) GetByNameOrEmail(user, email string) error {
	o := orm.NewOrm()
	err := o.Raw(`SELECT * FROM tenx_users WHERE user_name = ? or email = ?`, user, email).QueryRow(u)
	return err
}

// GetConflictUsers get user count by name or email
func (u *UserModel) GetLdapConflictUsers(name, email string) (int32, error) {
	o := orm.NewOrm()

	type result struct {
		Count int32 `orm:"column(count)"`
	}
	countResult := result{}
	sql := fmt.Sprintf(`SELECT count(*) as count FROM %s where (user_name = ? or email = ?) and type != ?`, u.TableName())
	err := o.Raw(sql, name, email, LDAPUser).QueryRow(&countResult)
	return countResult.Count, err
}

// GetLDAPUsers get all LDAP users
func (u *UserModel) GetLDAPUsers() ([]UserModel, error) {
	o := orm.NewOrm()
	var userModels []UserModel
	_, err := o.QueryTable(u.TableName()).Filter("type", LDAPUser).All(&userModels)
	return userModels, err
}

//GetByNamespace get user by Namespaze
func (u *UserModel) GetByNamespace(namespace string) (uint32, error) {
	o := orm.NewOrm()
	err := o.QueryTable(u.TableName()).Filter("namespace", namespace).One(u)
	return sqlstatus.ParseErrorCode(err)
}

func (u *UserModel) List(dataselect *common.DataSelectQuery) ([]UserModel, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT user_id FROM %s where %s %s %s;`, u.TableName(), dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var userModels []UserModel
	_, err := o.Raw(sql).QueryRows(&userModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}

	// check empty
	if len(userModels) == 0 {
		return []UserModel{}, sqlstatus.SQLSuccess, nil
	}

	// get the user IDs
	var userIDStr []string
	for _, um := range userModels {
		userIDStr = append(userIDStr, strconv.FormatUint(uint64(um.UserID), 10))
	}

	sql = fmt.Sprintf(`SELECT 
    *
FROM
    %s
WHERE
    user_id IN (%s) and %s %s;`, u.TableName(), strings.Join(userIDStr, ","), dataselect.FilterQuery, dataselect.SortQuery)
	_, err = o.Raw(sql).QueryRows(&userModels)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return userModels, errCode, err
}

func (u *UserModel) Count(dataselect *common.DataSelectQuery) (int, uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf(`SELECT 
    count(*) as total
FROM
    %s where %s;`, u.TableName(), dataselect.FilterQuery)
	meta := &common.ListMeta{}
	err := o.Raw(sql).QueryRow(meta)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return 0, errCode, err
	}
	return meta.Total, sqlstatus.SQLSuccess, nil
}

// ListByIDs returns users according to a list of user ids
func (u *UserModel) ListByIDs(userIDs []int32, dataselect *common.DataSelectQuery) ([]UserModel, uint32, error) {
	o := orm.NewOrm()

	// get the user IDs
	var userIDStr []string
	for _, item := range userIDs {
		userIDStr = append(userIDStr, strconv.FormatUint(uint64(item), 10))
	}

	sql := fmt.Sprintf(`SELECT * FROM %s WHERE user_id IN (%s) and %s %s %s;`, u.TableName(),
		strings.Join(userIDStr, ","), dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)

	var userModels []UserModel
	_, err := o.Raw(sql).QueryRows(&userModels)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return userModels, errCode, err
}

// ListIDsByIDs returns user ids according to a list of user ids and filter query
func (u *UserModel) ListIDsByIDs(userIDs []int32, dataselect *common.DataSelectQuery) ([]int32, uint32, error) {
	o := orm.NewOrm()

	// get the user IDs
	var userIDStr []string
	for _, item := range userIDs {
		userIDStr = append(userIDStr, strconv.FormatUint(uint64(item), 10))
	}

	sql := fmt.Sprintf(`SELECT user_id FROM %s WHERE user_id IN (%s) and %s;`, u.TableName(),
		strings.Join(userIDStr, ","), dataselect.FilterQuery)

	var userModels []UserModel
	_, err := o.Raw(sql).QueryRows(&userModels)
	if err != nil {
		errcode, err := sqlstatus.ParseErrorCode(err)
		return nil, errcode, err
	}
	userIDs = nil
	for _, um := range userModels {
		userIDs = append(userIDs, um.UserID)
	}

	return userIDs, sqlstatus.SQLSuccess, nil
}

// Delete removes a tenx_users record
func (u *UserModel) Delete() (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Delete(u)
	return sqlstatus.ParseErrorCode(err)
}

// RemoveLDAPUsers remove users that was synced from LDAP
func (u *UserModel) RemoveLDAPUsers(o orm.Ormer) (int64, error) {
	count, err := o.QueryTable(u.TableName()).
		Filter("type", LDAPUser).
		Delete()
	return count, err
}

type user struct {
	UserID    int32  `orm:"column(user_id)" json:"-"`
	UserName  string `orm:"column(user_name)" json:"userName"`
	Role      int    `json:"-"`
	Email     string `json:"email"`
	IsCreator bool   `json:"isCreator"`
	IsAdmin   bool   `json:"isAdmin"`
}

func (u *UserModel) GetAllUserInTeam(teamID string, userID int32) ([]user, error) {
	method := "GetAllUserInTeam"
	users := make([]user, 0, 1)
	emptyUsers := make([]user, 0, 1)
	o := orm.NewOrm()
	sql := `SELECT t1.user_id, t2.user_name, t1.role, t2.email
	FROM tenx_team_user_ref AS t1 INNER JOIN tenx_users AS t2
	ON t1.user_id = t2.user_id
	WHERE t1.team_id = ?`
	if _, err := o.Raw(sql, teamID).QueryRows(&users); err != nil {
		glog.Errorln(method, "get users failed.", err)
		return emptyUsers, err
	}

	// get team creator id
	var creatorID []orm.Params
	if _, err := o.QueryTable("tenx_teams").Filter("id", teamID).Values(&creatorID, "creator_id"); err != nil {
		glog.Errorln(method, "get creator id failed.", err)
		return emptyUsers, err
	}

	for i := range users {
		users[i].IsCreator = (int64(users[i].UserID) == creatorID[0]["CreatorID"].(int64))
		users[i].IsAdmin = (users[i].Role == 1 || users[i].Role == 2)
	}

	return users, nil
}

func (u *UserModel) GetInfoByEmail(email string) error {
	o := orm.NewOrm()
	return o.QueryTable(u.TableName()).Filter("email", email).One(u)
}
func (u *UserModel) UpdateActiveStatus(email string, status int8, orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	_, err := o.QueryTable(u.TableName()).Filter("email", email).Update(orm.Params{"active": status})
	return err
}

func (u *UserModel) IsEmailOrPhoneConflict(email, phone string) (int, error) {
	o := orm.NewOrm()
	meta := &common.ListMeta{}
	err := o.Raw(`SELECT count(*) as total FROM tenx_users WHERE email = ? or phone = ?`, email, phone).QueryRow(meta)
	if err != nil {
		return 0, err
	}
	return meta.Total, nil
}

// UpdateUserRole update user role
func (u *UserModel) UpdateUserRole(userID int32, role int32) (int64, error) {
	o := orm.NewOrm()
	n, err := o.QueryTable("tenx_users").Filter("user_id", userID).Update(
		orm.Params{
			"role": role,
		})
	return n, err
}

// UpdateByParams update by user specified params
func (u *UserModel) UpdateByParams(parms orm.Params) (int64, error) {
	o := orm.NewOrm()
	n, err := o.QueryTable("tenx_users").Filter("user_id", u.UserID).Update(parms)
	return n, err
}

func (u UserModel) ListNamespaces() ([]string, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf(`select namespace from %s;`, u.TableName())
	var namespaces []string
	_, err := o.Raw(sql).QueryRows(&namespaces)
	if err != nil {
		return nil, err
	}
	return namespaces, nil
}

// GetUserDirectorySummary get user directory summary
func (u *UserModel) GetUserDirectorySummary() ([]DirectorySummary, error) {
	o := orm.NewOrm()
	directoryRow := make([]DirectorySummary, 0, 5)
	sql := fmt.Sprintf(`select type, count(*) as count FROM %s group by type`, u.TableName())
	_, err := o.Raw(sql).QueryRows(&directoryRow)
	return directoryRow, err
}

func (u *UserModel) ToResetPassword() (email, code string, err error) {
	veryfyCode := rand.RandString(6)
	expireTime := time.Now().Unix() + int64(time.Hour)
	u.ConfirmCode = veryfyCode + "," + strconv.FormatInt(expireTime, 10)

	_, err = u.Update("confirm_code")
	if err != nil {
		return "", "", err
	}
	_, err = u.Get()
	if err != nil {
		return "", "", err
	}
	return u.Email, veryfyCode, nil
}
