/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package teamspace

import (
	"fmt"
	"time"

	"api-server/models/common"
	sqlstatus "api-server/models/sql/status"
	"api-server/models/teamspace2app"

	"github.com/golang/glog"
)

var (
	// SortByMapping is the mapping from display name to fields in tenx_team_space table
	SortByMapping = map[string]string{
		"spaceName":    "name",
		"creationTime": "creation_time",
	}
	// SortByAccountMapping is the mapping from display name to tenx_space_account table
	SortByAccountMapping = map[string]string{
		"balance": "balance",
	}
	// FilterByMapping is the mapping from display name to fields in tenx_team_space table
	FilterByMapping = map[string]string{
		"spaceName": "namespace", // currently use namespace, because the field exists in both tenx_team_space and tenx_team_space_account
	}

	// SortByAppMapping is the mapping from display name to temporary table fields
	SortByAppMapping = map[string]string{
		"appCount": "app_count",
	}
)

// Space tenx_users record
type Space struct {
	SpaceID      string    `json:"spaceID"`
	SpaceName    string    `json:"spaceName"`
	Namespace    string    `json:"namespace"`
	Description  string    `json:"description"`
	TeamID       string    `json:"teamID"`
	TeamName     string    `json:"teamName"`
	CreationTime time.Time `json:"creationTime"`
	Role         int32     `json:"role"`
	Balance      int64     `json:"balance"`
	UserCount    int64     `json:"userCount"`
	AppCount     int64     `json:"appCount"`
}

type SpaceList struct {
	ListMeta common.ListMeta `json:"listMeta"`
	Spaces   []Space         `json:"spaces"`
}

func ToSpace(s *SpaceModel, accountInfo *SpaceAccountModel, spaceAppInfo *teamspace2app.SpaceAppInfoModel) *Space {
	if s == nil {
		return nil
	}
	res := &Space{
		SpaceID:      s.SpaceID,
		SpaceName:    s.SpaceName,
		Namespace:    s.Namespace,
		Description:  s.Description,
		TeamID:       s.TeamID,
		CreationTime: s.CreationTime,
	}

	if accountInfo != nil {
		res.Balance = accountInfo.Balance
	}
	if spaceAppInfo != nil {
		res.AppCount = spaceAppInfo.AppCount
	}
	return res
}

func ToSpaces(spaces []SpaceModel, accounts []SpaceAccountModel, spaceAppInfos []teamspace2app.SpaceAppInfoModel, orderBy string) []Space {
	var res []Space

	// sort by account info, now balance
	if _, ok := SortByAccountMapping[orderBy]; ok {
		spaceMap := make(map[string]SpaceModel)
		for _, sp := range spaces {
			spaceMap[sp.SpaceID] = sp
		}
		spaceAppInfoMapping := make(map[string]teamspace2app.SpaceAppInfoModel)
		for _, sai := range spaceAppInfos {
			spaceAppInfoMapping[sai.SpaceID] = sai
		}

		for _, acc := range accounts {
			sp := spaceMap[acc.SpaceID]
			sai := spaceAppInfoMapping[acc.SpaceID]
			res = append(res, *ToSpace(&sp, &acc, &sai))
		}
	} else if _, ok := SortByAppMapping[orderBy]; ok { // sort by app infos
		spaceMap := make(map[string]SpaceModel)
		for _, sp := range spaces {
			spaceMap[sp.SpaceID] = sp
		}
		accMap := make(map[string]SpaceAccountModel)
		for _, acc := range accounts {
			accMap[acc.SpaceID] = acc
		}

		for _, sai := range spaceAppInfos {
			sp := spaceMap[sai.SpaceID]
			acc := accMap[sai.SpaceID]
			res = append(res, *ToSpace(&sp, &acc, &sai))

		}
	} else {
		// sort by name, creation_time
		accMap := make(map[string]SpaceAccountModel)
		for _, acc := range accounts {
			accMap[acc.SpaceID] = acc
		}
		spaceAppInfoMapping := make(map[string]teamspace2app.SpaceAppInfoModel)
		for _, sai := range spaceAppInfos {
			spaceAppInfoMapping[sai.SpaceID] = sai
		}

		glog.V(4).Infof("----------------> space number: %d\n", len(spaces))
		for _, sp := range spaces {
			acc := accMap[sp.SpaceID]
			sai := spaceAppInfoMapping[sp.SpaceID]
			res = append(res, *ToSpace(&sp, &acc, &sai))
		}
	}
	return res
}

