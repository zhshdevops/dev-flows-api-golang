/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package cluster

import (
	"encoding/json"
	"fmt"
	"time"

	sqlstatus "api-server/models/sql/status"
	"api-server/util/misc"

	"github.com/golang/glog"

	"api-server/models/common"
	"crypto/tls"

	"github.com/astaxie/beego/httplib"
)

var (
	// SortByMapping is the mapping from display name to fields in tenx_teams table
	SortByMapping = map[string]string{
		"clusterName":  "name",
		"creationTime": "creation_time",
	}
	// FilterByMapping is the mapping from display name to fields in tenx_teams table
	FilterByMapping = map[string]string{
		"clusterName": "name",
	}
)

// Cluster contains information excluding the sensitive info
type Cluster struct {
	ClusterID              string             `json:"clusterID"`
	ClusterName            string             `json:"clusterName"`
	APIProtocol            string             `json:"apiProtocol"`
	APIHost                string             `json:"apiHost"`
	APIToken               string             `json:"apiToken,omitempty"`
	APIVersion             string             `json:"apiVersion"`
	Description            string             `json:"description"`
	PublicIPs              string             `json:"publicIPs"`
	ListNodes              uint8              `json:"listNodes"`
	IsBuilder              bool               `json:"isBuilder"`
	BindingIPs             string             `json:"bindingIPs"`
	BindingDomains         string             `json:"bindingDomains"`
	StorageID              string             `json:"storage_id"`
	StorageTypes           []string           `json:"storage_types"`
	CreationTime           time.Time          `json:"creationTime"`
	ResourcePrice          map[string]float64 `json:"resource_price"`
	DisabledPlugins        map[string]bool    `json:"disabledPlugins,omitempty"`
	IsOk                   bool               `json:"isOk"`
	IsDefault              int8               `json:"isDefault"`
	NetworkPolicySupported *bool              `json:"networkPolicySupported,omitempty"`
}

const (
	StorageRBD      = "rbd"
	StorageHostpath = "hostPath"
)

type ClusterList struct {
	ListMeta common.ListMeta `json:"listMeta"`
	Clusters []Cluster       `json:"clusters"`
}

// 1 Yuan = 10000
var defaultPriceConfig = map[string]float64{
	"storage":  14,
	"1x":       250,
	"2x":       400,
	"4x":       800,
	"8x":       1500,
	"16x":      2880,
	"32x":      5160,
	"64x":      10320,
	"db_ratio": 1.2,
}

func ToCluster(t *ClusterModel, isAdmin bool) *Cluster {
	method := "ToCluster"
	if t == nil {
		return nil
	}
	// Use default strategy
	var priceConfig = defaultPriceConfig
	if t.ResourcePrice != "" {
		// Get the config from database value
		var customizedPrice map[string]float64
		err := json.Unmarshal([]byte(t.ResourcePrice), &customizedPrice)
		if err != nil {
			// If invalid price config, use the default one
			glog.Errorf("Failed to parse price config: %s", err)
			priceConfig = defaultPriceConfig
		} else {
			priceConfig = customizedPrice
			glog.Infof("Use customized price: %v", priceConfig)
		}
	}
	config := Config{}
	if err := json.Unmarshal([]byte(t.ConfigDetail), &config); err != nil {
		glog.Errorln(method, "unmarshal config detail failed. raw string:", t.ConfigDetail, ". err:", err)
	}
	// Standard edition -> Hide BindingIPs to prevent attacker
	if misc.IsStandardMode() {
		t.BindingIPs = "[]"
	}
	isOk := checkClusterHealthz(t)
	result := &Cluster{
		ClusterID:              t.ClusterID,
		ClusterName:            t.ClusterName,
		Description:            t.Description,
		APIProtocol:            t.APIProtocol,
		APIHost:                t.APIHost,
		APIVersion:             t.APIVersion,
		PublicIPs:              t.PublicIPs,
		BindingIPs:             t.BindingIPs,
		BindingDomains:         t.BindingDomains,
		StorageID:              t.StorageID,
		CreationTime:           t.CreationTime,
		ResourcePrice:          priceConfig,
		ListNodes:              config.ListNodes,
		IsBuilder:              config.IsBuilder == 1,
		IsOk:                   isOk,
		IsDefault:              t.IsDefault,
		NetworkPolicySupported: config.NetworkPolicySupported,
	}
	if isAdmin {
		result.APIToken = t.APIToken
	}
	return result
}

