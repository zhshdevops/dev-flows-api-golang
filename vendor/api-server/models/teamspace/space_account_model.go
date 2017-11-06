/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */
package teamspace

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"

	"api-server/models/common"
	sqlstatus "api-server/models/sql/status"

	"github.com/golang/glog"
)

type SpaceAccountModel struct {
	SpaceID             string    `orm:"pk;column(teamspace_id)"`
	Balance             int64     `orm:"column(balance)"`
	Namespace           string    `orm:"column(namespace);size(45)"`
	LastCost            int32     `orm:"column(last_cost)"`
	LastChargeAdminID   int32     `orm:"column(last_charge_admin_id)"`
	LastChargeAdminName string    `orm:"column(last_charge_admin_name)"`
	LastChargeAmount    int64     `orm:"column(last_charge_amount);digits(11);decimals(3)"`
	LastChargeTime      time.Time `orm:"column(last_charge_time)"`
	StorageLimit        int       `orm:"column(storage_limit)"`
}

// TableName returns the name of the table in database
func (m SpaceAccountModel) TableName() string {
	return "tenx_team_space_account"
}

func NewSpaceAccount() *SpaceAccountModel {
	return &SpaceAccountModel{}
}

// Insert adds a new user account to database
// the record's lifecycle is the same as the record with same teamspace_id in tenx_team_space
func (m *SpaceAccountModel) Insert(o orm.Ormer) (uint32, error) {
	m.LastChargeTime, _ = time.Parse("0000-00-00 00:00:00", "0000-00-00 00:00:00")
	switch common.DType {
	case common.DRMySQL:
		_, err := o.Insert(m)
		errcode, err := sqlstatus.ParseErrorCode(err)
		return errcode, err
	case common.DRPostgres:
		sql := fmt.Sprintf(`INSERT INTO "%s" ("teamspace_id", "balance", "namespace", "last_charge_time") VALUES (?, ?, ?, ?);`, m.TableName())
		_, err := o.Raw(sql, m.SpaceID, m.Balance, m.Namespace, m.LastChargeTime).Exec()
		errcode, err := sqlstatus.ParseErrorCode(err)
		return errcode, err
	}
	return sqlstatus.SQLErrUnAuthorized, fmt.Errorf("driver %v not supported", common.DType)
}

// List list user account, orderby : namespace, balance
func (m *SpaceAccountModel) List(dataselect *common.DataSelectQuery) ([]SpaceAccountModel, uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf(`SELECT  teamspace_id FROM %s where %s %s %s;`, m.TableName(), dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var userAccountModes []SpaceAccountModel
	_, err := o.Raw(sql).QueryRows(&userAccountModes)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}
	// check empty
	if len(userAccountModes) == 0 {
		return []SpaceAccountModel{}, sqlstatus.SQLSuccess, nil
	}

	// get the user IDs
	var SpaceIDStr []string
	for _, um := range userAccountModes {
		SpaceIDStr = append(SpaceIDStr, fmt.Sprintf("'%s'", um.SpaceID))
	}

	sql = fmt.Sprintf(`SELECT * FROM %s WHERE teamspace_id IN (%s) %s;`,
		m.TableName(), strings.Join(SpaceIDStr, ","), dataselect.SortQuery)
	_, err = o.Raw(sql).QueryRows(&userAccountModes)
	errCode, err := sqlstatus.ParseErrorCode(err)
	return userAccountModes, errCode, err
}

