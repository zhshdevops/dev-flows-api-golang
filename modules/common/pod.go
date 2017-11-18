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
	unversioned "k8s.io/client-go/1.4/pkg/api/unversioned"
	v1api "k8s.io/client-go/1.4/pkg/api/v1"
)

// FilterNamespacedPodsBySelector returns pods targeted by given resource label selector in given
// namespace.
func FilterNamespacedPodsBySelector(pods []v1api.Pod, namespace string,
	resourceSelector map[string]string) []v1api.Pod {

	var matchingPods []v1api.Pod
	for _, pod := range pods {
		if pod.ObjectMeta.Namespace == namespace &&
			IsSelectorMatching(resourceSelector, pod.Labels) {
			matchingPods = append(matchingPods, pod)
		}
	}

	return matchingPods
}

// FilterPodsBySelector returns pods targeted by given resource selector.
func FilterPodsBySelector(pods []v1api.Pod, resourceSelector map[string]string) []v1api.Pod {

	var matchingPods []v1api.Pod
	for _, pod := range pods {
		if IsSelectorMatching(resourceSelector, pod.Labels) {
			matchingPods = append(matchingPods, pod)
		}
	}
	return matchingPods
}

// FilterNamespacedPodsByLabelSelector returns pods targeted by given resource label selector in
// given namespace.
func FilterNamespacedPodsByLabelSelector(pods []v1api.Pod, namespace string,
	labelSelector *unversioned.LabelSelector) []v1api.Pod {

	var matchingPods []v1api.Pod
	for _, pod := range pods {
		if pod.ObjectMeta.Namespace == namespace &&
			IsLabelSelectorMatching(pod.Labels, labelSelector) {
			matchingPods = append(matchingPods, pod)
		}
	}
	return matchingPods
}

// FilterPodsByLabelSelector returns pods targeted by given resource label selector.
func FilterPodsByLabelSelector(pods []v1api.Pod, labelSelector *unversioned.LabelSelector) []v1api.Pod {

	var matchingPods []v1api.Pod
	for _, pod := range pods {
		if IsLabelSelectorMatching(pod.Labels, labelSelector) {
			matchingPods = append(matchingPods, pod)
		}
	}
	return matchingPods
}

// GetContainerImages returns container image strings from the given pod spec.
func GetContainerImages(podTemplate *v1api.PodSpec) []string {
	var containerImages []string
	for _, container := range podTemplate.Containers {
		containerImages = append(containerImages, container.Image)
	}
	return containerImages
}
