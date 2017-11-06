/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-04-19  @author zhangyongkang
 */

package cluster

import (
	"api-server/modules/tenx/id"
	"api-server/util"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
)

const (
	// proxy group type
	ProxyPublic    = "public"
	ProxyPrivate   = "private"
	LBGroupDefault = "group-default"
)

var typeList = []string{ProxyPublic, ProxyPrivate}

type NodeProxy struct {
	Host     string `json:"host"`
	Address  string `json:"address"`
	IsMaster bool   `json:"-"`
	Group    string `json:"group,omitempty"`
}

type ProxyGroupLite struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Address   string `json:"address"`
	Domain    string `json:"domain"`
	IsDefault bool   `json:"is_default,omitempty"`
}

type ProxyGroup struct {
	Nodes []NodeProxy `json:"nodes,omitempty"`
	*ProxyGroupLite
}

type Extention struct {
	Nodes  []NodeProxy      `json:"nodes"`
	Groups []ProxyGroupLite `json:"groups"`
}

func GetProxyGroupList(clusterID string, lite bool) ([]ProxyGroup, error) {
	groupDic, err := getProxyGroupDic(clusterID)
	if err != nil {
		return nil, err
	}
	groupList := make([]ProxyGroup, 0, len(groupDic))
	for _, group := range groupDic {
		if lite {
			group.Nodes = nil
		}
		for i := range group.Nodes {
			group.Nodes[i].Group = ""
		}

		groupList = append(groupList, group)
	}
	return groupList, nil
}

func GetProxyGroupByID(clusterID, groupID string, lite bool) (*ProxyGroup, error) {
	groupDic, err := getProxyGroupDic(clusterID)
	if err != nil {
		return nil, err
	}
	for key, group := range groupDic {
		if key == groupID {
			if lite {
				group.Nodes = nil
			}
			return &group, nil
		}
	}
	return nil, nil
}

func GroupListToExt(groups []ProxyGroup) (*Extention, error) {
	var ext Extention
	if len(groups) < 1 {
		return &ext, nil
	}
	// convert to Extention
	ext.Groups = make([]ProxyGroupLite, 0, len(groups))
	ext.Nodes = make([]NodeProxy, 0, len(groups))
	nodeDic := make(map[string]bool, len(groups))
	addrDic := make(map[string]bool, len(groups))
	// should only have one default, and must have one default
	var hasDefault bool
	for _, pg := range groups {
		if hasDefault && pg.IsDefault {
			pg.IsDefault = false
		}
		if !hasDefault && pg.IsDefault {
			hasDefault = true
		}
		if pg.ID == "" {
			pg.ID = id.NewProxyGroup()
		}
		if !util.StringInArry(pg.Type, typeList) {
			return nil, fmt.Errorf("invalid group type: %s", pg.Type)
		}
		ext.Groups = append(ext.Groups, *pg.ProxyGroupLite)
		for _, n := range pg.Nodes {
			if addrDic[n.Address] {
				return nil, fmt.Errorf("address %s show in more than one time", n.Address)
			}
			if nodeDic[n.Host] {
				return nil, fmt.Errorf("node %s show in more than one time", n.Host)
			}
			n.Group = pg.ID
			ext.Nodes = append(ext.Nodes, n)
			nodeDic[n.Host] = true
			addrDic[n.Address] = true
		}
	}
	if !hasDefault {
		ext.Groups[0].IsDefault = true
	}
	return &ext, nil
}

func UpdateProxyGroup(clusterID string, ext *Extention) error {
	cl := &ClusterModel{
		ClusterID: clusterID,
	}
	if ext == nil {
		cl.Extention = "{}"
	} else {
		b, err := json.Marshal(ext)
		if err != nil {
			return err
		}
		cl.Extention = string(b)
	}
	// Allow bind domain if default group has domain info
	bindingDomains := make([]string, 0)
	for _, p := range ext.Groups {
		if p.Domain != "" && p.IsDefault {
			bindingDomains = append(bindingDomains, p.Domain)
			break
		}
	}
	// update bindingDomain in database to enable/disable http
	strDomain, _ := json.Marshal(bindingDomains)
	cl.BindingDomains = string(strDomain)
	return cl.Update("binding_domain", "extention")
}

