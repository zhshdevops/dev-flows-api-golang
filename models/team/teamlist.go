/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package team

import (
	"fmt"
	"time"

	sqlstatus "dev-flows-api-golang/models/sql/status"

	"github.com/golang/glog"

	"dev-flows-api-golang/models/common"
	"dev-flows-api-golang/models/team2resource"
)

var (
	// SortByMapping is the mapping from display name to fields in tenx_teams table
	SortByMapping = map[string]string{
		"teamName":     "name",
		"creationTime": "creation_time",
	}
	// SortBySpaceInfoMapping is the fields defined by code
	SortBySpaceInfoMapping = map[string]string{
		"spaceCount": "space_count",
	}
	// SortByUserInfoMapping is the fields defined by code
	SortByUserInfoMapping = map[string]string{
		"userCount": "user_count",
	}
	// SortByResourceInfoMapping is the fileds defined by code
	SortByResourceInfoMapping = map[string]string{
		"clusterCount": "cluster_count",
	}
	// FilterByMapping is the mapping from display name to fields in tenx_teams table
	FilterByMapping = map[string]string{
		"teamName":  "tenx_teams.name",
		"creatorID": "tenx_teams.creator_id",
	}
)

type Team struct {
	TeamID       string    `json:"teamID"`
	TeamName     string    `json:"teamName"`
	Description  string    `json:"description"`
	CreatorID    int32     `json:"creatorID"`
	CreationTime time.Time `json:"creationTime"`
	SpaceCount   int       `json:"spaceCount"`
	UserCount    int       `json:"userCount"`
	ClusterCount int       `json:"clusterCount"`
	IsCreator    bool      `json:"isCreator"`
}

type TeamList struct {
	ListMeta common.ListMeta `json:"listMeta"`
	Teams    []Team          `json:"teams"`
}

func ToTeam(t *TeamModel, userInfo *TeamUserInfoModel, spaceInfo *TeamSpaceInfoModel, resourceInfo *team2resource.TeamResourceInfoModel) *Team {
	if t == nil {
		return nil
	}
	res := &Team{
		TeamID:       t.TeamID,
		TeamName:     t.TeamName,
		Description:  t.Description,
		CreatorID:    t.CreatorID,
		CreationTime: t.CreationTime,
	}
	if userInfo != nil {
		res.UserCount = userInfo.UserCount
	}
	if spaceInfo != nil {
		res.SpaceCount = spaceInfo.SpaceCount
	}
	if resourceInfo != nil {
		res.ClusterCount = resourceInfo.ClusterCount
	}
	return res
}

