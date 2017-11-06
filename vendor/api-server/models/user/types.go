/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package user

// UserRole defines some enums for the role field in tenx_users
type UserRole int

const (
	UserRoleNormal = iota
	UserRoleTeamManager
	UserRoleAdmin
)

const (
	// SuperUserName default super user name for now
	SuperUserName = "admin"
)
