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
	"github.com/golang/glog"
	v1core "k8s.io/client-go/1.4/kubernetes/typed/core/v1"
	v1api "k8s.io/client-go/1.4/pkg/api/v1" // v1api.ConfigMapList, v1v1v1api.Pod, v1api.PodPhase

	"dev-flows-api-golang/modules/common"

)

// PodDetail is a presentation layer view of Kubernetes PodDetail resource.
// This means it is PodDetail plus additional augumented data we can get
// from other sources (like services that target it).
type PodDetail struct {
	ObjectMeta common.ObjectMeta `json:"objectMeta"`
	TypeMeta   common.TypeMeta   `json:"typeMeta"`

	// Status of the Pod. See Kubernetes API for reference.
	PodPhase v1api.PodPhase `json:"podPhase"`

	// IP address of the Pod.
	PodIP string `json:"podIP"`

	// Name of the Node this Pod runs on.
	NodeName string `json:"nodeName"`

	// Count of containers restarts.
	RestartCount int32 `json:"restartCount"`

	// List of container of this pod.
	Containers []Container `json:"containers"`
}

// Container represents a docker/rkt/etc. container that lives in a pod.
type Container struct {
	// Name of the container.
	Name string `json:"name"`

	// Image URI of the container.
	Image string `json:"image"`

	// List of environment variables.
	Env []EnvVar `json:"env"`

	// Commands of the container
	Commands []string `json:"commands"`

	// Command arguments
	Args []string `json:"args"`
}

// EnvVar represents an environment variable of a container.
type EnvVar struct {
	// Name of the variable.
	Name string `json:"name"`

	// Value of the variable. May be empty if value from is defined.
	Value string `json:"value"`

	// Defined for derived variables. If non-null, the value is get from the reference.
	// Note that this is an API struct. This is intentional, as EnvVarSources are plain struct
	// references.
	ValueFrom *v1api.EnvVarSource `json:"valueFrom"`
}

// GetPodDetail returns the details (PodDetail) of a named Pod from a particular
// namespace.
func GetPodDetail(client v1core.CoreClient, namespace, name string) (*PodDetail, error) {

	glog.V(4).Infof("Getting details of %s pod in %s namespace", name, namespace)

	channels := &common.ResourceChannels{
		ConfigMapList: common.GetConfigMapListChannel(client, common.NewSameNamespaceQuery(namespace), 1),
		// PodMetrics:    common.GetPodMetricsChannel(heapsterClient, name, namespace),   // hide it for now
	}
	pod, err := client.Pods(namespace).Get(name)

	if err != nil {
		return nil, err
	}

	if err = <-channels.ConfigMapList.Error; err != nil {
		return nil, err
	}
	configMapList := <-channels.ConfigMapList.List

	podDetail := toPodDetail(pod, configMapList)
	return &podDetail, nil
}

func toPodDetail(pod *v1api.Pod, configMaps *v1api.ConfigMapList) PodDetail {

	var containers []Container
	for _, container := range pod.Spec.Containers {
		var vars []EnvVar
		for _, envVar := range container.Env {
			variable := EnvVar{
				Name:      envVar.Name,
				Value:     envVar.Value,
				ValueFrom: envVar.ValueFrom,
			}
			if variable.ValueFrom != nil {
				variable.Value = evalValueFrom(variable.ValueFrom, configMaps)
			}
			vars = append(vars, variable)
		}
		containers = append(containers, Container{
			Name:     container.Name,
			Image:    container.Image,
			Env:      vars,
			Commands: container.Command,
			Args:     container.Args,
		})
	}
	podDetail := PodDetail{
		ObjectMeta:   common.NewObjectMeta(pod.ObjectMeta),
		TypeMeta:     common.NewTypeMeta(common.ResourceKindPod),
		PodPhase:     pod.Status.Phase,
		PodIP:        pod.Status.PodIP,
		RestartCount: getRestartCount(*pod),
		NodeName:     pod.Spec.NodeName,
		Containers:   containers,
	}

	return podDetail
}

func evalValueFrom(src *v1api.EnvVarSource, configMaps *v1api.ConfigMapList) string {
	if src.ConfigMapKeyRef != nil {
		name := src.ConfigMapKeyRef.LocalObjectReference.Name

		for _, configMap := range configMaps.Items {
			if configMap.ObjectMeta.Name == name {
				return configMap.Data[src.ConfigMapKeyRef.Key]
			}
		}
	}
	return ""
}
