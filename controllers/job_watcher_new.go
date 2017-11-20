package controllers

import (
	"net/http"

	"github.com/golang/glog"

	"k8s.io/client-go/1.4/pkg/labels"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/watch"
	"encoding/json"
	"fmt"
	//"sync"

	"dev-flows-api-golang/models/common"
	"dev-flows-api-golang/modules/client"
	v1beta1 "k8s.io/client-go/1.4/pkg/apis/batch/v1"
	//"github.com/gorilla/websocket"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	//"golang.org/x/net/websocket"
	"strings"
)

//=================================new
type SocketsOfBuildNew struct {
	FlowId   string `json:"flowId"`
	Conn     net.Conn
	Op       ws.OpCode
	SocketId string
	StageId      string `json:"stageId"`
	StageBuildId string `json:"stageBuildId"`
	BuildStatus  int `json:"buildStatus"`
}

// 保存stage build id对应的所有socket
// stage build完成时，需要从该mapping中获取build对应的socket，从而进行通知
var SOCKETS_OF_BUILD_MAPPING_NEW = make(map[string]map[string]*SocketsOfBuildNew, 0)
// 保存socket对应的所有stage build id
// 删除指定socket对应的SOCKETS_OF_BUILD_MAPPING记录时，需要从此mapping中获取build id
// 当socket对应的所有build均完成通知之后须断开连接，根据此mapping来判断何时断开连接
var BUILDS_OF_SOCKET_MAPPING_NEW = make(map[string]map[string]bool, 0)
// 保存stage id 对应的所有socket
// 新建stage build时，需要通知哪个stage新建了build，根据此mapping来获取stage对应的socket
type SocketsOfStageNew struct {
	FlowId   string `json:"flowId"`
	SockerId string
	Conn     net.Conn
	Op       ws.OpCode
}

var SOCKETS_OF_STAGE_MAPPING_NEW = make(map[string]map[string]*SocketsOfStageNew, 0)
// 保存socket对应的所有stage
// 删除指定socket对应的SOCKETS_OF_STAGE_MAPPING记录时，需要从此mapping中获取stage id
var STAGES_OF_SOCKET_MAPPING_NEW = make(map[string]map[string]bool, 0)
// 保存flow id对应的所有socket

type Conn struct {
	Conn net.Conn
	Op   ws.OpCode
}

var SOCKETS_OF_FLOW_MAPPING_NEW = make(map[string]map[string]Conn, 0)
// 保存socket对应的所有flow
var FLOWS_OF_SOCKET_MAPPING_NEW = make(map[string]map[string]bool, 0)

type JobWatcherSocket struct {
	Handler http.Handler
}

func NewJobWatcherSocket() *JobWatcherSocket {
	return &JobWatcherSocket{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := "JobWatcherSocket"
			conn, _, _, err := ws.UpgradeHTTP(r, w, nil)
			if err != nil {
				glog.Errorf("%s connect JobWatcherSocket websocket failed:%v\n", method, err)
			}
			defer conn.Close()
			var watchMessage WatchBuildInfo
			var resp WatchBuildResp
			var flow FlowBuildStatusInfo

			CloseSocketChan := make(chan CloseSocketResp, 3)
			FlowResp := make(chan FlowBuilResp, 3)
			watchBuildInfo := make(chan []byte, 1)
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				glog.Errorf("msg:%v========err:%v\n", msg, err)
				resp.Status = 400
				resp.Results = "json unmarshal failed"
				data, _ := json.Marshal(resp)
				wsutil.WriteServerMessage(conn, op, data)
				return
			}

			glog.Infof("receive msg :%v\n", string(msg))

			watchBuildInfo <- msg

			go func() {
				for {
					select {
					case res, ok := <-FlowResp:
						glog.Infof("%s %s", res, ok)
						data, _ := json.Marshal(resp)
						wsutil.WriteServerMessage(conn, op, data)
					case resp, isOpen := <-CloseSocketChan:
						glog.Infof("%s %s", resp, isOpen)
						data, _ := json.Marshal(resp)
						wsutil.WriteServerMessage(conn, op, data)
						return
					case reqData, isOpen := <-watchBuildInfo:
						glog.Infof("%s %s", reqData, isOpen)
						if strings.Contains(string(msg), "watchedBuilds") {
							err := json.Unmarshal(reqData, &watchMessage)
							glog.Infof("%s message===StageBuildStatusSocket:%s\n", method, StageBuildStatusSocket)
							if err != nil {
								glog.Infof("%s message===:%v\n", method, msg)
								glog.Errorf("%s json unmarshal failed====> %v\n", method, err)
								resp.Status = 400
								resp.Results = "json unmarshal failed"
								data, _ := json.Marshal(resp)
								wsutil.WriteServerMessage(conn, op, data)
								return
							}

							if watchMessage.FlowId == "" || watchMessage.WatchedBuilds == nil {
								glog.Errorf("%s Missing Parameters====> %v\n", method, watchMessage)
								resp.Status = 400
								resp.Results = "Missing Parameters"
								data, _ := json.Marshal(resp)
								wsutil.WriteServerMessage(conn, op, data)
								return
							}

						} else {
							err := json.Unmarshal(msg, &flow)
							if err != nil {
								glog.Errorf("%s json unmarshal failed====> %v\n", method, err)
								resp.Status = 400
								resp.Results = "json unmarshal failed"
								data, _ := json.Marshal(resp)
								wsutil.WriteServerMessage(conn, op, data)
								return
							}

							if len(flow.FlowIds) == 0 {
								glog.Errorf("%s Missing Parameters====> %v\n", method, err)
								resp.Status = 400
								resp.Results = "Missing Parameters"
								data, _ := json.Marshal(resp)
								wsutil.WriteServerMessage(conn, op, data)
								return
							}

						}
					}

				}

			}()
		}),
	}
}

