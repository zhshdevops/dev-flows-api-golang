/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-27  @author mengyuan
 */

package label

import (
	"k8s.io/client-go/1.4/pkg/labels"
	"k8s.io/client-go/1.4/pkg/selection"
	"k8s.io/client-go/1.4/pkg/util/sets"
)

type Requirement struct {
	Key      string
	Operator selection.Operator
	Values   []string
}

func NewLabelSelector(rs []*Requirement) (labels.Selector, error) {
	var (
		lsel         = labels.NewSelector()
		requirements = make([]labels.Requirement, 0, len(rs))

		err error
		req *labels.Requirement
	)
	for _, v := range rs {
		req, err = labels.NewRequirement(v.Key, v.Operator, sets.NewString(v.Values...))
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, *req)
	}
	return lsel.Add(requirements...), nil
}