// validateSpaceDataSelect filter out invalid sort, filter options, and return valid ones
func validateSpaceDataSelect(old *common.DataSelectQuery) (*common.DataSelectQuery, error) {
	if old == nil {
		return common.NewDataSelectQuery(common.DefaultPagination, common.NewSortQuery([]string{"a", SortByMapping["spaceName"]}), common.NoFilter), nil
	}

	dataselect := common.NewDataSelectQuery(old.PaginationQuery, common.NoSort, common.NoFilter)
	if old.FilterQuery != nil {
		for _, f := range old.FilterQuery.FilterByList {
			prop, ok := FilterByMapping[f.Property]
			if !ok {
				glog.Errorf("space list, invalid filter by options: %s\n", f.Property)
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
				prop, ok = SortByAccountMapping[sq.Property]
				if !ok {
					prop, ok = SortByAppMapping[sq.Property]
					if !ok {
						glog.Errorf("space list, invalid sort options: %s\n", sq.Property)
						return nil, fmt.Errorf("invalid sort option: %s\n", sq.Property)
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

// GetSpaceListByIDs fetches spaces according to users
func GetSpaceListByIDs(spaceIDs []string, dataselect *common.DataSelectQuery) (*SpaceList, uint32, error) {
	if len(spaceIDs) == 0 {
		return &SpaceList{ListMeta: common.ListMeta{Total: 0}, Spaces: make([]Space, 0)}, sqlstatus.SQLSuccess, nil
	}

	var sortBy string
	if dataselect != nil && dataselect.SortQuery != nil && len(dataselect.SortQuery.SortByList) > 0 {
		sortBy = dataselect.SortQuery.SortByList[0].Property
	}

	dataselect, err := validateSpaceDataSelect(dataselect)
	if err != nil {
		glog.Errorf("Get space list by ids %v fails, error:%v\n", spaceIDs, err)
		return nil, sqlstatus.SQLErrSyntax, err
	}

	total := len(spaceIDs)
	sp := &SpaceModel{}
	spAcc := &SpaceAccountModel{}
	spaceAppInfoModel := &teamspace2app.SpaceAppInfoModel{}
	var spaces []SpaceModel
	var spaceAccounts []SpaceAccountModel
	var spaceAppInfos []teamspace2app.SpaceAppInfoModel

	var errcode uint32

	if _, ok := SortByAccountMapping[sortBy]; ok {
		// get space accounts first(because sort by balance)
		spaceAccounts, errcode, err = spAcc.ListByIDs(spaceIDs, dataselect)
		if err != nil {
			glog.Errorf("list spaces accounts of ids %v fails, error code:%d, error:%v\n", spaceIDs, errcode, err)
			return nil, errcode, err
		}

		spaceIDs = nil
		for _, acc := range spaceAccounts {
			spaceIDs = append(spaceIDs, acc.SpaceID)
		}

		// get spaces second
		spaces, errcode, err = sp.ListByIDs(spaceIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list spaces of ids %v fails, error code:%d, error:%v\n", spaceIDs, errcode, err)
			return nil, errcode, err
		}

		// get space's app infos
		spaceAppInfos, errcode, err = spaceAppInfoModel.ListByIDs(spaceIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list spaces of ids %v fails, error code:%d, error:%v\n", spaceIDs, errcode, err)
			return nil, errcode, err
		}

		tl := &SpaceList{
			Spaces: ToSpaces(spaces, spaceAccounts, spaceAppInfos, sortBy),
			ListMeta: common.ListMeta{
				Total: total,
			},
		}
		return tl, sqlstatus.SQLSuccess, nil
	} else if _, ok := SortByAppMapping[sortBy]; ok {
		spaceAppInfos, errcode, err = spaceAppInfoModel.ListByIDs(spaceIDs, dataselect)
		if err != nil {
			glog.Errorf("list space app infos of ids %v fails, error code:%d, error:%v\n", spaceIDs, errcode, err)
			return nil, errcode, err
		}
		spaceIDs = nil
		for _, spaceApp := range spaceAppInfos {
			spaceIDs = append(spaceIDs, spaceApp.SpaceID)
		}

		// get spaces second
		spaces, errcode, err = sp.ListByIDs(spaceIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list spaces of ids %v fails, error code:%d, error:%v\n", spaceIDs, errcode, err)
			return nil, errcode, err
		}

		// get space accounts
		spaceAccounts, errcode, err = spAcc.ListByIDs(spaceIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list spaces accounts of ids %v fails, error code:%d, error:%v\n", spaceIDs, errcode, err)
			return nil, errcode, err
		}

		tl := &SpaceList{
			Spaces: ToSpaces(spaces, spaceAccounts, spaceAppInfos, sortBy),
			ListMeta: common.ListMeta{
				Total: total,
			},
		}
		return tl, sqlstatus.SQLSuccess, nil
	} else {
		// sort by SpaceModel fields, or no sorting
		// get space list
		spaces, errcode, err = sp.ListByIDs(spaceIDs, dataselect)
		if err != nil {
			glog.Errorf("list spaces of ids %v fails, error code:%d, error:%v\n", spaceIDs, errcode, err)
			return nil, errcode, err
		}

		// get space accounts
		spaceAccounts, errcode, err = spAcc.ListByIDs(spaceIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list spaces accounts of ids %v fails, error code:%d, error:%v\n", spaceIDs, errcode, err)
			return nil, errcode, err
		}

		// get space's app infos
		spaceAppInfos, errcode, err = spaceAppInfoModel.ListByIDs(spaceIDs, common.NoDataSelect)
		if err != nil {
			glog.Errorf("list spaces of ids %v fails, error code:%d, error:%v\n", spaceIDs, errcode, err)
			return nil, errcode, err
		}

		tl := &SpaceList{
			Spaces: ToSpaces(spaces, spaceAccounts, spaceAppInfos, sortBy),
			ListMeta: common.ListMeta{
				Total: total,
			},
		}
		return tl, sqlstatus.SQLSuccess, nil
	}
}

// GetSpaceListByTeams fetches spaces according to team ids
func GetSpaceListByTeams(teamIDs []string, dataselect *common.DataSelectQuery) (*SpaceList, uint32, error) {
	if len(teamIDs) == 0 {
		return &SpaceList{Spaces: []Space{}}, sqlstatus.SQLSuccess, nil
	}

	sp := &SpaceModel{}
	spaceIDs, errcode, err := sp.ListSpaceIDsByTeams(teamIDs, dataselect)
	if err != nil {
		glog.Errorf("get space ids of teams %v fails, error code:%d, error:%v\n", teamIDs, errcode, err)
		return nil, errcode, err
	}

	return GetSpaceListByIDs(spaceIDs, dataselect)
}
