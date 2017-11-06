/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-12-12  @author Lei
 */

/*
 * WARNING: Only for Public Cloud usage
 * Used for user/enterprise certificate process
 *
 */

package models

import (
	"api-server/models/common"
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
)

// CertificateModel is the model for user/enterprise certificate
type CertificateModel struct {
	CertID               string    `orm:"pk;column(id)" json:"certID"`
	UserName             string    `orm:"size(50);;column(user_name)" json:"userName"`
	CertType             int       `orm:"column(cert_type)" json:"certType"`
	CertUserName         string    `orm:"size(50);column(cert_user_name)" json:"certUserName"`
	CertUserID           string    `orm:"size(25);column(cert_user_id)" json:"certUserID"`
	EnterpriseOwnerPhone string    `orm:"size(20);column(enterprise_owner_phone)" json:"enterpriseOwnerPhone"`
	UserHoldPic          string    `orm:"size(100);column(user_hold_pic)" json:"userHoldPic"`
	UserScanPic          string    `orm:"size(100);column(user_scan_pic)" json:"userScanPic"`
	EnterpriseName       string    `orm:"size(100);column(enterprise_name)" json:"enterpriseName"`
	EnterpriseCertID     string    `orm:"size(50);column(enterprise_cert_id)" json:"enterpriseCertID"`
	EnterpriseCertPic    string    `orm:"size(100);column(enterprise_cert_pic)" json:"enterpriseCertPic"`
	Status               int       `orm:"column(status)" json:"status"`
	FailureMessage       string    `orm:"size(200);column(failure_message)" json:"failureMessage"`
	CreateTime           time.Time `orm:"type(datetime);column(create_time)" json:"createTime"`
	CertificatedTime     time.Time `orm:"type(datetime);column(certificated_time)" json:"certificatedTime"`
}

// TableName returns the name of table in database
func (cm *CertificateModel) TableName() string {
	return "tenx_certificates"
}

// NewCertificate create a new certificate object
func NewCertificate() *CertificateModel {
	return &CertificateModel{}
}

// ListByUser list the certificates that this user created
func (cm *CertificateModel) ListByUser(user string) ([]CertificateModel, error) {
	o := orm.NewOrm()
	var results []CertificateModel
	_, err := o.QueryTable(cm.TableName()).Filter("user_name", user).All(&results)
	return results, err
}

// Insert insert template to the database
func (cm *CertificateModel) Insert() (int64, error) {
	o := orm.NewOrm()
	cm.CreateTime = time.Now()

	switch common.DType {
	case common.DRMySQL:
		count, err := o.Insert(cm)
		return count, err
	case common.DRPostgres:
		panic("Not implemented yet")
	}

	return 0, fmt.Errorf("Driver %s not supported", common.DType)
}

// Update updates one or more fields for one time
func (cm *CertificateModel) Update() error {
	o := orm.NewOrm()
	params := orm.Params{}
	params["cert_user_name"] = cm.CertUserName
	params["cert_user_id"] = cm.CertUserID
	params["enterprise_owner_phone"] = cm.EnterpriseOwnerPhone
	params["user_hold_pic"] = cm.UserHoldPic
	params["user_scan_pic"] = cm.UserScanPic
	params["enterprise_name"] = cm.EnterpriseName
	params["enterprise_cert_id"] = cm.EnterpriseCertID
	params["enterprise_cert_pic"] = cm.EnterpriseCertPic

	_, err := o.QueryTable(cm.TableName()).Filter("id", cm.CertID).Filter("user_name", cm.UserName).Update(params)

	return err
}

// GetCertCount check if the user already has the related cert type
func (cm *CertificateModel) GetCertCount(user string, certType int) (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable(cm.TableName()).Filter("id", cm.CertID).Filter("user_name", user).Filter("cert_type", certType).Count()

	return count, err
}

// GetByID get existing certificate by ID
func (cm *CertificateModel) GetByID(user string, certID string) error {
	o := orm.NewOrm()
	err := o.QueryTable(cm.TableName()).Filter("user_name", user).Filter("id", certID).One(cm)
	return err
}
