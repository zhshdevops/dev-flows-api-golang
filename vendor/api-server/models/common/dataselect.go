/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */
package common

import (
	"fmt"
	"strings"

	sqlutil "api-server/models/sql/util"
)

// PaginationQuery defines
type PaginationQuery struct {
	From int
	Size int
}

// NoPagination is the option for no pigination, currently a very large page
var NoPagination = NewPaginationQuery(-1, -1)

// DefaultPagination is the default options for pagination
var DefaultPagination = NewPaginationQuery(0, 10)

// NewPaginationQuery creates a new PaginationQuery
func NewPaginationQuery(from, size int) *PaginationQuery {
	return &PaginationQuery{
		From: from,
		Size: size,
	}
}

// IsValid checks if a PaginationQuery is valid
func (pq *PaginationQuery) IsValid() bool {
	if pq == nil {
		return false
	}
	if pq.From == -1 && pq.Size == -1 {
		return true
	}
	return pq.From >= 0 && pq.Size > 0
}

func (pq *PaginationQuery) String() string {
	if pq == nil {
		return ""
	}
	if pq.From == -1 || pq.Size == -1 {
		return ""
	}
	return fmt.Sprintf(" LIMIT %d OFFSET %d ", pq.Size, pq.From)
}

// SortQuery holds options for sort functionality of data select.
type SortQuery struct {
	SortByList []SortBy
}

// SortBy holds the name of the property that should be sorted and whether order should be ascending or descending.
type SortBy struct {
	Property  string
	Ascending bool
}

// NoSort is as option for no sort.
var NoSort = &SortQuery{
	SortByList: []SortBy{},
}

// NewSortQuery takes raw sort options list and returns SortQuery object. For example:
// ["a", "parameter1", "d", "parameter2"] - means that the data should be sorted by
// parameter1 (ascending) and later - for results that return equal under parameter 1 sort - by parameter2 (descending)
func NewSortQuery(sortByListRaw []string) *SortQuery {
	if sortByListRaw == nil || len(sortByListRaw)%2 == 1 {
		return &SortQuery{SortByList: []SortBy{}}
	}
	sortByList := []SortBy{}
	for i := 0; i+1 < len(sortByListRaw); i += 2 {
		// parse order option
		var ascending bool
		orderOption := sortByListRaw[i]
		if orderOption == "a" {
			ascending = true
		} else if orderOption == "d" {
			ascending = false
		} else {
			//  Invalid order option. Only ascending (a), descending (d) options are supported
			return &SortQuery{SortByList: []SortBy{}}
		}

		// parse property name
		propertyName := sortByListRaw[i+1]
		sortBy := SortBy{
			Property:  propertyName,
			Ascending: ascending,
		}
		// Add to the sort options.
		sortByList = append(sortByList, sortBy)
	}
	return &SortQuery{
		SortByList: sortByList,
	}
}

func (sq *SortQuery) String() string {
	if sq == nil {
		return ""
	}
	if len(sq.SortByList) == 0 {
		return ""
	}
	res := fmt.Sprintf(" ORDER BY %s ", sq.SortByList[0].Property)
	if sq.SortByList[0].Ascending == false {
		res += "desc "
	}
	return res
}

const (
	FilterTypeLikeSuffix     = "__like"
	FilterTypeEqualSuffix    = "__eq"
	FilterTypeNotEqualSuffix = "__neq"
)

type FilterType int

const (
	FilterTypeLike     = 0
	FilterTypeEqual    = 1
	FilterTypeNotEqual = 2
)

// FilterQuery holds options for filter functionality of data select.
// currently it only supports string match
type FilterQuery struct {
	FilterByList []FilterBy
}

// FilterBy holds the name of the property that should be .
type FilterBy struct {
	Property string
	Type     FilterType
	Value    string
}

// NoFilter is as option for no filter.
var NoFilter = &FilterQuery{
	FilterByList: []FilterBy{},
}

