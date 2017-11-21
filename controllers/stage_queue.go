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
	"sync"
)

type StageQueue struct {
	StageList     []models.CiStages
	User          *user.UserModel
	BuildReqbody  models.BuildReqbody
	FlowId        string
	TotalStage    int64
	Event         string
	Namespace     string
	CiFlow        *models.CiFlows
	FlowbuildLog  *models.CiFlowBuildLogs
	StageBuildLog *models.CiStageBuildLogs
	ImageBuilder  *models.ImageBuilder
}

func NewStageQueue(user *user.UserModel, buildReqbody models.BuildReqbody, event, namespace, flowId string, imageBuilder *models.ImageBuilder) (*StageQueue, interface{}, int) {
	//buildCluster = "CID-d7d3eb44c1db"
	flowBuildlog := &models.CiFlowBuildLogs{}
	stageBuildLog := &models.CiStageBuildLogs{}

	queue := &StageQueue{
		User:          user,
		BuildReqbody:  buildReqbody,
		FlowId:        flowId,
		Event:         event,
		Namespace:     namespace,
		FlowbuildLog:  flowBuildlog,
		StageBuildLog: stageBuildLog,
		ImageBuilder:  imageBuilder,
	}
	//校验是否存在该flow
	var resp FlowBuilResp
	method := "NewStageQueue"
	ciFlow := models.NewCiFlows()
	flow, err := ciFlow.FindFlowById(user.Namespace, flowId)
	if err != nil {
		parseCode, _ := sqlstatus.ParseErrorCode(err)
		if parseCode == sqlstatus.SQLErrNoRowFound {
			resp.Message = "Flow cannot be found"
			glog.Errorf("%s %s %v\n", method, "Failed to find flow "+flowId+" Of "+user.Namespace, err)
			return queue, resp, http.StatusNotFound
		} else {
			resp.Message = "Flow cannot be found"
			glog.Errorf("%s %s %v\n", method, "Failed to find flow "+flowId+" Of "+user.Namespace, err)
			return queue, resp, http.StatusBadRequest
		}
	}
	if flow.Name == "" {
		resp.Message = "Flow cannot be found"
		glog.Errorf("%s %s \n", method, "Failed to find flow "+flowId+" Of "+user.Namespace)
		return queue, resp, http.StatusNotFound
	}

	queue.CiFlow = &flow

	stageId := buildReqbody.StageId
	var stageList []models.CiStages
	stageList = make([]models.CiStages, 0)
	stageServer := models.NewCiStage()

	if stageId != "" {
		stage, err := stageServer.FindOneById(stageId)
		if err != nil {
			parseCode, _ := sqlstatus.ParseErrorCode(err)
			if parseCode == sqlstatus.SQLErrNoRowFound {
				resp.Message = "not found the stage!"
				glog.Errorf("%s %s %v\n", method, "Failed to find stage "+stageId+" Of "+user.Namespace, err)
				return queue, resp, http.StatusNotFound
			} else {
				resp.Message = "stage cannot be found"
				glog.Errorf("%s %s %v\n", method, "Failed to find stage "+stageId+" Of "+user.Namespace, err)
				return queue, resp, http.StatusBadRequest
			}

		}
		if stage.StageName == "" {
			resp.Message = "stage cannot be found"
			glog.Errorf("%s %s \n", method, "Failed to find stage "+stageId+" Of "+user.Namespace)
			return queue, resp, http.StatusNotFound
		}

		if stage.FlowId != flowId {
			glog.Errorf("%s %s\n", method, "Stage does not belong to Flow")
			resp.Message = "Stage does not belong to Flow"
			return queue, resp, http.StatusConflict
		}
		stages, _, err := stageServer.FindBuildEnabledStages(flowId)
		if err != nil {
			glog.Errorf("%s FindFirstOfFlow find stage failed from database: %v\n", method, err)
			resp.Message = "not find the stage of flow " + flowId
			return queue, resp, http.StatusNotFound
		}
		stageList = append(stageList, stage)
		for _, stageInfo := range stages {
			if stageInfo.Seq > stage.Seq {
				stageList = append(stageList, stageInfo)
			}
		}

		queue.TotalStage = int64(len(stageList))

	} else {
		stages, total, err := stageServer.FindBuildEnabledStages(flowId)
		if err != nil {
			glog.Errorf("%s FindFirstOfFlow find stage failed from database: %v\n", method, err)
			resp.Message = "not find the stage of flow " + flowId
			return queue, resp, http.StatusNotFound
		}
		queue.TotalStage = total

		stageList = append(stageList, stages...)

	}
	queue.StageList = stageList

	return queue, resp, http.StatusOK

}

func (queue *StageQueue) LengthOfStage() int {

	return len(queue.StageList)

}

func (queue *StageQueue) InsertLog() error {
	//开始执行 把执行日志插入到数据库
	flowBuildId := uuid.NewFlowBuildID()
	glog.Infof("====StartFlowBuild==before==flowBuildId===>%s\n", flowBuildId)
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
	if queue.BuildReqbody.Options.Branch != "" {
		stageBuildRec.BranchName = queue.BuildReqbody.Options.Branch
	}
	stageBuildRec.CreationTime = now
	// 如果 flow 开启了 “统一使用代码库” ，把构建时指定的 branch 保存在 stage 的 option 中
	// 在这个 stage 构建完成后，传递给下一个 stage 统一使用代码仓库UniformRepo
	if queue.CiFlow.UniformRepo == 0 {
		queue.StageList[0].Option = queue.BuildReqbody.Options
	}
	var flowBuildRec models.CiFlowBuildLogs
	flowBuildRec.BuildId = flowBuildId
	flowBuildRec.FlowId = queue.FlowId
	flowBuildRec.UserId = queue.User.UserID
	flowBuildRec.CreationTime = now
	flowBuildRec.StartTime = now
	//InsertBuildLog will update 执行状态
	err := models.InsertBuildLog(&flowBuildRec, &stageBuildRec, queue.StageList[0].StageId)
	if err != nil {
		return err
	}
	queue.StageBuildLog = &stageBuildRec
	queue.FlowbuildLog = &flowBuildRec
	return nil
}

