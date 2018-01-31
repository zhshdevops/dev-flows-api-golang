/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-24  @author Zhao Shuailong
 */

package pod

import (
	"github.com/golang/glog"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1" // v1core.CoreInterface
	resource "k8s.io/apimachinery/pkg/api/resource"       // resource.Quantity
	v1api "k8s.io/client-go/pkg/api/v1"                // v1v1api.Namespace, v1v1api.Pod
	"dev-flows-api-golang/modules/common"
	"dev-flows-api-golang/modules/dataselect"
)

// Gets restart count of given pod (total number of its containers restarts).
func getRestartCount(pod v1api.Pod) int32 {
	var restartCount int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		restartCount += containerStatus.RestartCount
	}
	return restartCount
}

// ToPod transforms Kubernetes pod object into object returned by v1api.
func ToPod(pod *v1api.Pod) Pod {
	podDetail := Pod{
		ObjectMeta:   common.NewObjectMeta(pod.ObjectMeta),
		TypeMeta:     common.NewTypeMeta(common.ResourceKindPod),
		PodPhase:     pod.Status.Phase,
		PodIP:        pod.Status.PodIP,
		RestartCount: getRestartCount(*pod),
		// TenxCloud: Add pod spec, as we'll use the volume mapping info
		Spec: pod.Spec,
	}

	return podDetail
}

// GetContainerImages returns container image strings from the given pod spec.
func GetContainerImages(podTemplate *v1api.PodSpec) []string {
	var containerImages []string
	for _, container := range podTemplate.Containers {
		containerImages = append(containerImages, container.Image)
	}
	return containerImages
}

// SetAnnotation set pod annotation
func SetAnnotation(client v1core.CoreV1Client, pod *v1api.Pod, key, value string) error {
	pod.Annotations[key] = value
	_, err := client.Pods(pod.GetNamespace()).Update(pod)
	return err
}

// The code below allows to perform complex data section on []v1api.Pod
// 下面的代码用来支持对 []v1api.Pod 的排序

type PodCell v1api.Pod

func (self PodCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

func toCells(std []v1api.Pod) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = PodCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []v1api.Pod {
	std := make([]v1api.Pod, len(cells))
	for i := range std {
		std[i] = v1api.Pod(cells[i].(PodCell))
	}
	return std
}

// RequestsAndLimits returns a dictionary of all defined resources summed up for all
// containers of the pod.
func RequestsAndLimits(pod *v1api.Pod) (reqs map[v1api.ResourceName]resource.Quantity, limits map[v1api.ResourceName]resource.Quantity, err error) {
	return podRequestsAndLimits(&pod.Spec)
}

func podRequestsAndLimits(spec *v1api.PodSpec) (reqs map[v1api.ResourceName]resource.Quantity, limits map[v1api.ResourceName]resource.Quantity, err error) {
	reqs, limits = map[v1api.ResourceName]resource.Quantity{}, map[v1api.ResourceName]resource.Quantity{}
	for _, container := range spec.Containers {
		for name, quantity := range container.Resources.Requests {
			if value, ok := reqs[name]; !ok {
				reqs[name] = *quantity.Copy()
			} else {
				value.Add(quantity)
				reqs[name] = value
			}
		}
		for name, quantity := range container.Resources.Limits {
			if value, ok := limits[name]; !ok {
				limits[name] = *quantity.Copy()
			} else {
				value.Add(quantity)
				limits[name] = value
			}
		}
	}
	// init containers define the minimum of any resource
	for _, container := range spec.InitContainers {
		for name, quantity := range container.Resources.Requests {
			value, ok := reqs[name]
			if !ok {
				reqs[name] = *quantity.Copy()
				continue
			}
			if quantity.Cmp(value) > 0 {
				reqs[name] = *quantity.Copy()
			}
		}
		for name, quantity := range container.Resources.Limits {
			value, ok := limits[name]
			if !ok {
				limits[name] = *quantity.Copy()
				continue
			}
			if quantity.Cmp(value) > 0 {
				limits[name] = *quantity.Copy()
			}
		}
	}
	return
}

// GetResourceRequested get pod reueested cpu and mem
func GetResourceRequested(spec *v1api.PodSpec) (cpurequest int64, memrequest int64) {
	request, _, err := podRequestsAndLimits(spec)
	if err != nil {
		return
	}
	if cpu, ok := request["cpu"]; ok {
		cpurequest = cpu.ScaledValue(resource.Milli)
	}
	if mem, ok := request["memory"]; ok {
		memrequest = mem.ScaledValue(resource.Kilo)
	}
	return
}

// GetResourceRange get resource request and limit
func GetResourceRange(spec *v1api.PodSpec) (request map[string]int64, limit map[string]int64) {
	req, lim, err := podRequestsAndLimits(spec)
	if err != nil {
		glog.Errorf("podRequestsAndLimits failed: %v", err)
		return
	}
	request, limit = make(map[string]int64, 2), make(map[string]int64, 2)
	if cpu, ok := req[v1api.ResourceCPU]; ok {
		request["cpu"] = cpu.ScaledValue(resource.Milli)
	}
	if mem, ok := req[v1api.ResourceMemory]; ok {
		request["memory"] = mem.ScaledValue(resource.Kilo)
	}
	if cpu, ok := lim[v1api.ResourceCPU]; ok {
		limit["cpu"] = cpu.ScaledValue(resource.Milli)
	}
	if mem, ok := lim[v1api.ResourceMemory]; ok {
		limit["memory"] = mem.ScaledValue(resource.Kilo)
	}
	return
}