// NewFilterQuery takes raw filter options list and returns FilterQuery object. For example:
// ["user_name", "test"] - means that the data should be filtered by
// user_name (%test%), means user_name contains 'test'
func NewFilterQuery(filterByListRaw []string) *FilterQuery {
	if filterByListRaw == nil || len(filterByListRaw)%2 == 1 {
		return &FilterQuery{FilterByList: []FilterBy{}}
	}
	filterByList := parseFilter(filterByListRaw)
	return &FilterQuery{
		FilterByList: filterByList,
	}
}

func (fq *FilterQuery) Add(filterByListRaw []string) {
	filterByList := parseFilter(filterByListRaw)
	fq.FilterByList = append(fq.FilterByList, filterByList...)
}
func parseFilter(filterByListRaw []string) []FilterBy {
	filterByList := []FilterBy{}
	for i := 0; i+1 < len(filterByListRaw); i += 2 {
		// parse property name and value
		propertyName := filterByListRaw[i]
		propertyValue := filterByListRaw[i+1]
		filterBy := FilterBy{
			Property: propertyName,
			Type:     FilterTypeLike,
			Value:    propertyValue,
		}
		if strings.HasSuffix(propertyName, FilterTypeEqualSuffix) {
			filterBy.Type = FilterTypeEqual
			filterBy.Property = strings.TrimSuffix(filterBy.Property, FilterTypeEqualSuffix)
		} else if strings.HasSuffix(propertyName, FilterTypeNotEqualSuffix) {
			filterBy.Type = FilterTypeNotEqual
			filterBy.Property = strings.TrimSuffix(filterBy.Property, FilterTypeNotEqualSuffix)
		} else {
			filterBy.Property = strings.TrimSuffix(filterBy.Property, FilterTypeLikeSuffix)
		}

		// Add to the sort options.
		filterByList = append(filterByList, filterBy)
	}

	return filterByList
}
func (fq *FilterQuery) String() string {
	if fq == nil || len(fq.FilterByList) == 0 {
		return " '1' = '1' "
	}

	var items []string
	for _, filterBy := range fq.FilterByList {
		if filterBy.Type == FilterTypeEqual {
			items = append(items, fmt.Sprintf(" %s='%s' ", filterBy.Property, sqlutil.EscapeStringBackslash(filterBy.Value)))
		} else if filterBy.Type == FilterTypeNotEqual {
			items = append(items, fmt.Sprintf(" %s!='%s' ", filterBy.Property, sqlutil.EscapeStringBackslash(filterBy.Value)))
		} else {
			items = append(items, fmt.Sprintf(" %s like '%%%s%%' ", filterBy.Property, sqlutil.EscapeUnderlineInLikeStatement(filterBy.Value)))
		}
	}

	return strings.Join(items, "and")
}

// DataSelectQuery currently included only Pagination and Sort options.
type DataSelectQuery struct {
	PaginationQuery *PaginationQuery
	SortQuery       *SortQuery
	FilterQuery     *FilterQuery
}

// NoDataSelect fetches all items with no sort.
var NoDataSelect = NewDataSelectQuery(NoPagination, NoSort, NoFilter)

// DefaultDataSelect fetches first 10 items from page 1 with no sort.
var DefaultDataSelect = NewDataSelectQuery(DefaultPagination, NoSort, NoFilter)

// NewDataSelectQuery creates DataSelectQuery object from simpler data select queries.
func NewDataSelectQuery(paginationQuery *PaginationQuery, sortQuery *SortQuery, filterQuery *FilterQuery) *DataSelectQuery {
	dataselect := &DataSelectQuery{
		PaginationQuery: &PaginationQuery{From: 0, Size: 10},
		SortQuery:       &SortQuery{SortByList: []SortBy{}},
		FilterQuery:     &FilterQuery{FilterByList: []FilterBy{}},
	}
	if paginationQuery != nil {
		pq := *paginationQuery
		dataselect.PaginationQuery = &pq
	}
	if sortQuery != nil && len(sortQuery.SortByList) > 0 {
		dataselect.SortQuery.SortByList = append(dataselect.SortQuery.SortByList, sortQuery.SortByList...)
	}
	if filterQuery != nil && len(filterQuery.FilterByList) > 0 {
		dataselect.FilterQuery.FilterByList = append(dataselect.FilterQuery.FilterByList, filterQuery.FilterByList...)
	}
	return dataselect
}
