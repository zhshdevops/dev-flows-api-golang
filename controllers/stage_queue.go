package controllers

import (
	"dev-flows-api-golang/models"
	"dev-flows-api-golang/models/user"
	"github.com/golang/glog"
	sqlstatus "dev-flows-api-golang/models/sql/status"
	"net/http"
	"dev-flows-api-golang/models/common"
	"dev-flows-api-golang/util/uuid"
	"fmt"
	"dev-flows-api-golang/models/cluster"
	"dev-flows-api-golang/modules/client"
	"time"
	"os"
	"encoding/json"
	"dev-flows-api-golang/ci/coderepo"
	"k8s.io/client-go/1.4/pkg/apis/batch/v1"
	apiv1 "k8s.io/client-go/1.4/pkg/api/v1"
	"strings"
)

type StageQueueNew struct {
	StageList        []models.CiStages
	User             *user.UserModel
	BuildReqbody     EnnFlow
	FlowId           string
	TotalStage       int64
	Event            string
	Namespace        string
	LoginUserName    string
	CiFlow           *models.CiFlows
	FlowbuildLog     *models.CiFlowBuildLogs
	StageBuildLog    *models.CiStageBuildLogs
	ImageBuilder     *models.ImageBuilder
	CurrentNamespace string
}

func NewStageQueueNew(buildReqbody EnnFlow, event, namespace, loginUserName, flowId string, imageBuilder *models.ImageBuilder) *StageQueueNew {
	method := "NewStageQueueNew"
	u := user.NewUserModel()
	parseCode, err := u.GetByName(loginUserName)
	if err != nil || parseCode == sqlstatus.SQLErrNoRowFound {
		glog.Errorf("%s get Login User failed:%v\n", method, err)

	}

	flowBuildlog := &models.CiFlowBuildLogs{}
	stageBuildLog := &models.CiStageBuildLogs{}

	queue := &StageQueueNew{
		User:             u,
		BuildReqbody:     buildReqbody,
		FlowId:           flowId,
		Event:            event,
		Namespace:        u.Namespace, //取当前登录用户的空间进行构建
		FlowbuildLog:     flowBuildlog,
		StageBuildLog:    stageBuildLog,
		ImageBuilder:     imageBuilder,
		LoginUserName:    loginUserName,
		CurrentNamespace: namespace,
	}

	ciFlow := models.NewCiFlows()
	//要根据当前的空间去查询EnnFlow
	flow, err := ciFlow.FindFlowById(namespace, flowId)
	if err != nil {
		parseCode, _ := sqlstatus.ParseErrorCode(err)
		if parseCode == sqlstatus.SQLErrNoRowFound {
			buildReqbody.Message = "Flow cannot be found"
			buildReqbody.Status = http.StatusNotFound
			buildReqbody.BuildStatus = common.STATUS_FAILED
			glog.Errorf("%s %s %v\n", method, "Failed to find flow "+flowId+" Of "+namespace, err)
			FlowMapping.Send(buildReqbody)
			return queue
		} else {
			buildReqbody.Message = "Flow cannot be found"
			buildReqbody.Status = http.StatusInternalServerError
			buildReqbody.BuildStatus = common.STATUS_FAILED
			glog.Errorf("%s %s %v\n", method, "Failed to find flow "+flowId+" Of "+namespace, err)
			FlowMapping.Send(buildReqbody)
			return queue
		}
	}

	if flow.Name == "" {
		buildReqbody.Message = "Flow cannot be found"
		buildReqbody.Status = http.StatusNotFound
		buildReqbody.BuildStatus = common.STATUS_FAILED
		glog.Errorf("%s %s %v\n", method, "Failed to find flow "+flowId+" Of "+namespace, err)
		FlowMapping.Send(buildReqbody)
		return queue
	}

	queue.CiFlow = &flow

	stageId := buildReqbody.StageId
	var stageList []models.CiStages
	stageList = make([]models.CiStages, 0)
	stageServer := models.NewCiStage()
	//指定stage
	if stageId != "" {
		stage, err := stageServer.FindOneById(stageId)
		if err != nil {
			parseCode, _ := sqlstatus.ParseErrorCode(err)
			if parseCode == sqlstatus.SQLErrNoRowFound {
				buildReqbody.Message = "not found the stage!"
				buildReqbody.Status = http.StatusNotFound
				buildReqbody.BuildStatus = common.STATUS_FAILED
				FlowMapping.Send(buildReqbody)
				glog.Errorf("%s %s %v\n", method, "Failed to find stage "+stageId+" Of "+namespace, err)
				return queue
			} else {
				buildReqbody.Message = "stage cannot be found"
				buildReqbody.Status = http.StatusInternalServerError
				buildReqbody.BuildStatus = common.STATUS_FAILED
				FlowMapping.Send(buildReqbody)
				glog.Errorf("%s %s %v\n", method, "Failed to find stage "+stageId+" Of "+namespace, err)
				return queue
			}

		}
		if stage.StageName == "" {
			buildReqbody.Message = "not found the stage!"
			buildReqbody.Status = http.StatusNotFound
			buildReqbody.BuildStatus = common.STATUS_FAILED
			FlowMapping.Send(buildReqbody)
			glog.Errorf("%s %s %v\n", method, "Failed to find stage "+stageId+" Of "+namespace, err)
			return queue
		}

		if stage.FlowId != flowId {
			glog.Errorf("%s %s\n", method, "Stage does not belong to Flow")
			buildReqbody.Message = "Stage does not belong to Flow!"
			buildReqbody.Status = http.StatusBadRequest
			buildReqbody.BuildStatus = common.STATUS_FAILED
			FlowMapping.Send(buildReqbody)
			return queue
		}
		stages, _, err := stageServer.FindAllFlowByFlowId(flowId)
		if err != nil {
			glog.Errorf("%s FindFirstOfFlow find stage failed from database: %v\n", method, err)
			buildReqbody.Message = "not find the stage of flow " + flowId
			buildReqbody.Status = http.StatusBadRequest
			buildReqbody.BuildStatus = common.STATUS_FAILED
			FlowMapping.Send(buildReqbody)
			return queue
		}
		stageList = append(stageList, stage)
		for _, stageInfo := range stages {
			if stageInfo.Seq > stage.Seq {
				stageList = append(stageList, stageInfo)
			}
		}

		queue.TotalStage = int64(len(stageList))
		//不指定stage
	} else {
		stages, total, err := stageServer.FindAllFlowByFlowId(flowId)
		if err != nil {
			glog.Errorf("%s FindFirstOfFlow find stage failed from database: %v\n", method, err)
			buildReqbody.Message = "not find the stage of flow " + flowId
			buildReqbody.Status = http.StatusBadRequest
			buildReqbody.BuildStatus = common.STATUS_FAILED
			FlowMapping.Send(buildReqbody)
			return queue
		}
		queue.TotalStage = total

		stageList = append(stageList, stages...)

	}

	queue.StageList = stageList

	return queue

}

