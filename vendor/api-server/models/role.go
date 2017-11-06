/*
* Licensed Materials - Property of tenxcloud.com
* (C) Copyright 2016 TenxCloud. All Rights Reserved.
* 2016-09-22  @author zhang_shouhong
 */

package models

import "github.com/astaxie/beego/orm"

// Role is the structure of a role both in mysql table and json output
type Role struct {
	ID          string `orm:"pk;column(id)" json:"id"`
	Name        string `orm:"size(20);column(name)" json:"name"`
	Description string `orm:"size(150);column(description)" json:"description"`
}

// GetRolesByUser get all permissions for a user in a teamspace.
func (m *Role) GetRolesByUser(userID int64, teamspaceID int64) (roles []Role, err error) {
	o := orm.NewOrm()
	sql := `SELECT 
		tenx_roles.id,
		tenx_roles.name,
		tenx_roles.description
	FROM
		tenx_roles
			INNER JOIN
		tenx_user_role_ref ON tenx_roles.id = tenx_user_role_ref.role_id AND 
		                      tenx_user_role_ref.user_id = ? AND
							  tenx_user_role_ref.teamspace_id = ?;`

	_, err = o.Raw(sql, userID, teamspaceID).QueryRows(&roles)

	return
}
