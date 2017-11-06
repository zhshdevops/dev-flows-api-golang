/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 */

package common

import (
	"dev-flows-api-golang/modules/label"

	"github.com/golang/glog"

	"fmt"

	"k8s.io/client-go/pkg/api"
	"k8s.io/apimachinery/pkg/fields"
)

// if fsel is str, should follow this form:
// <field><operator><value>,<filed><operator><value>,....
// demo:
// status.phase!=Succeeded,status.phase!=Failed
func NewOption(lsel []*label.Requirement, fsel interface{}) (*api.ListOptions, error) {
	labelSelector, err := label.NewLabelSelector(lsel)
	if err != nil {
		return nil, err
	}
	if fsel == nil {
		return &api.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: fields.Everything(),
		}, nil
	}
	var fieldSelector fields.Selector
	switch v := fsel.(type) {
	case fields.Selector:
		fieldSelector = v
	case map[string]string:
		fieldSelector = fields.SelectorFromSet(fields.Set(v))
	case fields.Set:
		fieldSelector = fields.SelectorFromSet(v)
	case string:
		glog.V(4).Infof("parsing %q to feild selector", v)
		fieldSelector, err = fields.ParseSelector(v)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupport type %T", v)
	}

	opt := &api.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	}
	return opt, nil
}
