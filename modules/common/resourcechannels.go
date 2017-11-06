// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	v1api "k8s.io/client-go/pkg/api/v1"
	v1alpha1apps "k8s.io/client-go/pkg/apis/apps/v1beta1"
	v1batch "k8s.io/client-go/pkg/apis/batch/v1"
	v1beta1extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceChannels struct holds channels to resource lists. Each list channel is paired with
// an error channel which *must* be read sequentially: first read the list channel and then
// the error channel.
//
// This struct can be used when there are multiple clients that want to process, e.g., a
// list of pods. With this helper, the list can be read only once from the backend and
// distributed asynchronously to clients that need it.
//
// When a channel is nil, it means that no resource list is available for getting.
//
// Each channel pair can be read up to N times. N is specified upon creation of the channels.
type ResourceChannels struct {
	// List and error channels to Replication Controllers.
	ReplicationControllerList ReplicationControllerListChannel

	// List and error channels to Replica Sets.
	ReplicaSetList ReplicaSetListChannel

	// List and error channels to Deployments.
	DeploymentList DeploymentListChannel

	// List and error channels to Daemon Sets.
	DaemonSetList DaemonSetListChannel

	// List and error channels to Jobs.
	JobList JobListChannel

	// List and error channels to Services.
	ServiceList ServiceListChannel

	// List and error channels to Pods.
	PodList PodListChannel

	// List and error channels to Events.
	EventList EventListChannel

	// List and error channels to Nodes.
	NodeList NodeListChannel

	// List and error channels to Namespaces.
	NamespaceList NamespaceListChannel

	// List and error channels to PetSets.
	PetSetList PetSetListChannel

	// List and error channels to PetSets.
	ConfigMapList ConfigMapListChannel

	// List and error channels to Secrets.
	SecretList SecretListChannel

	// // List and error channels to PodMetrics.
	// PodMetrics PodMetricsChannel

	// List and error channels to PersistentVolumes
	PersistentVolumeList PersistentVolumeListChannel

	// List and error channels to PersistentVolumeClaims
	PersistentVolumeClaimList PersistentVolumeClaimListChannel
}
// SecretListChannel is a list and error channels to Secrets.
type SecretListChannel struct {
	List  chan *v1api.SecretList
	Error chan error
}

