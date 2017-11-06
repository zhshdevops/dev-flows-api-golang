/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-12-12  @author mengyuan
 */

package team

import (
	"time"

	"errors"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

type InvitationStatus int

const (
	InvitationStatusWaiting = 0
	InvitationStatusCancel  = 1
	InvitationStatusAccept  = 2
)

func (s InvitationStatus) String() string {
	switch s {
	case InvitationStatusWaiting:
		return "Waiting"
	case InvitationStatusCancel:
		return "Cancel"
	case InvitationStatusAccept:
		return "Accept"
	default:
		return "Unknown"
	}
}

type Invitation struct {
	ID             int
	InvitationCode string           `orm:"column(invitation_code)"`
	Email          string           `orm:"column(email);size(100)"`
	TeamID         string           `orm:"column(team_id)"`
	TeamName       string           `orm:"column(team_name)"`
	UserID         int32            `orm:"column(user_id)"`
	UserName       string           `orm:"column(user_name)"`
	CreateTime     time.Time        `orm:"column(create_time)"`
	AcceptTime     time.Time        `orm:"column(accept_time)"`
	Status         InvitationStatus `orm:"column(status)"`
}

func NewInvitation() *Invitation {
	return &Invitation{}
}
func (t *Invitation) TableName() string {
	return "tenx_team_invitation"
}

type invitationInfo struct {
	Email string
}

func (t *Invitation) GetUsersEmail(teamID string, status int) ([]string, error) {
	method := "GetUsers"
	o := orm.NewOrm()
	result := make([]orm.Params, 0, 1)
	if _, err := o.QueryTable(t.TableName()).
		Filter("team_id", teamID).
		Filter("status", status).
		GroupBy("email").
		Values(&result, "email"); err != nil {
		glog.Error(method, "get invation info failed.", err)
		return nil, err
	}
	emails := make([]string, 0, 1)
	for i := range result {
		emails = append(emails, result[i]["Email"].(string))
	}

	return emails, nil
}

func (t *Invitation) GetByCode(code string) (orm.Params, error) {
	o := orm.NewOrm()
	var result []orm.Params
	_, err := o.QueryTable(t.TableName()).
		Filter("invitation_code", code).
		Values(&result, "invitation_code", "email", "team_id", "team_name", "status")
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return orm.Params{}, orm.ErrNoRows
	}
	return result[0], nil
}

// Add insert multi emails and invitation codes to DB, len(emails) must equal to len(codes)
func (t *Invitation) Add(emails []string, codes []string, teamID string, teamName string, userID int32, userName string) error {
	if len(emails) != len(codes) {
		return errors.New("emails and codes length is not equal")
	}
	invitations := make([]Invitation, 0, len(emails))
	now := time.Now()
	for i := range emails {
		invitation := Invitation{
			InvitationCode: codes[i],
			Email:          emails[i],
			TeamID:         teamID,
			TeamName:       teamName,
			UserID:         userID,
			UserName:       userName,
			CreateTime:     now,
			Status:         InvitationStatusWaiting,
		}
		invitations = append(invitations, invitation)
	}

	o := orm.NewOrm()
	_, err := o.InsertMulti(len(invitations), invitations)
	return err
}

func (t *Invitation) UpdateStatus(sts InvitationStatus, email string, teamID string, orms ...orm.Ormer) error {
	method := "UpdateStatus"
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	newInfo := orm.Params{"status": sts}
	if sts == InvitationStatusAccept {
		newInfo["accept_time"] = time.Now()
	}
	cnt, err := o.QueryTable(t.TableName()).
		Filter("email", email).
		Filter("team_id", teamID).
		Filter("status", InvitationStatusWaiting).
		Update(newInfo)
	glog.Infoln(method, "update status modify line number", cnt)
	return err
}
