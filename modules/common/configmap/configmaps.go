/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-28 15:12:38  @author huangxin
 */

package configmap

import (
	"k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	unversioned "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/golang/glog"

	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ConfigMapChannel struct {
	ConfigMap chan *apiv1.ConfigMap
	Error     chan error
}

func GetConfigMapChannel(client *kubernetes.Clientset,
	namespace, name string, numReads int) ConfigMapChannel {

	channel := newChannel(numReads)
	go func() {
		configMap, err := client.Core().ConfigMaps(namespace).Get(name)
		writeMulti(channel, configMap, err, numReads)
	}()

	return channel
}

func GetConfigMap(client *kubernetes.Clientset,
	namespace, name string) (*apiv1.ConfigMap, error) {

	return GetConfigMapChannel(client, namespace, name, 1).ReadOnceFromChannel()
}

func UpdateConfigMapChannel(client *kubernetes.Clientset,
	namespace string, configMap *apiv1.ConfigMap, numReads int) ConfigMapChannel {

	channel := newChannel(numReads)
	go func() {
		newConfigMap, err := client.Core().ConfigMaps(namespace).Update(configMap)
		writeMulti(channel, newConfigMap, err, numReads)
	}()

	return channel
}

func UpdateConfigMap(client *kubernetes.Clientset, namespace string,
	configMap *apiv1.ConfigMap) error {

	channel := UpdateConfigMapChannel(client, namespace, configMap, 1)
	<-channel.ConfigMap
	return <-channel.Error
}

func PatchConfigMap(client *kubernetes.Clientset, namespace string,
	configMap *apiv1.ConfigMap) error {
	body, err := json.Marshal(configMap)
	if nil != err {
		glog.Errorf("Failed to encode configmap to json: %v\n", err)
		return err
	}
	if _, err := client.ConfigMaps(namespace).Patch(configMap.Name, api.MergePatchType, body); nil != err {
		glog.Errorf("Failed to patch configmap: %v\n", err)
		return err
	}
	return nil
}

func CreateConfigMapChannel(client *kubernetes.Clientset,
	namespace string, configMap *apiv1.ConfigMap, numReads int) ConfigMapChannel {

	channel := newChannel(numReads)
	go func() {
		newConfigMap, err := client.Core().ConfigMaps(namespace).Create(configMap)
		writeMulti(channel, newConfigMap, err, numReads)
	}()

	return channel
}

func CreateConfigMap(client *kubernetes.Clientset, namespace string,
	configMap *apiv1.ConfigMap) error {

	channel := CreateConfigMapChannel(client, namespace, configMap, 1)
	<-channel.ConfigMap
	return <-channel.Error
}

func DeleteConfigMapChannel(client *kubernetes.Clientset,
	namespace, name string, numReads int) ConfigMapChannel {

	channel := newChannel(numReads)
	go func() {
		err := client.Core().ConfigMaps(namespace).Delete(name, nil)
		writeMulti(channel, nil, err, numReads)
	}()

	return channel
}

func DeleteConfigMaps(client *kubernetes.Clientset, namespace string,
	names []string) (map[string]error, error) {

	channels := make(map[string]ConfigMapChannel)
	//为每个configmap创建删除的channel来获取删除结果
	for _, name := range names {
		//避免创建重复channel
		if _, ok := channels[name]; false == ok {
			glog.V(2).Infof("Configmap %s will be deleted\n", name)
			channels[name] = DeleteConfigMapChannel(client, namespace, name, 1)
		}
	}

	deleted := make(map[string]error)
	timeout := time.After(3 * time.Minute)

	//收集删除结果
	//无限循环遍历names，直到接收到所有删除结果或超时
	for {
		for name, channel := range channels {
			//只接收尚未收到的删除结果
			if _, ok := deleted[name]; false == ok {
				glog.V(3).Infof("Waiting for result of deleting %s\n", name)
				select {
				case <-channel.ConfigMap:
					//接收到删除结果时，将结果保存到deleted中
					deleted[name] = <-channel.Error
					glog.V(3).Infof("Received result of deleting %s\n", name)
				case <-timeout:
					//删除超时，返回错误
					return nil,
						&k8serrors.StatusError{unversioned.Status{
							Status:  unversioned.StatusFailure,
							Code:    http.StatusInternalServerError,
							Reason:  unversioned.StatusReasonServerTimeout,
							Message: fmt.Sprintf("Delete configmaps timeout."),
						}}
				default:
					//未收到结果时，等待10ms后继续循环
					glog.V(3).Infof("No result of deleting %s\n", name)
					time.Sleep(10 * time.Millisecond)
				}
			}
		}
		if len(channels) == len(deleted) {
			//收集完成，返回删除结果
			return deleted, nil
		}
	}
}

func (channel ConfigMapChannel) ReadOnceFromChannel() (*apiv1.ConfigMap, error) {
	return <-channel.ConfigMap, <-channel.Error
}

func newChannel(numReads int) ConfigMapChannel {
	return ConfigMapChannel{
		ConfigMap: make(chan *apiv1.ConfigMap, numReads),
		Error:     make(chan error, numReads),
	}
}

func writeMulti(channel ConfigMapChannel, configMap *apiv1.ConfigMap, err error, numReads int) {
	for i := 0; i < numReads; i++ {
		channel.ConfigMap <- configMap
		channel.Error <- err
	}
}
