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
func Encrypt(src string) (string, error) {
	block, err := aes.NewCipher(seckey)
	if err != nil {
		return "", nil
	}

	arr := aesPadding([]byte(src), aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(arr, arr)
	return string(arr), nil
}
func Decrypt(src string) (string, error) {
	block, err := aes.NewCipher(seckey)
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

// func (this *CipherHandler) AesEnDecode(encode bool, plainText *string, cipherText *string) (err error) {

// 	defer func() {
// 		if r := recover(); r != nil {
// 			err = fmt.Errorf("aesEnDecode Panic is: %v", r)
// 		}
// 	}()

// 	var block cipher.Block

// 	if true == encode {

// 		block, err = aes.NewCipher(seckey)
// 		if err != nil {
// 			return
// 		}
// 		arr := []byte(*plainText)
// 		arr = this.aesPadding(arr, aes.BlockSize)
// 		mode := cipher.NewCBCEncrypter(block, iv)
// 		mode.CryptBlocks(arr, arr)
// 		*cipherText = this.base64Encode(arr)

// 	} else {

// 		if strings.EqualFold(cipherHolder, *cipherText) {
// 			*plainText = plainHolder
// 		} else {
// 			block, err = aes.NewCipher(seckey)
// 			if err != nil {
// 				return
// 			}
// 			arr, _ := this.base64Decode(*cipherText)
// 			mode := cipher.NewCBCDecrypter(block, iv)
// 			mode.CryptBlocks(arr, arr)
// 			arr = this.aesUnPadding(arr)
// 			*plainText = string(arr)
// 			plainHolder = *plainText
// 			cipherHolder = *cipherText
// 		}
// 	}

// 	return err
// }

// func (this *CipherHandler) DecodeData(serialStr *string, index string) (value string, err error) {

// 	method := "CipherHandler.GetDecodeData"

// 	var plainText string
// 	//var licenseData LicenseDataJson
// 	var licenseDataMap map[string] string

// 	err = this.aesEnDecode(false, &plainText, serialStr)
// 	if nil != err {
// 		return "", err
// 	}

// 	if err = json.Unmarshal([]byte(plainText), &licenseDataMap); nil == err {
// 		if "" == index {
// 			beego.Debug(method, " Decode Data --- ", plainText)
// 			return plainText, err
// 		} else {
// 			beego.Debug(method, " Decode Data --- ", licenseDataMap[index])
// 			return licenseDataMap[index], nil
// 		}
// 	} else {
// 		return "", err
// 	}
// }
// func (this *CipherHandler) GetDecodeData(serialStr *string, index string) (value string, err error) {

// 	method := "CipherHandler.GetDecodeData"

// 	var plainText string
// 	//var licenseData LicenseDataJson
// 	var licenseDataMap map[string] string

// 	err = this.aesEnDecode(false, &plainText, serialStr)
// 	if nil != err {
// 		return "", err
// 	}

// 	if err = json.Unmarshal([]byte(plainText), &licenseDataMap); nil == err {
// 		if "" == index {
// 			beego.Debug(method, " Decode Data --- ", plainText)
// 			return plainText, err
// 		} else {
// 			beego.Debug(method, " Decode Data --- ", licenseDataMap[index])
// 			return licenseDataMap[index], nil
// 		}
// 	} else {
// 		return "", err
// 	}
// }
// // Get limitation from license of specified item
// func (this *CipherHandler) GetLegalLimitValue(serialStr *string, limitItem string) (limit int) {

// 	method := "CipherHandler.GetLegalLimitValue"

// 	limit, ok := this.defaultLimitItems[limitItem]
// 	if !ok || "" == *serialStr {
// 		return 0
// 	}

// 	plainStr, err := this.GetDecodeData(serialStr, limitItem)
// 	if "" != plainStr && nil == err {
// 		limit, err = strconv.Atoi(plainStr)
// 	}

// 	beego.Debug(method, "Limit Item --- ", limitItem, " Value --- ", limit)

// 	return limit
// }

// // Get legal size of specified item. Return 'objLen' if it is less than limitation in license, or return limitation
// func (this *CipherHandler) GetLegalSizeValue(objLen int, serialStr *string, limitItem string) (limit int) {

// 	method := "CipherHandler.GetLegalSizeValue"

// 	limit, ok := this.defaultLimitItems[limitItem]
// 	if !ok || "" == *serialStr {
// 		return 0
// 	}

// 	plainStr, err := this.GetDecodeData(serialStr, limitItem)
// 	if "" != plainStr && nil == err {
// 		limit, err = strconv.Atoi(plainStr)
// 	}

// 	beego.Debug(method, "Limit Item --- ", limitItem, " Value --- ", limit)

// 	if limit > objLen {
// 		limit = objLen
// 	}

// 	return limit
// }
