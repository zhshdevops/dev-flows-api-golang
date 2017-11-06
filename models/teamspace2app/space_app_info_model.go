/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package teamspace2app

import (
	"fmt"
	"strings"

	sqlstatus "dev-flows-api-golang/models/sql/status"

	"github.com/astaxie/beego/orm"

	"dev-flows-api-golang/models/common"
)

// SpaceAppInfoModel defines the app summary information of a team space
// support filter by namespace, sort by app_count
type SpaceAppInfoModel struct {
	SpaceID   string `orm:"column(space_id);pk"`
	Namepsace string `orm:"column(namespace)"`
	AppCount  int64  `orm:"column(app_count)"`
}

func (u *SpaceAppInfoModel) Get(spaceID string) (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf(`explain SELECT 
    tenx_team_space.id AS space_id,
    tenx_team_space.namespace AS namespace,
    COUNT(tenx_app.id) AS app_count
FROM
    tenx_team_space
        LEFT JOIN
    tenx_app ON tenx_app.namespace = tenx_team_space.namespace
where tenx_team_space.id = ?;`)
	err := o.Raw(sql, spaceID).QueryRow(u)
	u.SpaceID = spaceID
	return sqlstatus.ParseErrorCode(err)
}

// List lists team space's app info
func (u *SpaceAppInfoModel) List(dataselect *common.DataSelectQuery) ([]SpaceAppInfoModel, uint32, error) {
	o := orm.NewOrm()
	if dataselect == nil {
		dataselect = common.DefaultDataSelect
	}
	sql := fmt.Sprintf(`SELECT 
    tenx_team_space.id AS space_id,
    tenx_team_space.namespace AS namespace,
    COUNT(tenx_app.id) AS app_count
FROM
    tenx_team_space
        LEFT JOIN
    tenx_app ON tenx_app.namespace = tenx_team_space.namespace
where %s 
GROUP BY tenx_team_space.id
%s %s;`, dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var appInfoModels []SpaceAppInfoModel
	_, err := o.Raw(sql).QueryRows(&appInfoModels)
	if err != nil {
		errcode, err := sqlstatus.ParseErrorCode(err)
		return nil, errcode, err
	}
	return appInfoModels, sqlstatus.SQLSuccess, nil
}

// List lists team space's app info
func (u *SpaceAppInfoModel) ListByIDs(spaceIDs []string, dataselect *common.DataSelectQuery) ([]SpaceAppInfoModel, uint32, error) {
	if len(spaceIDs) == 0 {
		return []SpaceAppInfoModel{}, sqlstatus.SQLSuccess, nil
	}

	spaceIDStr := fmt.Sprintf("'%s'", strings.Join(spaceIDs, "', '"))

	o := orm.NewOrm()
	if dataselect == nil {
		dataselect = common.NoDataSelect
	}

	sql := fmt.Sprintf(`SELECT 
    tenx_team_space.id AS space_id,
    tenx_team_space.namespace AS namespace,
    COUNT(tenx_app.id) AS app_count
FROM
    tenx_team_space
        LEFT JOIN
    tenx_app ON tenx_app.namespace = tenx_team_space.namespace
where tenx_team_space.id in (%s) and %s 
GROUP BY tenx_team_space.id
%s %s;`, spaceIDStr, dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var appInfoModels []SpaceAppInfoModel
	_, err := o.Raw(sql).QueryRows(&appInfoModels)
	if err != nil {
		errcode, err := sqlstatus.ParseErrorCode(err)
		return nil, errcode, err
	}
	return appInfoModels, sqlstatus.SQLSuccess, nil
}
