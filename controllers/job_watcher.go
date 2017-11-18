package controllers

import (
	"github.com/googollee/go-socket.io"
	"github.com/golang/glog"

	"k8s.io/client-go/1.4/pkg/labels"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/watch"
	"fmt"

	"dev-flows-api-golang/models/common"
	"dev-flows-api-golang/modules/client"
	v1beta1 "k8s.io/client-go/1.4/pkg/apis/batch/v1"
	//"github.com/gorilla/websocket"
	"sync"
	//"golang.org/x/net/websocket"
)

type SocketsOfBuild struct {
	FlowId  string `json:"flowId"`
	Socket  socketio.Socket
	StageId string `json:"stage_id"`
}

// 保存stage build id对应的所有socket
// stage build完成时，需要从该mapping中获取build对应的socket，从而进行通知
var SOCKETS_OF_BUILD_MAPPING = make(map[string]map[string]*SocketsOfBuild, 0)
var SOCKETS_OF_BUILD_MAPPING_MUTEX sync.RWMutex
// 保存socket对应的所有stage build id
// 删除指定socket对应的SOCKETS_OF_BUILD_MAPPING记录时，需要从此mapping中获取build id
// 当socket对应的所有build均完成通知之后须断开连接，根据此mapping来判断何时断开连接
var BUILDS_OF_SOCKET_MAPPING = make(map[string]map[string]bool, 0)
var BUILDS_OF_SOCKET_MAPPING_MUTEX sync.RWMutex
// 保存stage id 对应的所有socket
// 新建stage build时，需要通知哪个stage新建了build，根据此mapping来获取stage对应的socket
type SocketsOfStage struct {
	FlowId string `json:"flowId"`
	Socket socketio.Socket
}

var SOCKETS_OF_STAGE_MAPPING = make(map[string]map[string]*SocketsOfStage, 0)
var SOCKETS_OF_STAGE_MAPPING_MUTEX sync.RWMutex
// 保存socket对应的所有stage
// 删除指定socket对应的SOCKETS_OF_STAGE_MAPPING记录时，需要从此mapping中获取stage id
var STAGES_OF_SOCKET_MAPPING = make(map[string]map[string]bool, 0)
var STAGES_OF_SOCKET_MAPPING_MUTEX sync.RWMutex
// 保存flow id对应的所有socket

var SOCKETS_OF_FLOW_MAPPING = make(map[string]map[string]socketio.Socket, 0)
var SOCKETS_OF_FLOW_MAPPING_MUTEX sync.RWMutex
// 保存socket对应的所有flow
var FLOWS_OF_SOCKET_MAPPING = make(map[string]map[string]bool, 0)
var FLOWS_OF_SOCKET_MAPPING_MUTEX sync.RWMutex

type WatchedBuilds struct {
	StageBuildId string `json:"stageBuildId"`
	StageId      string `json:"stageId"`
}

type WatchBuildInfo struct {
	FlowId        string `json:"flowId"`
	WatchedBuilds []WatchedBuilds `json:"watchedBuilds"`
}

type Message struct {
	Message string `json:"message"`
}

type WatchBuildResp struct {
	Status  int `json:"status"`
	Results interface{} `json:"results"`
}

type FlowBuildStatusInfo struct {
	FlowIds []string `json:"flows"`
}

const StageBuildStatusSocket = "stageBuildStatus"
const FlowBuildStatus = "flowBuildStatus"
const StopWatch = "stopWatch"