// ListByIDs list all team_space_accounts by ids
func (m *SpaceAccountModel) ListByIDs(ids []string, dataselect *common.DataSelectQuery) ([]SpaceAccountModel, uint32, error) {
	if len(ids) == 0 {
		return []SpaceAccountModel{}, sqlstatus.SQLSuccess, nil
	}

	o := orm.NewOrm()
	var idstrs []string
	for _, id := range ids {
		idstrs = append(idstrs, "'"+id+"'")
	}

	sql := fmt.Sprintf(`select * from %s where teamspace_id IN (%s) and %s %s %s;`, m.TableName(), strings.Join(idstrs, ","), dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var vaccs []SpaceAccountModel
	_, err := o.Raw(sql).QueryRows(&vaccs)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return vaccs, errcode, err
}

// ListByNamespaces list all team_space_accounts by namespaces
func (m *SpaceAccountModel) ListByNamespaces(namespaces []string, dataselect *common.DataSelectQuery) ([]SpaceAccountModel, uint32, error) {
	if len(namespaces) == 0 {
		return []SpaceAccountModel{}, sqlstatus.SQLSuccess, nil
	}

	o := orm.NewOrm()
	var namespacestrs []string
	for _, namespace := range namespaces {
		namespacestrs = append(namespacestrs, "'"+namespace+"'")
	}

	sql := fmt.Sprintf(`select * from %s where namespace IN (%s) and %s %s %s;`, m.TableName(), strings.Join(namespacestrs, ","), dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var vaccs []SpaceAccountModel
	_, err := o.Raw(sql).QueryRows(&vaccs)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return vaccs, errcode, err
}

// Get returns one user account record by id
func (m *SpaceAccountModel) Get() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("SELECT  * FROM %s WHERE teamspace_id = ?;", m.TableName())
	err := o.Raw(sql, m.SpaceID).QueryRow(m)
	return sqlstatus.ParseErrorCode(err)
}

// Update updates one space account record by id
func (m SpaceAccountModel) Update(orms ...orm.Ormer) (uint32, error) {
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

// Update updates one space account record by id
func (m SpaceAccountModel) UpdateWithAdminInfo(orms ...orm.Ormer) (uint32, error) {
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
		Filter("namespace", m.Namespace).Update(newInfo)

	return sqlstatus.ParseErrorCode(err)
}

// Delete deletes a user or team space account
func (m *SpaceAccountModel) Delete() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("DELETE FROM %s where teamspace_id=?;", m.TableName())
	res, err := o.Raw(sql, m.SpaceID).Exec()
	if err != nil {
		return sqlstatus.ParseErrorCode(err)
	}
	if rowNumber, err := res.RowsAffected(); err != nil {
		return sqlstatus.ParseErrorCode(err)
	} else if rowNumber == 0 {
		return sqlstatus.SQLErrNoRowFound, nil
	}
	return sqlstatus.SQLSuccess, nil
}

// Count returns the number of  user_accounts by ids
func (m *SpaceAccountModel) Count(ids []int64) (int64, uint32, error) {
	o := orm.NewOrm()
	var idstrs []string
	for _, id := range ids {
		idstrs = append(idstrs, strconv.FormatInt(id, 10))
	}
	sql := fmt.Sprintf(`select count(*) as total from %s where teamspace_id IN (%s);`, m.TableName(), strings.Join(idstrs, ", "))
	c := common.ListMeta{}
	err := o.Raw(sql).QueryRow(&c)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return int64(c.Total), errcode, err
}

func (m *SpaceAccountModel) GetStorageLimit(space string) (int, error) {
	method := "GetStorageLimit"
	o := orm.NewOrm()
	result := SpaceAccountModel{}
	if err := o.QueryTable(m.TableName()).Filter("namespace", space).One(&result, "StorageLimit"); err != nil && err != orm.ErrNoRows {
		glog.Errorln(method, "failed.", err)
		return 0, err
	} else if err == orm.ErrNoRows {
		glog.Infoln(method, "get storage limit return zero row.", err)
		return 0, nil
	}
	return result.StorageLimit, nil
}

func (m *SpaceAccountModel) GetBalanceByNamespaceList(spaceList []string) (int, error) {
	method := "GetBalanceByNamespaceList"
	if len(spaceList) == 0 {
		return 0, nil
	}
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

func (m *SpaceAccountModel) GetBalanceByNamespace(namespace string) (int64, error) {
	method := "GetBalanceByNamespace"
	o := orm.NewOrm()
	result := SpaceAccountModel{}
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
func (m *SpaceAccountModel) LockRow(o orm.Ormer, namespace string) (string, int64, error) {
	var balances []int64
	var spaceID []string
	count, err := o.Raw("select teamspace_id, balance from "+m.TableName()+" where namespace = ? for update", namespace).QueryRows(&spaceID, &balances)
	if err != nil {
		return "", 0, err
	}
	if count > 0 {
		return spaceID[0], balances[0], nil
	}
	return "", 0, nil
}

func (t *SpaceAccountModel) DeleteByNamespace(namespaceList []string, orms ...orm.Ormer) (int64, error) {
	if len(namespaceList) == 0 {
		return 0, nil
	}
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	num, err := o.QueryTable(t.TableName()).
		Filter("namespace__in", namespaceList).
		Delete()
	return num, err
}
