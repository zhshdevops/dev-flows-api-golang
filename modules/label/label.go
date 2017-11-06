/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-27  @author mengyuan
 */

package label

import "strings"

// label
const (
    PetSetTenxCloudNameKey = "tenxcloud.com/petsetName"
	App             = "tenxcloud.com/appName"
	Service         = "tenxcloud.com/svcName"
	NodeRole        = "kubeadm.alpha.kubernetes.io/role"
	AlternativeRole = "alternative.tenxcloud.com/role"
	Plugin          = "plugin"
	ClusterID       = "ClusterID"
)

// annotation
const (
	Replicas       = "tenxcloud.com/replicas"
	LivenessProve  = "tenxcloud.com/livenessProbe"
	SchemaPortname = "tenxcloud.com/schemaPortname"
)

// DecodeSchemaPortname decode string like abcdef-1/TCP,abcdef-1/TCP/12345,aaaaa/HTTP
// to [][]string split by comma(,) and slash(/)
func DecodeSchemaPortname(str string) [][]string {
	portList := make([][]string, 0, 1)
	items := strings.Split(str, ",")
	for _, item := range items {
		sub := strings.Split(item, "/")
		portList = append(portList, sub)
	}

	return portList
}

// EncodeSchemaPortname join [][]string to string use comma(,) and slash(/)
func EncodeSchemaPortname(portList [][]string) string {
	strArr := make([]string, 0, len(portList))
	for _, port := range portList {
		strArr = append(strArr, strings.Join(port, "/"))
	}
	return strings.Join(strArr, ",")
}