func WatchNew(flowId string, watchBuildInfo WatchBuildInfo, socket net.Conn, socketId string, op ws.OpCode) {

	method := "jobwatch.WatchNew"
	var i int
	var watchedBuildslen = len(watchBuildInfo.WatchedBuilds)

	if watchedBuildslen < 1 {
		//未指定watchedBuilds时，当做只监听flow
		var conn Conn
		conn.Conn = socket
		conn.Op = op
		if _, exist := SOCKETS_OF_FLOW_MAPPING_NEW[flowId]; !exist {
			SOCKETS_OF_FLOW_MAPPING_NEW[flowId] = map[string]Conn{
				socketId: conn,
			}
		}

		if FLOWS_OF_SOCKET_MAPPING_NEW[socketId] == nil {
			FLOWS_OF_SOCKET_MAPPING_NEW[socketId] = map[string]bool{
				flowId: true,
			}
		}

		return
	}

	for _, stageBuild := range watchBuildInfo.WatchedBuilds {
		if stageBuild.StageId == "" { //如果是没有对应的stageId
			emitErrorNew(socket, flowId, stageBuild.StageId, stageBuild.StageBuildId, 400, "Stage id should be specified", op)
			return
		}

		//保存stage id对应的socket
		if _, exist := SOCKETS_OF_STAGE_MAPPING_NEW[stageBuild.StageId]; !exist {
			SOCKETS_OF_BUILD_MAPPING_MUTEX.RLock()
			socketsOfStageNew := &SocketsOfStageNew{}

			socketsOfStageNew.Op = op
			socketsOfStageNew.Conn = socket
			socketsOfStageNew.SockerId = socketId
			socketsOfStageNew.FlowId = flowId
			SOCKETS_OF_STAGE_MAPPING_NEW[stageBuild.StageId] = map[string]*SocketsOfStageNew{
				socketId: socketsOfStageNew,
			}

			SOCKETS_OF_BUILD_MAPPING_MUTEX.RUnlock()
		}

		//保存socket对应的stage id
		if _, exist := STAGES_OF_SOCKET_MAPPING_NEW[socketId]; !exist {
			STAGES_OF_SOCKET_MAPPING_NEW[socketId] = map[string]bool{
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
			emitErrorNew(socket, flowId, stageBuild.StageId, stageBuild.StageBuildId, 400, fmt.Sprintf("%s", err), op)
		} else if build.Status == common.STATUS_SUCCESS || build.Status == common.STATUS_FAILED {
			//状态为成功或失败时，返回状态
			emitStatusNew(socket, flowId, stageBuild.StageId, stageBuild.StageBuildId, int(build.Status), op)
		} else {
			//保存build与socket的映射关系
			saveSocketAndBuildNew(socket, socketId, stageBuild.StageBuildId, flowId, stageBuild.StageId, op)
		}
		i = i + 1
		if i == watchedBuildslen {
			//遍历完成时，处理不需要watch的socket
			handleNoWatchedExistNew(socket, socketId)
		}

	}

}

// 通知前端

func emitStatusNew(socket net.Conn, flowId, stageId, stageBuildId string, buildStatus int, op ws.OpCode) {

	glog.Infof("emitStatusNew notisyflowstatus flowsid=%s,status=%d ", flowId, buildStatus)

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

	data, _ := json.Marshal(messageResp)

	wsutil.WriteServerMessage(socket, op, data)
	return
}

func emitErrorNew(socket net.Conn, flowId, stageId, stageBuildId string, Status int, message string, op ws.OpCode) {
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

	data, _ := json.Marshal(resp)

	wsutil.WriteServerMessage(socket, op, data)
	return
}

func saveSocketAndBuildNew(socket net.Conn, socketId string, stageBuildId, flowId, stageId string, op ws.OpCode) {

	SOCKETS_OF_BUILD_MAPPING_MUTEX.RLock()
	defer SOCKETS_OF_BUILD_MAPPING_MUTEX.RUnlock()
	//保存build id对应的socket
	if SOCKETS_OF_BUILD_MAPPING_NEW[stageBuildId][socketId] == nil {
		socketsOfBuildNew := &SocketsOfBuildNew{
			Conn:    socket,
			FlowId:  flowId,
			StageId: stageId,
			Op:      op,
		}
		SOCKETS_OF_BUILD_MAPPING_NEW[stageBuildId] = map[string]*SocketsOfBuildNew{
			socketId: socketsOfBuildNew,
		}
	}

	//保存socket对应的build id
	if !BUILDS_OF_SOCKET_MAPPING_NEW[socketId][stageBuildId] {
		BUILDS_OF_SOCKET_MAPPING_NEW[socketId] = map[string]bool{
			stageBuildId: true,
		}
	}

}

func notifyFlowNew(flowId, flowBuildId string, status int) {
	SOCKETS_OF_BUILD_MAPPING_MUTEX.Lock()
	defer SOCKETS_OF_BUILD_MAPPING_MUTEX.Unlock()
	method := "notifyFlowNew"
	if flowId == "" || flowBuildId == "" {
		return
	}

	if socketidMap, ok := SOCKETS_OF_FLOW_MAPPING_NEW[flowId]; ok {
		for key, socketMap := range socketidMap {
			glog.Infof("%s the socket id is %s\n", method, key)
			emitStatusOfFlowNew(socketMap.Conn, flowId, flowBuildId, status, socketMap.Op)
		}
	}

}

func emitStatusOfFlowNew(socket net.Conn, flowId, flowBuildId string, buildStatus int, op ws.OpCode) {
	glog.Infof("Intoing emitStatusOfFlowNew flowsid=%s,status=%d ", flowId, buildStatus)
	message := struct {
		FlowId      string `json:"flowId"`
		FlowBuildId string `json:"flowBuildId"`
		BuildStatus int `json:"buildStatus"`
	}{
		FlowId:      flowId,
		FlowBuildId: flowBuildId,
		BuildStatus: buildStatus,
	}
	data, _ := json.Marshal(message)

	wsutil.WriteServerMessage(socket, op, data)
	return
}

func emitErrorOfFlowNew(socket net.Conn, flowId, flowBuildId, message string, status int, op ws.OpCode) {
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
	data, _ := json.Marshal(respToSocketMessage)

	wsutil.WriteServerMessage(socket, op, data)
	return
}

func notifyNewBuildNew(stageId, stageBuildId string, status int) {
	glog.Infof("==========>>>notifyNewBuildNew stageBuildId:%s,status=%d,stageId=%s\n", stageBuildId, status, stageId)
	method := "notifyNewBuildNew"
	if stageId == "" || stageBuildId == "" {
		return
	}

	if socketidMap, ok := SOCKETS_OF_STAGE_MAPPING_NEW[stageId]; ok {
		for key, socketMap := range socketidMap {
			glog.Infof("%s the socket id is %s\n", method, key)
			emitStatusNew(socketMap.Conn, socketMap.FlowId, stageId, stageBuildId, status, socketMap.Op)
			//保存新建build与socket的映射关系
			saveSocketAndBuildNew(socketMap.Conn, socketMap.SockerId, stageBuildId, socketMap.FlowId, stageId, socketMap.Op)
		}
	}

}

func notifyNew(stageBuildId string, status int) {
	glog.Infof("notifyNew=========>>stageBuildId:%s,status=%d\n", stageBuildId, status)
	method := "notify"
	if stageBuildId == "" {
		return
	}
	glog.Infof("notify stageBuildId:%s\n", stageBuildId)
	if socketidMap, ok := SOCKETS_OF_BUILD_MAPPING_NEW[stageBuildId]; ok {
		SOCKETS_OF_BUILD_MAPPING_MUTEX.RLock()
		defer SOCKETS_OF_BUILD_MAPPING_MUTEX.RUnlock()
		for key, socketMap := range socketidMap {
			glog.Infof("%s the socket id is %s\n", method, key)
			emitStatusNew(socketMap.Conn, socketMap.FlowId, socketMap.StageId, stageBuildId, status, socketMap.Op)

			if status != common.STATUS_BUILDING {
				// 删除socket对应的stage build
				delete(BUILDS_OF_SOCKET_MAPPING_NEW[key], stageBuildId)
				// 处理socket是否需要关闭
				handleNoWatchedExistNew(socketMap.Conn, socketMap.SocketId)
			}

		}

		if status != common.STATUS_BUILDING {
			// 清空stage build对应的socket
			delete(SOCKETS_OF_BUILD_MAPPING_NEW, stageBuildId)
		}
	}

}

func handleNoWatchedExistNew(socket net.Conn, socketId string) {
	if _, ok := BUILDS_OF_SOCKET_MAPPING_NEW[socketId]; !ok {
		return
	}
}

func removeStagesAndBuildsNew(scoketId string) bool {
	return removeFromMapping_StageMappingNew(scoketId) &&
		removeFromMapping_BuildMappingNew(scoketId)
}

//delete build
func removeFromMapping_BuildMappingNew(socketId string) bool {
	glog.Infof("removeFromMapping_BuildMappingNew=========>>socketId:%s\n", socketId)
	if BUILDS_OF_SOCKET_MAPPING_NEW[socketId] != nil {
		// 删除object对应的socket
		for buildId, _ := range BUILDS_OF_SOCKET_MAPPING_NEW[socketId] {
			delete(SOCKETS_OF_BUILD_MAPPING_NEW[buildId], socketId)
		}

		// 清空socket对应的object
		delete(BUILDS_OF_SOCKET_MAPPING_NEW, socketId)
		return true
	}
	return false
}

//delete stage
func removeFromMapping_StageMappingNew(socketId string) bool {
	glog.Infof("removeFromMapping_StageMappingNew=========>>socketId:%s\n", socketId)
	SOCKETS_OF_BUILD_MAPPING_MUTEX.RLock()
	defer SOCKETS_OF_BUILD_MAPPING_MUTEX.RUnlock()
	//socket没有对应的object时，不用删除
	if STAGES_OF_SOCKET_MAPPING_NEW[socketId] != nil {
		// 删除object对应的socket
		for stageId, _ := range STAGES_OF_SOCKET_MAPPING_NEW[socketId] {
			delete(SOCKETS_OF_STAGE_MAPPING_NEW[stageId], socketId)
		}

		// 清空socket对应的object
		delete(STAGES_OF_SOCKET_MAPPING_NEW, socketId)
		return true
	}
	return false
}

//delete flow
func removeFromMapping_FlowMappingNew(socketId string) bool {
	glog.Infof("removeFromMapping_FlowMappingNew=========>>socketId:%s\n", socketId)
	SOCKETS_OF_BUILD_MAPPING_MUTEX.RLock()
	defer SOCKETS_OF_BUILD_MAPPING_MUTEX.RUnlock()
	//socket没有对应的object时，不用删除
	if FLOWS_OF_SOCKET_MAPPING_NEW[socketId] != nil {
		// 删除object对应的socket
		for flowId, _ := range FLOWS_OF_SOCKET_MAPPING_NEW[socketId] {
			delete(SOCKETS_OF_FLOW_MAPPING_NEW[flowId], socketId)
		}

		// 清空socket对应的object
		delete(FLOWS_OF_SOCKET_MAPPING_NEW, socketId)
		return true
	}
	return false
}

func NotifyFlowStatusNew(flowId, flowBuildId string, status int) {
	glog.Infof("NotifyFlowStatusNew flowsid=%s,status=%d ", flowId, status)
	notifyFlowNew(flowId, flowBuildId, status)
}

func removeSocketNew(conn Conn, socketId string) {
	removeStagesAndBuildsNew(socketId)
	removeFromMapping_FlowMappingNew(socketId)
}

//func init() {
//	go doStartNew()
//}

func doStartNew() {
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

	watchInterface, err := client.KubernetesClientSet.BatchClient.Jobs("").Watch(listOptions)
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
				glog.Infof("%s A job is deleted:%v\n", event)
				if dm.ObjectMeta.Labels["stage-build-id"] != "" {
					if dm.Status.Succeeded >= 1 {
						//构建成功
						notifyNew(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_SUCCESS)
					} else {
						//其他情况均视为失败状态
						notifyNew(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_FAILED)
					}

				}

			} else if event.Type == watch.Added {
				//收到added事件，等待中的stage build开始构建
				notifyNewBuildNew(dm.ObjectMeta.Labels["stage-id"],
					dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_BUILDING)
			} else if dm.Status.Succeeded >= 1 {
				//job执行成功
				glog.Infof("===================>>Succeeded")
				notifyNew(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_SUCCESS)
			} else if dm.Status.Failed >= 1 {
				//job执行失败
				glog.Infof("===================>>failed:label stage-build-id=%s\n", dm.ObjectMeta.Labels["stage-build-id"])
				notifyNew(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_FAILED)
			} else if dm.Spec.Parallelism == Int32Toint32Point(0) {
				//停止job时
				//判断enncloud-builder-succeed label是否存在，从而确定执行成功或失败，并通知
				if dm.ObjectMeta.Labels["enncloud-builder-succeed"] != "1" {
					notifyNew(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_SUCCESS)
				} else {
					notifyNew(dm.ObjectMeta.Labels["stage-build-id"], common.STATUS_FAILED)
				}

			}

		}

	}

}
