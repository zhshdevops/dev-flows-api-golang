/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-27  @author liuyang
 */

package models

// AutoScaleInfo autoscale status
type AutoScaleInfo struct {
	MinCnt       uint32 `json:"min_cnt,omitempty"`
	MaxCnt       uint32 `json:"max_cnt"`
	CPUThreshold uint32 `json:"cpu_threshold,omitempty"`
}