// GetSecretListChannel returns a pair of channels to a Secret list and errors that
// both must be read numReads times.
func GetSecretListChannel(client *kubernetes.Clientset, nsQuery *NamespaceQuery, numReads int) SecretListChannel {

	channel := SecretListChannel{
		List:  make(chan *v1api.SecretList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.CoreV1Client.Secrets(nsQuery.ToRequestParam()).List(listEverything)
		var filteredItems []v1api.Secret
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}


// ServiceListChannel is a list and error channels to Services.
type ServiceListChannel struct {
	List  chan *v1api.ServiceList
	Error chan error
}

// GetServiceListChannel returns a pair of channels to a Service list and errors that both
// must be read numReads times.
func GetServiceListChannel(client *kubernetes.Clientset, nsQuery *NamespaceQuery, numReads int) ServiceListChannel {

	channel := ServiceListChannel{
		List:  make(chan *v1api.ServiceList, numReads),
		Error: make(chan error, numReads),
	}
	go func() {
		list, err := client.CoreV1Client.Services(nsQuery.ToRequestParam()).List(listEverything)
		var filteredItems []v1api.Service
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// NodeListChannel is a list and error channels to Nodes.
type NodeListChannel struct {
	List  chan *v1api.NodeList
	Error chan error
}

// GetNodeListChannel returns a pair of channels to a Node list and errors that both must be read
// numReads times.
func GetNodeListChannel(client *kubernetes.Clientset, numReads int) NodeListChannel {
	channel := NodeListChannel{
		List:  make(chan *v1api.NodeList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.CoreV1Client.Nodes().List(listEverything)
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// NamespaceListChannel is a list and error channels to Namespaces.
type NamespaceListChannel struct {
	List  chan *v1api.NamespaceList
	Error chan error
}

// GetNamespaceListChannel returns a pair of channels to a Namespace list and errors that both must be read
// numReads times.
func GetNamespaceListChannel(client *kubernetes.Clientset, numReads int) NamespaceListChannel {
	channel := NamespaceListChannel{
		List:  make(chan *v1api.NamespaceList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.CoreV1Client.Namespaces().List(listEverything)
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// EventListChannel is a list and error channels to Nodes.
type EventListChannel struct {
	List  chan *v1api.EventList
	Error chan error
}

// GetEventListChannel returns a pair of channels to an Event list and errors that both must be read
// numReads times.
func GetEventListChannel(client v1core.CoreV1Client,
	nsQuery *NamespaceQuery, numReads int) EventListChannel {
	return GetEventListChannelWithOptions(client, nsQuery, listEverything, numReads)
}

// GetEventListChannelWithOptions is GetEventListChannel plus list options.
func GetEventListChannelWithOptions(client v1core.CoreV1Client,
	nsQuery *NamespaceQuery, options v1.ListOptions, numReads int) EventListChannel {
	channel := EventListChannel{
		List:  make(chan *v1api.EventList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.Events(nsQuery.ToRequestParam()).List(options)
		var filteredItems []v1api.Event
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// PodListChannel is a list and error channels to Nodes.
type PodListChannel struct {
	List  chan *v1api.PodList
	Error chan error
}

// GetPodListChannel returns a pair of channels to a Pod list and errors that both must be read
// numReads times.
func GetPodListChannel(client v1core.CoreV1Client,
	nsQuery *NamespaceQuery, numReads int) PodListChannel {
	return GetPodListChannelWithOptions(client, nsQuery, listEverything, numReads)
}

// GetPodListChannelWithOptions is GetPodListChannel plus listing options.
func GetPodListChannelWithOptions(client v1core.CoreV1Client, nsQuery *NamespaceQuery,
	options v1.ListOptions, numReads int) PodListChannel {

	channel := PodListChannel{
		List:  make(chan *v1api.PodList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.Pods(nsQuery.ToRequestParam()).List(options)
		var filteredItems []v1api.Pod
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// ReplicationControllerListChannel is a list and error channels to Nodes.
type ReplicationControllerListChannel struct {
	List  chan *v1api.ReplicationControllerList
	Error chan error
}

// GetReplicationControllerListChannel Returns a pair of channels to a
// Replication Controller list and errors that both must be read
// numReads times.
func GetReplicationControllerListChannel(client *kubernetes.Clientset,
	nsQuery *NamespaceQuery, numReads int) ReplicationControllerListChannel {

	channel := ReplicationControllerListChannel{
		List:  make(chan *v1api.ReplicationControllerList, numReads),
		Error: make(chan error, numReads),
	}


	go func() {
		list, err := client.CoreV1Client.ReplicationControllers(nsQuery.ToRequestParam()).List(listEverything)
		var filteredItems []v1api.ReplicationController
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// DeploymentListChannel is a list and error channels to Deployments.
type DeploymentListChannel struct {
	List  chan *v1beta1extensions.DeploymentList
	Error chan error
}

// GetDeploymentListChannel returns a pair of channels to a Deployment list and errors
// that both must be read numReads times.
func GetDeploymentListChannel(client *kubernetes.Clientset,
	nsQuery *NamespaceQuery, numReads int) DeploymentListChannel {

	channel := DeploymentListChannel{
		List:  make(chan *v1beta1extensions.DeploymentList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.ExtensionsV1beta1Client.Deployments(nsQuery.ToRequestParam()).List(listEverything)
		var filteredItems []v1beta1extensions.Deployment
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// ReplicaSetListChannel is a list and error channels to Replica Sets.
type ReplicaSetListChannel struct {
	List  chan *v1beta1extensions.ReplicaSetList
	Error chan error
}

// GetReplicaSetListChannel returns a pair of channels to a ReplicaSet list and
// errors that both must be read numReads times.
func GetReplicaSetListChannel(client *kubernetes.Clientset,
	nsQuery *NamespaceQuery, numReads int) ReplicaSetListChannel {
	return GetReplicaSetListChannelWithOptions(client, nsQuery, listEverything, numReads)
}

// GetReplicaSetListChannelWithOptions returns a pair of channels to a ReplicaSet list filtered
// by provided options and errors that both must be read numReads times.
func GetReplicaSetListChannelWithOptions(client *kubernetes.Clientset,
	nsQuery *NamespaceQuery, options v1.ListOptions, numReads int) ReplicaSetListChannel {
	channel := ReplicaSetListChannel{
		List:  make(chan *v1beta1extensions.ReplicaSetList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.ExtensionsV1beta1Client.ReplicaSets(nsQuery.ToRequestParam()).List(options)
		var filteredItems []v1beta1extensions.ReplicaSet
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// DaemonSetListChannel is a list and error channels to Nodes.
type DaemonSetListChannel struct {
	List  chan *v1beta1extensions.DaemonSetList
	Error chan error
}

// GetDaemonSetListChannel returns a pair of channels to a DaemonSet list and errors that
// both must be read numReads times.
func GetDaemonSetListChannel(client *kubernetes.Clientset,
	nsQuery *NamespaceQuery, numReads int) DaemonSetListChannel {
	channel := DaemonSetListChannel{
		List:  make(chan *v1beta1extensions.DaemonSetList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.ExtensionsV1beta1Client.DaemonSets(nsQuery.ToRequestParam()).List(listEverything)
		var filteredItems []v1beta1extensions.DaemonSet
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// JobListChannel is a list and error channels to Nodes.
type JobListChannel struct {
	List  chan *v1batch.JobList // use the Job in batch instead of extensions
	Error chan error
}

// GetJobListChannel returns a pair of channels to a Job list and errors that
// both must be read numReads times.
func GetJobListChannel(client *kubernetes.Clientset,
	nsQuery *NamespaceQuery, numReads int) JobListChannel {
	channel := JobListChannel{
		List:  make(chan *v1batch.JobList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.BatchV1Client.Jobs(nsQuery.ToRequestParam()).List(listEverything)
		var filteredItems []v1batch.Job
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// PetSetListChannel is a list and error channels to Nodes.
type PetSetListChannel struct {
	List  chan *v1alpha1apps.StatefulSetList
	Error chan error
}

// GetPetSetListChannel returns a pair of channels to a PetSet list and errors that
// both must be read numReads times.
func GetPetSetListChannel(client *kubernetes.Clientset,
	nsQuery *NamespaceQuery, numReads int) PetSetListChannel {
	return GetPetSetListChannelWithOptions(client, nsQuery, listEverything, numReads)

}

// GetPetSetListChannelWithOptions is GetPetSetListChannel plus listing options.
func GetPetSetListChannelWithOptions(client *kubernetes.Clientset, nsQuery *NamespaceQuery,
	options v1.ListOptions, numReads int) PetSetListChannel {
	channel := PetSetListChannel{
		List:  make(chan *v1alpha1apps.StatefulSetList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		petSets, err := client.StatefulSets(nsQuery.ToRequestParam()).List(options)
		var filteredItems []v1alpha1apps.StatefulSet
		for _, item := range petSets.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		petSets.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- petSets
			channel.Error <- err
		}
	}()

	return channel
}

// ConfigMapListChannel is a list and error channels to ConfigMaps.
type ConfigMapListChannel struct {
	List  chan *v1api.ConfigMapList
	Error chan error
}

// GetConfigMapListChannel returns a pair of channels to a ConfigMap list and errors that
// both must be read numReads times.
func GetConfigMapListChannel(client v1core.CoreV1Client, nsQuery *NamespaceQuery, numReads int) ConfigMapListChannel {

	channel := ConfigMapListChannel{
		List:  make(chan *v1api.ConfigMapList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.ConfigMaps(nsQuery.ToRequestParam()).List(listEverything)
		var filteredItems []v1api.ConfigMap
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// PersistentVolumeListChannel is a list and error channels to PersistentVolumes.
type PersistentVolumeListChannel struct {
	List  chan *v1api.PersistentVolumeList
	Error chan error
}

// GetPersistentVolumeListChannel returns a pair of channels to a PersistentVolume list and errors that
// both must be read numReads times.
func GetPersistentVolumeListChannel(client *kubernetes.Clientset, numReads int) PersistentVolumeListChannel {
	channel := PersistentVolumeListChannel{
		List:  make(chan *v1api.PersistentVolumeList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.CoreV1Client.PersistentVolumes().List(listEverything)
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel
}

// PersistentVolumeClaimListChannel is a list and error channels to PersistentVolumeClaims.
type PersistentVolumeClaimListChannel struct {
	List  chan *v1api.PersistentVolumeClaimList
	Error chan error
}

// GetPersistentVolumeClaimListChannel returns a pair of channels to a PersistentVolumeClaim list and errors that
// both must be read numReads times.
func GetPersistentVolumeClaimListChannel(client *kubernetes.Clientset, nsQuery *NamespaceQuery,
	numReads int) PersistentVolumeClaimListChannel {

	return GetPersistentVolumeClaimListChannelWithOptions(client, nsQuery, listEverything, numReads)
}

// GetPetSetListChannelWithOptions is GetPetSetListChannel plus listing options.
func GetPersistentVolumeClaimListChannelWithOptions(client *kubernetes.Clientset, nsQuery *NamespaceQuery,
	options v1.ListOptions, numReads int) PersistentVolumeClaimListChannel {
	channel := PersistentVolumeClaimListChannel{
		List:  make(chan *v1api.PersistentVolumeClaimList, numReads),
		Error: make(chan error, numReads),
	}

	go func() {
		list, err := client.CoreV1Client.PersistentVolumeClaims(nsQuery.ToRequestParam()).List(options)
		var filteredItems []v1api.PersistentVolumeClaim
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems
		for i := 0; i < numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()

	return channel

}
var listEverything = v1.ListOptions{
	LabelSelector: labels.Everything().String(),
	FieldSelector: fields.Everything().String(),
}


