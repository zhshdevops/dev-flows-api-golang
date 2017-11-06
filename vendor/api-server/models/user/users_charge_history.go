/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-12-02  @author mengyuan
 */

package user

import (
	"api-server/models/common"
	"time"

	sqlstatus "api-server/models/sql/status"
	"fmt"

	"strconv"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

const (
	OrderTypePrivate             = 0   // private cloud
	OrderTypeWeixin              = 100 // 微信
	OrderTypeAlipay              = 101 // 支付宝
	OrderTypeRemittance          = 102 // 线下汇款
	OrderTypeRefundByDismissTeam = 103 // 解散团队的退款
)

type OrderStatus uint8

const (
	OrderStatusSuccess = 0
	OrderStatusNotPay  = 1
	OrderStatusFailed  = 2
)

// ChargePurpose for edtion upgrade only for now
// 0 = balance only and upgrade failed
// 1 = balance only and renew failed
// 2 = balance and upgrade success
// 3 = balance and renew success
type ChargePurpose struct {
	Upgrade    int       `json:"upgrade"`
	Duration   int       `json:"duratioin"`
	ChargeType int       `json:"charge_type"`
	EndTime    time.Time `json:"end_time"`
}

type UsersCharge struct {
	ID              int         `orm:"pk;column(id)" json:"-"`
	TenxOrderID     string      `orm:"column(tenx_order_id)" json:"-"`
	Namespace       string      `orm:"column(namespace)" json:"-"`
	NamespaceType   int32       `orm:"column(namespace_type)" json:"-"`
	OperatorID      int32       `orm:"column(operator_id)" json:"-"`
	OperatorName    string      `orm:"column(operator_name)" json:"operator"`
	OldBalance      int64       `orm:"column(old_balance)" json:"before"`   // unit fen
	ChargeAmount    int64       `orm:"column(charge_amount)" json:"charge"` // unit fen
	NewBalance      int64       `orm:"column(new_balance)" json:"after"`    // unit fen
	OrderID         string      `orm:"column(order_id)" json:"orderID"`
	OrderType       uint8       `orm:"column(order_type)" json:"orderType"`
	CreateTime      time.Time   `orm:"column(create_time)" json:"-"`
	ChargeTime      time.Time   `orm:"column(charge_time)" json:"time"`
	Status          OrderStatus `orm:"column(status)" json:"-"`
	Detail          string      `orm:"column(detail)" json:"detail"`
	VerificationKey string      `orm:"column(verification_key)" json:"verification_key"`
	Purpose         string      `orm:"column(purpose)" json:"-"`
}

func NewUsersCharge() *UsersCharge {
	return &UsersCharge{}
}

// IsValidOrderType return whether it's valid order type
func IsValidOrderType(orderType uint8) bool {
	switch orderType {
	case OrderTypePrivate, OrderTypeWeixin, OrderTypeAlipay, OrderTypeRemittance, OrderTypeRefundByDismissTeam:
		return true
	}
	return false
}

func (t *UsersCharge) TableName() string {
	return "tenx_users_charge_history"
}

// LockRow lock row to be updated
// Return current status and key for verification
func (t *UsersCharge) LockRow(o orm.Ormer) error {
	err := o.Raw("select namespace, operator_id, namespace_type, status, verification_key, new_balance, create_time, purpose from "+t.TableName()+" where tenx_order_id = ? for update", t.TenxOrderID).QueryRow(t)
	return err
}

// CreatePaymentRecord create a prepay record for update later
func (t *UsersCharge) CreatePaymentRecord(orms ...orm.Ormer) (int64, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	if t.CreateTime.IsZero() {
		t.CreateTime = time.Now()
	}

	switch common.DType {
	case common.DRMySQL:
		count, err := o.Insert(t)
		return count, err
	case common.DRPostgres:
		panic("Not implemented yet")
	}

	return 0, fmt.Errorf("Driver %s not supported", common.DType)
}

func (t *UsersCharge) Update(orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	// Support to update the attributes below
	newInfo := orm.Params{"old_balance": t.OldBalance}
	newInfo["new_balance"] = t.NewBalance
	newInfo["charge_amount"] = t.ChargeAmount
	newInfo["order_id"] = t.OrderID

	newInfo["charge_time"] = t.ChargeTime
	newInfo["status"] = t.Status
	newInfo["detail"] = t.Detail
	_, err := o.QueryTable(t.TableName()).
		Filter("tenx_order_id", t.TenxOrderID).Update(newInfo)
	return sqlstatus.ParseErrorCode(err)
}

func (t *UsersCharge) UpdateChargePurpose() (uint32, error) {
	o := orm.NewOrm()

	// Support to update the attributes below
	newInfo := orm.Params{"purpose": t.Purpose}
	_, err := o.QueryTable(t.TableName()).
		Filter("tenx_order_id", t.TenxOrderID).Update(newInfo)
	return sqlstatus.ParseErrorCode(err)
}

// GetRecordByOrderID get the record by tenxcloud order id
func (t *UsersCharge) GetRecordByOrderID(tenxOrderID string) error {
	o := orm.NewOrm()
	err := o.QueryTable(t.TableName()).Filter("tenx_order_id", tenxOrderID).One(t)
	return err
}

// GetChargeDetail get charge history
func (t *UsersCharge) GetChargeDetail(offset, size int, space string) ([]UsersCharge, int64, error) {
	method := "GetChargeDetail"

	o := orm.NewOrm()
	result := make([]UsersCharge, 0, size)
	// get charge record
	sql := `SELECT operator_name, old_balance, charge_amount, new_balance, charge_time, order_type, detail
	FROM %s
	WHERE namespace = ? AND status = %d
	ORDER BY charge_time DESC
	LIMIT ? OFFSET ?`
	sql = fmt.Sprintf(sql, t.TableName(), OrderStatusSuccess)
	_, err := o.Raw(sql, space, size, offset).QueryRows(&result)
	if err != nil {
		glog.Errorln(method, "get charge detail failed.", err)
		return nil, 0, err
	}
	// get total amount
	total, err := o.QueryTable(t.TableName()).
		Filter("namespace", space).
		Filter("status", OrderStatusSuccess).
		Count()
	if err != nil {
		glog.Errorln(method, "get charge total amount failed.", err)
		return nil, 0, err
	}
	return result, total, nil

}

// GetOrderID Return the order ID from TenxCloud
func (t *UsersCharge) GetOrderID() string {
	return time.Now().Format("20060102150405") + strconv.Itoa(time.Now().Nanosecond())
}