func (queue *StageQueueNew) LengthOfStage() int {

	return len(queue.StageList)

}

//第一次构建插入数据库
func (queue *StageQueueNew) InsertLog() error {
	//开始执行 把执行日志插入到数据库
	flowBuildId := uuid.NewFlowBuildID()
	stageBuildId := uuid.NewStageBuildID()
	codeBranch := ""
	if queue.Event != "" {
		//有待修改
		codeBranch = queue.Event //webhook 触发的 代码分枝
	}

	var stageBuildRec models.CiStageBuildLogs
	now := time.Now()
	stageBuildRec.FlowBuildId = flowBuildId
	stageBuildRec.StageId = queue.StageList[0].StageId
	stageBuildRec.BuildId = stageBuildId
	stageBuildRec.StageName = queue.StageList[0].StageName
	stageBuildRec.Status = common.STATUS_WAITING
	stageBuildRec.StartTime = now
	stageBuildRec.Namespace = queue.User.Namespace
	stageBuildRec.IsFirst = 1
	stageBuildRec.BranchName = codeBranch
	//如果不是事件推送,就执行DefaultBranch
	if codeBranch == "" {
		stageBuildRec.BranchName = queue.StageList[0].DefaultBranch
	}

	//代码分支，前端触发
	if queue.BuildReqbody.CodeBranch != "" {
		stageBuildRec.BranchName = queue.BuildReqbody.CodeBranch
	}
	stageBuildRec.CreationTime = now
	// 如果 flow 开启了 “统一使用代码库” ，把构建时指定的 branch 保存在 stage 的 option 中
	// 在这个 stage 构建完成后，传递给下一个 stage 统一使用代码仓库UniformRepo
	if queue.CiFlow.UniformRepo == 0 {
		//	queue.StageList[0].Option = queue.BuildReqbody.Options
	}

	var flowBuildRec models.CiFlowBuildLogs
	flowBuildRec.BuildId = flowBuildId
	flowBuildRec.FlowId = queue.FlowId
	flowBuildRec.UserId = queue.User.UserID
	flowBuildRec.CreationTime = now
	flowBuildRec.StartTime = now
	flowBuildRec.Branch = stageBuildRec.BranchName
	flowBuildRec.Creater = queue.LoginUserName
	//InsertBuildLog will update 执行状态
	err := models.InsertBuildLog(&flowBuildRec, &stageBuildRec, queue.StageList[0].StageId)
	if err != nil {
		return err
	}
	queue.StageBuildLog = &stageBuildRec
	queue.FlowbuildLog = &flowBuildRec
	return nil
}

