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

	"k8s.io/apimachinery/pkg/fake"
	v1api "k8s.io/client-go/pkg/api/v1" // v1api.Pod

	"dev-flows-api-golang/modules/common"
)

func TestToPodDetail(t *testing.T) {
	cases := []struct {
		pod      *v1api.PodList
		expected *PodDetail
	}{
		{
			pod: &v1api.PodList{
				Items: []v1api.Pod{
					{
						ObjectMeta: v1api.ObjectMeta{
							Name: "test-pod", Namespace: "test-namespace",
						}},
				},
			},
			expected: &PodDetail{
				TypeMeta: common.TypeMeta{Kind: common.ResourceKindPod},
				ObjectMeta: common.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			},
		},
	}

	for _, c := range cases {
		fakeClient := fake.NewSimpleClientset(&c.pod.Items[0])

		actual, err := GetPodDetail(fakeClient.Core(), "test-namespace", "test-pod")

		if err != nil {
			t.Errorf("GetPodDetail(%#v) == \ngot err %#v", c.pod, err)
		}
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("GetPodDetail(%#v) == \ngot      %#v, \nexpected %#v", c.pod, actual,
				c.expected)
		}
	}
}
