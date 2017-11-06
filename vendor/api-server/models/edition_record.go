/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-12-23  @author YangYuBiao
 */

package models

import (
	"fmt"
	"time"

	"api-server/models/user"
	"api-server/modules/transaction"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
	"github.com/pborman/uuid"
)

const (
	//EditionNormal is normal
	EditionNormal = iota //普通版
	//EditionStandard is standard
	EditionStandard //专业版
	//EditionProfessional is professional
	EditionProfessional //企业版
)

var (
	//StandardPrice is price of standard
	StandardPrice, _ = beego.AppConfig.Int64("standard")
	//ProfessionalPrice is price of professional
	ProfessionalPrice, _ = beego.AppConfig.Int64("professional")
)

//EditionRecord is table user_edition_record
type EditionRecord struct {
	ID                    int       `orm:"pk;column(id)" json:"-"`
	UserName              string    `orm:"column(user_name)" json:"user_name"`
	StartTime             time.Time `orm:"column(start_time)" json:"start_time"`
	EndTime               time.Time `orm:"column(end_time)" json:"end_time"`
	EnvEdition            uint8     `orm:"column(env_edition)" json:"env_edition"`
	ChargeAmount          int64     `orm:"column(charge_amount)" json:"charge_amount"`
	OpTime                time.Time `orm:"column(op_time)" json:"op_time"`
	Status                int       `orm:"column(status)" json:"status"`
	ConsumeptionHistoryID string    `orm:"column(consumeption_id)" json:"consumeption_id"`
}

type EditionResponse struct {
	Data  []EditionRecord `json:"data"`
	Total int64           `json:"total"`
}

func setPrice(month int) {
	StandardPrice = 990000
	if 3 <= month {
		StandardPrice = 890000
	}
	if 12 <= month {
		StandardPrice = 790000
	}
}

//EditionParam is param used by other package
type EditionParam struct {
	Namespace string
	Teamspace string
	Month     int
	Env       uint8
	SpaceType int
}

type sqlParam struct {
	tablename   string
	needBalance int64
	EditionParam
}

//NewEditionRecord get point of UserEditionRecord
func NewEditionRecord() *EditionRecord {
	return &EditionRecord{}
}

//TableName get current tablename
func (u *EditionRecord) TableName() string {
	return "tenx_user_edition_records"
}

func (u *EditionRecord) getUserAccountTable() string {
	return "tenx_user_account"
}

func (u *EditionRecord) getUserTable() string {
	return "tenx_users"
}

func (u *EditionRecord) getTeamspaceAccountTable() string {
	return "tenx_team_space_account"
}

//GetUserEditionRecord getUserEditionRecord
func (u *EditionRecord) GetUserEditionRecord(name string, filter map[string]int) (EditionResponse, error) {
	method := "GetEditionRecord"
	var offset int
	size := 1
	status := 1
	if filter != nil {
		offset = filter["offset"]
		size = filter["size"]
		status = filter["status"]
	}
	o := orm.NewOrm()
	result := make([]EditionRecord, size)
	var sql string
	sql = `select * from %s where user_name = ? and status = ? order by op_time desc limit ? offset ?`
	if status == 2 {
		sql = `select * from %s where user_name = ? and 2 = ? order by op_time desc limit ? offset ?`
	}
	sql = fmt.Sprintf(sql, u.TableName())
	_, err := o.Raw(sql, name, status, size, offset).QueryRows(&result)
	if err != nil {
		glog.Errorln(method, "get user edition records failed", err)
		return EditionResponse{
			Total: 0,
		}, err
	}
	if 0 == offset {
		return EditionResponse{
			Total: 1,
			Data:  result,
		}, err
	}
	total, err := o.QueryTable(u.TableName()).
		Filter("user_name", name).
		Count()
	if err != nil {
		glog.Errorln(method, "get user edition record total amount failed", err)
		return EditionResponse{
			Total: 0,
		}, err
	}
	if 0 == total {
		return EditionResponse{
			Total: 0,
		}, err
	}
	return EditionResponse{
		Data:  result,
		Total: total,
	}, err
}

