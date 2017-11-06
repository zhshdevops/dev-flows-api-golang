/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 */

package secure

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var (
	seckey       = []byte("T1X2C3D!F^I$G&H*T(E)R088") //24Byte
	base64Tables = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	iv           = []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	cipherHolder string
	plainHolder  string
)
const   (
	SECRET_KEY = "dazyunsecretkeysforuserstenx20141019generatedKey"
)


func base64Encode(src []byte) string {
	coder := base64.NewEncoding(base64Tables)
	return coder.EncodeToString(src)
}

func base64Decode(src string) ([]byte, error) {
	coder := base64.NewEncoding(base64Tables)
	return coder.DecodeString(src)
}

func aesPadding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func aesUnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}
//encryptContent
func Encrypt(src string) (string, error) {
	block, err := aes.NewCipher([]byte(SECRET_KEY))
	if err != nil {
		return "", nil
	}

	arr := aesPadding([]byte(src), aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(arr, arr)
	return string(arr), nil
}
func Decrypt(src string) (string, error) {
	block, err := aes.NewCipher([]byte(SECRET_KEY))
	if err != nil {
		return "", nil
	}
	arr := []byte(src)
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(arr, arr)
	arr = aesUnPadding(arr)
	return string(arr), nil

}

func EncryptAndBase64(src string) (string, error) {
	block, err := aes.NewCipher(seckey)
	if err != nil {
		return "", nil
	}

	arr := aesPadding([]byte(src), aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(arr, arr)
	return base64Encode(arr), nil
}
func Base64AndDecrypt(src string) (string, error) {
	block, err := aes.NewCipher(seckey)
	if err != nil {
		return "", nil
	}
	arr, _ := base64Decode(src)
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(arr, arr)
	arr = aesUnPadding(arr)
	return string(arr), nil

}