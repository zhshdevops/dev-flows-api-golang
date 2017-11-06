/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-22  @author shouhong.zhang
 */

package cluster

import (
	"api-server/models/common"
	"api-server/models/team2resource"
	"time"
)

//ClusterRequest cluster request
type ClusterRequest struct {
	TeamID        string    `json:"TeamID"`
	ResourceID    string    `json:"resourceID"`
	ResourceType  int       `json:"resourceType"`
	RequestUserID int32     `json:"requestUserID"`
	RequestTime   time.Time `json:"requestTime"`
	ApproveUserID int32     `json:"approveUserID"`
	ApproveTime   time.Time `json:"approveTime"`
	Status        int       `json:"status"`
}

//ClusterRequestList cluster request list
type ClusterRequestList struct {
	ListMeta        common.ListMeta  `json:"listMeta"`
	ClusterRequests []ClusterRequest `json:"requests"`
}

//ToClusterRequest convert TeamResourceRequestModel to ToClusterRequest
func ToClusterRequest(t *team2resource.TeamResourceRequestModel) *ClusterRequest {
	if t == nil {
		return nil
	}
	return &ClusterRequest{
		ResourceID:    t.ResourceID,
		ResourceType:  t.ResourceType,
		RequestUserID: t.RequestUserID,
		RequestTime:   t.RequestTime,
		ApproveUserID: t.ApproveUserID,
		ApproveTime:   t.ApproveTime,
		Status:        t.Status,
	}
}

//ToClusterRequests convert TeamResourceRequestModel list to ToClusterRequest list
func ToClusterRequests(requests []team2resource.TeamResourceRequestModel) []ClusterRequest {
	var res []ClusterRequest
	for _, t := range requests {
		res = append(res, *ToClusterRequest(&t))
	}
	return res
}