func (queue *StageQueue) WaitForBuildToComplete(job *v1.Job, stage models.CiStages, options BuildStageOptions) int {
	method := "WaitForBuildToComplete"
	registryConfig := common.HarborServerUrl
	glog.Infof("%s HarborServerUrl=[%s]\n", method, registryConfig)
	pod := apiv1.Pod{}
	var timeout bool
	var err error
	var errMsg string

	var wg sync.WaitGroup
	resultChan := make(chan bool, 1)
	wg.Add(1)
	go func() {
		//TODO 设置3分钟超时，如无法创建container则自动停止构建
		pod, timeout, err = HandleWaitTimeout(job, queue.ImageBuilder)
		if err != nil {
			glog.Errorf("%s HandleWaitTimeout get: %v\n", method, err)
		}
		resultChan <- false
		//检查是否超时
		select {
		case <-time.After(3 * time.Minute):
			wg.Done()
			glog.Infof("Kubernetes Job start timeout:%q\n", timeout)
		case <-resultChan:
			wg.Done()
			glog.Infof("Kubernetes Job not timeout:%q\n", timeout)
		}

	}()
	wg.Wait()

	statusMessage := queue.ImageBuilder.WaitForJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name, options.BuildWithDependency)

	if statusMessage.JobStatus.JobConditionType == models.ConditionUnknown {
		glog.Warningf("%s Waiting for job failed, try again %#v\n", method, statusMessage.JobStatus)
		statusMessage = queue.ImageBuilder.WaitForJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name, options.BuildWithDependency)
	}

	statusCode := 1
	//手动停止
	if statusMessage.JobStatus.ForcedStop {
		glog.Infof("Run job forced stop:%v\n", statusMessage.JobStatus)
		statusCode = 1
	} else {
		if statusMessage.JobStatus.Failed > 0 {
			glog.Infof("Run job failed:%v\n", statusMessage.JobStatus)
			//执行失败
			statusCode = 1
		} else if statusMessage.JobStatus.Succeeded > 0 {
			//执行成功
			glog.Infof("Run job success:%v\n", statusMessage.JobStatus)
			statusCode = 0
		}
	}

	glog.Infof("%s Wait ended normally... and the job status: %#v\n", method, statusMessage)

	queue.StageBuildLog.EndTime = time.Now()

	if statusCode == 0 {
		queue.StageBuildLog.Status = common.STATUS_SUCCESS
	} else {
		queue.StageBuildLog.Status = common.STATUS_FAILED
	}

	if pod.ObjectMeta.Name == "" {
		pod, err = queue.ImageBuilder.GetPod(job.ObjectMeta.Namespace, job.ObjectMeta.Name)
		if err != nil {
			glog.Errorf("%s get pod from kubernetes cluster failed:%v\n", method, err)
		}
		if pod.ObjectMeta.Name != "" {
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

			glog.Errorf("%s Failed to get a pod of job", method)
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
		if queue.StageBuildLog.FlowBuildId != "" && queue.StageBuildLog.BuildAlone != 1 { //不是单独构建
			errMsg = "构建成功将会构建下一个子任务"
			//TODO 通知下一个构建流程

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
				glog.Warningf("%s Flow build is finished, build of next stage stageId:[%s] will not start", method, stage.StageId)
				return http.StatusInternalServerError
			}

			glog.Infof(" will update status:%d,nodeName=%s,currentbuildId=%s\n", queue.StageBuildLog.Status, queue.StageBuildLog.NodeName, queue.StageBuildLog.BuildId)
			res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
			if err != nil {
				glog.Errorf("%s update stage status failed:%d, Err:%v\n", method, res, err)
				return http.StatusInternalServerError
			}
			glog.Infof("=======================>>update result:%d\n", res)
			return http.StatusInternalServerError
		}
		glog.Infof(errMsg)

		detail := &EmailDetail{
			Type:    "ci",
			Result:  "success",
			Subject: fmt.Sprintf(`'%s'构建成功`, stage.StageName),
			Body:    fmt.Sprintf(`构建流程%s成功完成一次构建`, stage.StageName),
		}
		detail.SendEmailUsingFlowConfig(queue.Namespace, stage.FlowId)
		return common.STATUS_SUCCESS

	}

	//执行失败时要停止相应的job
	glog.Warningf("%s Will Stop job: %s\n", method, job.ObjectMeta.Name)
	//执行失败时，终止job
	if !statusMessage.JobStatus.ForcedStop {
		glog.Infof("stop the run failed job job.ObjectMeta.Name=%s", job.ObjectMeta.Name)
		//不是手动停止
		errMsg = "程序停止构建job"
		_, err = queue.ImageBuilder.StopJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name, false, 0)
		if err != nil {
			glog.Errorf("%s Stop the job %s failed: %v\n", method, job.ObjectMeta.Name, err)
		}
	} else {
		glog.Infof("构建流程被用户手动停止")
		errMsg = "构建流程被用户手动停止"
	}

	glog.Infof("执行失败 Will Update State build PodName=====%d\n", queue.StageBuildLog.PodName)
	//UpdateStatusAndHandleWaiting(queue.User, stage, *queue.StageBuildLog, queue.StageBuildLog.BuildId, options.FlowOwner)
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
	detail.SendEmailUsingFlowConfig(queue.Namespace, stage.FlowId)
	glog.Infof("%s kubernetes run the job failed:%s\n", method, errMsg)

	return common.STATUS_FAILED

}

func (queue *StageQueue) InsertStageBuildLog(stage models.CiStages) error {
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

	glog.Infof("coming the startNextStageBuild =============>1")
	if queue.StageBuildLog.Status == common.STATUS_BUILDING {
		glog.Infof("coming the startNextStageBuild =============>2")
		return nil
	}

	return err

}

