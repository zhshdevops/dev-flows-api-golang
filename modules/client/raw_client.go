package client

import (
	"fmt"
	"io"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

)

// NewK8sClientSet creates a set of new Kubernetes Apiserver clients. When apiserverHost param is empty
// string the function assumes that it is running inside a Kubernetes cluster and attempts to
// discover the Apiserver. Otherwise, it connects to the Apiserver specified.
//
// apiserverHost param is in the format of protocol://address:port/pathPrefix, e.g.,
// http://localhost:8001.
func NewK8sClientSet(clusterName, apiserverProtocol, apiserverHost, apiserverToken, apiVersion string) (*kubernetes.Clientset, error) {
	cfg, err := NewConfig(clusterName, apiserverProtocol, apiserverHost, apiserverToken, apiVersion)

	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func NewConfig(clusterName, apiserverProtocol, apiserverHost, apiserverToken, apiVersion string) (*rest.Config, error) {
	config := clientcmdapi.NewConfig()
	config.Clusters[clusterName] = &clientcmdapi.Cluster{Server: fmt.Sprintf("%s://%s", apiserverProtocol, apiserverHost), InsecureSkipTLSVerify: true}
	config.AuthInfos[clusterName] = &clientcmdapi.AuthInfo{Token: apiserverToken}
	config.Contexts[clusterName] = &clientcmdapi.Context{
		Cluster:  clusterName,
		AuthInfo: clusterName,
	}
	config.CurrentContext = clusterName

	clientBuilder := clientcmd.NewNonInteractiveClientConfig(*config, clusterName, &clientcmd.ConfigOverrides{}, nil)

	return clientBuilder.ClientConfig()
}

type ClientSet struct {
	*kubernetes.Clientset
}

func NewClientSet(clusterName, apiserverProtocol, apiserverHost, apiserverToken, apiVersion string) (c *ClientSet, err error) {
	cs, err := NewK8sClientSet(clusterName, apiserverProtocol, apiserverHost, apiserverToken, apiVersion)
	if nil != err {
		return
	}
	c = &ClientSet{Clientset: cs}
	return
}

var timeout int64 = 518400

//Stream
type Stream struct {
	cs    *ClientSet
	k8sNs string
}

//GetStreamApi
func GetStreamApi(cs *ClientSet, namespace string) *Stream {
	return &Stream{
		cs:    cs,
		k8sNs: namespace,
	}
}

//WatchResource
func (s *Stream) WatchResource(resourceType string) (watch.Interface, error) {
	options := metav1.ListOptions{
		Watch:          true,
		TimeoutSeconds: &timeout,
	}
	var result watch.Interface
	var werr error
	if resourceType == "pod" {
		result, werr = s.cs.CoreV1Client.Pods("").Watch(options)
		// result, werr = s.cs.RESTClient().Get().Prefix("watch").Resource("pods").VersionedParams(&options, scheme.ParameterCodec).Watch()
	}
	if resourceType == "deployment" || resourceType == "app" {
		result, werr = s.cs.ExtensionsV1beta1().Deployments("").Watch(options)
		// result, werr = s.cs.RESTClient().Get().Prefix("watch").Resource("deployments").VersionedParams(&options, scheme.ParameterCodec).Watch()
	}
	if werr != nil {
		return nil, werr
	}
	return result, nil
}


func (s *Stream) FollowLog(podName, containerName string, tail int64) (io.ReadCloser, error) {
	logOption := &v1.PodLogOptions{
		Container:  containerName,
		Follow:     true,
		Timestamps: true,
		//SinceTime: &unversioned.Time{
		//	Time: time.Now(),
		//},
	}
	if 0 == tail {
		tail = 100
	}
	logOption.TailLines = &tail
	reader, err := s.cs.CoreV1Client.Pods(s.k8sNs).GetLogs(podName, logOption).Stream()
	return reader, err
}
