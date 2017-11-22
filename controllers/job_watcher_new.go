package controllers

import (
	"net/http"
	"github.com/golang/glog"
	"k8s.io/client-go/1.4/pkg/labels"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/api/v1"
	"k8s.io/client-go/1.4/pkg/watch"
	"encoding/json"
	"fmt"
	"dev-flows-api-golang/models/common"
	"dev-flows-api-golang/modules/client"
	v1beta1 "k8s.io/client-go/1.4/pkg/apis/batch/v1"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"sync"
)

type EnnFlow struct {
	FlowId        string `json:"flowId"`
	FlowBuildId   string `json:"flowBuildId"`
	StageId       string `json:"stageId"`
	StageBuildId  string `json:"stageBuildId"`
	BuildStatus   int `json:"buildStatus"`
	Status        int `json:"status"`
	Message       string `json:"message"`
	CodeBranch    string `json:"codeBranch"`
	Flag          int `json:"flag"` //1 表示flow 2表示stage
	Namespace     string `json:"namespace"`
	LoginUserName string `json:"loginUserName"`
	UserNamespace string `json:"userNamespace"`
	Event         string `json:"event"`
}

var SOCKETS_OF_BUILD_MAPPING_MUTEX sync.RWMutex
//前端记得要关闭websoccket 一个flow对应一个websocket
type Conn struct {
	Conn net.Conn
	Op   ws.OpCode
}

var SOCKETS_OF_FLOW_MAPPING_NEW = make(map[string]Conn, 0)

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
			go func() {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					glog.Errorf("读取客户端发送的数据包失败 connect JobLogSocket websocket failed: msg:%s,err:%v\n", msg, err)
					wsutil.WriteServerMessage(conn, op, []byte(`<font color="red">[Enn Flow API Error] 读取客户端发送的数据包失败</font>`))
					return
				}

				con.Op = op

				err = json.Unmarshal(msg, &buildMessage)
				if err != nil {
					glog.Errorf("反系列化数据库包失败 msg:%s,err:%v========err:%v\n", msg, err)
					wsutil.WriteServerMessage(conn, op, []byte(`<font color="red">[Enn Flow API Error] 反系列化数据库包失败</font>`))
					return
				}

				if message := CheckLogData(buildMessage); message != "" {
					glog.Errorf("Missing parameters====>>:[%v]\n", buildMessage)
					Send(message, con)
					return
				}

				GetStageBuildLogsFromK8S(buildMessage, con)

			}()
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
			//defer conn.Close()
			var flow EnnFlow
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				glog.Errorf("connect JobWatcherSocket websocket failed: msg:%s,err:%v\n", msg, err)
				flow.Status = 400
				flow.BuildStatus = common.STATUS_FAILED
				flow.Message = "读取客户端信息失败"
				data, _ := json.Marshal(flow)
				wsutil.WriteServerMessage(conn, op, data)
				return
			}

			err = json.Unmarshal(msg, &flow)
			if err != nil {
				glog.Errorf("msg:%v========err:%v\n", msg, err)
				flow.Status = 400
				flow.Message = "json unmarshal failed"
				data, _ := json.Marshal(flow)
				wsutil.WriteServerMessage(conn, op, data)
				return
			}

			if flow.FlowId != "" {
				SOCKETS_OF_BUILD_MAPPING_MUTEX.Lock()
				//存websocket,通过flowId获取某个Ennflow的websocket
				if _, ok := SOCKETS_OF_FLOW_MAPPING_NEW[flow.FlowId]; !ok {
					var connOfFlow Conn
					connOfFlow.Conn = conn
					connOfFlow.Op = op
					SOCKETS_OF_FLOW_MAPPING_NEW[flow.FlowId] = connOfFlow

				}
				flow.Status = http.StatusOK
				flow.Message = "建立websocket成功"
				data, _ := json.Marshal(flow)
				wsutil.WriteServerMessage(conn, op, data)
				SOCKETS_OF_BUILD_MAPPING_MUTEX.Unlock()
			} else {
				glog.Errorf("FlowId is empty")
				flow.Status = 400
				flow.Message = "FlowId is empty"
				data, _ := json.Marshal(flow)
				wsutil.WriteServerMessage(conn, op, data)
			}
			return
		}),
	}
}

