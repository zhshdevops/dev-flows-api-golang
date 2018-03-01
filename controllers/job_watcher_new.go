package controllers

import (
	"net/http"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/fields"
	//"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/apimachinery/pkg/watch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/json"
	"fmt"
	"dev-flows-api-golang/models/common"
	v1beta1 "k8s.io/client-go/pkg/apis/batch/v1"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"sync"
	"time"
	"dev-flows-api-golang/models"
)

type EnnFlow struct {
	FlowId           string `json:"flowId"`
	FlowBuildId      string `json:"flowBuildId"`
	StageId          string `json:"stageId"`
	StageBuildId     string `json:"stageBuildId"`
	BuildStatus      int `json:"buildStatus"`
	Status           int `json:"status"`
	Message          string `json:"message"`
	CodeBranch       string `json:"codeBranch"`
	Flag             int `json:"flag"` //1 表示flow 2表示stage
	Namespace        string `json:"namespace"`
	LoginUserName    string `json:"loginUserName"`
	UserNamespace    string `json:"userNamespace"`
	Event            string `json:"event"`
	WebSocketIfClose int `json:"webSocketIfClose"` //0表示不关闭 1表示关闭 2心跳
}

var FlowMapping *SocketsOfFlowMapping

type Conn struct {
	Conn     net.Conn
	Op       ws.OpCode
	ConnTime time.Time
}

func init() {
	FlowMapping = NewSocketsOfFlowMapping()
	go func() {
		for {
			select {
			case ennFlowInfo, ok := <-EnnFlowChan:
				if ok {
					flow, ok := ennFlowInfo.(EnnFlow)
					if ok {
						if flow.WebSocketIfClose == 1 {
							return
						} else {
							FlowMapping.Send(flow)
						}
					}
				} else {
					glog.Errorf(" EnnFlowChan is closed:%v\n", ennFlowInfo)
				}

			}
		}
	}()

}

var EnnFlowChan = make(chan interface{}, 10240)

type SocketsOfFlowMapping struct {
	SocketMutex sync.RWMutex
	FlowMap     map[string]map[string]Conn
}

func NewSocketsOfFlowMapping() *SocketsOfFlowMapping {
	return &SocketsOfFlowMapping{
		FlowMap: make(map[string]map[string]Conn, 0),
	}
}

func (soc *SocketsOfFlowMapping) ClearOrCloseConnect(flowId string, disconn net.Conn) {
	soc.SocketMutex.Lock()
	defer soc.SocketMutex.Unlock()

	conns, ok := soc.FlowMap[flowId]

	connsLen := len(conns)

	if ok {
		if connsLen != 0 {
			for key, conn := range conns {
				if conn.Conn == disconn {
					conn.Conn.Close()
					delete(conns, key)
				}
			}
		} else {
			delete(soc.FlowMap, flowId)
		}

	}

	return
}

func (soc *SocketsOfFlowMapping) Exist(flowId string) bool {
	soc.SocketMutex.Lock()
	defer soc.SocketMutex.Unlock()
	conns, ok := soc.FlowMap[flowId]
	connsLen := len(conns)
	if ok {
		if connsLen != 0 {
			for key, conn := range conns {
				if time.Now().Sub(conn.ConnTime) > 4*time.Hour {
					conn.Conn.Close()
					delete(conns, key)
				}
			}
		}
	}

	return ok
}

func (soc *SocketsOfFlowMapping) JoinConn(flowId, randId string, conn Conn) {
	soc.SocketMutex.Lock()
	defer soc.SocketMutex.Unlock()
	_, ok := soc.FlowMap[flowId]
	if ok {
		glog.Infof("the websocket is exist=======>>")
		soc.FlowMap[flowId][randId] = conn
		glog.Infof("exist soc.FlowMap[%s]=%v\n", flowId, soc.FlowMap[flowId])
		return
	}

	glog.Infof("the websocket is not exist=======>>")
	soc.FlowMap[flowId] = map[string]Conn{
		randId: conn,
	}

	glog.Infof("not exist soc.FlowMap[%s]=%v\n", flowId, soc.FlowMap[flowId])
	return
}

