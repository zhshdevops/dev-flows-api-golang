/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-10-20  @author liuyang
 */

package models

import (
	"encoding/json"

	"github.com/golang/glog"
)

const aesBlockSize = 16 // use aes 128

// ThirdPartyRegistryConfig third party registry config for user preference table
type ThirdPartyRegistryConfig struct {
	Title             string `json:"title"`
	RegistryURL       string `json:"registry_url"`
	AuthURL           string `json:"auth_url"`
	ServiceURL        string `json:"service_url"`
	Username          string `json:"username"`
	EncryptedPassword string `json:"password"` // Client should take care the encryption of the password
}

// NewThirdPartyRegistryConfig return a new config object
func NewThirdPartyRegistryConfig(title, username, encryptedPassword, registryURL string) *ThirdPartyRegistryConfig {
	return &ThirdPartyRegistryConfig{
		Title:             title,
		RegistryURL:       registryURL,
		Username:          username,
		EncryptedPassword: encryptedPassword,
	}
}

// Parse get third party registry config from json string
func (tpr *ThirdPartyRegistryConfig) Parse(configDetail string) error {
	method := "models.NewThirdPartyRegistryConfig"
	err := json.Unmarshal([]byte(configDetail), tpr)
	if err != nil {
		glog.Errorln(method, "failed to parse thire party registry config", err)
		return err
	}
	return nil
}

// String stringer for ThirdPartyRegistryConfig
func (tpr *ThirdPartyRegistryConfig) String() string {
	method := "models/ThirdPartyRegistryConfig.String"
	bytes, err := json.Marshal(tpr)
	if err != nil {
		glog.Errorln(method, "failed to marshal", err)
		return ""
	}
	return string(bytes)
}

/* Remove password encryption for now, as we'll leave it to client side

// ThirdPartyRegistryPassword encrypted third party registry password
type ThirdPartyRegistryPassword string

func (pwd *ThirdPartyRegistryPassword) String() string {
	return string(*pwd)
}

// newThirdPartyRegistryPassword encrypt third party registry password with aes and base64
func newThirdPartyRegistryPassword(token, password string) ThirdPartyRegistryPassword {
	aescipher := getAESCipher(token)

	cipherbytes := make([]byte, aesBlockSize)
	pwdbytes := make([]byte, aesBlockSize)
	copy(pwdbytes, password)

	aescipher.Encrypt(cipherbytes, pwdbytes)

	base64str := base64.StdEncoding.EncodeToString(cipherbytes)
	return ThirdPartyRegistryPassword(base64str)
}

// Decrypt decode registry password with base64 and aes
func (pwd *ThirdPartyRegistryPassword) Decrypt(token string) string {
	method := "models/ThirdPartyRegistryPassword.Decrypt"
	cipherbytes, err := base64.StdEncoding.DecodeString(string(*pwd))
	if err != nil {
		glog.Errorln(method, "failed to decode base64 string", *pwd)
		return ""
	}

	if len(cipherbytes) != aesBlockSize {
		glog.Errorln(method, "length not correct")
		return ""
	}

	aescipher := getAESCipher(token)
	pwdbytes := make([]byte, aesBlockSize)
	aescipher.Decrypt(pwdbytes, cipherbytes)

	return string(pwdbytes)
}

func getAESCipher(token string) cipher.Block {
	key := make([]byte, aesBlockSize)
	copy(key, []byte(token))
	cipher, _ := aes.NewCipher(key)
	return cipher
}
*/
