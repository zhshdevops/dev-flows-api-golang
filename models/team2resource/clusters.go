/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team2resource

import (
	sqlstatus "dev-flows-api-golang/models/sql/status"
)

// AddTeamClusters add team - cluster ref record
func AddTeamClusters(teamID string, clusterIDs []string) (uint32, error) {
	for _, clusterID := range clusterIDs {
		teamResourceModel := &TeamResourceModel{TeamID: teamID, ResourceID: clusterID, ResourceType: int(ResourceCluster)}
		if errcode, err := teamResourceModel.Insert(); err != nil {
			return errcode, err
		}
	}
	return sqlstatus.SQLSuccess, nil
}

// DeleteTeamClusters deletes team - cluster ref record
func DeleteTeamClusters(teamID string, clusterIDs []string) (uint32, error) {
	for _, clusterID := range clusterIDs {
		teamResourceModel := &TeamResourceModel{TeamID: teamID, ResourceID: clusterID, ResourceType: int(ResourceCluster)}
		if errcode, err := teamResourceModel.Delete(); err != nil {
			return errcode, err
		}
	}
	return sqlstatus.SQLSuccess, nil
}