func Watch(flowId string, watchBuildInfo WatchBuildInfo, socket socketio.Socket) {
	method := "jobwatch.Watch"
	var i int
	var watchedBuildslen = len(watchBuildInfo.WatchedBuilds)

	if watchedBuildslen < 1 {
		//未指定watchedBuilds时，当做只监听flow
		if _, exist := SOCKETS_OF_FLOW_MAPPING[flowId]; !exist {
			SOCKETS_OF_FLOW_MAPPING[flowId] = map[string]socketio.Socket{
				socket.Id(): socket,
			}
		}

		if FLOWS_OF_SOCKET_MAPPING[socket.Id()] == nil {
			FLOWS_OF_SOCKET_MAPPING[socket.Id()] = map[string]bool{
				flowId: true,
			}
		}

		return
	}

	for _, stageBuild := range watchBuildInfo.WatchedBuilds {
		if stageBuild.StageId == "" { //如果没有对应的stageId
			//通知前端
			emitError(socket, flowId, stageBuild.StageId, stageBuild.StageBuildId, 400, "Stage id should be specified")
			return
		}

		//保存stage id对应的socket
		if _, exist := SOCKETS_OF_STAGE_MAPPING[stageBuild.StageId]; !exist {
			var socketsOfStage SocketsOfStage
			socketsOfStage.Socket = socket
			socketsOfStage.FlowId = flowId
			SOCKETS_OF_STAGE_MAPPING[stageBuild.StageId] = map[string]*SocketsOfStage{
				socket.Id(): &socketsOfStage,
			}
		}

		//保存socket对应的stage id
		if _, exist := STAGES_OF_SOCKET_MAPPING[socket.Id()]; !exist {
			STAGES_OF_SOCKET_MAPPING[socket.Id()] = map[string]bool{
				stageBuild.StageId: true,
			}
		}

		if stageBuild.StageBuildId == "" {
			glog.Infof("%s stageBuildId is empty\n", method)
			return
		}
		build, err := GetValidStageBuild(flowId, stageBuild.StageId, stageBuild.StageBuildId)
		if err != nil {
			//未获取到build时，返回错误
			glog.Errorf("%s GetValidStageBuild failed==>:%v\n", method, err)
			emitError(socket, flowId, stageBuild.StageId, stageBuild.StageBuildId, 400, fmt.Sprintf("%s", err))

		} else if build.Status == common.STATUS_SUCCESS || build.Status == common.STATUS_FAILED {
			//状态为成功或失败时，返回状态
			emitStatus(socket, flowId, stageBuild.StageId, stageBuild.StageBuildId, int(build.Status))
		} else {
			//保存build与socket的映射关系
			saveSocketAndBuild(socket, stageBuild.StageBuildId, flowId, stageBuild.StageId)
		}
		i = i + 1
		if i == watchedBuildslen {
			//遍历完成时，处理不需要watch的socket
			handleNoWatchedExist(socket)
		}

	}

}

// 通知前端
func emitStatus(socket socketio.Socket, flowId, stageId, stageBuildId string, buildStatus int) {

	glog.Infof("通知前端 emitStatus notisyflowstatus flowsid=%s,status=%d ", flowId, buildStatus)

	message := struct {
		FlowId       string `json:"flowId"`
		StageId      string `json:"stageId"`
		StageBuildId string `json:"stageBuildId"`
		BuildStatus  int `json:"buildStatus"`
	}{
		FlowId:       flowId,
		StageBuildId: stageBuildId,
		BuildStatus:  buildStatus,
		StageId:      stageId,
	}
	messageResp := struct {
		Results struct {
			FlowId       string `json:"flowId"`
			StageId      string `json:"stageId"`
			StageBuildId string `json:"stageBuildId"`
			BuildStatus  int `json:"buildStatus"`
		} `json:"results"`
		Status int `json:"status"`
	}{
		Results: message,
		Status:  200,
	}
	socket.Emit(StageBuildStatusSocket, messageResp)
	return
}

func emitError(socket socketio.Socket, flowId, stageId, stageBuildId string, Status int, message string) {
	glog.Infof("emitError notisyflowstatus flowsid=%s,status=%d ", flowId, Status)
	var resp WatchBuildResp
	resp.Status = Status
	messageResp := struct {
		FlowId       string `json:"flowId"`
		StageId      string `json:"stageId"`
		StageBuildId string `json:"stageBuildId"`
		Message      string `json:"message"`
	}{
		FlowId:       flowId,
		StageBuildId: stageBuildId,
		StageId:      stageId,
		Message:      message,
	}

	resp.Results = messageResp

	socket.Emit(StageBuildStatusSocket, resp)
	return
}