func (queue *StageQueueNew) WaitForBuildToComplete(job *v1.Job, stage models.CiStages) int {

	method := "StageQueueNew/WaitForBuildToComplete"
	job, _ = queue.ImageBuilder.GetJob(job.GetNamespace(), job.GetName())
	pod := apiv1.Pod{}
	var err error
	var errMsg string

	statusCode := 1

	if job.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "true" && job.ObjectMeta.Labels["enncloud-builder-succeed"] != "1" {
		statusCode = 1
	} else if job.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "Timeout-OrRunFailed" && job.ObjectMeta.Labels["enncloud-builder-succeed"] != "0" {
		statusCode = 1
	} else if len(job.Status.Conditions) != 0 {
		if job.Status.Failed > 0 {
			glog.Infof("Run job failed:%v\n", job.Status)
			statusCode = 1
		} else if job.Status.Succeeded > 0 {
			glog.Infof("Run job success:%v\n", job.Status)
			statusCode = 0
		}
	}

	glog.Infof("%s Wait ended normally... and the job status: %#v\n", method, job.Status)

	queue.StageBuildLog.EndTime = time.Now()

	if statusCode == 0 {
		queue.StageBuildLog.Status = common.STATUS_SUCCESS
	} else {
		queue.StageBuildLog.Status = common.STATUS_FAILED
	}

	if pod.ObjectMeta.Name == "" {
		pod, err = queue.ImageBuilder.GetPodByPodName(job.ObjectMeta.Namespace, job.ObjectMeta.Name,
			queue.StageBuildLog.PodName, queue.StageBuildLog.BuildId)
		if err != nil {
			glog.Errorf("%s get pod from kubernetes cluster failed:%v\n", method, err)
		}

		if pod.GetName() != "" {
			//执行失败时，生成失败原因
			if statusCode == 1 {
				if len(pod.Status.ContainerStatuses) > 0 {
					for _, sontainerStatus := range pod.Status.ContainerStatuses {
						if sontainerStatus.Name == queue.ImageBuilder.BuilderName && sontainerStatus.State.Terminated != nil {
							errMsg = fmt.Sprintf(`运行构建的容器异常退出：exit code=%d，退出原因为:%s`, sontainerStatus.State.Terminated.ExitCode,
								sontainerStatus.State.Terminated.Message)
						}
					}
					if errMsg == "" && len(pod.Status.InitContainerStatuses) > 0 {
						for _, scmStatus := range pod.Status.InitContainerStatuses {
							if scmStatus.Name == queue.ImageBuilder.ScmName && scmStatus.State.Terminated != nil {
								errMsg = fmt.Sprintf(`代码拉取失败：exit code=%d，退出原因为:%s`, scmStatus.State.Terminated.ExitCode,
									scmStatus.State.Terminated.Message)
							}
						}
					}

				}
			}
		} else {
			statusCode = 1
			queue.StageBuildLog.Status = common.STATUS_FAILED
			glog.Errorf("%s Failed to get a pod of jobName:%s,podName=%s\n", method, job.ObjectMeta.Name, pod.GetName())
		}

	}

	queue.StageBuildLog.JobName = job.ObjectMeta.Name

	if pod.ObjectMeta.Name != "" {
		queue.StageBuildLog.PodName = pod.ObjectMeta.Name
		if queue.StageBuildLog.NodeName == "" {
			queue.StageBuildLog.NodeName = pod.Spec.NodeName
		}
	}

	//修改状态,并执行其他等待的子任务
	if queue.StageBuildLog.Status == common.STATUS_SUCCESS {
		if queue.StageBuildLog.FlowBuildId != "" && queue.StageBuildLog.BuildAlone != 1 {
			errMsg = "构建成功将会构建下一个子任务"
			flowBuild, err := models.NewCiFlowBuildLogs().FindOneById(queue.StageBuildLog.FlowBuildId)
			if err != nil {
				glog.Errorf("%s get flowbuild info failed from database err:%v\n", method, err)
				return http.StatusInternalServerError
			}
			if flowBuild.FlowId == "" {
				//flow构建不存在
				glog.Errorln("the flow not exist")
				return http.StatusInternalServerError
			}

			if flowBuild.Status < common.STATUS_BUILDING {
				glog.Infof("flowBuild status:%d flowBuild of id:%s\n", flowBuild.Status, flowBuild.BuildId)
				// flow构建已经被stop，此时不再触发下一步构建
				queue.StageBuildLog.Status = common.STATUS_FAILED
				glog.Warningf("%s Flow build is finished, build of next stage stageId:[%s] will not start", method, stage.StageId)
				return http.StatusInternalServerError
			}

			glog.Infof(" will update status:%d,nodeName=%s,currentbuildId=%s\n", queue.StageBuildLog.Status, queue.StageBuildLog.NodeName, queue.StageBuildLog.BuildId)

			res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
			if err != nil {
				glog.Errorf("%s update stage status failed:%d, Err:%v\n", method, res, err)
				return http.StatusInternalServerError
			}

		}
		glog.Infof(errMsg)

		detail := &EmailDetail{
			Type:    "ci",
			Result:  "success",
			Subject: fmt.Sprintf(`'%s'构建成功`, stage.StageName),
			Body:    fmt.Sprintf(`构建流程%s成功完成一次构建`, stage.StageName),
		}
		detail.SendEmailUsingFlowConfig(queue.CurrentNamespace, stage.FlowId)
		return common.STATUS_SUCCESS

	}

	//执行失败时要停止相应的job
	glog.Warningf("%s 构建失败 Will Stop job: %s\n", method, job.ObjectMeta.Name)
	//执行失败时，终止job

	if job.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "Timeout-OrRunFailed" && job.ObjectMeta.Labels["enncloud-builder-succeed"] != "0" {
		glog.Infof("stop the run failed job job.ObjectMeta.Name=%s", job.ObjectMeta.Name)
		//不是手动停止
		errMsg = "构建任务异常,已停止构建，请稍后重试"

	}

	if job.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "true" && job.ObjectMeta.Labels["enncloud-builder-succeed"] != "1" {
		glog.Infof("构建流程被用户手动停止")
		errMsg = "构建流程被用户手动停止"
	}

	glog.Infof("执行失败 Will Update State build PodName=====%d\n", queue.StageBuildLog.PodName)

	res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
	if err != nil {
		glog.Errorf("%s update stage status failed:%d, Err:%v\n", method, res, err)
		return http.StatusInternalServerError
	}

	if errMsg == "" {
		errMsg = "构建发生未知错误"
	}

	detail := &EmailDetail{
		Type:    "ci",
		Result:  "failed",
		Subject: fmt.Sprintf(`'%s'构建失败`, stage.StageName),
		Body:    fmt.Sprintf(`%s <br/>请点击<a href="%s?%s">此处</a>查看EnnFlow详情.`, errMsg, common.FlowDetailUrl, stage.FlowId),
	}
	detail.SendEmailUsingFlowConfig(queue.CurrentNamespace, stage.FlowId)
	return common.STATUS_FAILED

}

func (queue *StageQueueNew) InsertStageBuildLog(stage models.CiStages) error {
	method := "startNextStageBuild"
	queue.StageBuildLog.IsFirst = 0
	queue.StageBuildLog.BuildId = uuid.NewStageBuildID()
	queue.StageBuildLog.CreationTime = time.Now()
	queue.StageBuildLog.FlowBuildId = queue.FlowbuildLog.BuildId
	queue.StageBuildLog.StageId = stage.StageId
	queue.StageBuildLog.StageName = stage.StageName
	queue.StageBuildLog.Status = common.STATUS_WAITING
	queue.StageBuildLog.Namespace = queue.User.Namespace
	queue.StageBuildLog.StartTime = queue.StageBuildLog.CreationTime
	if queue.StageBuildLog.BranchName == "" {
		if stage.DefaultBranch != "" {
			queue.StageBuildLog.BranchName = stage.DefaultBranch
		}
		if stage.Option != nil {
			if stage.Option.Branch != "" {
				queue.StageBuildLog.BranchName = stage.Option.Branch
			}
		}
	}

	if queue.StageBuildLog.NodeName != "" {
		queue.StageBuildLog.NodeName = queue.StageBuildLog.NodeName
	}
	//查询这个stage有没有 正在创建中
	stageBuildLogs, result, err := models.NewCiStageBuildLogs().FindAllByIdWithStatus(stage.StageId, common.STATUS_BUILDING)
	if err != nil || result <= 0 {
		glog.Warningf("%s find next stage of flow %s failed from database %v", method, stage.FlowId, err)
		glog.V(5).Infof("%s stage build logs %v  err:%v\n", method, stageBuildLogs, err)
		//没有执行中的构建记录，则添加“执行中”状态的构建记录
		queue.StageBuildLog.Status = int8(common.STATUS_BUILDING)
	}
	//添加stage构建记录
	_, err = models.NewCiStageBuildLogs().InsertOne(*queue.StageBuildLog)

	if queue.StageBuildLog.Status == common.STATUS_BUILDING {
		return nil
	}

	return err

}

