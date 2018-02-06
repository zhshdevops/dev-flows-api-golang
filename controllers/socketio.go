package controllers

import (
	"github.com/golang/glog"
	"dev-flows-api-golang/models"
	"fmt"
	"dev-flows-api-golang/models/common"
	"k8s.io/client-go/pkg/api/v1"
	"text/template"
	"io"
	"k8s.io/apimachinery/pkg/fields"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/pkg/api"
	"strings"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"time"
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

const TailLines = 200
const POD_INIT = "pod-init"
const GET_LOG_RETRY_COUNT = 3
const GET_LOG_RETRY_MAX_INTERVAL = 30

//GetStageBuildLogsFromK8S
func GetStageBuildLogsFromK8S(buildMessage EnnFlow, conn Conn) {

	glog.Infoln("begin get log from kubernetes ")

	method := "GetStageBuildLogsFromK8S"

	imageBuilder := models.NewImageBuilder()

	build, err := GetValidStageBuild(buildMessage.FlowId, buildMessage.StageId, buildMessage.StageBuildId)
	if err != nil {
		glog.Errorf("%s GetValidStageBuild failed:%v\n", method, err)
		SendLog(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, err), conn)
		return
	}

	//正在等待中
	if build.Status == common.STATUS_WAITING {
		buildStatus := struct {
			BuildStatus string `json:"buildStatus"`
		}{
			BuildStatus: "waiting",
		}
		glog.Infof("%s the stage is onwaiting ===>%v\n", method, build)
		SendLog(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, buildStatus.BuildStatus), conn)
		return
	}

	podName, err := imageBuilder.GetPodName(build.Namespace, build.JobName, build.BuildId)
	if err != nil || podName == "" {
		glog.Errorf("%s 获取构建任务信息失败 get job name=[%s] pod name failed:%v\n", method, build.JobName, err)
		SendLog(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, "构建任务不存在或者日志信息已过期!"), conn)
		return
	}

	build.PodName = podName
	models.NewCiStageBuildLogs().UpdatePodNameById(podName, build.BuildId)
	GetLogsFromK8S(imageBuilder, build.Namespace, build.JobName, podName, conn, build.BuildId)
	return

}

//GetLogsFromK8S
func GetLogsFromK8S(imageBuilder *models.ImageBuilder, namespace, jobName, podName string, conn Conn, buildId string) {

	WatchEvent(imageBuilder, namespace, podName, conn)

	WaitForLogs(imageBuilder, namespace, podName, jobName, models.SCM_CONTAINER_NAME, conn, buildId)

	WaitForLogs(imageBuilder, namespace, podName, jobName, models.BUILDER_CONTAINER_NAME, conn, buildId)

}

func WatchEvent(imageBuild *models.ImageBuilder, namespace, podName string, conn Conn) {
	if podName == "" {
		glog.Errorf("the podName is empty")
		SendLog(fmt.Sprintf("%s", `<font color="red">[Enn Flow API Error]构建任务启动中</font>`), conn)
		return
	}
	method := "WatchEvent"

	fieldSelector, err := fields.ParseSelector(fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", podName))
	if nil != err {
		glog.Errorf("%s: Failed to parse field selector: %v\n", method, err)
		SendLog(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, err), conn)
		return
	}
	options := metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
		Watch:         true,
	}

	// 请求watch api监听pod发生的事件
	watchInterface, err := imageBuild.Client.Events(namespace).Watch(options)
	if err != nil {
		glog.Errorf("get event watchinterface failed===>%v\n", method, err)
		SendLog(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, err), conn)
		return
	}

	for {
		select {

		case event, isOpen := <-watchInterface.ResultChan():
			if !isOpen {
				glog.Infof("%s the event watch the chan is closed\n", method)
				SendLog(fmt.Sprintf(`<font color="red">[Enn Flow API Error]%s</font>`, err), conn)
				break
			}

			EventInfo, ok := event.Object.(*apiv1.Event)
			if ok {
				if strings.Index(EventInfo.Message, "PodInitializing") > 0 {
					SendLog(imageBuild.EventToLog(*EventInfo), conn)
					continue
				}
				SendLog(imageBuild.EventToLog(*EventInfo), conn)
			}
		case <-time.After(5 * time.Second):
			return

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
func WaitForLogs(imageBuild *models.ImageBuilder, namespace, podName, jobName, containerName string, conn Conn, buildId string) {
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
			SendLog(fmt.Sprintf(`<font color="red">[Enn Flow API Error]获取日志失败</font>`), conn)
			return
		}

		if containerName == models.SCM_CONTAINER_NAME {
			SendLog("---------------------------------------------------", conn)
			SendLog("--- 子任务容器: 仅显示最近 "+fmt.Sprintf("%d", TailLines)+" 条日志 ---", conn)
			SendLog("---------------------------------------------------", conn)
		}
		data := make([]byte, 1024*1024, 1024*1024)
		for {
			n, err := readCloser.Read(data)
			if nil != err {
				if err == io.EOF {
					glog.Infof("%s [Enn Flow API ] finish get log of %s.%s!\n", method, podName, containerName)
					glog.Infof("==========>>Get log successfully from socket.!!<<============\n")
					if containerName == models.BUILDER_CONTAINER_NAME {
						for {
							buildInfo, err := models.NewCiStageBuildLogs().FindOneById(buildId)
							if err != nil {
								glog.Errorf("get build info failed:%v\n", err)

							}
							if buildInfo.Status == common.STATUS_FAILED || buildInfo.Status == common.STATUS_SUCCESS {
								SendLog(fmt.Sprintf("%s", `<font color="#ffc20e">[Enn Flow API ] 日志读取结束</font>`), conn)
								break
							}
							continue
						}

					}
					return
				} else {
					SendLog(fmt.Sprintf(`<font color="red">[Enn Flow API ]获取日志失败%v!</font>`, err), conn)
					return
				}

			}

			job, _ := imageBuild.GetJob(namespace, jobName)
			_, ok := job.GetLabels()[common.MANUAL_STOP_LABEL]
			if job.Status.Failed >= 1 || job.Status.Succeeded >= 1 || ok {
				SendLog(fmt.Sprintf("%s", `<font  style="display:none;">[Enn Flow API ] 日志读取结束</font>`), conn)
				return
			}
			if strings.Contains(string(data[:n]), "rpc error:") {

				log := fmt.Sprintf(`<font color="#ffc20e">[%s]</font> %s <br/>`, time.Now().Format("2006/01/02 15:04:05"), string(data[:n]))
				SendLog(log, conn)

			} else {
				logInfo := strings.SplitN(template.HTMLEscapeString(string(data[:n])), "\n", -1)
				for _, logline := range logInfo {
					loglineArr := strings.SplitN(template.HTMLEscapeString(logline), " ", 2)
					if len(loglineArr) == 2 {
						logTime, _ := time.Parse(time.RFC3339, loglineArr[0])
						log := fmt.Sprintf(`<font color="#ffc20e">[%s]</font> %s <br/>`, logTime.Add(8 * time.Hour).Format("2006/01/02 15:04:05"), loglineArr[1])
						SendLog(log, conn)
					}

				}
			}

			//logInfo := strings.SplitN(template.HTMLEscapeString(string(data[:n])), " ", 2)
			//
			//log := fmt.Sprintf(`<font color="#ffc20e">[%s]</font> %s`, logInfo[0], logInfo[1])
			//SendLog(log, conn)

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
