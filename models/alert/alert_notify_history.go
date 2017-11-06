/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-13  @author mengyuan
 */
package alert

import (
	"time"

	"bytes"

	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

const (
	AlertNotNotified = 0
	AlertNotified    = 1
)
const (
	alertNotDeleted = 0
	alertDeleted    = 1
)

type NotifyHistory struct {
	ID            int64     `orm:"pk;column(id)" json:"-"`
	StrategyID    string    `orm:"column(strategy_id)" json:"strategyID"`
	ClusterID     string    `orm:"column(clusterid)" json:"clusterID"`
	Namespace     string    `orm:"column(namespace)" json:"namespace"`
	NamespaceType int       `orm:"column(namespace_type)" json:"namespaceType"`
	TargetName    string    `orm:"column(target_name)" json:"targetName"`
	TargetType    int       `orm:"column(target_type)" json:"targetType"`
	StrategyName  string    `orm:"size(45);column(strategy_name)" json:"strategyName"`
	AppName       string    `orm:"column(app_name)" json:"-"`
	TriggerRule   string    `orm:"column(trigger_rule)" json:"triggerRule"`
	TriggerValue  string    `orm:"column(trigger_value)" json:"triggerValue"`
	Status        int       `orm:"column(status)" json:"status"`
	CreateTime    time.Time `orm:"column(create_time)" json:"createTime"`
	Deleted       int       `orm:"column(deleted)" json:"-"`
}

type RecordListQuery struct {
	Cluster      *string
	Namespace    string
	StrategyName *string
	TargetName   *string
	TargetType   *int
	Status       *int
	BeginTime    time.Time
	EndTime      time.Time
	From         int
	Size         int
}
type CountUnit struct {
	TriggerRule string `json:"triggerRule"`
	RuleName    string `json:"ruleName"`
}

func (q RecordListQuery) String() string {
	buf := bytes.Buffer{}
	if q.Cluster == nil {
		buf.WriteString("Cluster[<nil>] ")
	} else {
		buf.WriteString(fmt.Sprintf("Cluster[%s] ", *q.Cluster))
	}
	buf.WriteString(fmt.Sprintf("Namespace[%s] ", q.Namespace))
	if q.StrategyName == nil {
		buf.WriteString("StrategyName[<nil>] ")
	} else {
		buf.WriteString(fmt.Sprintf("StrategyName[%s] ", *q.StrategyName))
	}
	if q.TargetName == nil {
		buf.WriteString("TargetName[<nil>] ")
	} else {
		buf.WriteString(fmt.Sprintf("TargetName[%s] ", *q.TargetName))
	}
	if q.TargetType == nil {
		buf.WriteString("TargetType[<nil>] ")
	} else {
		buf.WriteString(fmt.Sprintf("TargetType[%d] ", *q.TargetType))
	}
	if q.Status == nil {
		buf.WriteString("Status[<nil>] ")
	} else {
		buf.WriteString(fmt.Sprintf("Status[%d] ", *q.Status))
	}
	buf.WriteString(fmt.Sprintf("BeginTime[%s] ", q.BeginTime))
	buf.WriteString(fmt.Sprintf("EndTime[%s] ", q.EndTime))
	buf.WriteString(fmt.Sprintf("From[%d] ", q.From))
	buf.WriteString(fmt.Sprintf("Size[%d] ", q.Size))

	return buf.String()
}

func NewNotifyHistory() *NotifyHistory {
	return new(NotifyHistory)
}
func (t *NotifyHistory) TableName() string {
	return "tenx_alert_notify_history"
}

// CountByTriggerRules count by trigger_rule
// return ruleName-count(triggerRule) key-value pair
func (t *NotifyHistory) CountByTriggerRules(strategyID string, ruleAndName map[string]string) (map[string]int, error) {
	cap := len(ruleAndName)
	result := make(map[string]int, cap)
	var buffer bytes.Buffer
	for r, _ := range ruleAndName {
		buffer.WriteString(`,"` + r + `"`)
	}
	rules := buffer.String()
	rules = rules[1:]
	type counter struct {
		TriggerRule string `orm:"trigger_rule"`
		Count       int    `orm:"count"`
	}
	var counters []counter
	sql := fmt.Sprintf(`select trigger_rule, count(trigger_rule) as count from %s where strategy_id='%s' and trigger_rule in (%s) group by trigger_rule;`, t.TableName(), strategyID, rules)
	glog.V(4).Infof("sql:\t%s", sql)
	_, err := orm.NewOrm().Raw(sql).QueryRows(&counters)
	if err != nil {
		return nil, err
	}
	for _, c := range counters {
		result[ruleAndName[c.TriggerRule]] = c.Count
	}
	return result, nil
}

func (t *NotifyHistory) InsertMulti(records []NotifyHistory) error {
	if len(records) == 0 {
		return nil
	}
	o := orm.NewOrm()
	_, err := o.InsertMulti(len(records), records)
	return err
}
func (t *NotifyHistory) List(query RecordListQuery) ([]NotifyHistory, int64, error) {
	glog.Infoln(query)
	o := orm.NewOrm()
	qs := o.QueryTable(t.TableName()).
		Filter("create_time__gte", query.BeginTime).
		Filter("create_time__lte", query.EndTime).
		Filter("namespace", query.Namespace).
		Filter("deleted", alertNotDeleted).
		OrderBy("-create_time")
	if query.Cluster != nil {
		qs = qs.Filter("clusterid", *query.Cluster)
	}
	if query.StrategyName != nil {
		qs = qs.Filter("strategy_name", *query.StrategyName)
	}
	if query.TargetType != nil {
		qs = qs.Filter("target_type", *query.TargetType)
	}
	if query.TargetName != nil {
		qs = qs.Filter("target_name", *query.TargetName)
	}
	if query.Status != nil {
		qs = qs.Filter("status", *query.Status)
	}
	var items []NotifyHistory
	// get total count
	count, err := qs.Count()
	if err != nil {
		return items, 0, err
	}

	_, err = qs.Offset(query.From).Limit(query.Size).All(&items)

	return items, count, err
}

func (t *NotifyHistory) MarkDeleted(namespace string, strategyID string) (int64, error) {
	qs := orm.NewOrm().QueryTable(t.TableName()).
		Filter("namespace", namespace)
	if strategyID != "" {
		qs = qs.Filter("strategy_id", strategyID)
	}
	return qs.Update(orm.Params{"deleted": alertDeleted})
}

func (t *NotifyHistory) GetFilter(cluster string, namespace string) ([]NotifyHistory, error) {
	qs := orm.NewOrm().QueryTable(t.TableName()).
		Filter("deleted", alertNotDeleted).
		Filter("namespace", namespace).
		Filter("clusterid", cluster).
		GroupBy("strategy_id", "target_name", "target_type")
	var items []NotifyHistory
	_, err := qs.All(&items, "strategy_id", "target_name", "target_type", "strategy_name")
	return items, err
}