func (queue *StageQueueNew) UpdateStageBuidLogId() error {
	method := "UpdateStageBuidLogId"
	res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
	if err != nil {
		glog.Infof("%s UpdateStageBuidLogId:%v,result:%d\n", method, err, res)
		return err
	}
	return nil
}

func (queue *StageQueueNew) UpdateStageBuildLogStage(buildRec models.CiStageBuildLogs) error {

	updateResult, err := models.NewCiStageBuildLogs().UpdatePodNameAndJobNameByBuildId(buildRec, queue.StageBuildLog.BuildId)
	if err != nil || updateResult < 1 {
		glog.Errorf("%s update stage build failed updateResult=%d err=:%v\n", updateResult, err)
	}
	return nil

}

//修改flowBuildId状态
func (queue *StageQueueNew) UpdateById() error {

	endTime := time.Now()

	_, err := models.NewCiFlowBuildLogs().UpdateById(endTime, int(queue.StageBuildLog.Status), queue.StageBuildLog.FlowBuildId)
	if err != nil {
		return err
	}

	if 1 == queue.StageBuildLog.IsFirst {
		_, err = models.NewCiFlowBuildLogs().UpdateStartTimeAndStatusById(queue.StageBuildLog.StartTime,
			int(queue.StageBuildLog.Status), queue.StageBuildLog.FlowBuildId)
		if err != nil {
			return err
		}
	}

	return nil

}

//检查是否已经该EnnFlow已经在构建
func (queue *StageQueueNew) CheckIfBuiding(flowId string) error {

	recentflowsbuilds, total, err := models.NewCiFlowBuildLogs().FindAllOfFlow(flowId, 3)
	if err != nil {
		return err
	}
	flowsbuildingCount := 0
	if total > 0 {
		for _, recentflowsbuild := range recentflowsbuilds {
			if recentflowsbuild.Status == common.STATUS_BUILDING {
				flowsbuildingCount += 1
			}

			if flowsbuildingCount >= 1 {
				glog.Warningf("Too many waiting builds of %s \n", flowId)
				return fmt.Errorf("该EnnFlow已有任务在执行,请等待执行完再试")
			}

		}
	}

	return nil
}

func (queue *StageQueueNew) GetHarborServer() {
	method := "GetHarborServer"
	if common.HarborServerUrl == "" {
		configs := cluster.NewConfigs()
		harborServerUrl, err := configs.GetHarborServer()
		if err != nil {
			glog.Errorf("%s GetHarborServer failed:%v\n", method, err)

		}
		common.HarborServerUrl = harborServerUrl
	}
}

func (queue *StageQueueNew) SetFailedStatus() {
	method := "setFailedStatus"
	now := time.Now()
	queue.FlowbuildLog.EndTime = now
	queue.FlowbuildLog.Status = common.STATUS_FAILED

	if queue.StageBuildLog.FlowBuildId != "" {

		flowbuild, err := models.NewCiFlowBuildLogs().FindOneById(queue.StageBuildLog.FlowBuildId)
		if err != nil {
			glog.Errorf("%s find flow %s build failed from database err:%v \n", method, queue.StageBuildLog.FlowBuildId, err)
			return
		}

		//非独立构建stage时，更新flow构建的状态
		if flowbuild.FlowId != "" {
			_, err = models.NewCiFlowBuildLogs().UpdateById(queue.FlowbuildLog.EndTime, int(queue.FlowbuildLog.Status), queue.StageBuildLog.FlowBuildId)
			if err != nil {
				glog.Errorf("%s update stagebuild failed: %v\n", method, err)
			}
		}

	}

	if queue.StageBuildLog.JobName != "" {
		queue.ImageBuilder.StopJob(queue.StageBuildLog.Namespace, queue.StageBuildLog.JobName, false, 0)

	}

}

func (queue *StageQueueNew) SetSuncessStatus() {
	method := "SetSuncessStatus"
	now := time.Now()
	queue.FlowbuildLog.EndTime = now
	queue.FlowbuildLog.Status = common.STATUS_SUCCESS

	if queue.StageBuildLog.FlowBuildId != "" { //是否单独构建 1是 0否
		//不是单独构建
		flowbuild, err := models.NewCiFlowBuildLogs().FindOneById(queue.StageBuildLog.FlowBuildId)
		if err != nil {
			glog.Errorf("%s find flow %s build failed from database err:%v \n", method, queue.StageBuildLog.FlowBuildId, err)
			return
		}

		//非独立构建stage时，更新flow构建的状态
		if flowbuild.FlowId != "" {
			_, err = models.NewCiFlowBuildLogs().UpdateById(queue.FlowbuildLog.EndTime, int(queue.FlowbuildLog.Status), queue.StageBuildLog.FlowBuildId)
			if err != nil {
				glog.Errorf("%s update stagebuild failed: %v\n", method, err)
			}
		}

	}
}