func saveSocketAndBuild(socket socketio.Socket, stageBuildId, flowId, stageId string) {
	glog.Infof("saveSocketAndBuild flowsid=%s,stageId=%s\n ", flowId, stageId)
	SOCKETS_OF_BUILD_MAPPING_MUTEX.RLock()
	defer SOCKETS_OF_BUILD_MAPPING_MUTEX.RUnlock()
	//保存build id对应的socket
	if SOCKETS_OF_BUILD_MAPPING[stageBuildId][socket.Id()] == nil {
		SOCKETS_OF_BUILD_MAPPING[stageBuildId] = map[string]*SocketsOfBuild{
			socket.Id(): &SocketsOfBuild{
				Socket:  socket,
				FlowId:  flowId,
				StageId: stageId,
			},
		}
	}

	//保存socket对应的build id
	if !BUILDS_OF_SOCKET_MAPPING[socket.Id()][stageBuildId] {
		BUILDS_OF_SOCKET_MAPPING[socket.Id()] = map[string]bool{
			stageBuildId: true,
		}
	}

}

func notifyFlow(flowId, flowBuildId string, status int) {
	glog.Infof("通知 flow 执行结果 :notifyFlow flowsid=%s,status=%d ", flowId, status)
	SOCKETS_OF_BUILD_MAPPING_MUTEX.Lock()
	defer SOCKETS_OF_BUILD_MAPPING_MUTEX.Unlock()
	method := "notifyFlow"
	if flowId == "" || flowBuildId == "" {
		return
	}

	if socketidMap, ok := SOCKETS_OF_FLOW_MAPPING[flowId]; ok {
		for key, socketMap := range socketidMap {

			glog.Infof("%s 存在对应的阶段任务: the socket id is %s\n", method, key)
			emitStatusOfFlow(socketMap, flowId, flowBuildId, status)
		}
	}

}

func emitStatusOfFlow(socket socketio.Socket, flowId, flowBuildId string, buildStatus int) {
	glog.Infof("通知前端某个子任务的状态 Intoing notisyflowstatus flowsid=%s,status=%d ", flowId, buildStatus)
	message := struct {
		FlowId      string `json:"flowId"`
		FlowBuildId string `json:"flowBuildId"`
		BuildStatus int `json:"buildStatus"`
	}{
		FlowId:      flowId,
		FlowBuildId: flowBuildId,
		BuildStatus: buildStatus,
	}
	socket.Emit(FlowBuildStatus, message)
	return
}

func emitErrorOfFlow(socket socketio.Socket, flowId, flowBuildId, message string, status int) {
	respToSocketMessage := struct {
		FlowId      string `json:"flowId"`
		FlowBuildId string `json:"flowBuildId"`
		BuildStatus int `json:"buildStatus"`
		Message     string `json:"message"`
	}{
		FlowId:      flowId,
		FlowBuildId: flowBuildId,
		BuildStatus: status,
		Message:     message,
	}
	socket.Emit(FlowBuildStatus, respToSocketMessage)
	return
}

func notifyNewBuild(stageId, stageBuildId string, status int) {
	glog.Infof("==========>>>notifyNewBuild stageBuildId:%s,status=%d\n", stageBuildId, status)
	method := "notifyNewBuild"
	if stageId == "" || stageBuildId == "" {
		return
	}

	if socketidMap, ok := SOCKETS_OF_STAGE_MAPPING[stageId]; ok {
		for key, socketMap := range socketidMap {
			glog.Infof("%s the socket id is %s\n", method, key)
			emitStatus(socketMap.Socket, socketMap.FlowId, stageId, stageBuildId, status)
			//保存新建build与socket的映射关系
			saveSocketAndBuild(socketMap.Socket, stageBuildId, socketMap.FlowId, stageId)
		}
	}

}

