/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-14  @author mengyuan
 */
package alert

import (
	"time"

	"fmt"

	"api-server/util/misc"

	"api-server/models/sql/util"

	"github.com/astaxie/beego/orm"
)

const (
	InvitationStatusNotAccepted = 0
	InvitationStatusAccepted    = 1
	InvitationStatusNotSend     = 2
)

type NotifyReceiverInvitation struct {
	ID         int64     `orm:"pk;column(id)"`
	User       string    `orm:"column(user)"`
	Email      string    `orm:"column(email)"`
	Code       string    `orm:"column(code)"`
	Status     int       `orm:"column(status)"`
	CreateTime time.Time `orm:"column(create_time)"`
	ModifyTime time.Time `orm:"column(modify_time)"`
}

func NewNotifyReceiverInvitation() *NotifyReceiverInvitation {
	return new(NotifyReceiverInvitation)
}
func (t *NotifyReceiverInvitation) TableName() string {
	return "tenx_alert_notify_receiver_invitation"
}

func (t *NotifyReceiverInvitation) GetMultiLock(user string, emails []string, o orm.Ormer) ([]NotifyReceiverInvitation, error) {
	safeEmailList := util.EscapeSliceForInjection(emails)
	emailsStr := misc.SliceToString(safeEmailList, "'", ",")
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE user = ? AND email IN (%s) FOR UPDATE`, t.TableName(), emailsStr)
	invitations := make([]NotifyReceiverInvitation, 0)
	_, err := o.Raw(sql, user).QueryRows(&invitations)
	return invitations, err
}

func (t *NotifyReceiverInvitation) ModifyStatus(code string, status int) (int64, error) {
	now := time.Now()
	return orm.NewOrm().QueryTable(t.TableName()).
		Filter("code", code).
		Update(orm.Params{
			"status":      status,
			"modify_time": now,
		})
}

func (t *NotifyReceiverInvitation) GetStatus(user string, emails []string) (map[string]int, error) {
	var items []NotifyReceiverInvitation
	_, err := orm.NewOrm().QueryTable(t.TableName()).
		Filter("user", user).
		Filter("email__in", emails).
		All(&items, "Email", "Status")
	result := make(map[string]int)
	for i := range items {
		result[items[i].Email] = items[i].Status
	}
	return result, err

}

func (t *NotifyReceiverInvitation) GetEmailStatus(user string, emails []string) (status map[string]int, err error) {
	invitations := make([]NotifyReceiverInvitation, 0)
	status = make(map[string]int)
	_, err = orm.NewOrm().QueryTable(t.TableName()).
		Filter("user", user).
		Filter("email__in", emails).
		All(&invitations)
	if err != nil {
		return
	}
	for _, invitation := range invitations {
		status[invitation.Email] = invitation.Status
	}
	for _, email := range emails {
		if _, find := status[email]; !find {
			status[email] = InvitationStatusNotSend
		}
	}
	return
}
