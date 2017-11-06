/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-07  @author mengyuan
 */
package alert

import (
	"errors"
	"time"

	"fmt"

	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

const (
	StrategyDisabled = iota
	StrategyEnabled
	StrategyAlerting
	StrategyIgnored

	StrategyNotSendEmail = 0
)

const (
	StrategyTypeService = 0
	StrategyTypeNode    = 1
)
const notifyStrategyTable = "tenx_alert_notify_strategy"

var (
	columMapping = map[string]string{
		"clusterID":    "clusterid",
		"namespace":    "namespace",
		"targetName":   "target_name",
		"targetType":   "target_type",
		"appName":      "app_name",
		"strategyName": "strategy_name",
		"strategyID":   "strategy_id",
	}
)

type NotifyStrategyUnion struct {
	NotifyStrategy
	Receivers  string `orm:"column(name)" json:"receivers"`
	StatusCode int    `json:"statusCode"`
}

type NotifyStrategy struct {
	ID                   int64     `orm:"pk;column(id)" json:"-"`
	StrategyID           string    `orm:"column(strategy_id)" json:"strategyID"`
	Enable               int       `orm:"column(enable)" json:"enable"`        // 0: disable, 1: enable
	SendEmail            int       `orm:"column(send_email)" json:"sendEmail"` // 0: not send, 1: send
	ClusterID            string    `orm:"column(clusterid)" json:"clusterID"`
	Namespace            string    `orm:"column(namespace)" json:"namespace"`
	NamespaceType        int       `orm:"column(namespace_type)" json:"namespaceType"`
	TargetName           string    `orm:"column(target_name)" json:"targetName"`
	TargetType           int       `orm:"column(target_type)" json:"targetType"`
	StrategyName         string    `orm:"size(45);column(strategy_name)" json:"strategyName"`
	AppName              string    `orm:"column(app_name)" json:"appName"`
	ReceiversGroup       string    `orm:"column(receivers_group)" json:"receiversGroup"`
	RepeatInterval       int       `orm:"column(repeat_interval)" json:"repeatInterval"` // seconds
	Description          string    `orm:"column(description)" json:"description"`
	CreateTime           time.Time `orm:"column(create_time)" json:"createTime"`
	ModifyTime           time.Time `orm:"column(modify_time)" josn:"modityTime"`
	Creator              string    `orm:"column(creator)" json:"creator"`
	Updater              string    `orm:"column(updater)" json:"updater"`
	DisableNotifyEndTime time.Time `orm:"column(disable_notify_end_time)" json:"disableNotifyEndTime"`
}

func NewNotifyStrategy() *NotifyStrategy {
	return new(NotifyStrategy)
}
func (t *NotifyStrategy) TableName() string {
	return notifyStrategyTable
}
func (t *NotifyStrategy) GetStrategy(namespace, targetName, targetType, strategyName, clusterID string) (NotifyStrategy, error) {
	typ := 0
	rule := NotifyStrategy{}
	switch targetType {
	case "service":
		typ = 0
	case "node":
		typ = 1
	default:
		return rule, fmt.Errorf("invalide target type[%s]", targetType)
	}
	o := orm.NewOrm()
	err := o.QueryTable(t.TableName()).
		Filter("namespace", namespace).
		Filter("target_name", targetName).
		Filter("target_type", typ).
		Filter("strategy_name", strategyName).
		Filter("clusterid", clusterID).
		One(&rule)
	return rule, err
}
func (t *NotifyStrategy) GetByID(ids []string) (items []NotifyStrategy, err error) {
	o := orm.NewOrm()
	_, err = o.QueryTable(notifyStrategyTable).Filter("id__in", ids).All(&items)
	return
}
func (t *NotifyStrategy) GetStrategiesByGroup(creator string, groupIDs []string) ([]NotifyStrategy, error) {
	strategies := []NotifyStrategy{}
	qs := orm.NewOrm().QueryTable(t.TableName()).
		Filter("receivers_group__in", groupIDs)
	if creator != "" {
		qs = qs.Filter("creator", creator)
	}
	_, err := qs.All(&strategies)
	return strategies, err
}

func (t *NotifyStrategy) GetByStrategyIDs(ids []string) ([]NotifyStrategy, error) {
	items := []NotifyStrategy{}
	_, err := orm.NewOrm().QueryTable(t.TableName()).
		Filter("strategy_id__in", ids).
		All(&items)
	return items, err
}

// DeleteByStrategyIDs delete record by strategy_ids
func (t *NotifyStrategy) DeleteByStrategyIDs(ids []string) (int64, error) {
	return orm.NewOrm().QueryTable(t.TableName()).
		Filter("strategy_id__in", ids).
		Delete()
}

func (t *NotifyStrategy) Insert() (int64, error) {
	o := orm.NewOrm()
	return o.Insert(t)
}

func (t *NotifyStrategy) Update(cls ...string) error {
	o := orm.NewOrm()

	_, err := o.Update(t, cls...)
	return err
}

func (t *NotifyStrategy) UpdateByStrategyID(params map[string]interface{}) error {
	if t.StrategyID == "" {
		return errors.New("strategyID is empty !")
	}
	_, err := orm.NewOrm().QueryTable(t.TableName()).
		Filter("strategy_id", t.StrategyID).Update(params)
	return err
}

func (t *NotifyStrategy) GetByCondition(conditions map[string]interface{}, from, size int) ([]*NotifyStrategyUnion, int64, error) {
	items := []NotifyStrategy{}
	qs := orm.NewOrm().QueryTable(t.TableName())
	for key, val := range conditions {
		col, ok := columMapping[key]
		if !ok {
			return nil, 0, fmt.Errorf("column %s not found", key)
		}
		qs = qs.Filter(col, val)
	}
	total, err := qs.Count()
	if err != nil {
		return nil, 0, err
	}
	_, err = qs.OrderBy("-modify_time").Limit(size, from).All(&items)
	if err != nil {
		return nil, 0, err
	}
	result, err := populateUnion(items)
	return result, total, err
}

// SearchByStrategyName fuzzy search by strategyName
func (t *NotifyStrategy) SearchByStrategyName(strategy string, from, size int) ([]*NotifyStrategyUnion, int64, error) {
	items := []NotifyStrategy{}
	qs := orm.NewOrm().QueryTable(t.TableName()).
		Filter("clusterid", t.ClusterID).
		Filter("namespace", t.Namespace).
		Filter("strategy_name__icontains", strategy)
	total, err := qs.Count()
	if err != nil {
		return nil, 0, err
	}
	_, err = qs.OrderBy("-modify_time").Limit(size, from).All(&items)
	if err != nil {
		return nil, 0, err
	}
	result, err := populateUnion(items)
	return result, total, err
}

// GroupStrategties exactly search by targetNames or appNames
func (t *NotifyStrategy) GroupStrategties(clusterID, namespace string, targetNames, appNames []string) ([]NotifyStrategy, error) {
	items := []NotifyStrategy{}
	qs := orm.NewOrm().QueryTable(t.TableName()).
		Filter("clusterid", clusterID).
		Filter("namespace", namespace)
	if len(targetNames) > 0 {
		qs = qs.Filter("target_name__in", targetNames)
	}
	if len(appNames) > 0 {
		qs = qs.Filter("app_name__in", appNames)
	}
	_, err := qs.OrderBy("-modify_time").All(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func populateUnion(items []NotifyStrategy) ([]*NotifyStrategyUnion, error) {
	method := "populateUnion"
	result := []*NotifyStrategyUnion{}
	if len(items) < 1 {
		glog.Warningf("%s no record found", method)
		return nil, nil
	}
	groubIDs := make([]string, 0, len(items))
	var (
		wg  sync.WaitGroup
		err error
	)
	for _, item := range items {
		groubIDs = append(groubIDs, item.ReceiversGroup)

		union := &NotifyStrategyUnion{}
		union.NotifyStrategy = item
		wg.Add(1)
		go func(*NotifyStrategyUnion) {
			defer func() {
				if r := recover(); r != nil {
					glog.Errorf("%s recovered from panic: %v", method, r)
				}
				wg.Done()
			}()
			union.StatusCode = union.NotifyStrategy.alertStatus()
		}(union)
		result = append(result, union)
	}
	var receivers []NotifyReceiverGroup
	group := NewNotifyReceiverGroup()
	_, err = orm.NewOrm().QueryTable(group.TableName()).
		Filter("group_id__in", groubIDs).
		All(&receivers)
	if err != nil {
		return nil, err
	}
	if len(receivers) > 0 {
		receiverMapping := make(map[string]string, 1)
		for _, r := range receivers {
			receiverMapping[r.GroupID] = r.Name
		}
		for _, item := range result {
			item.Receivers = receiverMapping[item.ReceiversGroup]
		}
	}
	wg.Wait()

	return result, err
}

// alertStatus 检查策略状态
func (t *NotifyStrategy) alertStatus() int {
	if t.Enable == StrategyDisabled {
		return StrategyDisabled
	}
	now := time.Now()
	if now.Before(t.DisableNotifyEndTime) {
		return StrategyIgnored
	}
	statusCode := StrategyEnabled
	history := NewNotifyHistory()
	lastCheckTime := now.Add(time.Duration(t.RepeatInterval * -1 * int(time.Second)))
	count, err := orm.NewOrm().QueryTable(history.TableName()).
		Filter("strategy_id", t.StrategyID).
		Filter("create_time__gt", lastCheckTime).Count()
	if err != nil {
		glog.Errorf("query strategy histoy to check strategy status failed: %v", err)
		return statusCode
	}
	if count > 0 {
		statusCode = StrategyAlerting
	}
	return statusCode
}

func (t *NotifyStrategy) IsGroupsUsing(groupList []string) bool {
	return orm.NewOrm().QueryTable(t.TableName()).
		Filter("receivers_group__in", groupList).
		Exist()
}

func (t *NotifyStrategy) IsAlreadyExist() (bool, error) {
	var count int64
	sql := fmt.Sprintf("select count(id) from %s where clusterid=%q and namespace=%q and strategy_name=%q", t.TableName(), t.ClusterID, t.Namespace, t.StrategyName)
	err := orm.NewOrm().Raw(sql).QueryRow(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
