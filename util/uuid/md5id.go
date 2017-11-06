/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-13  @author yangle
 */

package uuid

import (
	"dev-flows-api-golang/util/secure"
	"fmt"
)

const (
	trunclen = 12
)

// newShortID generate short uuid and add type prefix
func generateFromString(str string, length int) string {
	md5id := secure.EncodeMD5(str)
	return md5id[:length]
}

// newLongID generate long uuid and add type prefix
func combinedId(prefix, content string) string {
	id := generateFromString(content, trunclen)
	return fmt.Sprintf("%s%s", prefix, id)
}

func NewMd5ClusterId(content string) string {
	return combinedId(clusterID, content)
}

func CombineClusterUUId(apiurl, apitoken string) string {
	word := fmt.Sprintf("%s-%s", apiurl, apitoken)
	return NewMd5ClusterId(word)
}
