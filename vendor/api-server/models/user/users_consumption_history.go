/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-29  @author mengyuan
 */

package user

import (
	"api-server/models/common"
	"time"

	"fmt"
	"strings"

	"api-server/util/misc"
	"errors"

	"api-server/models/sql/util"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

type UsersConsumption struct {
	ConsumptionID  string    `orm:"pk;column(consumption_id)"`
	Namespace      string    `orm:"column(namespace)"`
	Type           string    `orm:"column(consumption_type)"`
	Amount         int       `orm:"column(consumption_amount)"` // unit fen
	StartTime      time.Time `orm:"column(consumption_start_time)"`
	TotalTime      int       `orm:"column(consumption_total_time)"` // unit minute
	Config         string    `orm:"column(consumption_config)"`
	Name           string    `orm:"column(consumption_name)"`
	HostingCluster string    `orm:"column(hosting_cluster)"`
	Price          int       `orm:"column(consumption_price)"` // unit fen
	Label          string    `orm:"column(consumption_label)"`
	Way            string    `orm:"column(Consumption_way)"`
	CreateTime     time.Time `orm:"column(create_time)" json:"createTime"`
}
type ConsumptionDetail struct {
	ConsumptionID  string    `orm:"pk;column(consumption_id)" json:"id"`
	Type           string    `orm:"column(consumption_type)" json:"type"`
	Amount         int       `orm:"column(consumption_amount)" json:"amount"` // unit fen
	StartTime      time.Time `orm:"column(consumption_start_time)" json:"startTime"`
	TotalTime      int       `orm:"column(consumption_total_time)" json:"continueTime"` // unit minute
	CreateTime     time.Time `orm:"column(create_time)" json:"createTime"`
	Name           string    `orm:"column(consumption_name)" json:"consumptionName"`
	HostingCluster string    `orm:"column(hosting_cluster)" json:"cluserID"`
	Price          int       `orm:"column(consumption_price)"json:"unitPrice"` // unit fen
	ClusterName    string    `orm:"column(clusterName)" json:"clusterName"`
}

type ConsumptionItem struct {
	Sum      int64
	BackName string // used as name if Name field is empty
	Name     string
}

func (t *UsersConsumption) TableName() string {
	return "tenx_users_consumption_history"
}

func NewUsersConsumption() *UsersConsumption {
	return &UsersConsumption{}
}

// GetSummaryByCluster get consumption, select all clusters data if cluster parameter is empty
func (t *UsersConsumption) GetSummaryGroupByCluster(spaceList []string, cluster string, beginTime, endTime string) ([]ConsumptionItem, error) {
	method := "GetSummaryGroupByCluster"
	sql := `SELECT SUM(t1.consumption_amount) AS sum, t1.hosting_cluster as backname , t2.name
	FROM tenx_users_consumption_history AS t1 LEFT JOIN tenx_clusters AS t2
	ON t1.hosting_cluster  = t2.id
	WHERE t1.create_time >= ? AND t1.create_time < ?
	AND t1.namespace IN (%s) AND t1.consumption_type != 6 `
	if cluster != "" {
		sql += `AND hosting_cluster = ? `
	}
	sql += `GROUP BY t1.hosting_cluster
	ORDER BY sum DESC`
	safeSpaceList := util.EscapeSliceForInjection(spaceList)
	spaceListWithQuote := misc.SliceToString(safeSpaceList, "'", ",")
	sql = fmt.Sprintf(sql, spaceListWithQuote)
	o := orm.NewOrm()
	result := make([]ConsumptionItem, 0, 1)
	var rs orm.RawSeter
	if cluster != "" {
		rs = o.Raw(sql, beginTime, endTime, cluster)
	} else {
		rs = o.Raw(sql, beginTime, endTime)
	}
	if _, err := rs.QueryRows(&result); err != nil {
		glog.Errorln(method, "failed.", err)
		return nil, err
	}
	if cluster == "" && misc.IsStandardMode() {
		sql := `SELECT SUM(consumption_amount) AS sum
			FROM tenx_users_consumption_history
			WHERE create_time >= ? AND create_time < ?
			AND namespace IN (%s) AND consumption_type = 6`
		sql = fmt.Sprintf(sql, spaceListWithQuote)
		ret := ConsumptionItem{
			Name: "__PRO_EDITION",
		}
		if err := o.Raw(sql, beginTime, endTime).QueryRow(&ret); err != nil {
			glog.Errorln(method, "failed.", err)
			return nil, err
		}
		if ret.Sum != 0 {
			result = append(result, ret)
		}
	}

	// handle the situation of consumption_type = 6 (pro version consumption, not belong to any cluster)
	return result, nil
}

// GetSummaryByNamespace get consumption, select all clusters data if cluster parameter is empty
func (t *UsersConsumption) GetSummaryGroupByNamespace(spaceList []string, cluster string, beginTime, endTime string) ([]ConsumptionItem, error) {
	method := "GetSummaryByNamespace"
	var spaceListCopy []string
	for _, space := range spaceList {
		spaceListCopy = append(spaceListCopy, "'"+space+"'")
	}
	sql := `SELECT SUM(t1.consumption_amount) AS sum, t1.namespace as backname, t2.name
	FROM tenx_users_consumption_history AS t1 LEFT JOIN tenx_team_space AS t2
	ON t1.namespace = t2.namespace
	WHERE t1.create_time >= ? AND t1.create_time < ?
	AND t1.namespace IN (%s) `
	if cluster != "" {
		sql += `AND hosting_cluster = ? `
	}
	sql += `GROUP BY t1.namespace
	ORDER BY sum DESC`
	sql = fmt.Sprintf(sql, strings.Join(spaceListCopy, ","))
	o := orm.NewOrm()
	result := make([]ConsumptionItem, 0, 1)
	var rs orm.RawSeter
	if cluster != "" {
		rs = o.Raw(sql, beginTime, endTime, cluster)
	} else {
		rs = o.Raw(sql, beginTime, endTime)
	}
	if _, err := rs.QueryRows(&result); err != nil {
		glog.Errorln(method, "failed.", err)
		return nil, err
	}
	return result, nil
}

func (t *UsersConsumption) GetAllItemDetail(offset, size int, beginTime, endTime string, space string, typeList []string) ([]ConsumptionDetail, int64, error) {
	method := "GetDetail"
	qb, err := orm.NewQueryBuilder(common.DType.String())
	if err != nil {
		glog.Errorln(method, "NewQueryBuilder failed.", err)
		return nil, 0, err
	}
	safeTypeList := util.EscapeSliceForInjection(typeList)
	// get items
	qb = qb.Select("t1.*,t2.name as clusterName").
		From(t.TableName() + " AS t1").
		LeftJoin("tenx_clusters AS t2").
		On("t1.hosting_cluster = t2.id").
		Where("t1.create_time >= ?").
		And("t1.create_time < ?").
		And("t1.namespace = ?")
	if len(safeTypeList) > 0 {
		qb.And("t1.consumption_type").In(misc.AddSurroundToSlice(safeTypeList, "'")...)
	}
	sql := qb.OrderBy("t1.create_time").Desc().
		Limit(size).Offset(offset).String()
	o := orm.NewOrm()
	result := make([]ConsumptionDetail, 0, 1)
	_, err = o.Raw(sql, beginTime, endTime, space).QueryRows(&result)
	if err != nil {
		glog.Errorln(method, "failed.", err)
		return nil, 0, err
	}

	location := beego.AppConfig.String("location")
	var begin, end interface{}
	if location == "" {
		begin = beginTime
		end = endTime
	} else {
		loc, err := time.LoadLocation(location)
		if err != nil {
			glog.Errorln(method, "load location failed.", err)
			return nil, 0, err
		}
		begin, err = time.ParseInLocation("2006-01-02 15:04:05", beginTime, loc)
		if err != nil {
			glog.Errorln(method, "parse beginTime failed.", err)
			return nil, 0, err
		}
		end, err = time.ParseInLocation("2006-01-02 15:04:05", endTime, loc)
		if err != nil {
			glog.Errorln(method, "parse endTime failed.", err)
			return nil, 0, err
		}
	}
	// get total count
	qs := o.QueryTable(t.TableName()).
		Filter("create_time__gte", begin).
		Filter("create_time__lt", end).
		Filter("namespace", space)
	if len(safeTypeList) > 0 {
		qs = qs.Filter("consumption_type__in", safeTypeList)
	}
	total, err := qs.Count()
	if err != nil {
		glog.Errorln(method, "failed.", err)
		return nil, 0, err
	}
	return result, total, nil
}

type summary struct {
	Time string `json:"time"`
	Cost int64  `json:"cost"`
}

func (t *UsersConsumption) GetDetailInDuration(spaceList []string, beginTime, endTime string, interval string) ([]summary, error) {
	var timePrefixLen int
	if interval == "month" {
		timePrefixLen = 7
	} else if interval == "day" {
		timePrefixLen = 10
	} else {
		return nil, errors.New("invalid interval parameter")
	}

	sql := `SELECT LEFT(create_time,%d) as time,
	SUM(consumption_amount) AS cost
	FROM %s
	WHERE create_time >= ? AND
	create_time < ? AND
	namespace IN (%s)
	GROUP BY LEFT(create_time,%d)
	ORDER BY time`

	safeSpaceList := util.EscapeSliceForInjection(spaceList)
	spaceListWithQuote := misc.SliceToString(safeSpaceList, "'", ",")
	sql = fmt.Sprintf(sql, timePrefixLen, t.TableName(), spaceListWithQuote, timePrefixLen)
	o := orm.NewOrm()
	result := make([]summary, 0, 1)
	_, err := o.Raw(sql, beginTime, endTime).QueryRows(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
