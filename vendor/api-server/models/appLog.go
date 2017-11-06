/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-27  @author mengyuan
 */

package models

import (
	"fmt"
	"time"

	"api-server/models/common"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

// app operations
const (
	AppOpCreateApp   = 0
	AppOpMod         = 1
	AppOpAddSvc      = 2
	AppOpDelSvc      = 3
	AppOpStopApp     = 4
	AppOpStartApp    = 5
	AppOpRestartApp  = 6
	AppOpDeleteApp   = 7
	AppOpStopSvc     = 8
	AppOpStartSvc    = 9
	AppOpRedeploySvc = 10
)

// app operate result
const (
	AppOpResultSucc = 1
	AppOpResultFail = 2
)

// AppLog is the structrue of table tenx_app
type AppLog struct {
	ID              uint64    `orm:"pk;column(id)"`
	UserID          int32     `orm:"column(user_id)"`
	Namespace       string    `orm:"size(255);column(namespace)"`
	HostCluster     string    `orm:"size(45);column(hosting_cluster)"`
	AppID           string    `orm:"size(36);column(appid)"`
	Name            string    `orm:"size(45);column(name)"`
	Operation       uint8     `orm:"column(operation)"`
	OperationDetail string    `orm:"size(1000);column(operation_detail)"`
	Result          uint8     `orm:"column(result)"` // 1: success, 2: fail
	CreateTime      time.Time `orm:"type(datetime);column(creation_time)"`
}

// AppOperationLog app operation log
type AppOperationLog struct {
	OperationCode uint8  `json:"operation_code"`
	Operation     string `json:"operation"`
	Result        string `json:"result"`
	Time          string `json:"time"`
	Detail        string `json:"detail"`
}

// TableName get table name
func (t *AppLog) TableName() string {
	return "tenx_app_log"
}

func NewAppLog() *AppLog {
	return &AppLog{}
}

// Insert insert one log
func (t *AppLog) Insert(orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	_, err := o.Insert(t)
	if err != nil {
		glog.Errorln("Insert app table failed.", err)
	}
	return err
}

// InsertMulti insert multi logs
func (t *AppLog) InsertMulti(logs []AppLog, orms ...orm.Ormer) (int64, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	num, err := o.InsertMulti(len(logs), logs)
	if err != nil {
		glog.Errorln("Insert app_log table failed.", err)
	} else if num != int64(len(logs)) {
		glog.Errorf("Insert app_log table failed. insert %v, success %v\n", len(logs), num)
	}
	return num, err
}
func init() {
	orm.RegisterModel(new(AppLog))
}

// get app operation log
func (t *AppLog) get(appName, namespace, clusterID string, from, size int) ([]AppLog, error) {
	o := orm.NewOrm()

	var app App
	err := o.QueryTable(app.TableName()).Filter("name", appName).Filter("namespace", namespace).Filter("hosting_cluster", clusterID).One(&app)
	if err != nil {
		return nil, err
	}

	var logs []AppLog
	_, err = o.QueryTable(t.TableName()).Filter("appid", app.AppID).Filter("namespace", namespace).Filter("hosting_cluster", clusterID).OrderBy("-creation_time").Offset(from).Limit(size).All(&logs)
	return logs, err
}

// Get get app operation log
func (t *AppLog) Get(appName, namespace, clusterID string, from, size int) ([]AppOperationLog, error) {
	var logs []AppOperationLog
	logRecords, err := t.get(appName, namespace, clusterID, from, size)

	for _, r := range logRecords {
		logs = append(logs, AppOperationLog{
			OperationCode: r.Operation,
			Operation:     GetOpStr(int(r.Operation)),
			Detail:        r.OperationDetail,
			Result:        GetOpResultStr(int(r.Result)),
			Time:          r.CreateTime.Format(TimeLayout),
		})
	}
	return logs, err
}

// GetOpStr convert operation code to string
func GetOpStr(code int) string {
	switch code {
	case AppOpCreateApp:
		return "create app"
	case AppOpMod:
		return "modify app"
	case AppOpAddSvc:
		return "add service"
	case AppOpDelSvc:
		return "delete service"
	case AppOpStopApp:
		return "stop app"
	case AppOpStartApp:
		return "start app"
	case AppOpRestartApp:
		return "restart app"
	case AppOpDeleteApp:
		return "delete app"
	case AppOpStopSvc:
		return "stop service"
	case AppOpStartSvc:
		return "start service"
	case AppOpRedeploySvc:
		return "redeploy service"
	default:
		return "unknown operation"
	}
}

// GetOpResultStr get operation result string
func GetOpResultStr(code int) string {
	switch code {
	case AppOpResultSucc:
		return "success"
	case AppOpResultFail:
		return "fail"
	default:
		return "unknown result"
	}
}

func (t *AppLog) GetCountByType(namespaceList, clusterList []string, beginTime, endTime string) (map[string]int, error) {
	method := "GetCountByType"

	qb, err := orm.NewQueryBuilder(common.DType.String())
	if err != nil {
		glog.Errorln(method, "NewQueryBuilder failed.", err)
		return nil, err
	}
	qb.Select("operation, COUNT(*) as cnt").
		From(t.TableName()).
		Where("creation_time > ?").
		And("creation_time <= ?").
		And(fmt.Sprintf("result =  %v", AppOpResultSucc))

	// copy slice to avoid modify original parameters
	// qb.In() doesnot add quote (sql is like namespace IN (space1, space2), it's wrong
	// so we add qoute ourself
	var spaceListCopy []string
	var clusterListCopy []string
	for _, space := range namespaceList {
		spaceListCopy = append(spaceListCopy, "'"+space+"'")
	}
	for _, cluster := range clusterList {
		clusterListCopy = append(clusterListCopy, "'"+cluster+"'")
	}
	if len(spaceListCopy) > 0 {
		qb.And("namespace").In(spaceListCopy...)
	}
	if len(clusterListCopy) > 0 {
		qb.And("hosting_cluster").In(clusterListCopy...)
	}
	qb.GroupBy("operation")
	sql := qb.String()

	o := orm.NewOrm()
	var result []struct {
		Operation int
		Cnt       int
	}
	o.Raw(sql, beginTime, endTime).QueryRows(&result)

	// slice to map
	countMap := make(map[string]int)
	for _, r := range result {
		countMap[convertAppOp(r.Operation)] = r.Cnt
	}
	return countMap, nil
}

// convertAppOp convert int code to readable string
func convertAppOp(code int) string {
	opMap := map[int]string{
		AppOpCreateApp:   "appCreate",
		AppOpMod:         "appModify",
		AppOpAddSvc:      "svcCreate",
		AppOpDelSvc:      "svcDelete",
		AppOpStopSvc:     "svcStop",
		AppOpStartSvc:    "svcStart",
		AppOpRedeploySvc: "svcRedeploy",
		AppOpStopApp:     "appStop",
		AppOpStartApp:    "appStart",
		AppOpRestartApp:  "appRedeploy",
		AppOpDeleteApp:   "appDelete",
	}
	op, ok := opMap[code]
	if !ok {
		return "unknown"
	}
	return op
}
