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

	"dev-flows-api-golang/models/common"
	sqlstatus "dev-flows-api-golang/models/sql/status"

	"github.com/golang/glog"
)

type UserAccountModel struct {
	UserID              int32     `orm:"pk;column(user_id)"`
	Namespace           string    `orm:"column(namespace);size(45)"`
	Balance             int64     `orm:"column(balance)"`
	PayBalance          int64     `orm:"column(pay_balance)"`
	LastCost            int32     `orm:"column(last_cost)"`
	LastChargeAdminID   int32     `orm:"column(last_charge_admin_id)"`
	LastChargeAdminName string    `orm:"column(last_charge_admin_name)"`
	LastChargeAmount    int64     `orm:"column(last_charge_amount);digits(11);decimals(3)"`
	LastChargeTime      time.Time `orm:"column(last_charge_time)"`
	StorageLimit        int       `orm:"column(storage_limit)"`
	LastEmailTime       time.Time `orm:"column(last_email_time)"`
}

// TableName returns the name of the table in database
func (m UserAccountModel) TableName() string {
	return "tenx_user_account"
}

func NewUserAccount() *UserAccountModel {
	return &UserAccountModel{}
}

// Insert adds a new user account to database
// the record's lifecycle is the same as the record with same user_id in tenx_users
func (m *UserAccountModel) Insert(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	switch common.DType {
	case common.DRMySQL:
		sql := `INSERT INTO %s (user_id, balance, namespace) VALUES (?, ?, ?);`
		sql = fmt.Sprintf(sql, m.TableName())
		_, err := o.Raw(sql, m.UserID, m.Balance, m.Namespace).Exec()
		errcode, err := sqlstatus.ParseErrorCode(err)
		return errcode, err
	case common.DRPostgres:
		sql := fmt.Sprintf(`INSERT INTO "%s" ("user_id", "balance", "namespace") VALUES (?, ?, ?);`, m.TableName())
		_, err := o.Raw(sql, m.UserID, m.Balance, m.Namespace).Exec()
		errcode, err := sqlstatus.ParseErrorCode(err)
		return errcode, err
	}
	return sqlstatus.SQLErrUnAuthorized, fmt.Errorf("driver %v not supported", common.DType)
}

