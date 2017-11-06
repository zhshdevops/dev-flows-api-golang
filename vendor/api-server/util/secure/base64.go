/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-06
 */

package secure

import (
	"encoding/base64"
)

const (
	base64Table = "123QRSTUabcdVWXYZHijKLAWDCABDstEFGuvwxyzGHIJklmnopqr234560178912"
)

var coder = base64.NewEncoding(base64Table)

func Base64Encode(src []byte) []byte {
	return []byte(coder.EncodeToString(src))
}

func Base64Decode(src []byte) ([]byte, error) {
	return coder.DecodeString(string(src))
}

func Base64StdDecode(src string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(src)
}

func Base64StdEncode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}
