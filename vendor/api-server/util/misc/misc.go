/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-12-01  @author mengyuan
 */

package misc

import (
	"api-server/modules/tenx/id"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego"
)

const (
	StandardDeploymentMode   = "standard"
	RunMode_Dev              = "Dev"
	EnterpriseDeploymentMode = "enterprise"
)

func AddSurroundToSlice(raw []string, s string) []string {
	out := make([]string, 0, len(raw))
	for i := range raw {
		item := s + raw[i] + s
		out = append(out, item)
	}
	return out
}
func SliceToString(raw []string, surround, sep string) string {
	return strings.Join(AddSurroundToSlice(raw, surround), sep)
}

// PrettyPrint only for debug
func PrettyPrint(data interface{}, prefix ...string) {
	tmp, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(prefix, string(tmp))
}

func RemoveDuplication(dup []string) []string {
	m := make(map[string]struct{})
	for _, item := range dup {
		m[item] = struct{}{}
	}
	uniq := make([]string, 0, len(m))
	for item := range m {
		uniq = append(uniq, item)
	}
	return uniq
}

// CheckCode check code format
func CheckCode(code string) bool {
	if len(code) != 20 {
		return false
	}
	raw := []byte(code[:16])
	checkBs, err := base64.URLEncoding.DecodeString(code[16:])
	if err != nil {
		return false
	}
	maskList := [...]byte{6, 8, 7, 5, 6, 9}
	indexList := [...]byte{20, 15, 4, 19, 5, 6}
	for i := 0; i < len(indexList); i += 2 {
		index := indexList[i]
		shift := 4 * ((index + 1) & 1)
		index >>= 1
		high := ((raw[index] >> shift) & 0x0F) ^ (checkBs[i/2] >> 4)
		if maskList[i] != high {
			return false
		}

		index = indexList[i+1]
		shift = 4 * ((index + 1) & 1)
		index >>= 1
		low := (((raw[index] >> shift) & 0x0F) ^ (checkBs[i/2] & 0x0F))
		if maskList[i+1] != low {
			return false
		}
	}
	return true
}

// GenSelfCheckCode generate 20 length string which can self check
func GenSelfCheckCode() string {
	raw := id.New16LengthsCode() // 16 characters ascii, 128 bit
	bs := []byte(raw)
	maskList := [...]byte{6, 8, 7, 5, 6, 9}
	indexList := [...]byte{20, 15, 4, 19, 5, 6}
	checkBits := make([]byte, 0, 3)
	for i := 0; i < len(indexList); i += 2 {
		index := indexList[i]
		shift := 4 * ((index + 1) & 1)
		index >>= 1
		high := (((bs[index] >> shift) & 0x0F) ^ maskList[i]) << 4

		index = indexList[i+1]
		shift = 4 * ((index + 1) & 1)
		index >>= 1
		low := (((bs[index] >> shift) & 0x0F) ^ maskList[i+1])
		checkBits = append(checkBits, high|low)
	}
	return raw + base64.URLEncoding.EncodeToString(checkBits)
}

// IsStandardMode return whether it's in standard mode (public cloud)
func IsStandardMode() bool {
	deploymentMode := beego.AppConfig.String("deployment_mode")
	if deploymentMode == StandardDeploymentMode {
		return true
	}
	return false
}
func IsDebug() bool {
	RunMode := beego.AppConfig.String("Run_Mode")
	if RunMode == RunMode_Dev {
		return true
	}
	return false
}

// Int32Ptr returns a pointer to the passed int32.
func Int32Ptr(i int32) *int32 {
	return &i
}
