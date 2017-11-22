package controllers

import (
	"github.com/golang/glog"
	"encoding/json"
	"dev-flows-api-golang/models"
	"fmt"
	"dev-flows-api-golang/models/common"
	"k8s.io/client-go/1.4/pkg/api/v1"
	"text/template"
	"io"
	"k8s.io/client-go/1.4/pkg/fields"
	"k8s.io/client-go/1.4/pkg/api"
	"strings"
	apiv1 "k8s.io/client-go/1.4/pkg/api/v1"
)

//判断stage构建状态，如果已为失败或构建完成，则从ElasticSearch中获取日志
//如构建中，则从k8s API获取实时日志
type BuildMessage struct {
	FlowId        string `json:"flowId"`
	FlowBuildId   string `json:"flowBuildId"`
	StageBuildId  string `json:"stageBuildId"`
	StageId       string `json:"stageId"`
	ContainerName string `json:"containerName"`
	PodName       string `json:"podName"`
	JobName       string `json:"jobName"`
	NodeName      string `json:"nodeName"`
	Status        int    `json:"status"`
	ControllerUid string `json:"controller_id"`
	LogData       string `json:"logData"`
	ClusterId     string `json:"cluster_id"`
}

const CILOG = "ciLogs"
const TailLines = 200
const POD_INIT = "pod-init"
const GET_LOG_RETRY_COUNT = 3
const GET_LOG_RETRY_MAX_INTERVAL = 30

//GetStageBuildLogsFromK8S
func GetStageBuildLogsFromK8S(buildMessage EnnFlow, conn Conn) {
	glog.Infoln("开始从kubernetes搜集实时日志======================>>")
	method := "GetStageBuildLogsFromK8S"

	imageBuilder := models.NewImageBuilder()

	build, err := GetValidStageBuild(buildMessage.FlowId, buildMessage.StageId, buildMessage.StageBuildId)
	if err != nil {
		glog.Errorf("%s GetValidStageBuild failed===>%v\n", method, err)
		Send(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, err), conn)
		return
	}
	glog.Infoln("build info ==========>>jobName:%s\n", build.JobName)
	//正在等待中
	if build.Status == common.STATUS_WAITING {
		buildStatus := struct {
			BuildStatus string `json:"buildStatus"`
		}{
			BuildStatus: "waiting",
		}
		glog.Infof("%s the stage is onwaiting ===>%v\n", method, build)
		Send(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, buildStatus.BuildStatus), conn)
		return
	}

	if build.PodName == "" {
		podName, err := imageBuilder.GetPodName(build.Namespace, build.JobName)
		if err != nil || podName == "" {
			glog.Errorf("%s 获取构建任务信息失败 get job name=[%s] pod name failed:======>%v\n", method, build.JobName, err)
			Send(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, "获取构建任务信息失败"), conn)
			return
		}
		models.NewCiStageBuildLogs().UpdatePodNameById(podName, build.BuildId)
		GetLogsFromK8S(imageBuilder, build.Namespace, build.JobName, build.PodName, conn)
		return
	}

	GetLogsFromK8S(imageBuilder, build.Namespace, build.JobName, build.PodName, conn)
	return

}

//GetLogsFromK8S
func GetLogsFromK8S(imageBuilder *models.ImageBuilder, namespace, jobName, podName string, conn Conn) {

	WatchEvent(imageBuilder, namespace, podName, conn)

	WaitForLogs(imageBuilder, namespace, podName, models.SCM_CONTAINER_NAME, conn)

	WaitForLogs(imageBuilder, namespace, podName, models.BUILDER_CONTAINER_NAME, conn)

}

func WatchEvent(imageBuild *models.ImageBuilder, namespace, podName string, conn Conn) {
	if podName == "" {
		glog.Errorf("the podName is empty")
	}
	method := "WatchEvent"
	glog.Infoln("Begin watch kubernetes Event=====>>")
	fieldSelector, err := fields.ParseSelector(fmt.Sprintf("involvedObject.kind=pod,involvedObject.name=%s", podName))
	if nil != err {
		glog.Errorf("%s: Failed to parse field selector: %v\n", method, err)
		return
	}
	options := api.ListOptions{
		FieldSelector: fieldSelector,
		Watch:         true,
	}

	// 请求watch api监听pod发生的事件
	watchInterface, err := imageBuild.Client.Events(namespace).Watch(options)
	if err != nil {
		glog.Errorf("get event watchinterface failed===>%v\n", method, err)
		Send(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, err), conn)
		return
	}
	//TODO pod 不存在的情况
	for {
		select {
		case event, isOpen := <-watchInterface.ResultChan():
			if !isOpen {
				glog.Infof("%s the event watch the chan is closed\n", method)
				Send(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, err), conn)
				break
			}
			glog.Infof("the pod event type=%s\n", event.Type)
			EventInfo, ok := event.Object.(*apiv1.Event)
			if ok {
				if strings.Index(EventInfo.Message, "PodInitializing") > 0 {
					Send(imageBuild.EventToLog(*EventInfo), conn)
					continue
				}
				Send(imageBuild.EventToLog(*EventInfo), conn)
			}

		}

	}

	return

}

