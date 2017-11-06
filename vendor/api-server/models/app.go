/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-27  @author mengyuan
 */

package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"

	"api-server/models/common"
	sqlstatus "api-server/models/sql/status"
	sqlutil "api-server/models/sql/util"
)

// App is the structure of table app
type App struct {
	ID           uint64    `orm:"pk;column(id)"`
	AppID        string    `orm:"size(45);column(appid)"`
	UserID       int32     `orm:"column(user_id)"`
	Namespace    string    `orm:"size(255);column(namespace)"`
	HostCluster  string    `orm:"size(45);column(hosting_cluster)"`
	Name         string    `orm:"size(63);column(name)"`
	Description  string    `orm:"size(1024);column(description)"`
	CreateTime   time.Time `orm:"type(datetime);column(creation_time)"`
	ModifyTime   time.Time `orm:"type(datetime);column(modify_time)"`
	EntryService string    `orm:"size(255);column(entry_service)"`
	Relations    string    `orm:"type(text);column(relations)"`
	Phase        uint8     `orm:"type(tinyint);column(phase)"`
}

var (
	// SortByMapping is the mapping from display name to fields in app table
	SortByMapping = map[string]string{
		"name":         "name",
		"namespace":    "namespace",
		"creationTime": "creation_time",
	}
	// FilterByMapping is the mapping from display name to fields in app table
	FilterByMapping = map[string]string{
		"name":      "name",
		"namespace": "namespace",
		"clusterID": "hosting_cluster",
	}
)

// validateAppDataSelect filter out invalid sort, filter options, and return valid ones
func validateAppDataSelect(old *common.DataSelectQuery) (*common.DataSelectQuery, error) {
	if old == nil {
		return common.NewDataSelectQuery(common.DefaultPagination, common.NoSort, common.NoFilter), nil
	}

	dataselect := common.NewDataSelectQuery(old.PaginationQuery, common.NoSort, common.NoFilter)
	if old.FilterQuery != nil {
		for _, f := range old.FilterQuery.FilterByList {
			prop, ok := FilterByMapping[f.Property]
			if !ok {
				glog.Errorf("team list, invalid filter by options: %s\n", f.Property)
				return nil, fmt.Errorf("Invalid filter option: %s", f.Property)
			}
			f.Property = prop
			dataselect.FilterQuery.FilterByList = append(dataselect.FilterQuery.FilterByList, f)
		}
	}

	if old.SortQuery != nil {
		for _, sq := range old.SortQuery.SortByList {
			prop, ok := SortByMapping[sq.Property]
			if !ok {
				glog.Errorf("team list, invalid sort options: %s\n", sq.Property)
				return nil, fmt.Errorf("invalid sort option: %s\n", sq.Property)
			}
			sq.Property = prop
			dataselect.SortQuery.SortByList = append(dataselect.SortQuery.SortByList, sq)

		}
	}
	if old.PaginationQuery != nil {
		dataselect.PaginationQuery = common.NewPaginationQuery(old.PaginationQuery.From, old.PaginationQuery.Size)
	}
	return dataselect, nil
}

// 应用创建阶段
const (
	AppCreatePhaseProcessing = 1 // creating k8s resources
	AppCreatePhaseFinish     = 2 // finish
)

// TableName return table name
func (t *App) TableName() string {
	return "tenx_app"
}

func NewApp() *App {
	return &App{}
}

func (t *App) Exist(cluster string, namespace string, appName string) bool {
	o := orm.NewOrm()
	return o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("hosting_cluster", cluster).
		Filter("name", appName).Exist()
}