func notify(stageBuildId string, status int) {
	glog.Infof("通知前端:stageBuildId的状态==============>>%s,status=%d\n",stageBuildId,status)
	method := "notify"
	if stageBuildId == "" {
		return
	}
	if socketidMap, ok := SOCKETS_OF_BUILD_MAPPING[stageBuildId]; ok {
		SOCKETS_OF_BUILD_MAPPING_MUTEX.RLock()
		defer SOCKETS_OF_BUILD_MAPPING_MUTEX.RUnlock()
		for key, socketMap := range socketidMap {
			glog.Infof("%s the socket id is %s\n", method, key)
			emitStatus(socketMap.Socket, socketMap.FlowId, socketMap.StageId, stageBuildId, status)

			if status != common.STATUS_BUILDING {
				// 删除socket对应的stage build
				delete(BUILDS_OF_SOCKET_MAPPING[key], stageBuildId)
				// 处理socket是否需要关闭
				handleNoWatchedExist(socketMap.Socket)
			}

		}

		if status != common.STATUS_BUILDING {
			// 清空stage build对应的socket
			delete(SOCKETS_OF_BUILD_MAPPING, stageBuildId)
		}
	}

}
func handleNoWatchedExist(socket socketio.Socket) {
	// if (!BUILDS_OF_SOCKET_MAPPING[socket.id] ||
	//       Object.keys(BUILDS_OF_SOCKET_MAPPING[socket.id]).length < 1) {
	//   socket.disconnect()
	// }
}

func removeStagesAndBuilds(socket socketio.Socket) bool {
	glog.Infof("删除 stage build 的状态 removeStagesAndBuilds:%s\n",socket)
	return removeFromMapping_StageMapping(socket.Id()) &&
		removeFromMapping_BuildMapping(socket.Id())
}

//delete build
func removeFromMapping_BuildMapping(socketId string) bool {
	glog.Infof("removeFromMapping_BuildMapping:%s\n",socketId)
	if BUILDS_OF_SOCKET_MAPPING[socketId] != nil {
		// 删除object对应的socket
		for buildId, _ := range BUILDS_OF_SOCKET_MAPPING[socketId] {
			delete(SOCKETS_OF_BUILD_MAPPING[buildId], socketId)
		}

		// 清空socket对应的object
		delete(BUILDS_OF_SOCKET_MAPPING, socketId)
		return true
	}
	return false
}

//delete stage
func removeFromMapping_StageMapping(socketId string) bool {
	glog.Infof("removeFromMapping_StageMapping:%s\n",socketId)
	SOCKETS_OF_BUILD_MAPPING_MUTEX.RLock()
	defer SOCKETS_OF_BUILD_MAPPING_MUTEX.RUnlock()
	//socket没有对应的object时，不用删除
	if STAGES_OF_SOCKET_MAPPING[socketId] != nil {
		// 删除object对应的socket
		for stageId, _ := range STAGES_OF_SOCKET_MAPPING[socketId] {
			delete(SOCKETS_OF_STAGE_MAPPING[stageId], socketId)
		}

		// 清空socket对应的object
		delete(STAGES_OF_SOCKET_MAPPING, socketId)
		return true
	}
	return false
}

//delete flow
func removeFromMapping_FlowMapping(socketId string) bool {
	glog.Infof("删除flow的 removeFromMapping_FlowMapping:%s\n",socketId)
	SOCKETS_OF_BUILD_MAPPING_MUTEX.RLock()
	defer SOCKETS_OF_BUILD_MAPPING_MUTEX.RUnlock()
	//socket没有对应的object时，不用删除
	if FLOWS_OF_SOCKET_MAPPING[socketId] != nil {
		// 删除object对应的socket
		for flowId, _ := range FLOWS_OF_SOCKET_MAPPING[socketId] {
			delete(SOCKETS_OF_FLOW_MAPPING[flowId], socketId)
		}

		// 清空socket对应的object
		delete(FLOWS_OF_SOCKET_MAPPING, socketId)
		return true
	}
	return false
}

