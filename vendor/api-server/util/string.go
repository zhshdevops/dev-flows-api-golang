/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-06  @author zhangyongkang
 */
package util

import (
	"strconv"
	"strings"
)

func SplitAndTrim(str, sep string) []string {
	parts := strings.Split(str, sep)
	result := make([]string, 0, len(parts))
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		result = append(result, s)
	}
	return result
}

// StringInArry check if str exist in arr
func StringInArry(str string, arr []string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

// UniqArr remove duplicate item
func UniqArr(arrs ...[]string) []string {
	dic := make(map[string]struct{}, 5)
	for _, arr := range arrs {
		for _, item := range arr {
			dic[item] = struct{}{}
		}
	}
	result := make([]string, 0, len(dic))
	for k := range dic {
		result = append(result, k)
	}
	return result
}

func StringToFloat64(in string) float64 {
	out, _ := strconv.Atoi(in)
	return float64(out)
}
