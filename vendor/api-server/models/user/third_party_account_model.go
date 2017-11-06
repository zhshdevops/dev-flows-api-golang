/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2017-1-10  @author zhang shouhong
 */

package user

import (
	"api-server/models/common"
	sqlstatus "api-server/models/sql/status"
	"fmt"

	"github.com/astaxie/beego/orm"
)

// ThirdPartyAccountModel tenx_user_3rd_account, namespace and account_type together as primary key
type ThirdPartyAccountModel struct {
	Namespace     string `orm:"column(namespace);size(200);pk"`
	AccountType   string `orm:"column(account_type)";size(45)`
	AccountID     string `orm:"column(account_id)";size(100)`
	AccountDetail string `orm:"column(account_detail)";size(600)`
}

// TableName return tenx_user_3rd_account
func (m *ThirdPartyAccountModel) TableName() string {
	return "tenx_user_3rd_account"
}

// Insert adds a new third party account to database
func (m *ThirdPartyAccountModel) Insert(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	switch common.DType {
	case common.DRMySQL:
		sql := `INSERT INTO %s (namespace, account_type, account_id, account_detail) VALUES (?, ?, ?, ?);`
		sql = fmt.Sprintf(sql, m.TableName())
		_, err := o.Raw(sql, m.Namespace, m.AccountType, m.AccountID, m.AccountDetail).Exec()
		errcode, err := sqlstatus.ParseErrorCode(err)
		return errcode, err
	case common.DRPostgres:
		sql := fmt.Sprintf(`INSERT INTO "%s" ("namespace", "account_type", "account_id", "account_detail") VALUES (?, ?, ?, ?);`, m.TableName())
		_, err := o.Raw(sql, m.Namespace, m.AccountType, m.AccountID, m.AccountDetail).Exec()
		errcode, err := sqlstatus.ParseErrorCode(err)
		return errcode, err
	}
	return sqlstatus.SQLErrUnAuthorized, fmt.Errorf("driver %v not supported", common.DType)
}

// Exist check whether the namespace exists
func (m *ThirdPartyAccountModel) NamespaceExist() bool {
	o := orm.NewOrm()
	return o.QueryTable(m.TableName()).Filter("namespace", m.Namespace).
		Filter("account_type", m.AccountType).Exist()
}

// Exist check whether the third party account exists
func (m *ThirdPartyAccountModel) Exist() bool {
	o := orm.NewOrm()
	return o.QueryTable(m.TableName()).Filter("account_type", m.AccountType).
		Filter("account_id", m.AccountID).Exist()
}

// GetByAccountTypeAndID fetch a third party record by account type and id
func (m *ThirdPartyAccountModel) GetByAccountTypeAndID() (uint32, error) {
	o := orm.NewOrm()
	err := o.QueryTable(m.TableName()).Filter("account_type", m.AccountType).Filter("account_id", m.AccountID).One(m)
	return sqlstatus.ParseErrorCode(err)
}

// DeleteByNamespaceAndAccountType deletes a third party account record
func (m *ThirdPartyAccountModel) DeleteByNamespaceAndAccountType() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("DELETE FROM %s where namespace=? AND account_type=?;", m.TableName())
	_, err := o.Raw(sql, m.Namespace, m.AccountType).Exec()
	return sqlstatus.ParseErrorCode(err)
}

// GetThirdPartyAccountsByNamespace get all third party accounts for a user
func (m *ThirdPartyAccountModel) GetThirdPartyAccountsByNamespace() ([]ThirdPartyAccountModel, uint32, error) {
	o := orm.NewOrm()
	var result []ThirdPartyAccountModel
	_, err := o.QueryTable(m.TableName()).Filter("namespace", m.Namespace).All(&result)
	errcode, err := sqlstatus.ParseErrorCode(err)
	return result, errcode, err
}