func UpdateProxyGroupByID(clusterID string, group *ProxyGroup) error {
	groupDic, err := getProxyGroupDic(clusterID)
	if err != nil {
		return err
	}
	if _, found := groupDic[group.ID]; !found {
		return fmt.Errorf("group by id %s not found", group.ID)
	}
	groupDic[group.ID] = *group
	groups := make([]ProxyGroup, 0, len(groupDic))
	for _, p := range groupDic {
		groups = append(groups, p)
	}
	ext, err := GroupListToExt(groups)
	if err != nil {
		return err
	}
	return UpdateProxyGroup(clusterID, ext)
}

func SetDefaultGroup(clusterID, groupID string) error {
	groupDic, err := getProxyGroupDic(clusterID)
	if err != nil {
		return err
	}
	group, found := groupDic[groupID]
	if !found {
		return fmt.Errorf("group by id %s not found", groupID)
	}
	group.IsDefault = true
	groupDic[groupID] = group
	groups := make([]ProxyGroup, 0, len(groupDic))
	for id, group := range groupDic {
		if id != groupID && group.IsDefault {
			group.IsDefault = false
		}
		groups = append(groups, group)
	}
	ext, err := GroupListToExt(groups)
	if err != nil {
		return err
	}
	return UpdateProxyGroup(clusterID, ext)
}

func GetNodeProxies(clusterID string) ([]NodeProxy, error) {
	ext, err := getExt(clusterID)
	if err != nil {
		return nil, err
	}
	if ext == nil {
		return nil, nil
	}
	return ext.Nodes, nil
}

func groupLiteByID(id string, groups []ProxyGroupLite) *ProxyGroupLite {
	for _, g := range groups {
		if g.ID == id {
			return &g
		}
	}
	return nil
}

func getExt(clusterID string) (*Extention, error) {
	cl := &ClusterModel{}
	_, err := cl.Get(clusterID)
	if err != nil {
		return nil, err
	}
	var ext Extention
	if cl.Extention == "" {
		return nil, nil
	}
	if cl.Extention[0] != '{' {
		cl.Extention = "{ \"nodes\":" + cl.Extention + "}"
	}
	err = json.Unmarshal([]byte(cl.Extention), &ext)
	if err != nil {
		return nil, err
	}
	if len(ext.Groups) < 1 {
		defaultGroup := ProxyGroupLite{
			ID:        LBGroupDefault,
			Name:      "默认网络",
			Type:      ProxyPublic,
			IsDefault: true,
		}
		defaultGroup.Domain, err = jsonArrayToString(cl.BindingDomains)
		if err != nil {
			glog.Errorf("jsonArrayToString(%s) failed: %v", cl.BindingDomains, err)
		}
		defaultGroup.Address, err = jsonArrayToString(cl.BindingIPs)
		if err != nil {
			glog.Errorf("jsonArrayToString(%s) failed: %v", cl.BindingIPs, err)
		}
		ext.Groups = []ProxyGroupLite{defaultGroup}
	}
	return &ext, nil
}

func jsonArrayToString(str string) (string, error) {
	if str == "" || str == "[]" || str == `[""]` {
		return "", nil
	}
	if !strings.Contains(str, "[") && !strings.Contains(str, "]") {
		return str, nil
	}
	array := make([]string, 0, 1)
	if err := json.Unmarshal([]byte(str), &array); err != nil {
		return "", err
	}
	return strings.Join(array, ","), nil
}

func getProxyGroupDic(clusterID string) (map[string]ProxyGroup, error) {
	ext, err := getExt(clusterID)
	if err != nil {
		return nil, err
	}
	if ext == nil {
		return nil, nil
	}
	var defaultGroupID = ext.Groups[0].ID
	var proxyGroups = make(map[string]ProxyGroup, len(ext.Groups))
	groupDic := make(map[string][]NodeProxy, len(ext.Groups))
	for _, node := range ext.Nodes {
		if node.Group == "" {
			node.Group = defaultGroupID
		}
		addrs := groupDic[node.Group]
		if len(addrs) < 1 {
			addrs = make([]NodeProxy, 0, 1)
		}
		addrs = append(addrs, node)
		groupDic[node.Group] = addrs
	}
	for k, v := range groupDic {
		pg := ProxyGroup{v, groupLiteByID(k, ext.Groups)}
		proxyGroups[k] = pg
	}
	return proxyGroups, nil
}