func Int64Toint64Point(input int64) *int64 {
	tmp := new(int64)
	*tmp = int64(input)
	return tmp

}

//WaitForLogs websocket get logs
func WaitForLogs(imageBuild *models.ImageBuilder, namespace, podName, containerName string, conn Conn) {
	method := "WaitForLogs"
	follow := false
	previous := true
	if conn.Conn != nil {
		follow = true
		previous = false
	}
	opt := &v1.PodLogOptions{
		Container:  containerName,
		TailLines:  Int64Toint64Point(TailLines),
		Previous:   previous, //
		Follow:     follow,
		Timestamps: true,
	}
	//websocket的请求
	if conn.Conn != nil {
		readCloser, err := imageBuild.Client.Pods(namespace).GetLogs(podName, opt).Stream()
		if err != nil {
			glog.Errorf("%s socket get pods log readCloser faile from kubernetes:==>%v\n", method, err)
			Send(fmt.Sprintf(`<font color="red">[Enn Flow API Error] Failed to get log of %s</font>\n`, podName), conn)
			return
		}

		if containerName == models.BUILDER_CONTAINER_NAME {
			//glog.Infof("==============>> user stop_receive_log user <<===========\n")
			//Send( fmt.Sprintf(`<font color="#ffc20e">[Enn Flow API] 您停止了接收日志</font>\n`),conn)
			//return
			Send("---------------------------------------------------", conn)
			Send("--- 子任务容器: 仅显示最近 "+fmt.Sprintf("%d", TailLines)+" 条日志 ---", conn)
			Send("---------------------------------------------------", conn)
		}
		data := make([]byte, 1024*1024, 1024*1024)
		for {
			n, err := readCloser.Read(data)
			if nil != err {
				if err == io.EOF {
					glog.Infof("%s [Enn Flow API ] finish get log of %s.%s!\n", method, podName, containerName)
					glog.Infof("==========>>Get log successfully from socket.!!<<============\n")
					Send(fmt.Sprintf(`<font color="red">[Enn Flow API ] 日志读取结束  %s.%s!</font>\n`, podName, containerName), conn)
					return
				}
				return
			}
			glog.Infof("=======the log is ===>>string(data[:n])==>%s\n", string(data[:n]))
			logMessage := &LogMessage{
				Name: containerName,
				Log:  template.HTMLEscapeString(string(data[:n])),
			}
			message, err := json.Marshal(logMessage)
			if nil != err {
				glog.Warningf("%s [Enn Flow API Error] Parse container log failed, container name is %s.%s Error:==>%v\n", method, podName, containerName, err)
				Send(fmt.Sprintf(`<font color="red">[Enn Flow API Error] 日志读取失败,请重试  %s.%s!</font>\n`, podName, containerName), conn)
				return
			}

			Send(message, conn)

		}
	} else {

		glog.Errorf("the socket is nil\n")

	}
	return

}

func GetValidStageBuild(flowId, stageId, stageBuildId string) (models.CiStageBuildLogs, error) {
	var build models.CiStageBuildLogs
	method := "SocketLogController.GetValidStageBuild"
	stage, err := models.NewCiStage().FindOneById(stageId)
	if err != nil {
		glog.Errorf("%s find stage by stageId failed or not exist from database: %v\n", method, err)
		return build, err
	}
	if flowId != stage.FlowId {

		return build, fmt.Errorf("Stage is not %s in the flow\n", stageId)
	}

	build, err = models.NewCiStageBuildLogs().FindOneById(stageBuildId)
	if err != nil {
		glog.Errorf("%s find stagebuild by StageBuildId failed or not exist from database: %v\n", method, err)
		return build, err
	}

	if stage.StageId != build.StageId {

		return build, fmt.Errorf("Build is not %s one of the stage \n", build.BuildId)

	}

	return build, nil
}