//WatchJob  watch  the job event
func (queue *StageQueueNew) WatchJob(namespace, jobName string) *v1beta1.Job {
	method := "WatchJob"
	var job *v1beta1.Job

	glog.Infof("%s begin watch job jobName=[%s]  namespace=[%s]\n", method, jobName, namespace)

	labelsStr := fmt.Sprintf("stage-build-id=%s", queue.StageBuildLog.BuildId)
	labelsSel, err := labels.Parse(labelsStr)
	if err != nil {
		glog.Errorf("%s label parse failed==>:%v\n", method, err)
		job.Status.Conditions[0].Status = v1.ConditionUnknown
		return job
	}
	listOptions := api.ListOptions{
		LabelSelector: labelsSel,
		Watch:         true,
	}
	watchInterface, err := client.KubernetesClientSet.BatchClient.Jobs(namespace).Watch(listOptions)
	if err != nil {
		glog.Errorf("%s,err: %v\n", method, err)
		job.Status.Conditions[0].Status = v1.ConditionUnknown
		return job
	}

	for {
		select {
		case event, isOpen := <-watchInterface.ResultChan():
			if isOpen == false {
				glog.Errorf("%s the watch job chain is close\n", method)
				job.Status.Conditions[0].Status = v1.ConditionUnknown
				return job
			}
			dm, parseIsOk := event.Object.(*v1beta1.Job)
			if false == parseIsOk {
				glog.Errorf("%s job %s\n", method, ">>>>>>断言失败<<<<<<")
				job.Status.Conditions[0].Status = v1.ConditionUnknown
				return job
			}

			glog.Infof("%s job event.Type=%s\n", method, event.Type)
			glog.Infof("%s job event.Status=%#v\n", method, dm.Status)

			if event.Type == watch.Added {
				//收到deleted事件，job可能被第三方删除
				glog.Infof("%s %s,status:%v\n", method, "收到ADD事件,开始起job进行构建", dm.Status)
				//成功时并且已经完成时
			} else if event.Type == watch.Deleted {
				//收到deleted事件，job可能被第三方删除
				glog.Errorf("%s  %s status:%v\n", method, " 收到deleted事件，job可能被第三方删除", dm.Status)
				return dm
				//成功时并且已经完成时
			} else if dm.Status.Succeeded >= 1 &&
				dm.Status.CompletionTime != nil && len(dm.Status.Conditions) != 0 {
				glog.Infof("%s %s,status:%v\n", method, "构建成功", dm.Status)
				return dm
				//} else if dm.Status.Failed >=1 && dm.Spec.Completions == Int32Toint32Point(1) &&
				//	dm.Status.CompletionTime == nil && dm.Status.Succeeded==0{
			} else if dm.Status.Failed >= 1 {
				glog.Infof("%s %s,status:%v\n", method, "构建失败", dm.Status)
				return dm
				//手动停止job
			} else if dm.Spec.Parallelism == Int32Toint32Point(0) {
				//有依赖服务，停止job时 不是手动停止 1 表示手动停止
				if dm.ObjectMeta.Labels["enncloud-builder-succeed"] != "1" && dm.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "true" {
					glog.Infof("%s %s,status:%v\n", method, "用户停止了构建任务", dm)
					return dm
					//没有依赖服务时
				} else {
					glog.Infof("%s %s,status:%v\n", method, "job执行失败程序发送了停止构建任务的命令", dm)
					return dm
				}
			}

		}
	}
	return job
}

func Send(flow interface{}, conn Conn) {
	if conn.Conn != nil {
		data, _ := json.Marshal(flow)
		wsutil.WriteServerMessage(conn.Conn, conn.Op, data)
		return
	}

}

func Int32Toint32Point(input int32) *int32 {
	tmp := new(int32)
	*tmp = int32(input)
	return tmp

}