func (queue *StageQueueNew) StartStageBuild(stage models.CiStages, index int) int {
	method := "controller/StartStageBuild"
	var ennFlow EnnFlow
	//project 查询失败
	project := models.NewCiManagedProjects()
	if stage.ProjectId != "" {
		err := project.FindProjectById(queue.CurrentNamespace, stage.ProjectId)
		if err != nil || project.Id == "" {
			//project不存在，更新构建状态为失败
			glog.Errorf("%s find project failed: project:%v  err:%v\n", method, project, err)
			ennFlow.Status = http.StatusOK
			ennFlow.BuildStatus = common.STATUS_FAILED
			ennFlow.StageId = stage.StageId
			ennFlow.FlowId = stage.FlowId
			ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
			ennFlow.StageBuildId = queue.StageBuildLog.BuildId
			ennFlow.Flag = 2
			ennFlow.Message = "Project is inactive"
			EnnFlowChan <- ennFlow
			return common.STATUS_FAILED
		}
	}

	glog.Infof("index=%d,FlowId=%s,StageId=%s,FlowBuildId=%s,BuildId=%s,podName=%s\n", index, stage.FlowId, stage.StageId,
		queue.StageBuildLog.FlowBuildId, queue.StageBuildLog.BuildId, queue.StageBuildLog.PodName)

	//获取存贮volume
	volumeMapping, message, respCode := models.GetVolumeSetting(stage.FlowId, stage.StageId,
		queue.StageBuildLog.FlowBuildId, queue.StageBuildLog.BuildId)
	if respCode != http.StatusOK {
		glog.Errorf("get volumeMapping failed: %s\n", message.Message)
		//修改状态并判断是否是单独构建 并执行下一个等待的stage
		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_FAILED
		ennFlow.Message = message.Message
		ennFlow.StageId = stage.StageId
		ennFlow.FlowId = stage.FlowId
		ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
		ennFlow.StageBuildId = queue.StageBuildLog.BuildId
		ennFlow.Flag = 2
		EnnFlowChan <- ennFlow
		queue.SetFailedStatus()
		return common.STATUS_FAILED
	}
	//get harbor server url
	queue.GetHarborServer()

	var buildInfo models.BuildInfo
	buildInfo.ClusterID = client.ClusterID
	buildInfo.BUILD_INFO_TYPE = 0
	buildInfo.RepoUrl = project.Address
	buildInfo.IsCodeRepo = 1
	if queue.StageBuildLog.BranchName != "" {
		buildInfo.Branch = queue.StageBuildLog.BranchName
	} else {
		buildInfo.Branch = stage.DefaultBranch
	}
	//代码仓库类型
	buildInfo.RepoType = models.DepotToRepoType(project.RepoType)
	//获取代码仓库的镜像地址
	cloneImage := os.Getenv("CICD_REPO_CLONE_IMAGE")
	if cloneImage == "" {
		cloneImage = common.CICD_REPO_CLONE_IMAGE
	}
	buildInfo.ScmImage = common.HarborServerUrl + "/" + cloneImage
	//克隆到镜像的目录 默认是 /app
	buildInfo.Clone_location = common.CLONE_LOCATION
	// Only build under user namespace or the owner of project(CI case) TODO
	namespace := ""
	if queue.Namespace != "" {
		namespace = queue.Namespace
	} else {
		namespace = project.Owner
	}
	buildInfo.Namespace = namespace
	//编译的镜像
	buildInfo.Build_image = stage.Image
	//创建镜像标识 3
	buildInfo.BuildImageFlag = stage.Type == common.BUILD_IMAGE_STAGE_TYPE
	buildInfo.FlowName = stage.FlowId
	buildInfo.StageName = stage.StageId
	buildInfo.FlowBuildId = queue.StageBuildLog.FlowBuildId
	buildInfo.StageBuildId = queue.StageBuildLog.BuildId
	buildInfo.Type = stage.Type
	buildInfo.ImageOwner = strings.ToLower(queue.Namespace)

	//镜像的创建相关信息
	if stage.BuildInfo != "" {
		err := json.Unmarshal([]byte(stage.BuildInfo), &buildInfo.TargetImage)
		if err != nil {
			glog.Errorf("%s json unmarshal failed:%v\n", method, err)
			ennFlow.Status = http.StatusOK
			ennFlow.BuildStatus = common.STATUS_FAILED
			ennFlow.StageId = stage.StageId
			ennFlow.FlowId = stage.FlowId
			ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
			ennFlow.StageBuildId = queue.StageBuildLog.BuildId
			ennFlow.Flag = 2
			ennFlow.Message = fmt.Sprintf("json 解析 buildInfo 失败:%s", err)
			EnnFlowChan <- ennFlow
			return common.STATUS_FAILED
		}
		// Image name should be project/image-name, user should specify the target project
		// If not specified, use default public one 镜像仓库
		if strings.TrimSpace(buildInfo.TargetImage.Project) == "" || buildInfo.TargetImage.Project == "" {
			buildInfo.TargetImage.Project = common.Default_push_project
		}
		buildInfo.TargetImage.Image = buildInfo.TargetImage.Project + "/" + buildInfo.TargetImage.Image
		if common.CUSTOM_REGISTRY == buildInfo.TargetImage.RegistryType {
			//自定义仓库时 时速云没有给相关的表
			//TODO
		}
		//Dockerfile from where 2 onlinedockerdile code dockerfile
		if ONLINE == buildInfo.TargetImage.DockerfileFrom {
			//获取在线Dockerfile
			dockerfileOL, err := models.NewCiDockerfile().GetDockerfile(queue.CurrentNamespace, stage.FlowId, stage.StageId)
			if err != nil {
				glog.Errorf("%s get dockerfile content failed from database err==>:%v\n", method, err)
				ennFlow.Status = http.StatusOK
				ennFlow.BuildStatus = common.STATUS_FAILED
				ennFlow.StageId = stage.StageId
				ennFlow.FlowId = stage.FlowId
				ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
				ennFlow.StageBuildId = queue.StageBuildLog.BuildId
				ennFlow.Flag = 2
				ennFlow.Message = "Online Dockerfile should be created before starting a build"
				EnnFlowChan <- ennFlow
				return common.STATUS_FAILED
			}

			if dockerfileOL.Content != "" {
				buildInfo.TargetImage.DockerfileOL = dockerfileOL.Content
			} else {
				glog.Infof("%s Online Dockerfile should be created before starting a build\n", method)
				ennFlow.Status = http.StatusOK
				ennFlow.BuildStatus = common.STATUS_FAILED
				ennFlow.StageId = stage.StageId
				ennFlow.FlowId = stage.FlowId
				ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
				ennFlow.StageBuildId = queue.StageBuildLog.BuildId
				ennFlow.Flag = 2
				ennFlow.Message = "Online Dockerfile should be created before starting a build"
				EnnFlowChan <- ennFlow
				return common.STATUS_FAILED
			}

		}
	}
	//指定节点
	if queue.StageBuildLog.NodeName != "" {
		buildInfo.NodeName = queue.StageBuildLog.NodeName
	}
	//查看是否有依赖
	buildWithDependency := false
	var containerInfo models.Container
	if stage.ContainerInfo != "" {
		err := json.Unmarshal([]byte(stage.ContainerInfo), &containerInfo)
		if err != nil {
			glog.Errorf("%s json unmarshal containerInfo failed:%v\n", method, err)
			ennFlow.Status = http.StatusOK
			ennFlow.BuildStatus = common.STATUS_FAILED
			ennFlow.StageId = stage.StageId
			ennFlow.FlowId = stage.FlowId
			ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
			ennFlow.StageBuildId = queue.StageBuildLog.BuildId
			ennFlow.Flag = 2
			ennFlow.Message = "json 解析 ContainerInfo 信息失败"
			EnnFlowChan <- ennFlow
			return common.STATUS_FAILED
		}
		if containerInfo.Scripts_id != "" {
			MakeScriptEntryEnvForInitContainer(queue.User, &containerInfo)
		}
		//容器的启动命令
		if len(containerInfo.Command) != 0 {
			buildInfo.Command = containerInfo.Command
		}
		//镜像args命令
		buildInfo.Build_command = containerInfo.Args
		//镜像的环境变量
		buildInfo.Env = containerInfo.Env

		if len(containerInfo.Dependencies) > 0 {
			buildWithDependency = true
			buildInfo.Dependencies = make([]models.Dependencie, 0)
			var dependencies models.Dependencie
			for _, info := range containerInfo.Dependencies {
				dependencies.Env = info.Env
				dependencies.Service = common.HarborServerUrl + "/" + info.Service
				buildInfo.Dependencies = append(buildInfo.Dependencies, dependencies)
			}
		}

	}

	if buildWithDependency {
		return common.STATUS_FAILED
	}

	//仓库类型
	depot := project.RepoType
	// For private svn repository
	if depot == "svn" && project.IsPrivate == 1 {
		repo := models.NewCiRepos()
		err := repo.FindOneRepo(project.Namespace, models.DepotToRepoType(depot))
		if err != nil {
			parseCode, err := sqlstatus.ParseErrorCode(err)
			if parseCode == sqlstatus.SQLErrNoRowFound {

				glog.Errorf("%s find one repo failed namespace:%s, err:%v\n", method, queue.Namespace, err)
				ennFlow.Status = http.StatusOK
				ennFlow.BuildStatus = common.STATUS_FAILED
				ennFlow.StageId = stage.StageId
				ennFlow.FlowId = stage.FlowId
				ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
				ennFlow.StageBuildId = queue.StageBuildLog.BuildId
				ennFlow.Flag = 2
				ennFlow.Message = "No repo auth info found"
				EnnFlowChan <- ennFlow
				return common.STATUS_FAILED
			} else {
				glog.Errorf("%s  find one repo failed err:%v\n", method, err)
				ennFlow.Status = http.StatusOK
				ennFlow.BuildStatus = common.STATUS_FAILED
				ennFlow.StageId = stage.StageId
				ennFlow.FlowId = stage.FlowId
				ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
				ennFlow.StageBuildId = queue.StageBuildLog.BuildId
				ennFlow.Flag = 2
				ennFlow.Message = "search repo " + depot + " failed "
				EnnFlowChan <- ennFlow
				return common.STATUS_FAILED
			}

		}

		if repo.Namespace == "" {
			glog.Errorf("%s  find one repo failed err:%v\n", method, err)
			ennFlow.Status = http.StatusOK
			ennFlow.BuildStatus = common.STATUS_FAILED
			ennFlow.StageId = stage.StageId
			ennFlow.FlowId = stage.FlowId
			ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
			ennFlow.StageBuildId = queue.StageBuildLog.BuildId
			ennFlow.Flag = 2
			ennFlow.Message = "No repo auth info found"
			EnnFlowChan <- ennFlow
			return common.STATUS_FAILED
		}

		username := ""
		var repoUserInfos []coderepo.UserInfo
		if repo.UserInfo != "" {
			err = json.Unmarshal([]byte(repo.UserInfo), &repoUserInfos)
			if err != nil {
				glog.Errorf("%s  json unmarshal repo user failed err:%v\n", method, err)
			}
			username = repoUserInfos[0].Login
		} else {
			username = repo.AccessUserName
		}
		buildInfo.Svn_username = username
		buildInfo.Svn_password = repo.AccessToken

	} else if (depot == "gitlab" && project.IsPrivate == 1) ||
		(project.Address != "" && strings.Index(project.Address, "@") > 0) {
		// Handle private githlab
		buildInfo.Git_repo_url = strings.Split(strings.Split(project.Address, "@")[1], ":")[0]
		buildInfo.PublicKey = project.PublicKey
		buildInfo.PrivateKey = project.PrivateKey
	}

	//设置构建集群
	var ciConfig models.CiConfig
	buildCluster := ""
	if stage.CiConfig != "" {
		err := json.Unmarshal([]byte(stage.CiConfig), &ciConfig)
		if err != nil {
			glog.Errorf("%s json unmarshal ciconfig failed: %v\n", method, err)
			ennFlow.Status = http.StatusOK
			ennFlow.BuildStatus = common.STATUS_FAILED
			ennFlow.Message = "CiConfig json 解析失败"
			ennFlow.StageId = stage.StageId
			ennFlow.FlowId = stage.FlowId
			ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
			ennFlow.StageBuildId = queue.StageBuildLog.BuildId
			ennFlow.Flag = 2
			EnnFlowChan <- ennFlow
			return common.STATUS_FAILED
		}
		if ciConfig.BuildCluster != "" {
			buildCluster = ciConfig.BuildCluster
		}

	}

	//如果没有下一个构建 BUILD_INFO_TYPE =1 用来是否删除编译好的二进制文件或者 满足构建多个构建镜像
	if (queue.TotalStage - 1) != int64(index) {
		buildInfo.BUILD_INFO_TYPE = 1 //显示还有下一个stage
	} else {
		buildInfo.BUILD_INFO_TYPE = 2 //显示没有下一个stage
	}

	glog.Infoln("buildCluster=", buildCluster)
	if buildCluster != "" {
		queue.ImageBuilder = models.NewImageBuilder(buildCluster)
	}

	//构建job的参数以及执行job命令
	job, err := queue.ImageBuilder.BuildImage(buildInfo, volumeMapping, common.HarborServerUrl)
	if index == 0 {
		//开始使用websocket通知前端,开始构建
		queue.BuildReqbody.Message = "开始构建:" + queue.FlowId
		queue.BuildReqbody.Status = http.StatusOK
		queue.BuildReqbody.BuildStatus = common.STATUS_BUILDING
		queue.BuildReqbody.FlowBuildId = queue.FlowbuildLog.BuildId
		queue.BuildReqbody.FlowId = queue.CiFlow.FlowId
		queue.BuildReqbody.StageBuildId = queue.StageBuildLog.BuildId
		queue.BuildReqbody.Flag = 1
		EnnFlowChan <- queue.BuildReqbody
	}

	if err != nil || job == nil {

		queue.StageBuildLog.Status = common.STATUS_FAILED
		queue.StageBuildLog.EndTime = time.Now()
		glog.Errorf("%s BuildImage create job failed Err: %v\n", method, err)

		pod, err := queue.ImageBuilder.GetPod(job.ObjectMeta.Namespace, job.ObjectMeta.Name,
			queue.StageBuildLog.BuildId)
		if err != nil {
			glog.Errorf("%s get pod info of %s from kubernetes failed:%v\n", method, job.ObjectMeta.Name, err)
		}

		if pod.ObjectMeta.Name != "" {
			queue.StageBuildLog.PodName = pod.ObjectMeta.Name
			queue.StageBuildLog.NodeName = pod.Spec.NodeName
			queue.StageBuildLog.JobName = job.ObjectMeta.Name
		}
		if queue.StageBuildLog.PodName != "" {
			updateResult, err := models.NewCiStageBuildLogs().UpdatePodNameAndJobNameByBuildId(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
			if err != nil || updateResult < 1 {
				glog.Errorf("%s update stage build failed updateResult=%d err=:%v\n", method, updateResult, err)
			}
		}

		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_FAILED
		ennFlow.Message = fmt.Sprintf("构建任务%s失败\n", stage.StageName)
		ennFlow.StageId = stage.StageId
		ennFlow.FlowId = stage.FlowId
		ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
		ennFlow.StageBuildId = queue.StageBuildLog.BuildId
		ennFlow.Flag = 2
		EnnFlowChan <- ennFlow
		queue.SetFailedStatus()
		detail := &EmailDetail{
			Type:    "ci",
			Result:  "failed",
			Subject: fmt.Sprintf(`'%s'构建失败`, stage.StageName),
			Body:    "发生未知错误，构建失败",
		}
		detail.SendEmailUsingFlowConfig(queue.CurrentNamespace, stage.FlowId)

		return common.STATUS_FAILED
	}

	//for i := 0; i < 5; i++ {
	timeOut, err := queue.WatchPod(job.ObjectMeta.Namespace, job.ObjectMeta.Name, stage)
	if err != nil && !timeOut {
		glog.Errorf(" WatchPod failed:%s\n", err)

	}
	if timeOut && err == nil {
		queue.ImageBuilder.StopJob(job.GetNamespace(), job.GetName(), false, 0)
		detail := &EmailDetail{
			Type:    "ci",
			Result:  "failed",
			Subject: fmt.Sprintf(`'%s'构建失败`, stage.StageName),
			Body:    "任务启动超时",
		}
		detail.SendEmailUsingFlowConfig(queue.CurrentNamespace, stage.FlowId)
	}
	//}
	job, err = queue.ImageBuilder.GetJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name)
	if err != nil || job == nil {
		glog.Errorf("%s, get job from kubernetes failed:%v\n", method, err)
		queue.StageBuildLog.Status = common.STATUS_FAILED
		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_FAILED
		ennFlow.Message = fmt.Sprintf("构建任务%s失败\n", stage.StageName)
		ennFlow.StageId = stage.StageId
		ennFlow.FlowId = stage.FlowId
		ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
		ennFlow.StageBuildId = queue.StageBuildLog.BuildId
		ennFlow.Flag = 2
		EnnFlowChan <- ennFlow
		queue.SetStageBuildStatusFailed()
		return common.STATUS_FAILED
	}

	err = queue.WatchOneJob(job.GetNamespace(), job.GetName())
	if err != nil {
		glog.Errorf("%s WatchOneJob from kubernetes failed:%v\n", method, err)
		queue.StageBuildLog.Status = common.STATUS_FAILED
		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_FAILED
		ennFlow.Message = fmt.Sprintf("构建任务%s失败\n", stage.StageName)
		ennFlow.StageId = stage.StageId
		ennFlow.FlowId = stage.FlowId
		ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
		ennFlow.StageBuildId = queue.StageBuildLog.BuildId
		ennFlow.Flag = 2
		EnnFlowChan <- ennFlow
		queue.SetStageBuildStatusFailed()
		return common.STATUS_FAILED

	}

	job, err = queue.ImageBuilder.GetJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name)
	if err != nil || job == nil {
		glog.Errorf("%s, get job from kubernetes failed:%v\n", method, err)
		queue.StageBuildLog.Status = common.STATUS_FAILED
		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_FAILED
		ennFlow.Message = fmt.Sprintf("构建任务%s失败\n", stage.StageName)
		ennFlow.StageId = stage.StageId
		ennFlow.FlowId = stage.FlowId
		ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
		ennFlow.StageBuildId = queue.StageBuildLog.BuildId
		ennFlow.Flag = 2
		EnnFlowChan <- ennFlow
		queue.SetStageBuildStatusFailed()
		return common.STATUS_FAILED
	}

	status := queue.WaitForBuildToComplete(job, stage)
	<-time.After(10 * time.Second)
	if status >= common.STATUS_FAILED {
		queue.SetStageBuildStatusFailed()
		glog.Infof("%s run failed:%s\n", method, job.ObjectMeta.Name)
		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_FAILED
		ennFlow.Message = fmt.Sprintf("构建任务%s失败\n", stage.StageName)
		ennFlow.StageId = stage.StageId
		ennFlow.FlowId = stage.FlowId
		ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
		ennFlow.StageBuildId = queue.StageBuildLog.BuildId
		ennFlow.Flag = 2
		EnnFlowChan <- ennFlow
		return common.STATUS_FAILED
	} else if status == common.STATUS_SUCCESS {
		queue.SetStageBuildStatusSuccess()
		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_SUCCESS
		ennFlow.Message = fmt.Sprintf("构建任务%s成功\n", stage.StageName)
		ennFlow.StageId = stage.StageId
		ennFlow.FlowId = stage.FlowId
		ennFlow.FlowBuildId = queue.FlowbuildLog.BuildId
		ennFlow.StageBuildId = queue.StageBuildLog.BuildId
		ennFlow.Flag = 2
		EnnFlowChan <- ennFlow
		return common.STATUS_SUCCESS
	}
	return common.STATUS_FAILED
}