func NotifyFlowStatus(flowId, flowBuildId string, status int) {
	glog.Infof("通知前端 flow 的状态 :Intoing notisyflowstatus flowsid=%s,status=%d ", flowId, status)
	notifyFlow(flowId, flowBuildId, status)
}

func removeSocket(socket socketio.Socket) {
	removeStagesAndBuilds(socket)
	removeFromMapping_FlowMapping(socket.Id())
}

func init() {
	go doStart()
}

func doStart() {
Begin:
	method := "jobWatcher/doStart"

	glog.Infof("%s Job watcher starting with config kubetnetes clusterId:%s\n", method, client.ClusterID)
	//watch含有stage-build-id label的jobs
	labelsStr := fmt.Sprintf("stage-build-id%s", "")
	labelsSel, err := labels.Parse(labelsStr)
	if err != nil {
		glog.Errorf("%s label parse failed==>:%v\n", method, err)
		return
	}
	listOptions := api.ListOptions{
		LabelSelector: labelsSel,
	}

	watchInterface, err := client.KubernetesClientSet.BatchClient.Jobs("qinzhao").Watch(listOptions)
	if err != nil {
		glog.Errorf("%s get watchInterface failed %v\n", method, err)
		return
	}

	glog.Infof("%s Job watcher is ready\n", method)

	for {
		select {
		case event, isOpen := <-watchInterface.ResultChan():
			if !isOpen {
				glog.Errorf("%s the pod watch the chan is closed\n", method)
				goto Begin
				return
			}
			glog.Infof("Job watcher is ready the job event type=%s\n", event.Type)

			dm, parseIsOk := event.Object.(*v1beta1.Job)
			if !parseIsOk {
				glog.Errorf("%s 断言失败 \n", method)
				continue
			}

			if event.Type == watch.Deleted {
				glog.Infof("构建成功 %s A job is deleted:%v\n", event)
				if dm.ObjectMeta.Labels["stage-build-id"] != "" {
					if dm.Status.Succeeded >= 1 {
						//构建成功
						notify(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_SUCCESS)
					} else {
						//其他情况均视为失败状态
						notify(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_FAILED)
					}

				}

			} else if event.Type == watch.Added {
				glog.Infoln("收到added事件，等待中的stage build开始构建 namespace=",dm.ObjectMeta.Namespace)
				//收到added事件，等待中的stage build开始构建
				notifyNewBuild(dm.ObjectMeta.Labels["stage-id"],
					dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_BUILDING)
			} else if dm.Status.Succeeded >= 1 {
				//job执行成功
				glog.Infoln("namespace=",dm.ObjectMeta.Namespace)
				glog.Infof("===================>>Succeeded")
				notify(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_SUCCESS)
			} else if dm.Status.Failed >= 1 {
				//job执行失败
				glog.Infoln("namespace=",dm.ObjectMeta.Namespace)
				glog.Infof("===================>>failed")
				notify(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_FAILED)
			} else if dm.Spec.Parallelism == Int32Toint32Point(0) {
				glog.Infoln("namespace=",dm.ObjectMeta.Namespace)
				//停止job时
				//判断enncloud-builder-succeed label是否存在，从而确定执行成功或失败，并通知
				if dm.ObjectMeta.Labels["enncloud-builder-succeed"] != "1" {
					notify(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_SUCCESS)
				} else {
					notify(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_FAILED)
				}

			}

		}

	}

}

func Int32Toint32Point(input int32) *int32 {
	tmp := new(int32)
	*tmp = int32(input)
	return tmp

}