func (soc *SocketsOfFlowMapping) Send(flow interface{}) {

	soc.SocketMutex.Lock()
	defer soc.SocketMutex.Unlock()

	flowInfo, ok := flow.(EnnFlow)
	if !ok {
		return
	}

	ConnLen := len(soc.FlowMap[flowInfo.FlowId])
	if ConnLen != 0 {
		for key, conn := range soc.FlowMap[flowInfo.FlowId] {
			if conn.Conn != nil {
				for j := 1; j <= 5; j++ {
					err := SendRetry(flow, conn)
					if err != nil {
						glog.Errorf("key=%s retry  %d times send msg to client:%v\n", key, j, err)
						if j == 5 {
							break

						}
						continue
					} else {
						break
					}

				}

			}
		}

	}

}

type JobLogSocket struct {
	Handler http.Handler
}

func NewJobLogSocket() *JobLogSocket {
	return &JobLogSocket{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := "JobLogSocket"
			glog.Infof("%s connect user build log  获取构建实时日志 \n", method)
			//判断是否存在
			conn, _, _, err := ws.UpgradeHTTP(r, w, nil)
			if err != nil {
				glog.Errorf("%s 建立Websocket日志链接失败 connect JobLogSocket websocket failed:%v\n", method, err)
				w.Write([]byte(`<font color="red">[Enn Flow API Error] 建立Websocket日志链接失败</font>`))
				return
			}

			defer conn.Close()
			var con Conn

			con.Conn = conn
			var buildMessage EnnFlow

			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {

				glog.Errorf("读取客户端发送的数据包失败 connect JobLogSocket websocket failed: msg:%s,err:%v\n", string(msg), err)
				wsutil.WriteServerMessage(conn, op, []byte(`<font color="red">[Enn Flow API Error] 读取客户端发送的数据包失败</font>`))
				return
			}

			con.Op = op

			err = json.Unmarshal(msg, &buildMessage)
			if err != nil {
				glog.Errorf("反系列化数据库包失败 msg:%s,err:%v========err:%v\n", string(msg), err)
				wsutil.WriteServerMessage(conn, op, []byte(`<font color="red">[Enn Flow API Error] 反系列化数据库包失败</font>`))
				return
			}

			if message := CheckLogData(buildMessage); message != "" {
				glog.Errorf("Missing parameters====>>:[%v]\n", buildMessage)
				SendLog(message, con)
				return
			}

			GetStageBuildLogsFromK8S(buildMessage, con)

		}),
	}
}

type JobWatcherSocket struct {
	Handler http.Handler
}

func NewJobWatcherSocket() *JobWatcherSocket {
	return &JobWatcherSocket{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := "JobWatcherSocket"
			glog.Infof("%s connect JobWatcherSocket websocket ", method)
			//判断是否存在
			conn, _, _, err := ws.UpgradeHTTP(r, w, nil)
			if err != nil {
				glog.Errorf("%s connect JobWatcherSocket websocket failed:%v\n", method, err)
				w.Write([]byte("建立Websocket状态链接失败"))
				return
			}

			for {
				var flow EnnFlow
				var randId string = newId(r)
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil && len(msg) != 0 {
					glog.Errorf("connect JobWatcherSocket websocket failed: msg:%s,err:%v\n", msg, err)
					flow.Status = 400
					flow.Flag = 1
					flow.BuildStatus = common.STATUS_FAILED
					flow.Message = "读取客户端信息失败"
					glog.Infof("response info:%v\n", flow)
					data, _ := json.Marshal(flow)
					err := wsutil.WriteServerMessage(conn, op, data)
					if err != nil {
						glog.Errorf("=======err:%v\n", err)
					}
					return
				}

				err = json.Unmarshal(msg, &flow)
				if err != nil {
					glog.Errorf("request msg:%v err:%v\n", string(msg), err)
					flow.Status = 400
					flow.Flag = 1
					flow.Message = "反系列化失败"
					data, _ := json.Marshal(flow)
					err := wsutil.WriteServerMessage(conn, op, data)
					if err != nil {
						glog.Errorf("=======err:%v\n", err)
					}
					return
				}

				glog.Infof("Flow info ====>>%#v\n", flow.FlowId)

				if flow.FlowId != "" {

					if flow.WebSocketIfClose == 0 {
						var connOfFlow Conn
						connOfFlow.Conn = conn
						connOfFlow.Op = op
						connOfFlow.ConnTime = time.Now()
						flow.Status = http.StatusOK
						flow.Message = "success"
						flow.Flag = 1
						Retry(flow, connOfFlow)
						FlowMapping.JoinConn(flow.FlowId, randId, connOfFlow)
						continue

					} else if flow.WebSocketIfClose == 1 {
						//释放资源
						glog.Infof("the websocket is closeing=======>>%v\n", FlowMapping.FlowMap[flow.FlowId])
						FlowMapping.ClearOrCloseConnect(flow.FlowId, conn)
						glog.Infof("the websocket is closeed=======>>%v\n", FlowMapping.FlowMap[flow.FlowId])
						return
					} else if flow.WebSocketIfClose == 2 {
						flow.Status = 200
						flow.Message = "success"
						data, _ := json.Marshal(flow)
						err := wsutil.WriteServerMessage(conn, op, data)
						if err != nil {
							glog.Errorf("=======err:%v\n", err)
						}
						continue
					}

				} else {
					glog.Errorf("FlowId is empty")
					flow.Status = 400
					flow.Message = "FlowId is empty"
					flow.Flag = 1
					data, _ := json.Marshal(flow)
					err := wsutil.WriteServerMessage(conn, op, data)
					if err != nil {
						glog.Errorf("WriteServerMessage to client failed:%v\n", err)
					}
					return
				}

			}

		}),
	}
}

