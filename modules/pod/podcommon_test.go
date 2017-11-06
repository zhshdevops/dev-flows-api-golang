/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-24  @author Zhao Shuailong
 */

/*
 *  the original version of code is from github.com/kubernetes/dashbaord
 */

package pod

import (
	"reflect"
	"testing"

	v1api "k8s.io/client-go/pkg/api/v1" // v1v1api.Namespace, v1v1api.Pod

	"dev-flows-api-golang/modules/common"
)

func TestGetPodDetail(t *testing.T) {
	cases := []struct {
		pod      *v1api.Pod
		expected Pod
	}{
		{
			pod: &v1api.Pod{}, expected: Pod{
				TypeMeta: common.TypeMeta{Kind: common.ResourceKindPod},
			},
		}, {
			pod: &v1api.Pod{
				ObjectMeta: v1api.ObjectMeta{
					Name: "test-pod", Namespace: "test-namespace",
				}},
			expected: Pod{
				TypeMeta: common.TypeMeta{Kind: common.ResourceKindPod},
				ObjectMeta: common.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			},
		},
	}

	for _, c := range cases {
		actual := ToPod(c.pod)

		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("ToPod(%#v) == \ngot %#v, \nexpected %#v", c.pod, actual,
				c.expected)
		}
	}
}
