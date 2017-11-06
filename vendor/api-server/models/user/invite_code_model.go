/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhang shouhong
 */

package user

import (
	"time"

	"github.com/astaxie/beego/orm"
)

// InviteCodeModel tenx_users record
type InviteCodeModel struct {
	Code      string    `orm:"column(code);pk"`
	Email     string    `orm:"column(email);size(100)"`
	BoundTime time.Time `orm:"column(bound_time);type(datetime)"`
}

// TableName return tenx_invite_codes
func (u *InviteCodeModel) TableName() string {
	return "tenx_invite_codes"
}

// IsUserInvited check whether the user is invited
func IsUserInvited(email, code string) (bool, error) {
	o := orm.NewOrm()
	if o.QueryTable("tenx_invite_codes").Filter("email", email).Exist() {
		return true, nil
	}
	if o.QueryTable("tenx_invite_codes").Filter("code", code).Filter("email", "").Exist() {
		sql := `UPDATE tenx_invite_codes SET email=?, bound_time=? WHERE code=? `
		_, err := o.Raw(sql, email, time.Now(), code).Exec()
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}