func GetEnnFlow(job *v1beta1.Job, buildStatus int) {
	glog.Infof("GetEnnFlow buildStatus=%d\n", buildStatus)
	var ennFlow EnnFlow
	ennFlow.Status = http.StatusOK
	ennFlow.Flag = 2
	ennFlow.BuildStatus = buildStatus
	ennFlow.Message = fmt.Sprintf("构建任务stage-build-id=%s\n", job.GetLabels()["stage-build-id"])
	ennFlow.StageId = job.GetLabels()["stage-id"]
	ennFlow.FlowId = job.GetLabels()["flow-id"]
	ennFlow.FlowBuildId = job.GetLabels()["flow-build-id"]
	ennFlow.StageBuildId = job.GetLabels()["stage-build-id"]
	if buildStatus == common.STATUS_SUCCESS || buildStatus == common.STATUS_FAILED {
		models.NewCiStageBuildLogs().UpdateStageBuildStatusById(buildStatus, ennFlow.StageBuildId)
	}
	EnnFlowChan <- ennFlow
}

//WatchJob  watch  the job event fieldSelectorStr := "status.phase!=Succeeded,status.phase!=Failed"
func jobWatcher() {
	method := "jobWatcher"

	labelsStr := fmt.Sprintf("system/jobType=%s", "devflows")

	labelsSel, err := labels.Parse(labelsStr)
	if err != nil {
		glog.Errorf("%s label parse failed==>:%v\n", method, err)
		return
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelsSel.String(),
		Watch:         true,
	}

	watchInterface, err := models.NewImageBuilder().Client.BatchV1Client.Jobs("").Watch(listOptions)
	if err != nil {
		glog.Errorf("%s,err: %v\n", method, err)
		return
	}

	for {
		select {
		case event, isOpen := <-watchInterface.ResultChan():
			if isOpen == false {
				glog.Errorf("%s the watch job chain is close\n", method)
				jobWatcher()
			}
			dm, parseIsOk := event.Object.(*v1beta1.Job)
			if false == parseIsOk {
				glog.Errorf("%s job %s\n", method, ">>>>>>断言失败<<<<<<")
				continue
			}
			glog.Infof("%s job event.Type=%s, namespace=%s ,jobName=%s\n", method, event.Type, dm.GetNamespace(),
				dm.GetName())
			glog.Infof("%s job event.Status=%#v\n", method, dm.Status)
			if event.Type == watch.Added {
				//收到deleted事件，job可能被第三方删除
				glog.Infof("%s %s,status:%v\n", method, "收到ADD事件,开始起job进行构建", dm.Status)
				GetEnnFlow(dm, common.STATUS_BUILDING)
				continue
			} else if event.Type == watch.Deleted {
				//收到deleted事件，job可能被第三方删除
				glog.Errorf("%s  %s status:%v\n", method, " 收到deleted事件，job可能被第三方删除", dm.Status)
				continue
				//成功时并且已经完成时
			} else if dm.Status.Succeeded >= 1 &&
				dm.Status.CompletionTime != nil && len(dm.Status.Conditions) != 0 {
				glog.Infof("%s %s,status:%v\n", method, "构建成功", dm.Status)
				GetEnnFlow(dm, common.STATUS_SUCCESS)
				continue
				//} else if dm.Status.Failed >=1 && dm.Spec.Completions == Int32Toint32Point(1) &&
				//	dm.Status.CompletionTime == nil && dm.Status.Succeeded==0{
			} else if dm.Status.Failed >= 1 {
				glog.Infof("%s %s,status:%v\n", method, "构建失败", dm.GetName())
				GetEnnFlow(dm, common.STATUS_FAILED)
				continue
				//手动停止job
			} else if dm.Spec.Parallelism == Int32Toint32Point(0) {
				//有依赖服务，停止job时 不是手动停止 1 表示手动停止
				if dm.ObjectMeta.Labels["enncloud-builder-succeed"] != "1" && dm.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "true" {
					glog.Infof("%s %s,status:%v\n", method, "用户停止了构建任务", dm)
					GetEnnFlow(dm, common.STATUS_FAILED)
					continue
					//没有依赖服务时
				} else {
					GetEnnFlow(dm, common.STATUS_FAILED)
					glog.Infof("%s %s,status:%v\n", method, "job执行失败程序发送了停止构建任务的命令", dm)
					continue
				}
			}

		}
	}

}

