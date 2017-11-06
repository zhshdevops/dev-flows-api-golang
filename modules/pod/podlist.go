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

	v1core "k8s.io/client-go/kubernetes/typed/core/v1" // v1core.CoreInterface
	v1api "k8s.io/client-go/pkg/api/v1"                // v1api.ConfigMapList, v1v1v1api.Pod, v1api.PodPhase

	k8sclient "dev-flows-api-golang/modules/client"
	"dev-flows-api-golang/modules/common"
	"dev-flows-api-golang/modules/dataselect"
	"dev-flows-api-golang/modules/label"

	"k8s.io/apimachinery/pkg/fields"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// ReplicationSetList contains a list of Pods in the cluster.
type PodList struct {
	ListMeta common.ListMeta `json:"listMeta"`

	// Unordered list of Pods.
	Pods []Pod `json:"pods"`
}

// Pod is a presentation layer view of Kubernetes Pod resource. This means
// it is Pod plus additional augumented data we can get from other sources
// (like services that target it).
type Pod struct {
	ObjectMeta common.ObjectMeta `json:"objectMeta"`
	TypeMeta   common.TypeMeta   `json:"typeMeta"`

	// Status of the Pod. See Kubernetes API for reference.
	PodPhase v1api.PodPhase `json:"podPhase"`
	// TenxCloud: Add pod spec, as we'll use the volume mapping info
	Spec v1api.PodSpec `json:"podSpec"`

	// IP address of the Pod.
	PodIP string `json:"podIP"`

	// Count of containers restarts.
	RestartCount int32 `json:"restartCount"`
}

// GetPodList returns a list of all Pods in the cluster.
func GetPodList(client v1core.CoreV1Client, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*PodList, error) {
	glog.V(4).Infof("Getting list of all pods in the cluster\n")

	channels := &common.ResourceChannels{
		PodList: common.GetPodListChannelWithOptions(client, nsQuery,v1.ListOptions{}, 1),
	}

	return GetPodListFromChannels(channels, dsQuery)
}

//  TODO: should change function name

// GetPodListBySelector get pod list by labels and fields, both use "==" operator
func GetPodListBySelector(client v1core.CoreV1Client, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery, labelSet, fieldSet map[string]string) (*PodList, error) {
	glog.V(4).Infof("Getting pod list by LabelSelector and fieldSet in the cluster\n")
	option := &v1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set(labelSet)).String(),
		FieldSelector: fields.SelectorFromSet(fields.Set(fieldSet)).String(),
	}

	return GetPodListByOption(client, nsQuery, dsQuery, option)
}

// GetPodListByOption get pod list by option
func GetPodListByOption(client v1core.CoreV1Client, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery, option *v1.ListOptions) (*PodList, error) {
	channels := &common.ResourceChannels{
		PodList: common.GetPodListChannelWithOptions(client, nsQuery, *option, 1),
	}

	return GetPodListFromChannels(channels, dsQuery)
}

// GetPodListFromChannels returns a list of all Pods in the cluster
// reading required resource list once from the channels.
func GetPodListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery) (*PodList, error) {

	pods := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	podList := CreatePodList(pods.Items, dsQuery)
	return &podList, nil
}

// CreatePodList filters a list of pods using data select query
func CreatePodList(pods []v1api.Pod, dsQuery *dataselect.DataSelectQuery) PodList {
	podList := PodList{
		Pods:     make([]Pod, 0),
		ListMeta: common.ListMeta{TotalItems: len(pods)},
	}

	podCells := dataselect.GenericDataSelect(toCells(pods), dsQuery)
	pods = fromCells(podCells)

	for _, pod := range pods {
		podDetail := ToPod(&pod)
		podList.Pods = append(podList.Pods, podDetail)
	}
	return podList
}

func GetCountByNamespace(k8sList []*k8sclient.ClientSet, namespaceList []string) (int, error) {
	method := "ListByNamespace"
	selector, err := labels.Parse(label.App)
	if err != nil {
		glog.Errorln(method, "Parse selector string failed.", err)
		return 0, err
	}
	cnt := 0
	for _, namespace := range namespaceList {
		for _, k8s := range k8sList {
			podList, err := k8s.CoreV1Client.Pods(namespace).List(v1.ListOptions{
				LabelSelector: selector.String(),
			})
			if err != nil {
				glog.Errorln(method, "get pods by namespace failed.", err)
				return 0, err
			}
			cnt += len(podList.Items)
		}
	}
	return cnt, nil
}