// validateTeamDataSelect filter out invalid sort, filter options, and return valid ones
func validateTeamDataSelect(isAdmin bool, old *common.DataSelectQuery) (*common.DataSelectQuery, error) {
	if old == nil {
		return common.NewDataSelectQuery(common.DefaultPagination, common.NewSortQuery([]string{"a", SortByMapping["teamName"]}), common.NoFilter), nil
	}

	dataselect := common.NewDataSelectQuery(old.PaginationQuery, common.NoSort, common.NoFilter)
	if old.FilterQuery != nil {
		for _, f := range old.FilterQuery.FilterByList {
			// Admin user should not check creatorID and view all teams
			if isAdmin && f.Property == "creatorID" {
				continue
			}
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
				prop, ok = SortByUserInfoMapping[sq.Property]
				if !ok {
					prop, ok = SortBySpaceInfoMapping[sq.Property]
					if !ok {
						prop, ok = SortByResourceInfoMapping[sq.Property]
						if !ok {
							glog.Errorf("team list, invalid sort options: %s\n", sq.Property)
							return nil, fmt.Errorf("invalid sort option: %s\n", sq.Property)
						}
					}
				}
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

func ToTeams(teams []TeamModel, userInfos []TeamUserInfoModel, spaceInfos []TeamSpaceInfoModel, resourceInfos []team2resource.TeamResourceInfoModel, sortBy string) []Team {
	var res []Team
	if _, ok := SortByResourceInfoMapping[sortBy]; ok {
		teamModelMapping := make(map[string]TeamModel)
		for _, t := range teams {
			teamModelMapping[t.TeamID] = t
		}

		spaceModelMapping := make(map[string]TeamSpaceInfoModel)
		for _, t := range spaceInfos {
			spaceModelMapping[t.TeamID] = t
		}

		userModelMapping := make(map[string]TeamUserInfoModel)
		for _, t := range userInfos {
			userModelMapping[t.TeamID] = t
		}

		for _, ri := range resourceInfos {
			t, ok := teamModelMapping[ri.TeamID]
			ui, ok2 := userModelMapping[ri.TeamID]
			si, ok3 := spaceModelMapping[ri.TeamID]
			if ok && ok2 && ok3 {
				newT := ToTeam(&t, &ui, &si, &ri)
				res = append(res, *newT)
			}
		}

	} else if _, ok := SortByUserInfoMapping[sortBy]; ok {
		teamModelMapping := make(map[string]TeamModel)
		for _, t := range teams {
			teamModelMapping[t.TeamID] = t
		}

		spaceModelMapping := make(map[string]TeamSpaceInfoModel)
		for _, t := range spaceInfos {
			spaceModelMapping[t.TeamID] = t
		}

		resourceModelMapping := make(map[string]team2resource.TeamResourceInfoModel)
		for _, t := range resourceInfos {
			resourceModelMapping[t.TeamID] = t
		}

		for _, ui := range userInfos {
			t, ok := teamModelMapping[ui.TeamID]
			si, ok2 := spaceModelMapping[ui.TeamID]
			ri, ok3 := resourceModelMapping[ui.TeamID]
			if ok && ok2 && ok3 {
				newT := ToTeam(&t, &ui, &si, &ri)
				res = append(res, *newT)
			}
		}
	} else if _, ok := SortBySpaceInfoMapping[sortBy]; ok {
		teamModelMapping := make(map[string]TeamModel)
		for _, t := range teams {
			teamModelMapping[t.TeamID] = t
		}

		userModelMapping := make(map[string]TeamUserInfoModel)
		for _, t := range userInfos {
			userModelMapping[t.TeamID] = t
		}

		resourceModelMapping := make(map[string]team2resource.TeamResourceInfoModel)
		for _, t := range resourceInfos {
			resourceModelMapping[t.TeamID] = t
		}

		for _, si := range spaceInfos {
			t, ok := teamModelMapping[si.TeamID]
			ui, ok2 := userModelMapping[si.TeamID]
			ri, ok3 := resourceModelMapping[si.TeamID]
			if ok && ok2 && ok3 {
				newT := ToTeam(&t, &ui, &si, &ri)
				res = append(res, *newT)
			}
		}
	} else {
		userModelMapping := make(map[string]TeamUserInfoModel)
		for _, t := range userInfos {
			userModelMapping[t.TeamID] = t
		}

		spaceModelMapping := make(map[string]TeamSpaceInfoModel)
		for _, t := range spaceInfos {
			spaceModelMapping[t.TeamID] = t
		}

		resourceModelMapping := make(map[string]team2resource.TeamResourceInfoModel)
		for _, t := range resourceInfos {
			resourceModelMapping[t.TeamID] = t
		}

		for _, t := range teams {
			ui, ok := userModelMapping[t.TeamID]
			si, ok2 := spaceModelMapping[t.TeamID]
			ri, ok3 := resourceModelMapping[si.TeamID]
			if ok && ok2 && ok3 {
				newT := ToTeam(&t, &ui, &si, &ri)
				res = append(res, *newT)
			}
		}
	}
	return res
}

// GetTeamList fetches teams (no resource summary information included, e.g. cluster number)
func GetTeamList(userID int32, dataselect *common.DataSelectQuery) (*TeamList, uint32, error) {
	newdataselect, err := validateTeamDataSelect(false, dataselect)
	if err != nil {
		glog.Errorf("GetTeamList fails, error:%s\n", err)
		return nil, sqlstatus.SQLErrSyntax, err
	}

	team := &TeamModel{}
	teamIDs, errcode, err := team.ListIDs(userID, newdataselect)
	if err != nil {
		glog.Errorf("get team ids fails, error code:%d, error:%v\n", errcode, err)
		return nil, errcode, err
	}

	return GetTeamListByIDs(false, teamIDs, dataselect)
}

// GetTeamList fetches teams (no resource summary information included, e.g. cluster number)
func GetTeamListByAdmin(isAdmin bool, dataselect *common.DataSelectQuery) (*TeamList, uint32, error) {
	team := &TeamModel{}
	teamIDs, errcode, err := team.ListAllIDs()
	if err != nil {
		glog.Errorf("get team ids fails, error code:%d, error:%v\n", errcode, err)
		return nil, errcode, err
	}
	return GetTeamListByIDs(isAdmin, teamIDs, dataselect)
}

// GetTeamListByIDs fetches teams according to IDs  (no resource summary information included, e.g. cluster number)
func GetTeamListByIDs(isAdmin bool, teamIDs []string, dataselect *common.DataSelectQuery) (*TeamList, uint32, error) {
	if len(teamIDs) == 0 {
		return &TeamList{ListMeta: common.ListMeta{Total: 0}, Teams: make([]Team, 0)}, sqlstatus.SQLSuccess, nil
	}

	var sortby string
	if dataselect != nil && dataselect.SortQuery != nil && len(dataselect.SortQuery.SortByList) > 0 {
		sortby = dataselect.SortQuery.SortByList[0].Property
	}

	dataselect, err := validateTeamDataSelect(isAdmin, dataselect)
	if err != nil {
		glog.Errorf("GetTeamList fails, error:%s\n", err)
		return nil, sqlstatus.SQLErrSyntax, err
	}
	var teams []TeamModel
	var teamUserInfos []TeamUserInfoModel
	var teamSpaceInfos []TeamSpaceInfoModel
	var teamResourceInfos []team2resource.TeamResourceInfoModel

	teamModel := &TeamModel{}
	teamUserInfoModel := &TeamUserInfoModel{}
	teamSpaceInfoModel := &TeamSpaceInfoModel{}
	resourceInfoModel := &team2resource.TeamResourceInfoModel{}

	var errcode uint32

	teamNumber := len(teamIDs)
	if _, ok := SortByResourceInfoMapping[sortby]; ok {
		// if sort by cluster_count
		teamResourceInfos, errcode, err = resourceInfoModel.ListByTeams(teamIDs, common.NewDataSelectQuery(dataselect.PaginationQuery, dataselect.SortQuery, common.NoFilter))
		if err != nil {
			glog.Errorf("get team resource summary list fails, error code:%d, error:%v\n", errcode, err)
			return nil, errcode, err
		}
		if len(teamResourceInfos) == 0 {
			return &TeamList{Teams: []Team{}}, sqlstatus.SQLSuccess, nil
		}
		var teamIDs []string
		for _, ti := range teamResourceInfos {
			teamIDs = append(teamIDs, ti.TeamID)
		}
		// get team list
		teams, errcode, err = teamModel.ListByIDs(teamIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list teams of ids %v fails, error code:%d, error:%v\n", teamIDs, errcode, err)
			return nil, errcode, err
		}
		teamSpaceInfos, errcode, err = teamSpaceInfoModel.ListByTeamIDs(teamIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list team space info of ids %v fails, error code:%d, error:%v\n", teamIDs, errcode, err)
			return nil, errcode, err
		}
		teamUserInfos, errcode, err = teamUserInfoModel.List(common.NoDataSelect)
		if err != nil {
			glog.Errorf("list team information fails, error code:%d, error:%v\n", errcode, err)
			return nil, errcode, err
		}
	} else if _, ok := SortByUserInfoMapping[sortby]; ok {
		// if sort by space_number/user_number
		teamUserInfos, errcode, err = teamUserInfoModel.ListByTeamIDs(teamIDs, dataselect)
		if err != nil {
			glog.Errorf("list team information fails, error code:%d, error:%v\n", errcode, err)
			return nil, errcode, err
		}

		if len(teamUserInfos) == 0 {
			return &TeamList{Teams: []Team{}}, sqlstatus.SQLSuccess, nil
		}

		var teamIDs []string
		for _, ti := range teamUserInfos {
			teamIDs = append(teamIDs, ti.TeamID)
		}

		// get team list
		teams, errcode, err = teamModel.ListByIDs(teamIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list teams of ids %v fails, error code:%d, error:%v\n", teamIDs, errcode, err)
			return nil, errcode, err
		}

		teamSpaceInfos, errcode, err = teamSpaceInfoModel.ListByTeamIDs(teamIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list team space info of ids %v fails, error code:%d, error:%v\n", teamIDs, errcode, err)
			return nil, errcode, err
		}
		teamResourceInfos, errcode, err = resourceInfoModel.ListByTeams(teamIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get team resource summary list fails, error code:%d, error:%v\n", errcode, err)
			return nil, errcode, err
		}
	} else if _, ok := SortBySpaceInfoMapping[sortby]; ok {
		teamSpaceInfos, errcode, err = teamSpaceInfoModel.ListByTeamIDs(teamIDs, dataselect)
		if err != nil {
			glog.Errorf("list team user info fails, error code:%d, error:%v\n", errcode, err)
			return nil, errcode, err
		}
		if len(teamSpaceInfos) == 0 {
			return &TeamList{Teams: []Team{}, ListMeta: common.ListMeta{Total: teamNumber}}, sqlstatus.SQLSuccess, nil
		}

		var teamIDs []string
		for _, ti := range teamSpaceInfos {
			teamIDs = append(teamIDs, ti.TeamID)
		}

		// get team list
		teams, errcode, err = teamModel.ListByIDs(teamIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list teams of ids %v fails, error code:%d, error:%v\n", teamIDs, errcode, err)
			return nil, errcode, err
		}

		teamUserInfos, errcode, err = teamUserInfoModel.List(common.NoDataSelect)
		if err != nil {
			glog.Errorf("list team information fails, error code:%d, error:%v\n", errcode, err)
			return nil, errcode, err
		}

		teamResourceInfos, errcode, err = resourceInfoModel.ListByTeams(teamIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get team resource summary list fails, error code:%d, error:%v\n", errcode, err)
			return nil, errcode, err
		}
	} else {
		// get team list
		teams, errcode, err = teamModel.ListByIDs(teamIDs, dataselect)
		if err != nil {
			glog.Errorf("list teams of ids %v fails, error code:%d, error:%v\n", teamIDs, errcode, err)
			return nil, errcode, err
		}
		if len(teams) == 0 {
			return &TeamList{Teams: []Team{}, ListMeta: common.ListMeta{Total: teamNumber}}, sqlstatus.SQLSuccess, nil
		}

		// get team info list
		teamIDs = nil
		for _, t := range teams {
			teamIDs = append(teamIDs, t.TeamID)
		}
		// user summary info
		teamUserInfos, errcode, err = teamUserInfoModel.ListByTeamIDs(teamIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list team infos fails, error code: %d, error:%v\n", errcode, err)
			return nil, errcode, err
		}

		// space summary info
		teamSpaceInfos, errcode, err = teamSpaceInfoModel.ListByTeamIDs(teamIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list team space info of ids %v fails, error code:%d, error:%v\n", teamIDs, errcode, err)
			return nil, errcode, err
		}

		// resource infos, e.g. cluster_count
		teamResourceInfos, errcode, err = resourceInfoModel.ListByTeams(teamIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("get team resource summary list fails, error code:%d, error:%v\n", errcode, err)
			return nil, errcode, err
		}
	}

	tl := &TeamList{
		ListMeta: common.ListMeta{
			Total: teamNumber,
		},
		Teams: ToTeams(teams, teamUserInfos, teamSpaceInfos, teamResourceInfos, sortby),
	}
	return tl, sqlstatus.SQLSuccess, nil
}