func ToClusters(clusters []ClusterModel, isAdmin bool) []Cluster {
	var res []Cluster
	for _, t := range clusters {
		res = append(res, *ToCluster(&t, isAdmin))
	}
	return res
}

// validateClusterDataSelect filter out invalid sort, filter options, and return valid ones
func validateClusterDataSelect(old *common.DataSelectQuery) (*common.DataSelectQuery, error) {
	if old == nil {
		return common.NewDataSelectQuery(common.DefaultPagination, common.NewSortQuery([]string{"a", SortByMapping["clusterName"]}), common.NoFilter), nil
	}

	dataselect := common.NewDataSelectQuery(old.PaginationQuery, common.NoSort, common.NoFilter)
	if old.FilterQuery != nil {
		for _, f := range old.FilterQuery.FilterByList {
			prop, ok := FilterByMapping[f.Property]
			if !ok {
				glog.Errorf("cluster list, invalid filter by options: %s\n", f.Property)
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
				glog.Errorf("cluster list, invalid sort options: %s\n", sq.Property)
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

// GetClusterList fetches clusters according to IDs
func GetClusterList(dataselect *common.DataSelectQuery) (*ClusterList, uint32, error) {
	// get cluster list, currently dataselect is not used in List method
	clusterCount, clusters, errCode, err := GetClusterModelList(dataselect)
	if err != nil {
		glog.Errorf("list clusters fails, error code:%d, error:%v\n", errCode, err)
		return nil, errCode, err
	}

	tl := &ClusterList{
		ListMeta: common.ListMeta{
			Total: clusterCount,
		},
		Clusters: ToClusters(clusters, false),
	}

	return tl, sqlstatus.SQLSuccess, nil
}

func GetClusterModelList(dataselect *common.DataSelectQuery) (int, []ClusterModel, uint32, error) {
	clusterModel := &ClusterModel{}

	dataselect, err := validateClusterDataSelect(dataselect)
	if err != nil {
		glog.Errorf("get cluster list fails, error:%v\n", err)
		return 0, nil, sqlstatus.SQLErrSyntax, err
	}

	clusterCount, errcode, err := clusterModel.Count(dataselect)
	if err != nil {
		glog.Errorf("get cluster number fails, error code:%d, error:%v\n", errcode, err)
		return 0, nil, errcode, err
	}

	models, errcode, err := clusterModel.List(dataselect)
	// get cluster list, currently dataselect is not used in List method
	return clusterCount, models, errcode, err
}

// GetClusterListByIDs fetches clusters according to IDs
func GetClusterListByIDs(clusterIDs []string, dataselect *common.DataSelectQuery) (*ClusterList, uint32, error) {
	if len(clusterIDs) == 0 {
		return &ClusterList{ListMeta: common.ListMeta{Total: 0}, Clusters: make([]Cluster, 0)}, sqlstatus.SQLSuccess, nil
	}

	if dataselect == nil {
		dataselect = common.DefaultDataSelect
	}

	clusterModel := &ClusterModel{}

	// get cluster list
	clusters, errCode, err := clusterModel.ListByIDs(clusterIDs, dataselect)
	if err != nil {
		glog.Errorf("list clusters of ids %v fails, error code:%d, error:%v\n", clusterIDs, errCode, err)
		return nil, errCode, err
	}

	tl := &ClusterList{
		ListMeta: common.ListMeta{
			Total: len(clusterIDs),
		},
		Clusters: ToClusters(clusters, false),
	}

	return tl, sqlstatus.SQLSuccess, nil
}

func checkClusterHealthz(cluster *ClusterModel) bool {
	url := fmt.Sprintf("%s://%s/healthz", cluster.APIProtocol, cluster.APIHost)
	req := httplib.Get(url).SetTimeout(10*time.Second, 10*time.Second)
	if cluster.APIProtocol == "https" {
		req.Header("Authorization", fmt.Sprintf("bearer %s", cluster.APIToken)).
			SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	resp, err := req.String()
	if err != nil {
		glog.Infof("cluster [%s] get healthz info by url[%s] failed, error: %s", cluster.ClusterID, url, err)
		return false
	}
	if resp != "ok" {
		glog.Infof("cluster [%s] healthz info by url[%s] is not ok [%s]", cluster.ClusterID, url, resp)
		return false
	}
	return true
}
