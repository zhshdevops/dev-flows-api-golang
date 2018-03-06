package client

import (
	"github.com/golang/glog"
	clustermodel "dev-flows-api-golang/models/cluster"
	sqlstatus "dev-flows-api-golang/models/sql/status"
	"fmt"
	"dev-flows-api-golang/models/common"
	"encoding/json"
	"os"
	"path/filepath"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
)

var (
	KubernetesClientSet *ClientSet
	ClusterID           string
	Token               string
)
//get cluster configuration from database
//automatically response to client if error happens
func GetClusterOrRespErr(clusterId string) (*clustermodel.ClusterModel, bool) {
	cluster := &clustermodel.ClusterModel{}
	if statusCode, err := cluster.Get(clusterId); statusCode != sqlstatus.SQLSuccess {
		if statusCode == sqlstatus.SQLErrNoRowFound {
			glog.Errorf("Cluster %s does not exist\n", clusterId)
			return nil, false
		}
		glog.Errorf("Failed to query cluster from db, error:%s\n", err)
		return nil, false
	}
	return cluster, true
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func GetClientSetOrRespErr(clusterId string) (*ClientSet, bool) {
	// 获取cluster配置
	cluster, ok := GetClusterOrRespErr(clusterId)
	if false == ok {
		glog.Errorf("Failed to get cluster configuration: %s\n", clusterId)
		return nil, false
	}

	config := filepath.Join(homeDir(), "config")
	file, err := os.Create(config)
	if err != nil {
		glog.Errorf("create k8s config file err: %v", err)
		return nil, false
	}

	if cluster.APIToken == "" {

		if err = file.Truncate(0); err != nil {
			glog.Errorf("clean k8s tmp config file err: %v", err)
			return nil, false
		}
		if _, err = file.Write([]byte(cluster.Content)); err != nil {
			glog.Errorf("write k8s tmp config file err: %v", err)
			return nil, false
		}

		cs, err := newClientsetByConfile(config)
		if err != nil {
			glog.Errorf("create k8s client err: %v, who's clusterID is %q", err, cluster.ClusterID)
		}

		c := &ClientSet{Clientset: cs}
		return c, true

	}

	//初始化client
	cs, err := NewClientSet(cluster.ClusterName, cluster.APIProtocol, cluster.APIHost, cluster.APIToken, cluster.APIVersion)
	if nil != err {
		glog.Errorf("Failed to init k8s api of %s, error:%s\n", clusterId, err)
		return nil, false
	}

	glog.Infof("get k8s client %s://%s APIVersion=%s \n", cluster.APIProtocol, cluster.APIHost, cluster.APIVersion)
	return cs, true
}

func newClientsetByConfile(config string) (*kubernetes.Clientset, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config)
	if err != nil {
		return nil, err
	}
	cs, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// GetK8sConnection get k8s client and cluster info
// 1. check cluster id
// 2. get cluster info from db
// 3. connect to k8s
// return k8s client and cluster info
func GetK8sConnection(clusterID string) *ClientSet {
	method := "controllers/BaseController.GetK8sConnection"
	// get cluster info from db
	cluster, _ := GetClusterOrRespErr(clusterID)
	if cluster == nil {
		return nil
	}

	// get k8s client
	k8s, _ := GetClientSetOrRespErr(clusterID)
	if k8s == nil {
		return nil
	}
	glog.Infof("%s %s", method, k8s)
	return k8s
}

func Initk8sClient() {
	method := "Initk8sClient"
	cluster := clustermodel.NewClusterModel()
	clusters, err := cluster.GetAllCluster()
	if err != nil {
		fmt.Errorf("%s %v\n", method, err)
		return
	}

	var config clustermodel.Config
	for _, clu := range clusters {
		err = json.Unmarshal([]byte(clu.ConfigDetail), &config)
		if err != nil {
			fmt.Errorf(" json 解析失败 %s %v\n", method, err)
			return
		}
		if config.IsBuilder == 1 {
			ClusterID = clu.ClusterID
			clientSet, ok := GetClientSetOrRespErr(clu.ClusterID)
			KubernetesClientSet = clientSet
			Token = clu.APIToken
			if !ok {
				fmt.Errorf("get kubernetes client failed %s %v\n", method, err)
			}

			break
		}
	}

	return

}

func GetHarborServer() {
	method := "GetHarborServer"
	configs := clustermodel.NewConfigs()
	harborServerUrl, err := configs.GetHarborServer()
	if err != nil {
		fmt.Errorf("%s %v\n", method, err)
		return
	}

	common.HarborServerUrl = harborServerUrl

	fmt.Printf("HarborServerUrl=[%s] \n", harborServerUrl)
	return

}