//PayEdition payedition
func (u *EditionRecord) PayEdition(params EditionParam) (*EditionRecord, bool, bool, error) {
	setPrice(params.Month)
	var price int64
	if params.Env == EditionStandard {
		price = StandardPrice
	}
	if params.Env == EditionProfessional {
		price = ProfessionalPrice
	}
	needBalance := price * int64(params.Month)
	return u.commitByTranaction(sqlParam{
		EditionParam: params,
		needBalance:  needBalance,
	}, params.Env)
}

func (u *EditionRecord) commitByTranaction(sqlParam sqlParam, env uint8) (*EditionRecord, bool, bool, error) {
	method := "commitByTranaction"
	t := transaction.New()
	o := t.O()
	var err error
	var isOk bool
	var balance int64
	var historyID string
	var now time.Time
	var retuenResult *EditionRecord
	var isInsert bool = false
	t.Do(func() {
		balance, err = u.getBalanceByNamespace(o, sqlParam)
		if err != nil {
			t.Rollback(method, fmt.Sprintf("get user balance failed, user is %s", sqlParam.Namespace), err)
			isOk = false
			return
		}
		if balance < int64(sqlParam.needBalance) {
			t.Rollback()
			isOk = false
		}
	}).Do(func() {
		var name string
		var tablename string
		if 0 == sqlParam.SpaceType {
			name = sqlParam.Namespace
			tablename = u.getUserAccountTable()
		} else {
			name = sqlParam.Teamspace
			tablename = u.getTeamspaceAccountTable()
		}
		sql := fmt.Sprintf("update %s set balance = balance - ? where namespace = ? ", tablename)
		sqlResult, err := o.Raw(sql, sqlParam.needBalance, name).Exec()
		if err != nil {
			isOk = false
			t.Rollback(method, fmt.Sprintf("failed to update balance, user is %s", sqlParam.Namespace), err)
			return
		}
		r, err := sqlResult.RowsAffected()
		if err != nil {
			isOk = false
			t.Rollback(method, fmt.Sprintf("failed to update balance, user is %s", sqlParam.Namespace), err)
			return
		}
		if r == 0 {
			isOk = false
			t.Rollback(method, fmt.Sprintf("failed to update balance, user is %s", sqlParam.Namespace))
			return
		}
	}).Do(func() {
		//get new balance
		balance, err = u.getBalanceByNamespace(o, sqlParam)
		if err != nil {
			isOk = false
			t.Rollback(method, "get new balance failed, namespace is %s, teamspace is %s", sqlParam.Namespace, sqlParam.Teamspace)
			return
		}
		//update env_edition
		if balance >= 0 {
			value := orm.Params{
				"env_edition": env,
			}
			_, err := o.QueryTable(u.getUserTable()).Filter("namespace", sqlParam.Namespace).Update(value)
			if err != nil {
				isOk = false
				t.Rollback(method, fmt.Sprintf("update user env edition failed, namespace is %s", sqlParam.Namespace), err)
				return
			}
			return
		}
		isOk = false
		t.Rollback(method, fmt.Sprintf("update user env edition failed,the balance is not enough, namespace is %s", sqlParam.Namespace))
	}).Do(func() {
		//update user_edition_record
		var result EditionRecord
		historyID = uuid.New()
		err = o.QueryTable(u.TableName()).Filter("user_name", sqlParam.Namespace).Filter("status", 1).One(&result)
		if nil != err && orm.ErrNoRows != err {
			t.Rollback(method, fmt.Sprintf("get user edition record failed, namespace is %s", sqlParam.Namespace), err)
			isOk = false
			return
		}
		now = time.Now()
		if 0 == result.ID {
			isInsert = true
			insertEntity := &EditionRecord{
				UserName:              sqlParam.Namespace,
				StartTime:             now,
				EndTime:               now.AddDate(0, sqlParam.Month, 0),
				EnvEdition:            sqlParam.Env,
				ChargeAmount:          sqlParam.needBalance,
				OpTime:                now,
				Status:                1,
				ConsumeptionHistoryID: historyID,
			}
			retuenResult = insertEntity
			_, err = o.Insert(insertEntity)
			if nil != err {
				t.Rollback(method, fmt.Sprintf("insert edition record failed, namespace is %s", sqlParam.Namespace), err)
				isOk = false
				return
			}
		} else {
			//update old record to status = 0
			value := orm.Params{
				"status": 0,
			}
			_, err = o.QueryTable(u.TableName()).Filter("user_name", sqlParam.Namespace).Filter("status", 1).Update(value)
			if nil != err {
				t.Rollback(method, fmt.Sprintf("update edition record failed, namespace is %s", sqlParam.Namespace), err)
				isOk = false
				return
			}
			newRecord := &EditionRecord{
				UserName:              sqlParam.Namespace,
				StartTime:             result.StartTime,
				EndTime:               result.EndTime.AddDate(0, sqlParam.Month, 0),
				EnvEdition:            sqlParam.Env,
				ChargeAmount:          sqlParam.needBalance,
				OpTime:                time.Now(),
				Status:                1,
				ConsumeptionHistoryID: historyID,
			}
			isInsert = false
			retuenResult = newRecord
			_, err := o.Insert(newRecord)
			if nil != err {
				t.Rollback(method, fmt.Sprintf("insert  new edition record failed, namespace is %s", sqlParam.Namespace), err)
				isOk = false
				return
			}
		}
	}).Do(func() {
		var price int64
		if sqlParam.Env == EditionStandard {
			price = StandardPrice
		} else {
			price = ProfessionalPrice
		}

		var envName = "专业版"
		if 2 == sqlParam.Env {
			envName = "企业版"
		}
		op := "升级版本至"
		if false == isInsert {
			op = "续费"
		}
		//insert into pay history
		consumption := user.UsersConsumption{
			ConsumptionID:  historyID,
			Namespace:      sqlParam.Namespace,
			Type:           "6",
			Amount:         int(sqlParam.needBalance),
			StartTime:      now,
			TotalTime:      sqlParam.Month,
			Name:           fmt.Sprintf("%s%s", op, envName),
			Price:          int(price),
			HostingCluster: "-",
			CreateTime: time.Now(),
		}
		_, err := o.Insert(&consumption)
		if nil != err {
			t.Rollback(method, fmt.Sprintf("insert tenx_users_consumption_history failed, namespace is %s", sqlParam.Namespace), err)
			isOk = false
			return
		}
	}).Done()
	if t.IsCommit() == true {
		return retuenResult, true, isInsert, nil
	}
	return nil, false, false, nil
}

func (u *EditionRecord) getBalanceByNamespace(o orm.Ormer, sqlParam sqlParam) (int64, error) {
	spaceType := sqlParam.SpaceType
	var err error
	var balance int64
	if 0 == int32(spaceType) {
		userModel := user.NewUserAccount()
		balance, err = userModel.GetBalanceByNamespace(sqlParam.Namespace)
	} else {
		balance, err = getTeamBalanceByNamespace(sqlParam.Teamspace)
	}
	return balance, err
}

//SpaceAccountModel teamspace balance
type SpaceAccountModel struct {
	Balance int64 `orm:"column(balance)"`
}

func getTeamBalanceByNamespace(namespace string) (int64, error) {
	method := "GetBalanceByNamespace"
	o := orm.NewOrm()
	result := SpaceAccountModel{}
	if err := o.QueryTable("tenx_team_space_account").Filter("namespace", namespace).One(&result, "balance"); err != nil && err != orm.ErrNoRows {
		glog.Errorln(method, "failed.", err)
		return 0, err
	} else if err == orm.ErrNoRows {
		glog.Errorln(method, "no balance record", err)
		return 0, nil
	}
	return result.Balance, nil
}