func (queue *StageQueueNew) SetStageBuildStatusFailed() {
	res, err := models.NewCiStageBuildLogs().UpdateStageBuildStatusById(common.STATUS_FAILED, queue.StageBuildLog.BuildId)
	if err != nil {
		glog.Errorf(" update result=%d,err:%v\n", res, err)
	}
}

func (queue *StageQueueNew) SetStageBuildStatusSuccess() {
	res, err := models.NewCiStageBuildLogs().UpdateStageBuildStatusById(common.STATUS_SUCCESS, queue.StageBuildLog.BuildId)
	if err != nil {
		glog.Errorf(" update result=%d,err:%v\n", res, err)
	}
}

func (queue *StageQueueNew) Run() {
	method := "StageQueueNew"

	go func() {
		for index, stage := range queue.StageList {
			if index != 0 {
				queue.InsertStageBuildLog(stage)
			}

			glog.Infof("第%d个子任务构建,构建任务名称为:%s\n", index+1, stage.StageName)
			//开始使用websocket通知前端,开始构建 开始此次构建
			if common.STATUS_BUILDING == queue.StageBuildLog.Status {
				// Only use namespace for teamspace scope
				queue.BuildReqbody.Message = "开始构建"
				queue.BuildReqbody.Status = http.StatusOK
				queue.BuildReqbody.BuildStatus = common.STATUS_BUILDING
				queue.BuildReqbody.FlowBuildId = queue.FlowbuildLog.BuildId
				queue.BuildReqbody.StageBuildId = queue.StageBuildLog.BuildId
				queue.BuildReqbody.StageId = stage.StageId
				queue.BuildReqbody.Flag = 2 //1 表示stage构建  2表示flow构建
				EnnFlowChan <- queue.BuildReqbody

				status := queue.StartStageBuild(stage, index)

				if status == common.STATUS_FAILED {
					res, err := models.NewCiStageBuildLogs().UpdateStageBuildStatusById(common.STATUS_FAILED, queue.StageBuildLog.BuildId)
					if err != nil {
						glog.Errorf("%s, update result=%d,err:%v\n", method, res, err)
					}
					queue.BuildReqbody.Message = "构建失败:" + queue.FlowId
					queue.BuildReqbody.Status = http.StatusOK
					queue.BuildReqbody.BuildStatus = common.STATUS_FAILED
					queue.BuildReqbody.FlowBuildId = queue.FlowbuildLog.BuildId
					queue.BuildReqbody.FlowId = queue.CiFlow.FlowId
					queue.BuildReqbody.StageBuildId = queue.StageBuildLog.BuildId
					queue.BuildReqbody.Flag = 1
					queue.SetFailedStatus()
					EnnFlowChan <- queue.BuildReqbody
					return
				} else {
					res, err := models.NewCiStageBuildLogs().UpdateStageBuildStatusById(common.STATUS_SUCCESS, queue.StageBuildLog.BuildId)
					if err != nil {
						glog.Errorf("%s, update result=%d,err:%v\n", method, res, err)
					}

					if index == (queue.LengthOfStage() - 1) {
						//通知EnnFlow 成功构建
						queue.SetSuncessStatus()
						queue.BuildReqbody.Message = "构建成功"
						queue.BuildReqbody.Status = http.StatusOK
						queue.BuildReqbody.BuildStatus = common.STATUS_SUCCESS
						queue.BuildReqbody.FlowBuildId = queue.FlowbuildLog.BuildId
						queue.BuildReqbody.StageBuildId = queue.StageBuildLog.BuildId
						queue.BuildReqbody.Flag = 1
						EnnFlowChan <- queue.BuildReqbody
						return
					}
				}
			}
		}
	}()

	return

}
