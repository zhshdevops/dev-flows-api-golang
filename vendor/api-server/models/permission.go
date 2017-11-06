/*
* Licensed Materials - Property of tenxcloud.com
* (C) Copyright 2016 TenxCloud. All Rights Reserved.
* 2016-09-22  @author zhang_shouhong
 */

package models

import "github.com/astaxie/beego/orm"

// Permission is the structure of a permission both in mysql table and json output
type Permission struct {
	ID        string `orm:"pk;column(id)" json:"id"`
	Resource  string `orm:"size(10);column(resource)" json:"resource"`
	Operation string `orm:"size(10);column(operation)" json:"operation"`
}

// TableName returns the name of table in database
func (m *Permission) TableName() string {
	return "tenx_permissions"
}

// ListPermissions returns all permissions added
func (m *Permission) ListPermissions() ([]Permission, error) {
	o := orm.NewOrm()
	var permissions []Permission
	_, err := o.QueryTable(m.TableName()).All(&permissions)
	return permissions, err
}

// GetPermissionsByUser get all permissions for a user in a teamspace.
func (m *Permission) GetPermissionsByUser(userID int64, teamspaceID int64) (permissions []Permission, err error) {
	o := orm.NewOrm()
	sql := `SELECT 
		tenx_permissions.id,
		tenx_permissions.resource,
		tenx_permissions.operation
	FROM
		tenx_permissions
			INNER JOIN
		tenx_role_permission_ref ON tenx_permissions.id = tenx_role_permission_ref.permission_id
			INNER JOIN
		tenx_roles ON tenx_roles.id = tenx_role_permission_ref.role_id
			INNER JOIN
		tenx_user_role_ref ON tenx_roles.id = tenx_user_role_ref.role_id AND 
		                      tenx_user_role_ref.user_id = ? AND
							  tenx_user_role_ref.teamspace_id = ?;`

	_, err = o.Raw(sql, userID, teamspaceID).QueryRows(&permissions)

	return
}
