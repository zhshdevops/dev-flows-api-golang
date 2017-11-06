/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 */

package util

// IntInArry check if str exist in arr
func IntInArry(n int, arr []int) bool {
	for _, v := range arr {
		if v == n {
			return true
		}
	}
	return false
}