func Retry(flow interface{}, conn Conn) {

	for i := 1; i <= 5; i++ {
		err := SendRetry(flow, conn)
		if err != nil {
			glog.Errorf("retry %d time send msg to client\n", i)
			if i == 5 {
				flowInfo, ok := flow.(EnnFlow)
				if ok {
					FlowMapping.ClearOrCloseConnect(flowInfo.FlowId, conn.Conn)
				}
				break

			}
			continue
		} else {
			break
		}

	}

}

func SendLog(flow string, conn Conn) {
	if conn.Conn != nil {
		glog.Infof("websocket response flow build log info:%v\n", flow)
		wsutil.WriteServerMessage(conn.Conn, conn.Op, []byte(flow))
		return
	}

}

func SendRetry(flow interface{}, conn Conn) error {

	if conn.Conn != nil {
		data, _ := json.Marshal(flow)
		glog.Infof("websocket response flow build log info:%v\n", flow)
		return wsutil.WriteServerMessage(conn.Conn, conn.Op, data)

	}

	return nil
}

func Int32Toint32Point(input int32) *int32 {
	tmp := new(int32)
	*tmp = int32(input)
	return tmp

}

func (queue *StageQueueNew) WatchPod(namespace, jobName string, stage models.CiStages) (bool, error) {
	method := "WatchPod"
	var ennFlow EnnFlow
	var timeOut bool = false
	glog.Infof("Begin watch Pod \n")
	ennFlow.Status = http.StatusOK
	ennFlow.BuildStatus = common.STATUS_BUILDING
	ennFlow.Message = fmt.Sprintf("构建任务%s进行中\n", stage.StageName)
	ennFlow.StageId = stage.StageId
	ennFlow.FlowId = stage.FlowId
	ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
	ennFlow.StageBuildId = queue.StageBuildLog.BuildId
	ennFlow.Flag = 2

	labelsStr := fmt.Sprintf("job-name=%s", jobName)
	labelSelector, err := labels.Parse(labelsStr)
	if err != nil {
		glog.Errorf("%s label parse failed: %v\n", method, err)
		return timeOut, err
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector.String(),
		Watch:         true,
	}
	// 请求watch api监听pod发生的事件
	watchInterface, err := queue.ImageBuilder.Client.Pods(namespace).Watch(listOptions)
	if err != nil {
		glog.Errorf("%s get pod watchInterface failed: %v\n", method, err)
		EnnFlowChan <- ennFlow
		return timeOut, err
	}
	podCount := 0
	for {
		select {
		case event, isOpen := <-watchInterface.ResultChan():
			if !isOpen {
				glog.Errorf("the pod watch the chan is closed\n")
				EnnFlowChan <- ennFlow
				return timeOut, fmt.Errorf("%s", "the pod watch the chan is closed")
			}

			pod, parseIsOk := event.Object.(*v1.Pod)
			if !parseIsOk {
				continue
			}

			glog.Infof("%s jobName=[%s],The pod [%s] event type=%s\n", method, jobName, pod.GetName(), event.Type)

			if len(pod.Status.ContainerStatuses) > 0 {
				IsContainerCreated(queue.ImageBuilder.ScmName, pod.Status.InitContainerStatuses)
				IsContainerCreated(queue.ImageBuilder.BuilderName, pod.Status.ContainerStatuses)
			}

			if queue.SelectAndUpdateStatus() {
				return timeOut, nil
			}

			if event.Type == watch.Added {
				glog.Infof("EventType ADDED %s jobName=[%s],The pod [%s] event type=%s\n", method, jobName, pod.GetName(), event.Type)

				if podCount >= 1 {
					return timeOut, nil
				}
				podCount = podCount + 1
				if pod.ObjectMeta.Name != "" {
					queue.StageBuildLog.PodName = pod.ObjectMeta.Name
					queue.StageBuildLog.NodeName = pod.Spec.NodeName
					queue.StageBuildLog.JobName = jobName
					queue.StageBuildLog.Status = common.STATUS_BUILDING
				}
				queue.UpdateStageBuidLogId()
				glog.Infof("%s pod of job %s is add with final status podName: %#v\n", method, jobName, pod.Status, pod.GetName())
				EnnFlowChan <- ennFlow

				continue
			} else if event.Type == watch.Modified {

				if pod.Status.Phase == v1.PodSucceeded {

					if pod.ObjectMeta.Name != "" {
						queue.StageBuildLog.PodName = pod.ObjectMeta.Name
						queue.StageBuildLog.NodeName = pod.Spec.NodeName
						queue.StageBuildLog.JobName = jobName
						queue.StageBuildLog.Status = common.STATUS_BUILDING
					}

					queue.UpdateStageBuidLogId()
					EnnFlowChan <- ennFlow
					return timeOut, nil
				} else if pod.Status.Phase == v1.PodFailed {
					if pod.ObjectMeta.Name != "" {
						queue.StageBuildLog.PodName = pod.ObjectMeta.Name
						queue.StageBuildLog.NodeName = pod.Spec.NodeName
						queue.StageBuildLog.JobName = jobName
						queue.StageBuildLog.Status = common.STATUS_BUILDING
					}
					queue.UpdateStageBuidLogId()
					EnnFlowChan <- ennFlow
					return timeOut, nil
				} else if pod.Status.Phase == v1.PodRunning {
					if pod.ObjectMeta.Name != "" {
						queue.StageBuildLog.PodName = pod.ObjectMeta.Name
						queue.StageBuildLog.NodeName = pod.Spec.NodeName
						queue.StageBuildLog.JobName = jobName
						queue.StageBuildLog.Status = common.STATUS_BUILDING
					}
					queue.UpdateStageBuidLogId()
					EnnFlowChan <- ennFlow
					return timeOut, nil
				} else if pod.Status.Phase == v1.PodUnknown {
					glog.Infof("The pod PodPhase=%s\n", v1.PodUnknown)
					return timeOut, nil

				} else if pod.Status.Phase == v1.PodPending {
					continue
				}

				glog.Infof("%s pod of job %s is Modified with final status: %#v\n", method, jobName, pod.Status)
				continue
			} else if event.Type == watch.Deleted {
				glog.Warningf("%s pod of job %s is deleted with final status: %#v\n", method, jobName, pod.Status)
				//收到deleted事件，pod可能被删除
				if pod.ObjectMeta.Name != "" {
					queue.StageBuildLog.PodName = pod.ObjectMeta.Name
					queue.StageBuildLog.NodeName = pod.Spec.NodeName
					queue.StageBuildLog.JobName = jobName
					queue.StageBuildLog.Status = common.STATUS_BUILDING
				}
				queue.UpdateStageBuidLogId()
				return timeOut, nil
			} else if event.Type == watch.Error {
				glog.Warningf("%s %s %v\n", method, "call watch api of pod of "+jobName+" error:", event.Object)
				//创建失败
				if pod.ObjectMeta.Name != "" {
					queue.StageBuildLog.PodName = pod.ObjectMeta.Name
					queue.StageBuildLog.NodeName = pod.Spec.NodeName
					queue.StageBuildLog.JobName = jobName
					queue.StageBuildLog.Status = common.STATUS_BUILDING
				}
				queue.UpdateStageBuidLogId()
				EnnFlowChan <- ennFlow
				return timeOut, nil
			}

		case <-time.After(6 * time.Minute):
			queue.ImageBuilder.StopJob(namespace, jobName, false, 0)
			timeOut = true
			return timeOut, nil

		}

	}
	return timeOut, nil

}

