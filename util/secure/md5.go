/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 */

package secure

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io"
)

func DecodeBase64(str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func MD5Sum(str ...string) string {
	h := md5.New()
	for _, s := range str {
		io.WriteString(h, s)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func EncodeMD5(str string) string {
	h := md5.New()
	io.WriteString(h, str)

	strSum := h.Sum(nil)
	res := hex.EncodeToString(strSum)
	return res
}
