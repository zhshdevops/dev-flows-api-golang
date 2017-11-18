/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-10-26  @author zhao shuailong
 */

/*
 *  the original version of code is from github.com/kubernetes/dashboard
 */

package common

import (
	"k8s.io/client-go/1.4/pkg/api/unversioned"
	v1api "k8s.io/client-go/1.4/pkg/api/v1"
)

// ObjectMeta is metadata about an instance of a resource.
// 任意 kubernetes 对象 都共有的属性
type ObjectMeta struct {
	// Name is unique within a namespace. Name is primarily intended for creation
	// idempotence and configuration definition.
	Name string `json:"name,omitempty"`

	// Namespace defines the space within which name must be unique. An empty namespace is
	// equivalent to the "default" namespace, but "default" is the canonical representation.
	// Not all objects are required to be scoped to a namespace - the value of this field for
	// those objects will be empty.
	Namespace string `json:"namespace,omitempty"`

	// Labels are key value pairs that may be used to scope and select individual resources.
	// Label keys are of the form:
	//     label-key ::= prefixed-name | name
	//     prefixed-name ::= prefix '/' name
	//     prefix ::= DNS_SUBDOMAIN
	//     name ::= DNS_LABEL
	// The prefix is optional.  If the prefix is not specified, the key is assumed to be private
	// to the user.  Other system components that wish to use labels must specify a prefix.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are unstructured key value data stored with a resource that may be set by
	// external tooling. They are not queryable and should be preserved when modifying
	// objects.  Annotation keys have the same formatting restrictions as Label keys. See the
	// comments on Labels for details.
	Annotations map[string]string `json:"annotations,omitempty"`

	// CreationTimestamp is a timestamp representing the server time when this object was
	// created. It is not guaranteed to be set in happens-before order across separate operations.
	// Clients may not set this value. It is represented in RFC3339 form and is in UTC.
	CreationTimestamp unversioned.Time `json:"creationTimestamp,omitempty"`
}

// TypeMeta describes an individual object in an API response or request with strings representing
// the type of the object.
// Kubenretes 中支持的所有资源类型
type TypeMeta struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// In smalllettercase.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#types-kinds
	Kind ResourceKind `json:"kind,omitempty"`
}

// ListMeta describes list of objects, i.e. holds information about pagination options set for the
// list.
// 获取对象列表时，用来存储分页相关的信息（目前还不支持分页）
type ListMeta struct {
	// Total number of items on the list. Used for pagination.
	TotalItems int `json:"total"`
}

// NewObjectMeta returns internal endpoint name for the given service properties, e.g.,
// NewObjectMeta creates a new instance of ObjectMeta struct based on K8s object meta.
func NewObjectMeta(k8SObjectMeta v1api.ObjectMeta) ObjectMeta {
	return ObjectMeta{
		Name:              k8SObjectMeta.Name,
		Namespace:         k8SObjectMeta.Namespace,
		Labels:            k8SObjectMeta.Labels,
		CreationTimestamp: k8SObjectMeta.CreationTimestamp,
		Annotations:       k8SObjectMeta.Annotations,
	}
}

// NewTypeMeta creates new type mete for the resource kind.
func NewTypeMeta(kind ResourceKind) TypeMeta {
	return TypeMeta{
		Kind: kind,
	}
}

// ResourceKind is an unique name for each resource. It can used for API discovery and generic
// code that does things based on the kind. For example, there may be a generic "deleter"
// that based on resource kind, name and namespace deletes it.
type ResourceKind string

// List of all resource kinds supported by the UI.
const (
	ResourceKindReplicaSet              = "ReplicaSet"
	ResourceKindService                 = "Service"
	ResourceKindIngress                 = "Ingress"
	ResourceKindDeployment              = "Deployment"
	ResourceKindPod                     = "Pod"
	ResourceKindEvent                   = "Event"
	ResourceKindReplicationController   = "ReplicationController"
	ResourceKindDaemonSet               = "DaemonSet"
	ResourceKindJob                     = "Job"
	ResourceKindPetSet                  = "PetSet"
	ResourceKindNamespace               = "Namespace"
	ResourceKindNode                    = "Node"
	ResourceKindSecret                  = "Secret"
	ResourceKindConfigMap               = "ConfigMap"
	ResourceKindPersistentVolume        = "PersistentVolume"
	ResourceKindPersistentVolumeClaim   = "PersistentVolumeClaim"
	ResourceKindHorizontalPodAutoscaler = "HorizontalPodAutoscaler"
)

// IsSelectorMatching returns true when an object with the given
// selector targets the same Resources (or subset) that
// the tested object with the given selector.
func IsSelectorMatching(labelSelector map[string]string,
	testedObjectLabels map[string]string) bool {

	// If service has no selectors, then assume it targets different Resource.
	if len(labelSelector) == 0 {
		return false
	}
	for label, value := range labelSelector {
		if rsValue, ok := testedObjectLabels[label]; !ok || rsValue != value {
			return false
		}
	}
	return true
}

// IsLabelSelectorMatching returns true when a resource with the given selector targets the same
// Resources(or subset) that a tested object selector with the given selector.
func IsLabelSelectorMatching(selector map[string]string,
	labelSelector *unversioned.LabelSelector) bool {

	// If the resource has no selectors, then assume it targets different Pods.
	if len(selector) == 0 {
		return false
	}
	for label, value := range selector {
		if rsValue, ok := labelSelector.MatchLabels[label]; !ok || rsValue != value {
			return false
		}
	}
	return true
}
