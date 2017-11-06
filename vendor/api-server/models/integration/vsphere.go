/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-22  @author liuyang
 */

package integration

import (
	"encoding/json"
	"time"

	shortid "api-server/modules/tenx/id"
)

// VSphere vsphere config
type VSphere struct {
	ID          string `json:"id,omitempty"` // for output
	Name        string `json:"name"`
	URL         string `json:"url"`
	Username    string `json:"username"`
	Password    string `json:"password,omitempty"` // set to empty when return to user
	Description string `json:"description"`
}

func (v *VSphere) String() (string, error) {
	bytes, err := json.Marshal(v)
	return string(bytes), err
}

// NewVSphere get vsphere config from config
func NewVSphere(config string) (*VSphere, error) {
	v := &VSphere{}
	err := json.Unmarshal([]byte(config), v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetRecord convert vsphere config to tenx_integration record to insert to db
func (v *VSphere) GetRecord(username, namespace string) (*Record, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return &Record{
		ID:         shortid.NewIntegration(),
		Name:       v.Name,
		Type:       RecordTypeVSphere,
		Username:   username,
		Namespace:  namespace,
		Config:     string(bytes),
		CreateTime: time.Now(),
	}, nil
}
