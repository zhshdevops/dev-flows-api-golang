package cache

import (
	"fmt"
)
//这个包暂时没有用到
//考虑日志需不需要缓存,另外把sql语句比较复杂的也考虑缓存
//用一个协程去处理Job的清理工作
type Builder interface {
	List(namespace string) ([]BuilderCache, error)
	Add(namespace string, info BuilderCache) error
	Delete(namespace string, info BuilderCache) error
	Sync() error
}

// key namespace
var FlowsOfNamespace = make(map[string][]BuilderCache)

type BuilderCache struct {
	FlowId        string `json:"flowId"`
	FlowBuildId   string `json:"flowBuildId"`
	StageBuildId  string `json:"stageBuildId"`
	StageId       string `json:"stageId"`
	ContainerName string `json:"containerName"`
	PodName       string `json:"podName"`
	JobName       string `json:"jobName"`
	NodeName      string `json:"nodeName"`
	Status        int `json:"status"`
	ControllerUid string `json:"controller_id"`
	LogData       string `json:"logData"`
}

func newBuilderCache() Builder {
	return &BuilderCache{}
}

func (cache *BuilderCache) List(namespace string) (cacheList []BuilderCache, err error) {
	if namespace == "" {
		err = fmt.Errorf("%s", "namespace is empty.")
		return
	}

	cacheList =FlowsOfNamespace[namespace]

	return
}

func (cache *BuilderCache) Add(namespace string,info BuilderCache) error {
	if namespace == "" {
		return fmt.Errorf("%s", "namespace is empty.")
	}
	FlowsOfNamespace[namespace]=append(FlowsOfNamespace[namespace],info)

	return nil
}

func (cache *BuilderCache) Delete(namespace string,info BuilderCache) error{
	if namespace == "" {
		return fmt.Errorf("%s", "namespace is empty.")
	}
	delete(FlowsOfNamespace,namespace)

	return nil

}
//同步k8s里面的信息
func (cache *BuilderCache) Sync() error {

	return nil
}
