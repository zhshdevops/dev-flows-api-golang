/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-22  @author huangxin
 */
package deployment

import (
	"k8s.io/client-go/kubernetes"
	v1beta1extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// 可能存在类似于resourcechannels.go中的场景，为了方便封装类似的channels，故此处提供DeploymentChannel
type DeploymentChannel struct {
	Deployment chan *v1beta1extensions.Deployment
	Error      chan error
}

// 获取get方法channel
func GetDeploymentChannel(client *kubernetes.Clientset,
	namespace, name string, numReads int) DeploymentChannel {

	channel := newChannel(numReads)
	go func() {
		deployment, err := client.Extensions().Deployments(namespace).Get(name)
		writeMulti(channel, deployment, err, numReads)
	}()

	return channel
}

// 获取update方法channel
func UpdateDeploymentChannel(client *kubernetes.Clientset, namespace string,
	deployment *v1beta1extensions.Deployment, numReads int) DeploymentChannel {

	channel := newChannel(numReads)
	go func() {
		deployment, err := client.Extensions().Deployments(namespace).Update(deployment)
		writeMulti(channel, deployment, err, numReads)
	}()
	return channel
}

// 获取deployment
func GetDeployment(client *kubernetes.Clientset,
	namespace, name string) (*v1beta1extensions.Deployment, error) {

	return GetDeploymentChannel(client, namespace, name, 1).ReadOnceFromChannel()
}

// 更新deployement
func UpdateDeployment(client *kubernetes.Clientset, namespace string,
	deployment *v1beta1extensions.Deployment) (*v1beta1extensions.Deployment, error) {

	return UpdateDeploymentChannel(client, namespace, deployment, 1).ReadOnceFromChannel()
}

func (channel DeploymentChannel) ReadOnceFromChannel() (*v1beta1extensions.Deployment, error) {
	return <-channel.Deployment, <-channel.Error
}

func newChannel(numReads int) DeploymentChannel {
	return DeploymentChannel{
		Deployment: make(chan *v1beta1extensions.Deployment, numReads),
		Error:      make(chan error, numReads),
	}
}

func writeMulti(channel DeploymentChannel, deployment *v1beta1extensions.Deployment, err error, numReads int) {
	for i := 0; i < numReads; i++ {
		channel.Deployment <- deployment
		channel.Error <- err
	}
}