//WatchJob  watch  the job event fieldSelectorStr := "status.phase!=Succeeded,status.phase!=Failed"
func (queue *StageQueueNew) WatchOneJob(namespace, jobName string) error {

	method := "WatchOneJob"
	//labelsStr := fmt.Sprintf("stage-build-id=%s", queue.StageBuildLog.BuildId)
	//labelsSel, err := labels.Parse(labelsStr)
	fieldSelector, err := fields.ParseSelector(fmt.Sprintf("metadata.name=%s", jobName))
	if err != nil {
		glog.Errorf("%s fieldSelector parse failed:%v\n", method, err)
		return err
	}

	listOptions := metav1.ListOptions{
		//LabelSelector: labelsSel,
		FieldSelector: fieldSelector.String(),
		Watch:         true,
	}
GoOnWatch:
	watchInterface, err := queue.ImageBuilder.Client.BatchV1Client.Jobs(namespace).Watch(listOptions)
	if err != nil {
		glog.Errorf("%s,err: %v\n", method, err)
		return err
	}

	for {
		select {
		case event, isOpen := <-watchInterface.ResultChan():
			if isOpen == false {
				glog.Warningf("%s the watch job chain is close will watch again\n", method)
				goto GoOnWatch
			}

			dm, parseIsOk := event.Object.(*v1beta1.Job)
			if false == parseIsOk {
				glog.Errorf("%s job %s\n", method, ">>>>>>断言失败<<<<<<")
				continue
			}
			glog.Infof("%s job event.Type=%s, namespace=%s ,jobName=%s\n", method, event.Type, dm.GetNamespace(),
				dm.GetName())
			glog.Infof("%s job event.Status=%#v\n", method, dm.Status)
			if event.Type == watch.Added {
				if queue.SelectAndUpdateStatus() {
					return fmt.Errorf("%s", "the user stop build")
				}
				//收到deleted事件，job可能被第三方删除
				glog.Infof("%s %s,status:%v\n", method, "收到ADD事件,开始起job进行构建", dm.Status)
				GetEnnFlow(dm, common.STATUS_BUILDING)
				if dm.Status.Failed >= 1 || dm.Status.Succeeded >= 1 {
					return nil
				}
				continue
			} else if event.Type == watch.Deleted {
				if queue.SelectAndUpdateStatus() {
					return fmt.Errorf("%s", "the user stop build")
				}
				//收到deleted事件，job可能被第三方删除
				glog.Errorf("%s  %s status:%v\n", method, " 收到deleted事件，job可能被第三方删除", dm.Status)
				return nil
				//成功时并且已经完成时
			} else if dm.Status.Succeeded >= 1 &&
				dm.Status.CompletionTime != nil && len(dm.Status.Conditions) != 0 {
				if queue.SelectAndUpdateStatus() {
					return fmt.Errorf("%s", "the user stop build")
				}
				glog.Infof("%s %s,status:%v\n", method, "构建成功", dm.Status)
				GetEnnFlow(dm, common.STATUS_BUILDING)
				return nil
				//} else if dm.Status.Failed >=1 && dm.Spec.Completions == Int32Toint32Point(1) &&
				//	dm.Status.CompletionTime == nil && dm.Status.Succeeded==0{
			} else if dm.Status.Failed >= 1 {
				if queue.SelectAndUpdateStatus() {
					return fmt.Errorf("%s", "the user stop build")
				}
				glog.Infof("%s %s,status:%v\n", method, "构建失败", dm.GetName())
				GetEnnFlow(dm, common.STATUS_BUILDING)
				return nil
				//手动停止job
			} else if dm.Spec.Parallelism == Int32Toint32Point(0) {
				if queue.SelectAndUpdateStatus() {
					return fmt.Errorf("%s", "the user stop build")
				}
				//有依赖服务，停止job时 不是手动停止 1 表示手动停止
				if dm.ObjectMeta.Labels["enncloud-builder-succeed"] != "1" && dm.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "true" {
					glog.Infof("%s %s,status:%v\n", method, "用户停止了构建任务", dm)
					GetEnnFlow(dm, common.STATUS_BUILDING)
					return nil
					//没有依赖服务时
				} else {
					GetEnnFlow(dm, common.STATUS_BUILDING)
					glog.Infof("%s %s,status:%v\n", method, "job执行失败程序发送了停止构建任务的命令", dm)
					return nil
				}
			}

		}
	}

}