// List list user account, orderby : namespace, balance
func (m *UserAccountModel) List(dataselect *common.DataSelectQuery) ([]UserAccountModel, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT  user_id FROM %s %s %s;`, m.TableName(), dataselect.SortQuery, dataselect.PaginationQuery)
	var userAccountModes []UserAccountModel
	_, err := o.Raw(sql).QueryRows(&userAccountModes)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}

	if len(userAccountModes) == 0 {
		return nil, sqlstatus.SQLSuccess, nil
	}

	// get the user IDs
	var userIDStr []string
	for _, um := range userAccountModes {
		userIDStr = append(userIDStr, strconv.FormatUint(uint64(um.UserID), 10))
	}

	sql = fmt.Sprintf(`SELECT * FROM %s WHERE user_id IN (%s) %s;`,
		m.TableName(), strings.Join(userIDStr, ","), dataselect.SortQuery)
	_, err = o.Raw(sql).QueryRows(&userAccountModes)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return userAccountModes, errCode, err
}

// ListByIDs all user_accounts by ids
func (m *UserAccountModel) ListByIDs(ids []int32, dataselect *common.DataSelectQuery) ([]UserAccountModel, uint32, error) {
	if len(ids) == 0 {
		return nil, sqlstatus.SQLSuccess, nil
	}

	o := orm.NewOrm()
	var idstrs []string
	for _, id := range ids {
		idstrs = append(idstrs, strconv.FormatInt(int64(id), 10))
	}

	sql := fmt.Sprintf(`select * from %s where user_id IN (%s) and %s %s %s;`, m.TableName(), strings.Join(idstrs, ", "), dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var vaccs []UserAccountModel
	_, err := o.Raw(sql).QueryRows(&vaccs)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return vaccs, errcode, err
}

// ListByNamespaces all user_accounts by namespaces
func (m *UserAccountModel) ListByNamespaces(namespaces []string, dataselect *common.DataSelectQuery) ([]UserAccountModel, uint32, error) {
	if len(namespaces) == 0 {
		return nil, sqlstatus.SQLSuccess, nil
	}

	o := orm.NewOrm()
	var namespacestrs []string
	for _, namespace := range namespaces {
		namespacestrs = append(namespacestrs, "'"+namespace+"'")
	}

	sql := fmt.Sprintf(`select * from %s where namespace IN (%s) and %s %s %s;`, m.TableName(), strings.Join(namespacestrs, ", "), dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var vaccs []UserAccountModel
	_, err := o.Raw(sql).QueryRows(&vaccs)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return vaccs, errcode, err
}

// Get returns one user account record by id
func (m *UserAccountModel) Get(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	sql := fmt.Sprintf("SELECT  * FROM %s WHERE user_id = ?;", m.TableName())
	err := o.Raw(sql, m.UserID).QueryRow(m)
	return sqlstatus.ParseErrorCode(err)
}

// Update updates one user account record by id
func (m UserAccountModel) Update(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	// Only update balance
	newInfo := orm.Params{"balance": m.Balance}
	_, err := o.QueryTable(m.TableName()).
		Filter("user_id", m.UserID).Update(newInfo)

	return sqlstatus.ParseErrorCode(err)
}

// Update updates one user account record by id
func (m UserAccountModel) UpdateWithAdminInfo(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	// Only update balance
	newInfo := orm.Params{"balance": m.Balance, "last_charge_admin_id": m.LastChargeAdminID, "last_charge_admin_name": m.LastChargeAdminName,
		"last_charge_amount": m.LastChargeAmount, "last_charge_time": m.LastChargeTime}
	_, err := o.QueryTable(m.TableName()).
		Filter("user_id", m.UserID).Update(newInfo)

	return sqlstatus.ParseErrorCode(err)
}

// UpdateBalanceByNamespace update updates one user account record by namespace
func (m UserAccountModel) UpdateBalanceByNamespace(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	// Only update balance
	newInfo := orm.Params{"balance": m.Balance}
	_, err := o.QueryTable(m.TableName()).
		Filter("namespace", m.Namespace).Update(newInfo)

	return sqlstatus.ParseErrorCode(err)
}

// UpdateBlance updates one user account record by id
func (m *UserAccountModel) AddBalance(balance int64, namespace string, orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	sql := `UPDATE ` + m.TableName() + ` SET balance=balance+?,
	last_charge_amount=?,
	last_charge_time=?
	WHERE namespace=?`
	_, err := o.Raw(sql, balance, balance, time.Now(), namespace).Exec()
	return err
}

// DeleteByID deletes a user or team space account
func (m *UserAccountModel) DeleteByID(userID int32) (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("DELETE FROM %s where user_id=?;", m.TableName())
	_, err := o.Raw(sql, userID).Exec()
	return sqlstatus.ParseErrorCode(err)
}

// Count returns the number of  user_accounts by ids
func (m *UserAccountModel) Count(ids []int32) (int64, uint32, error) {
	o := orm.NewOrm()
	var idstrs []string
	for _, id := range ids {
		idstrs = append(idstrs, strconv.FormatInt(int64(id), 10))
	}
	sql := fmt.Sprintf(`select count(*) as total from %s where user_id IN (%s);`, m.TableName(), strings.Join(idstrs, ", "))
	c := common.ListMeta{}
	err := o.Raw(sql).QueryRow(&c)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return int64(c.Total), errcode, err
}

func (m *UserAccountModel) GetStorageLimit(space string) (int, error) {
	method := "GetStorageLimit"
	o := orm.NewOrm()
	result := UserAccountModel{}
	if err := o.QueryTable(m.TableName()).Filter("namespace", space).One(&result, "StorageLimit"); err != nil && err != orm.ErrNoRows {
		glog.Errorln(method, "failed.", err)
		return 0, err
	} else if err == orm.ErrNoRows {
		glog.Infoln(method, "get storage limit return zero row.", err)
		return 0, nil
	}
	return result.StorageLimit, nil
}

func (m *UserAccountModel) GetBalanceByNamespaceList(spaceList []string) (int, error) {
	method := "GetBalanceByNamespaceList"
	qb, err := orm.NewQueryBuilder(common.DType.String())
	if err != nil {
		glog.Errorln(method, "NewQueryBuilder failed.", err)
		return 0, err
	}
	var spaceListCopy []string
	for _, space := range spaceList {
		spaceListCopy = append(spaceListCopy, "'"+space+"'")
	}
	sql := qb.Select("SUM(balance) as sum").
		From(m.TableName()).
		Where("namespace").In(spaceListCopy...).String()
	o := orm.NewOrm()
	result := struct {
		Sum int64
	}{}
	if err := o.Raw(sql).QueryRow(&result); err != nil {
		glog.Errorln(method, "failed.", err)
		return 0, err
	}
	return int(result.Sum), nil
}

func (m *UserAccountModel) GetBalanceByNamespace(namespace string) (int64, error) {
	method := "GetBalanceByNamespace"
	o := orm.NewOrm()
	result := UserAccountModel{}
	if err := o.QueryTable(m.TableName()).Filter("namespace", namespace).One(&result, "balance"); err != nil && err != orm.ErrNoRows {
		glog.Errorln(method, "failed.", err)
		return 0, err
	} else if err == orm.ErrNoRows {
		glog.Errorln(method, "no balance record", err)
		return 0, nil
	}
	return result.Balance, nil
}

// LockRow lock row to be updated
func (m *UserAccountModel) LockRow(o orm.Ormer, namespace string) (int64, error) {
	var balances []int64
	count, err := o.Raw("select balance from "+m.TableName()+" where namespace = ? for update", namespace).QueryRows(&balances)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return balances[0], nil
	}
	return 0, nil
}