// Insert insert one
func (t *App) Insert(orms ...orm.Ormer) error {
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

// LockRow lock one row and return the app name
func (t *App) LockRow(o orm.Ormer, cluster string, namespace string, appName string) (string, error) {
	method := "models.app.LockRow"
	var names []string
	count, err := o.Raw("select name from "+t.TableName()+" where namespace = ? and hosting_cluster = ? and name = ? for update", namespace, cluster, appName).
		QueryRows(&names)
	glog.V(2).Infof("%s names:%v, count:%v, error:%v\n", method, names, count, err)
	if err != nil {
		return "", err
	}
	if count > 0 {
		return names[0], nil
	}
	return "", nil
}

// GetAppID get app <appName> ID
func (t *App) GetAppID(cluster string, namespace string, appNames []string) (map[string]string, error) {
	method := "AppModel.GetAppID"
	o := orm.NewOrm()
	var res []orm.Params
	var idMap = make(map[string]string)
	_, err := o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("hosting_cluster", cluster).
		Filter("name__in", appNames).
		Values(&res, "name", "appid")
	if err != nil {
		return idMap, err
	}
	for _, row := range res {
		var name, id interface{}
		var nameStr, idStr string
		var ok bool
		name, ok = row["Name"]
		id, ok = row["AppID"]
		nameStr, ok = name.(string)
		if !ok {
			glog.Errorln(method, "error")
			continue
		}
		idStr, ok = id.(string)
		if !ok {
			glog.Errorln(method, "error")
			continue
		}
		idMap[nameStr] = idStr
	}
	return idMap, nil
}

// DeleteMulti delete multi row
func (t *App) DeleteMulti(cluster string, namespace string, appName []string, orms ...orm.Ormer) (int64, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	num, err := o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("hosting_cluster", cluster).
		Filter("name__in", appName).
		Delete()
	return num, err
}

// ListByNamespace list apps by namespace only
func (t *App) ListByNamespace(namespace string, dataselect *common.DataSelectQuery) ([]App, uint32, error) {
	o := orm.NewOrm()
	o.Using(common.DatabaseDefault)

	dataselect, err := validateAppDataSelect(dataselect)
	dataselect.FilterQuery = common.NewFilterQuery([]string{"namespace__eq", namespace})

	if err != nil {
		glog.Errorf("list apps by namespace %s fails, error:%v\n", namespace, err)
		return []App{}, sqlstatus.SQLErrSyntax, err
	}

	var apps []App
	sql := fmt.Sprintf(`SELECT * FROM tenx_app where %s %s %s;`, dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	_, err = o.Raw(sql).QueryRows(&apps)
	if err != nil {
		errcode, err := sqlstatus.ParseErrorCode(err)
		glog.Errorf("get app under namspace %s fails, error code:%d, error:%v\n", namespace, errcode, err)
		return nil, errcode, err
	}
	return apps, sqlstatus.SQLSuccess, nil
}

// List list apps by namespace and cluster id
func (t *App) List(namespace, clusterID string, from, size int, filterName string, reverse bool) ([]App, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(t.TableName()).Filter("namespace", namespace).Filter("hosting_cluster", clusterID)

	var apps []App

	if filterName != "" {
		qs = qs.Filter("name__contains", sqlutil.EscapeUnderlineInLikeStatement(filterName))
	}
	total, err := qs.Count()
	if err != nil {
		return apps, 0, err
	}
	if reverse {
		qs = qs.OrderBy("creation_time")
	} else {
		qs = qs.OrderBy("-creation_time")
	}
	if from > 0 {
		qs = qs.Offset(from)
	}
	if size > 0 {
		qs = qs.Limit(size)
	}
	_, err = qs.All(&apps)
	return apps, total, err
}

// Get get app by name and cluster id
func (t *App) Get() error {
	o := orm.NewOrm()
	err := o.QueryTable(t.TableName()).Filter("name", t.Name).Filter("namespace", t.Namespace).Filter("hosting_cluster", t.HostCluster).One(t)
	return err
}

// UpdatePhase update phase field
func (t *App) UpdatePhase(phase uint8) (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable(t.TableName()).
		Filter("namespace", t.Namespace).
		Filter("hosting_cluster", t.HostCluster).
		Filter("name", t.Name).
		Update(orm.Params{"phase": phase})
}

func (t *App) UpdateDesc() (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable(t.TableName()).
		Filter("namespace", t.Namespace).
		Filter("hosting_cluster", t.HostCluster).
		Filter("name", t.Name).
		Update(orm.Params{"description": t.Description})
}

// GetTopology Get Topology
func (t *App) GetTopology(appName, namespace, clusterID string) (string, error) {
	o := orm.NewOrm()
	err := o.QueryTable(t.TableName()).Filter("name", appName).Filter("namespace", namespace).Filter("hosting_cluster", clusterID).One(t)
	return t.Relations, err
}

// SetTopology set topology
func (t *App) SetTopology(appName, namespace, clusterID, topology string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable(t.TableName()).Filter("name", appName).Filter("namespace", namespace).Filter("hosting_cluster", clusterID).Update(orm.Params{"relations": topology})
	return err
}

// SetEntryService set entry_service
func (t *App) SetEntryService(appName, namespace, clusterID, entryService string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable(t.TableName()).
		Filter("name", appName).
		Filter("namespace", namespace).
		Filter("hosting_cluster", clusterID).
		Update(orm.Params{"entry_service": entryService})
	return err
}

func (t *App) GetCount(namespaces, clusterIDs []string) (int64, error) {
	if len(namespaces) == 0 && len(clusterIDs) == 0 {
		return 0, errors.New("invalid parameters")
	}
	o := orm.NewOrm()
	qs := o.QueryTable(t.TableName())
	if len(namespaces) != 0 {
		qs = qs.Filter("namespace__in", namespaces)
	}
	if len(clusterIDs) != 0 {
		qs = qs.Filter("hosting_cluster__in", clusterIDs)
	}
	return qs.Count()
}

func (t *App) DeleteAllByNamespace(namespaceList []string, orms ...orm.Ormer) (int64, error) {
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
func (t *App) GetNamesByNamespaceAndCluster(namespace string, clusterID string) ([]string, error) {
	appNames := make([]string, 0)
	o := orm.NewOrm()
	sql := `SELECT name FROM %s WHERE namespace = ? AND hosting_cluster = ?`
	sql = fmt.Sprintf(sql, t.TableName())
	_, err := o.Raw(sql, namespace, clusterID).QueryRows(&appNames)
	return appNames, err
}
