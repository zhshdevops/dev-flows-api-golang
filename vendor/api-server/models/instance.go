/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-27  @author liuyang
 */

package models

import (
	"time"
)

// InstanceInfo status of an instance
type InstanceInfo struct {
	Status        string    `json:"status"`
	CreateTime    time.Time `json:"create_time"`
}