func (queue *StageQueue) UpdateStageBuidLogId() error {
	res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
	if err != nil {
		glog.Infof("%s kubernetes run the job failed:%s\n", res)
		return err
	}
	return nil
}

func (queue *StageQueue) UpdateStageBuildLogStage(buildRec models.CiStageBuildLogs) error {

	updateResult, err := models.NewCiStageBuildLogs().UpdatePodNameAndJobNameByBuildId(buildRec, queue.StageBuildLog.BuildId)
	if err != nil || updateResult < 1 {
		glog.Errorf("%s update stage build failed updateResult=%d err=:%v\n", updateResult, err)
	}
	return nil

}

//修改flowBuildId状态
func (queue *StageQueue) UpdateById() error {

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
func (queue *StageQueue) CheckIfBuiding(flowId string) error {

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

func (queue *StageQueue) GetHarborServer() {
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

func (queue *StageQueue) SetFailedStatus() {
	method := "setFailedStatus"
	now := time.Now()
	queue.FlowbuildLog.EndTime = now
	queue.FlowbuildLog.Status = common.STATUS_FAILED
	//if queue.StageBuildLog.FlowBuildId != "" && 1 != queue.StageBuildLog.BuildAlone { //是否单独构建 1是 0否
	if queue.StageBuildLog.FlowBuildId != "" { //是否单独构建 1是 0否
		//不是单独构建
		flowbuild, err := models.NewCiFlowBuildLogs().FindOneById(queue.StageBuildLog.FlowBuildId)
		if err != nil {
			glog.Errorf("%s find flow %s build failed from database err:%v \n", method, queue.StageBuildLog.FlowBuildId, err)
			return
		}
		glog.Infof("%s Flow build log id:%s\n", method, flowbuild.FlowId)
		//非独立构建stage时，更新flow构建的状态
		if flowbuild.FlowId != "" {
			_, err = models.NewCiFlowBuildLogs().UpdateById(queue.FlowbuildLog.EndTime, int(queue.FlowbuildLog.Status), queue.StageBuildLog.FlowBuildId)
			if err != nil {
				glog.Errorf("%s update stagebuild failed: %v\n", method, err)
			}
		}

	}
}

func (queue *StageQueue) SetSuncessStatus() {
	method := "setFailedStatus"
	now := time.Now()
	queue.FlowbuildLog.EndTime = now
	queue.FlowbuildLog.Status = common.STATUS_SUCCESS
	//if queue.StageBuildLog.FlowBuildId != "" && 1 != queue.StageBuildLog.BuildAlone { //是否单独构建 1是 0否
	if queue.StageBuildLog.FlowBuildId != "" { //是否单独构建 1是 0否
		//不是单独构建
		flowbuild, err := models.NewCiFlowBuildLogs().FindOneById(queue.StageBuildLog.FlowBuildId)
		if err != nil {
			glog.Errorf("%s find flow %s build failed from database err:%v \n", method, queue.StageBuildLog.FlowBuildId, err)
			return
		}
		glog.Infof("%s Flow build log id:%s\n", method, flowbuild.FlowId)
		//非独立构建stage时，更新flow构建的状态
		if flowbuild.FlowId != "" {
			_, err = models.NewCiFlowBuildLogs().UpdateById(queue.FlowbuildLog.EndTime, int(queue.FlowbuildLog.Status), queue.StageBuildLog.FlowBuildId)
			if err != nil {
				glog.Errorf("%s update stagebuild failed: %v\n", method, err)
			}
		}

	}
}

//开始使用websocket通知前端,开始构建
func (queue *StageQueue) NotifyFlowStatus(flowId, flowBuildId string, status int) {

	NotifyFlowStatus(flowId, flowBuildId, status)
	//NotifyFlowStatusNew(flowId, flowBuildId, status)

}

func (queue *StageQueue) StartStageBuild(stage models.CiStages, index int) (StageBuildResp, int) {
	method := "controller/StartStageBuild"
	var stageBuildResp StageBuildResp
	//project 查询失败
	project := models.NewCiManagedProjects()
	if stage.ProjectId != "" {
		err := project.FindProjectById(queue.User.Namespace, stage.ProjectId)
		if err != nil || project.Id == "" {
			//project不存在，更新构建状态为失败
			glog.Errorf("%s find project failed:==> project:%v  err:%v\n", method, project, err)
			stageBuildResp.Message = "Project is inactive"
			return stageBuildResp, http.StatusForbidden
		}
	}

	glog.Infof("FlowId=%s,StageId=%s,FlowBuildId=%s,BuildId=%s\n", stage.FlowId, stage.StageId,
		queue.StageBuildLog.FlowBuildId, queue.StageBuildLog.BuildId)
	//获取存贮volume
	volumeMapping, message, respCode := models.GetVolumeSetting(stage.FlowId, stage.StageId,
		queue.StageBuildLog.FlowBuildId, queue.StageBuildLog.BuildId)
	if respCode != http.StatusOK {
		glog.Errorf("get volumeMapping failed: %s\n", message.Message)
		//修改状态并判断是否是单独构建 并执行下一个等待的stage
		stageBuildResp.Message = message.Message
		stageBuildResp.Setting = volumeMapping
		return stageBuildResp, respCode
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
			stageBuildResp.Message = fmt.Sprintf("json unmarshal failed:%s", err)
			return stageBuildResp, http.StatusBadRequest
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
			dockerfileOL, err := models.NewCiDockerfile().GetDockerfile(queue.Namespace, stage.FlowId, stage.StageId)
			if err != nil {
				glog.Errorf("%s get dockerfile content failed from database err==>:%v\n", method, err)
				stageBuildResp.Message = "Online Dockerfile should be created before starting a build"
				return stageBuildResp, http.StatusConflict
			}

			if dockerfileOL.Content != "" {
				buildInfo.TargetImage.DockerfileOL = dockerfileOL.Content
			} else {

				glog.Infof("%s Online Dockerfile should be created before starting a build\n", method)
				stageBuildResp.Message = "Online Dockerfile should be created before starting a build"
				return stageBuildResp, http.StatusBadRequest
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
			stageBuildResp.Message = "json 解析 ContainerInfo 信息失败"
			return stageBuildResp, http.StatusInternalServerError
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
	//仓库类型
	depot := project.RepoType
	// For private svn repository
	if depot == "svn" && project.IsPrivate == 1 {
		repo := models.NewCiRepos()
		err := repo.FindOneRepo(queue.Namespace, models.DepotToRepoType(depot))
		if err != nil {
			parseCode, err := sqlstatus.ParseErrorCode(err)
			if parseCode == sqlstatus.SQLErrNoRowFound {
				glog.Errorf("%s find one repo failed err:%v\n", method, err)
				stageBuildResp.Message = "No repo auth info found"
				return stageBuildResp, http.StatusNotFound
			} else {
				glog.Errorf("%s  find one repo failed err:%v\n", method, err)
				stageBuildResp.Message = "search repo " + depot + " failed"
				return stageBuildResp, http.StatusForbidden
			}

		}

		if repo.Namespace == "" {
			glog.Errorf("%s  find one repo failed err:%v\n", method, err)
			stageBuildResp.Message = "No repo auth info found"
			return stageBuildResp, http.StatusNotFound
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
			stageBuildResp.Message = " CiConfig json 解析失败"
			return stageBuildResp, http.StatusInternalServerError
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

	//buildCluster = "CID-f794208bc85f"
	glog.Infoln("buildCluster===================", buildCluster)
	//queue.ImageBuilder = models.NewImageBuilder(buildCluster)

	//构建job的参数以及执行job命令
	job, err := queue.ImageBuilder.BuildImage(buildInfo, volumeMapping, common.HarborServerUrl)
	if err != nil || job == nil {
		queue.StageBuildLog.Status = common.STATUS_FAILED
		queue.StageBuildLog.EndTime = time.Now()
		glog.Errorf("%s BuildImage create job failed Err: %v\n", method, err)
		stageBuildResp.Message = "Failed to create job"

		pod, err := queue.ImageBuilder.GetPod(job.ObjectMeta.Namespace, job.ObjectMeta.Name)
		if err != nil {
			glog.Errorf("%s get pod info of %s from kubernetes failed:%v\n", method, job.ObjectMeta.Name, err)
			stageBuildResp.Message = "get pod failed from kubernetes"
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

		res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
		if err != nil {
			glog.Errorf("%s, update result=%d,err:%v\n", method, res, err)
		}

		return stageBuildResp, http.StatusInternalServerError
	}

	res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
	if err != nil {
		glog.Errorf("%s, update result=%d,err:%v\n", method, res, err)
	}

	jobdata, _ := json.Marshal(job)
	glog.Infof("%s build images job=%v\n", method, string(jobdata))
	var options BuildStageOptions
	options.BuildWithDependency = buildWithDependency
	options.FlowOwner = queue.Namespace
	options.ImageName = buildInfo.TargetImage.Image
	options.UseCustomRegistry = false //不是客户的镜像仓库
	queue.StageBuildLog.JobName = job.ObjectMeta.Name
	queue.StageBuildLog.Namespace = job.ObjectMeta.Namespace
	//等待构建完成
	status := queue.WaitForBuildToComplete(job, stage, options)
	if status == common.STATUS_FAILED {

		glog.Infof("%s run failed:%s\n", method, job.ObjectMeta.Name)
		stageBuildResp.Message = fmt.Sprintf("构建任务%s失败\n", stage.StageId)
		respCode = http.StatusConflict

	} else {
		stageBuildResp.Message = fmt.Sprintf("构建任务%s成功\n", stage.StageId)
		respCode = http.StatusOK
	}

	return stageBuildResp, respCode

}

func (queue *StageQueue) Run() (FlowBuilResp, int) {
	method := "StageQueue"
	var resp FlowBuilResp
	//判断是否该EnnFlow当前有执行中
	err := queue.CheckIfBuiding(queue.FlowId)
	if err != nil {
		glog.Warningf("%s Too many waiting builds of:  %v\n", method, err)
		if strings.Contains(fmt.Sprintf("%s", err), "该EnnFlow已有任务在执行,请等待执行完再试") {
			resp.Message = "该EnnFlow" + queue.CiFlow.Name + "已有任务在执行,请等待执行完再试"
			return resp, http.StatusForbidden
		} else {
			resp.Message = "not find the stage of flow " + queue.FlowId
			return resp, http.StatusNotFound
		}
	}
	//开始执行 把执行日志插入到数据库
	queue.InsertLog()
	//开始使用websocket通知前端,开始构建
	queue.NotifyFlowStatus(queue.FlowId, queue.StageBuildLog.FlowBuildId, common.STATUS_BUILDING)

	go func() {
		for index, stage := range queue.StageList {
			if index != 0 {
				queue.InsertStageBuildLog(stage)
			}
			glog.Infof("第%d次构建================<<<\n", index+1)
			//开始使用websocket通知前端,开始构建 开始此次构建
			if common.STATUS_BUILDING == queue.StageBuildLog.Status {
				glog.Infof("%s ======get build status============stageBuildRec:%#v\n", method, queue.StageBuildLog)
				// Only use namespace for teamspace scope
				respStage, code := queue.StartStageBuild(stage, index)
				glog.Infof("%s resp code:%d,respStage:%v\n", method, code, respStage)
				//构建失败
				if code != 200 {
					glog.Infof("%s Run failed respStage:%v\n", method, respStage)
					//修改FlowBuildLog
					queue.SetFailedStatus()
					//通知websocket 失败
					NotifyFlowStatus(queue.FlowId, queue.StageBuildLog.FlowBuildId, common.STATUS_FAILED)
					//NotifyFlowStatusNew(queue.FlowId, queue.StageBuildLog.FlowBuildId, common.STATUS_FAILED)
					return
				}
				if index == (queue.LengthOfStage() - 1) {
					queue.SetSuncessStatus()
					//NotifyFlowStatusNew(queue.FlowId, queue.StageBuildLog.FlowBuildId, common.STATUS_SUCCESS)
					NotifyFlowStatus(queue.FlowId, queue.StageBuildLog.FlowBuildId, common.STATUS_SUCCESS)
					glog.Infof("resp.StageBuildId=%s\n", resp.StageBuildId)
					return
				}
			}
		}
	}()

	resp.FlowBuildId = queue.FlowbuildLog.BuildId
	resp.StageBuildId = queue.StageBuildLog.BuildId
	return resp, http.StatusOK

}

//StageQueueNew============================>>>

type StageQueueNew struct {
	StageList     []models.CiStages
	User          *user.UserModel
	BuildReqbody  EnnFlow
	FlowId        string
	TotalStage    int64
	Event         string
	Namespace     string
	LoginUserName string
	CiFlow        *models.CiFlows
	FlowbuildLog  *models.CiFlowBuildLogs
	StageBuildLog *models.CiStageBuildLogs
	ImageBuilder  *models.ImageBuilder
	Conn          Conn
}

func NewStageQueueNew(buildReqbody EnnFlow, event, namespace, loginUserName, flowId string, imageBuilder *models.ImageBuilder, conn Conn) *StageQueueNew {
	method := "NewStageQueueNew"
	u := user.NewUserModel()
	parseCode, err := u.GetByName(loginUserName)
	if err != nil || parseCode == sqlstatus.SQLErrNoRowFound {
		glog.Errorf("%s get Login User failed:%v\n", method, err)

	}

	flowBuildlog := &models.CiFlowBuildLogs{}
	stageBuildLog := &models.CiStageBuildLogs{}

	queue := &StageQueueNew{
		User:          u,
		BuildReqbody:  buildReqbody,
		FlowId:        flowId,
		Event:         event,
		Namespace:     namespace,
		FlowbuildLog:  flowBuildlog,
		StageBuildLog: stageBuildLog,
		ImageBuilder:  imageBuilder,
		Conn:          conn,
		LoginUserName: loginUserName,
	}

	ciFlow := models.NewCiFlows()
	flow, err := ciFlow.FindFlowById(namespace, flowId)
	if err != nil {
		parseCode, _ := sqlstatus.ParseErrorCode(err)
		if parseCode == sqlstatus.SQLErrNoRowFound {
			buildReqbody.Message = "Flow cannot be found"
			buildReqbody.Status = http.StatusNotFound
			buildReqbody.BuildStatus = common.STATUS_FAILED
			glog.Errorf("%s %s %v\n", method, "Failed to find flow "+flowId+" Of "+namespace, err)
			Send(buildReqbody, conn)
			return queue
		} else {
			buildReqbody.Message = "Flow cannot be found"
			buildReqbody.Status = http.StatusInternalServerError
			buildReqbody.BuildStatus = common.STATUS_FAILED
			glog.Errorf("%s %s %v\n", method, "Failed to find flow "+flowId+" Of "+namespace, err)
			Send(buildReqbody, conn)
			return queue
		}
	}

	if flow.Name == "" {
		buildReqbody.Message = "Flow cannot be found"
		buildReqbody.Status = http.StatusNotFound
		buildReqbody.BuildStatus = common.STATUS_FAILED
		glog.Errorf("%s %s %v\n", method, "Failed to find flow "+flowId+" Of "+namespace, err)
		Send(buildReqbody, conn)
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
				Send(buildReqbody, conn)
				glog.Errorf("%s %s %v\n", method, "Failed to find stage "+stageId+" Of "+namespace, err)
				return queue
			} else {
				buildReqbody.Message = "stage cannot be found"
				buildReqbody.Status = http.StatusInternalServerError
				buildReqbody.BuildStatus = common.STATUS_FAILED
				Send(buildReqbody, conn)
				glog.Errorf("%s %s %v\n", method, "Failed to find stage "+stageId+" Of "+namespace, err)
				return queue
			}

		}
		if stage.StageName == "" {
			buildReqbody.Message = "not found the stage!"
			buildReqbody.Status = http.StatusNotFound
			buildReqbody.BuildStatus = common.STATUS_FAILED
			Send(buildReqbody, conn)
			glog.Errorf("%s %s %v\n", method, "Failed to find stage "+stageId+" Of "+namespace, err)
			return queue
		}

		if stage.FlowId != flowId {
			glog.Errorf("%s %s\n", method, "Stage does not belong to Flow")
			buildReqbody.Message = "Stage does not belong to Flow!"
			buildReqbody.Status = http.StatusBadRequest
			buildReqbody.BuildStatus = common.STATUS_FAILED
			Send(buildReqbody, conn)
			return queue
		}
		stages, _, err := stageServer.FindBuildEnabledStages(flowId)
		if err != nil {
			glog.Errorf("%s FindFirstOfFlow find stage failed from database: %v\n", method, err)
			buildReqbody.Message = "not find the stage of flow " + flowId
			buildReqbody.Status = http.StatusBadRequest
			buildReqbody.BuildStatus = common.STATUS_FAILED
			Send(buildReqbody, conn)
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
		stages, total, err := stageServer.FindBuildEnabledStages(flowId)
		if err != nil {
			glog.Errorf("%s FindFirstOfFlow find stage failed from database: %v\n", method, err)
			buildReqbody.Message = "not find the stage of flow " + flowId
			buildReqbody.Status = http.StatusBadRequest
			buildReqbody.BuildStatus = common.STATUS_FAILED
			Send(buildReqbody, conn)
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
	glog.Infof("====StartFlowBuild==before==flowBuildId===>%s\n", flowBuildId)
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
	//InsertBuildLog will update 执行状态
	err := models.InsertBuildLog(&flowBuildRec, &stageBuildRec, queue.StageList[0].StageId)
	if err != nil {
		return err
	}
	queue.StageBuildLog = &stageBuildRec
	queue.FlowbuildLog = &flowBuildRec
	return nil
}

func (queue *StageQueueNew) WaitForBuildToComplete(job *v1.Job, stage models.CiStages, options BuildStageOptions) int {
	method := "WaitForBuildToComplete"
	registryConfig := common.HarborServerUrl
	glog.Infof("%s HarborServerUrl=[%s]\n", method, registryConfig)
	pod := apiv1.Pod{}
	var timeout bool
	var err error
	var errMsg string

	var wg sync.WaitGroup
	resultChan := make(chan bool, 1)
	wg.Add(1)
	go func() {
		//TODO 设置3分钟超时，如无法创建container则自动停止构建
		pod, timeout, err = HandleWaitTimeout(job, queue.ImageBuilder)
		if err != nil {
			glog.Errorf("%s HandleWaitTimeout get: %v\n", method, err)
		}
		resultChan <- false
		//检查是否超时
		select {
		case <-time.After(3 * time.Minute):
			wg.Done()
			glog.Infof("Kubernetes Job start timeout:%q\n", timeout)
		case <-resultChan:
			wg.Done()
			glog.Infof("Kubernetes Job not timeout:%q\n", timeout)
		}

	}()
	wg.Wait()

	jobWatch := queue.WatchJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name)

	if len(jobWatch.Status.Conditions) != 0 {
		if jobWatch.Status.Conditions[0].Status == apiv1.ConditionUnknown {
			glog.Warningf("%s Waiting for job failed, try again %#v\n", method, jobWatch)
			jobWatch = queue.WatchJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name)
		}
	}

	statusCode := 1
	//手动停止
	if jobWatch.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "true" && job.ObjectMeta.Labels["enncloud-builder-succeed"] != "1" {
		glog.Infof("用户停止job:Run job forced stop:%v\n", "")
		statusCode = 1
	} else if jobWatch.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "Timeout-OrRunFailed" && job.ObjectMeta.Labels["enncloud-builder-succeed"] != "0" {
		glog.Infof("程序停止job:Run job forced stop:%v\n", "")
		statusCode = 1
	} else if len(jobWatch.Status.Conditions) != 0 {
		if jobWatch.Status.Failed > 0 {
			glog.Infof("Run job failed:%v\n", job.Status)
			//执行失败
			statusCode = 1
		} else if jobWatch.Status.Succeeded > 0 {
			//执行成功
			glog.Infof("Run job success:%v\n", job.Status)
			statusCode = 0
		}
	}

	glog.Infof("%s Wait ended normally... and the job status: %#v\n", method, jobWatch)

	queue.StageBuildLog.EndTime = time.Now()

	if statusCode == 0 {
		queue.StageBuildLog.Status = common.STATUS_SUCCESS
	} else {
		queue.StageBuildLog.Status = common.STATUS_FAILED
	}

	if pod.ObjectMeta.Name == "" {
		pod, err = queue.ImageBuilder.GetPod(job.ObjectMeta.Namespace, job.ObjectMeta.Name)
		if err != nil {
			glog.Errorf("%s get pod from kubernetes cluster failed:%v\n", method, err)
		}
		if pod.ObjectMeta.Name != "" {
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

			glog.Errorf("%s Failed to get a pod of job", method)
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
		if queue.StageBuildLog.FlowBuildId != "" && queue.StageBuildLog.BuildAlone != 1 { //不是单独构建
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
				glog.Warningf("%s Flow build is finished, build of next stage stageId:[%s] will not start", method, stage.StageId)
				return http.StatusInternalServerError
			}

			glog.Infof(" will update status:%d,nodeName=%s,currentbuildId=%s\n", queue.StageBuildLog.Status, queue.StageBuildLog.NodeName, queue.StageBuildLog.BuildId)
			res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
			if err != nil {
				glog.Errorf("%s update stage status failed:%d, Err:%v\n", method, res, err)
				return http.StatusInternalServerError
			}
			glog.Infof("=======================>>update result:%d\n", res)

		}
		glog.Infof(errMsg)

		detail := &EmailDetail{
			Type:    "ci",
			Result:  "success",
			Subject: fmt.Sprintf(`'%s'构建成功`, stage.StageName),
			Body:    fmt.Sprintf(`构建流程%s成功完成一次构建`, stage.StageName),
		}
		detail.SendEmailUsingFlowConfig(queue.Namespace, stage.FlowId)
		return common.STATUS_SUCCESS

	}

	//执行失败时要停止相应的job
	glog.Warningf("%s Will Stop job: %s\n", method, job.ObjectMeta.Name)
	//执行失败时，终止job

	if jobWatch.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "Timeout-OrRunFailed" && job.ObjectMeta.Labels["enncloud-builder-succeed"] != "0" {
		glog.Infof("stop the run failed job job.ObjectMeta.Name=%s", job.ObjectMeta.Name)
		//不是手动停止
		errMsg = "程序停止构建job"
		_, err = queue.ImageBuilder.StopJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name, false, 0)
		if err != nil {
			glog.Errorf("%s Stop the job %s failed: %v\n", method, job.ObjectMeta.Name, err)
		}
	}
	if jobWatch.ObjectMeta.Labels[common.MANUAL_STOP_LABEL] == "true" && job.ObjectMeta.Labels["enncloud-builder-succeed"] != "1" {
		glog.Infof("构建流程被用户手动停止")
		errMsg = "构建流程被用户手动停止"
	}

	glog.Infof("执行失败 Will Update State build PodName=====%d\n", queue.StageBuildLog.PodName)
	//UpdateStatusAndHandleWaiting(queue.User, stage, *queue.StageBuildLog, queue.StageBuildLog.BuildId, options.FlowOwner)
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

	detail.SendEmailUsingFlowConfig(queue.Namespace, stage.FlowId)
	glog.Infof("%s kubernetes run the job failed:%s\n", method, errMsg)

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

	glog.Infof("coming the startNextStageBuild =============>1")
	if queue.StageBuildLog.Status == common.STATUS_BUILDING {
		glog.Infof("coming the startNextStageBuild =============>2")
		return nil
	}

	return err

}

func (queue *StageQueueNew) UpdateStageBuidLogId() error {
	res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
	if err != nil {
		glog.Infof("%s kubernetes run the job failed:%s\n", res)
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
	//if queue.StageBuildLog.FlowBuildId != "" && 1 != queue.StageBuildLog.BuildAlone { //是否单独构建 1是 0否
	if queue.StageBuildLog.FlowBuildId != "" { //是否单独构建 1是 0否
		//不是单独构建
		flowbuild, err := models.NewCiFlowBuildLogs().FindOneById(queue.StageBuildLog.FlowBuildId)
		if err != nil {
			glog.Errorf("%s find flow %s build failed from database err:%v \n", method, queue.StageBuildLog.FlowBuildId, err)
			return
		}
		glog.Infof("%s Flow build log id:%s\n", method, flowbuild.FlowId)
		//非独立构建stage时，更新flow构建的状态
		if flowbuild.FlowId != "" {
			_, err = models.NewCiFlowBuildLogs().UpdateById(queue.FlowbuildLog.EndTime, int(queue.FlowbuildLog.Status), queue.StageBuildLog.FlowBuildId)
			if err != nil {
				glog.Errorf("%s update stagebuild failed: %v\n", method, err)
			}
		}

	}
}

func (queue *StageQueueNew) SetSuncessStatus() {
	method := "setFailedStatus"
	now := time.Now()
	queue.FlowbuildLog.EndTime = now
	queue.FlowbuildLog.Status = common.STATUS_SUCCESS
	//if queue.StageBuildLog.FlowBuildId != "" && 1 != queue.StageBuildLog.BuildAlone { //是否单独构建 1是 0否
	if queue.StageBuildLog.FlowBuildId != "" { //是否单独构建 1是 0否
		//不是单独构建
		flowbuild, err := models.NewCiFlowBuildLogs().FindOneById(queue.StageBuildLog.FlowBuildId)
		if err != nil {
			glog.Errorf("%s find flow %s build failed from database err:%v \n", method, queue.StageBuildLog.FlowBuildId, err)
			return
		}
		glog.Infof("%s Flow build log id:%s\n", method, flowbuild.FlowId)
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
		err := project.FindProjectById(queue.User.Namespace, stage.ProjectId)
		if err != nil || project.Id == "" {
			queue.SetFailedStatus()
			//project不存在，更新构建状态为失败
			glog.Errorf("%s find project failed:==> project:%v  err:%v\n", method, project, err)
			ennFlow.Status = http.StatusOK
			ennFlow.BuildStatus = common.STATUS_FAILED
			ennFlow.Message = "Project is inactive"
			Send(ennFlow, queue.Conn)
			return common.STATUS_FAILED
		}
	}

	glog.Infof("FlowId=%s,StageId=%s,FlowBuildId=%s,BuildId=%s\n", stage.FlowId, stage.StageId,
		queue.StageBuildLog.FlowBuildId, queue.StageBuildLog.BuildId)

	//获取存贮volume
	volumeMapping, message, respCode := models.GetVolumeSetting(stage.FlowId, stage.StageId,
		queue.StageBuildLog.FlowBuildId, queue.StageBuildLog.BuildId)
	if respCode != http.StatusOK {
		queue.SetFailedStatus()
		glog.Errorf("get volumeMapping failed: %s\n", message.Message)
		//修改状态并判断是否是单独构建 并执行下一个等待的stage
		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_FAILED
		ennFlow.Message = message.Message
		Send(ennFlow, queue.Conn)
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
			queue.SetFailedStatus()
			glog.Errorf("%s json unmarshal failed:%v\n", method, err)
			ennFlow.Status = http.StatusOK
			ennFlow.BuildStatus = common.STATUS_FAILED
			ennFlow.Message = fmt.Sprintf("json unmarshal failed:%s", err)
			Send(ennFlow, queue.Conn)
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
			dockerfileOL, err := models.NewCiDockerfile().GetDockerfile(queue.Namespace, stage.FlowId, stage.StageId)
			if err != nil {
				queue.SetFailedStatus()
				glog.Errorf("%s get dockerfile content failed from database err==>:%v\n", method, err)
				ennFlow.Status = http.StatusOK
				ennFlow.BuildStatus = common.STATUS_FAILED
				ennFlow.Message = "Online Dockerfile should be created before starting a build"
				Send(ennFlow, queue.Conn)
				return common.STATUS_FAILED
			}

			if dockerfileOL.Content != "" {
				buildInfo.TargetImage.DockerfileOL = dockerfileOL.Content
			} else {
				queue.SetFailedStatus()
				glog.Infof("%s Online Dockerfile should be created before starting a build\n", method)
				ennFlow.Status = http.StatusOK
				ennFlow.BuildStatus = common.STATUS_FAILED
				ennFlow.Message = "Online Dockerfile should be created before starting a build"
				Send(ennFlow, queue.Conn)
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
			queue.SetFailedStatus()
			glog.Errorf("%s json unmarshal containerInfo failed:%v\n", method, err)
			ennFlow.Status = http.StatusOK
			ennFlow.BuildStatus = common.STATUS_FAILED
			ennFlow.Message = "json 解析 ContainerInfo 信息失败"
			Send(ennFlow, queue.Conn)
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
	//仓库类型
	depot := project.RepoType
	// For private svn repository
	if depot == "svn" && project.IsPrivate == 1 {
		repo := models.NewCiRepos()
		err := repo.FindOneRepo(queue.Namespace, models.DepotToRepoType(depot))
		if err != nil {
			queue.SetFailedStatus()
			parseCode, err := sqlstatus.ParseErrorCode(err)
			if parseCode == sqlstatus.SQLErrNoRowFound {

				glog.Errorf("%s find one repo failed err:%v\n", method, err)
				ennFlow.Status = http.StatusOK
				ennFlow.BuildStatus = common.STATUS_FAILED
				ennFlow.Message = "No repo auth info found"
				Send(ennFlow, queue.Conn)
				return common.STATUS_FAILED
			} else {
				queue.SetFailedStatus()
				glog.Errorf("%s  find one repo failed err:%v\n", method, err)
				ennFlow.Status = http.StatusOK
				ennFlow.BuildStatus = common.STATUS_FAILED
				ennFlow.Message = "search repo " + depot + " failed "
				Send(ennFlow, queue.Conn)
				return common.STATUS_FAILED
			}

		}

		if repo.Namespace == "" {
			queue.SetFailedStatus()
			glog.Errorf("%s  find one repo failed err:%v\n", method, err)
			ennFlow.Status = http.StatusOK
			ennFlow.BuildStatus = common.STATUS_FAILED
			ennFlow.Message = "No repo auth info found"
			Send(ennFlow, queue.Conn)
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
			queue.SetFailedStatus()
			glog.Errorf("%s json unmarshal ciconfig failed: %v\n", method, err)
			ennFlow.Status = http.StatusOK
			ennFlow.BuildStatus = common.STATUS_FAILED
			ennFlow.Message = "CiConfig json 解析失败"
			Send(ennFlow, queue.Conn)
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

	//buildCluster = "CID-f794208bc85f"
	glog.Infoln("buildCluster===================", buildCluster)
	//queue.ImageBuilder = models.NewImageBuilder(buildCluster)

	//构建job的参数以及执行job命令
	job, err := queue.ImageBuilder.BuildImage(buildInfo, volumeMapping, common.HarborServerUrl)
	if err != nil || job == nil {

		queue.StageBuildLog.Status = common.STATUS_FAILED
		queue.StageBuildLog.EndTime = time.Now()
		glog.Errorf("%s BuildImage create job failed Err: %v\n", method, err)
		//stageBuildResp.Message = "Failed to create job"

		pod, err := queue.ImageBuilder.GetPod(job.ObjectMeta.Namespace, job.ObjectMeta.Name)
		if err != nil {
			glog.Errorf("%s get pod info of %s from kubernetes failed:%v\n", method, job.ObjectMeta.Name, err)
			//stageBuildResp.Message = "get pod failed from kubernetes"
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

		res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
		if err != nil {
			glog.Errorf("%s, update result=%d,err:%v\n", method, res, err)
		}
		queue.SetFailedStatus()
		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_FAILED
		ennFlow.Message = fmt.Sprintf("构建任务%s失败\n", stage.StageName)
		Send(ennFlow, queue.Conn)
		queue.SetFailedStatus()
		return common.STATUS_FAILED

	}

	queue.StageBuildLog.JobName = job.ObjectMeta.Name
	res, err := models.NewCiStageBuildLogs().UpdateById(*queue.StageBuildLog, queue.StageBuildLog.BuildId)
	if err != nil {
		glog.Errorf("%s, update result=%d,err:%v\n", method, res, err)
	}

	jobdata, _ := json.Marshal(job)
	glog.Infof("%s build images job=%v\n", method, string(jobdata))
	var options BuildStageOptions
	options.BuildWithDependency = buildWithDependency
	options.FlowOwner = queue.CiFlow.Owner
	options.ImageName = buildInfo.TargetImage.Image
	options.UseCustomRegistry = false //不是客户的镜像仓库
	queue.StageBuildLog.JobName = job.ObjectMeta.Name
	queue.StageBuildLog.Namespace = job.ObjectMeta.Namespace
	//等待构建完成
	status := queue.WaitForBuildToComplete(job, stage, options)
	if status >= common.STATUS_FAILED {
		glog.Infof("%s run failed:%s\n", method, job.ObjectMeta.Name)
		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_FAILED
		ennFlow.Message = fmt.Sprintf("构建任务%s失败\n", stage.StageName)
		Send(ennFlow, queue.Conn)
		queue.SetFailedStatus()
		return common.STATUS_FAILED
	} else if status == common.STATUS_SUCCESS {
		ennFlow.Status = http.StatusOK
		ennFlow.BuildStatus = common.STATUS_SUCCESS
		ennFlow.Message = fmt.Sprintf("构建任务%s成功\n", stage.StageName)
		Send(ennFlow, queue.Conn)
		return common.STATUS_SUCCESS
	}

	return common.STATUS_FAILED

}

func (queue *StageQueueNew) Run() {
	method := "StageQueue"

	//开始使用websocket通知前端,开始构建
	queue.BuildReqbody.Message = "开始构建:" + queue.FlowId
	queue.BuildReqbody.Status = http.StatusOK
	queue.BuildReqbody.BuildStatus = common.STATUS_BUILDING
	queue.BuildReqbody.FlowBuildId = queue.FlowbuildLog.BuildId
	queue.BuildReqbody.FlowId = queue.CiFlow.FlowId
	queue.BuildReqbody.StageBuildId = queue.StageBuildLog.BuildId
	queue.BuildReqbody.Flag = 1
	Send(queue.BuildReqbody, queue.Conn)

	go func() {
		for index, stage := range queue.StageList {
			if index != 0 {
				queue.InsertStageBuildLog(stage)
			}
			glog.Infof("第%d次构建================<<<\n", index+1)
			//开始使用websocket通知前端,开始构建 开始此次构建
			if common.STATUS_BUILDING == queue.StageBuildLog.Status {
				glog.Infof("%s ======get build status============stageBuildRec:%#v\n", method, queue.StageBuildLog)
				// Only use namespace for teamspace scope
				queue.BuildReqbody.Message = "开始构建"
				queue.BuildReqbody.Status = http.StatusOK
				queue.BuildReqbody.BuildStatus = common.STATUS_BUILDING
				queue.BuildReqbody.FlowBuildId = queue.FlowbuildLog.BuildId
				queue.BuildReqbody.StageBuildId = queue.StageBuildLog.BuildId
				queue.BuildReqbody.Flag = 2 //1 表示stage构建  2表示flow构建
				Send(queue.BuildReqbody, queue.Conn)

				status := queue.StartStageBuild(stage, index)
				if status == common.STATUS_FAILED {
					return
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
					Send(queue.BuildReqbody, queue.Conn)
					return
				}
			}
		}
	}()

	return

}
