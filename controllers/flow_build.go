package controllers

import (
	"github.com/golang/glog"
	"fmt"
	"os"
	"dev-flows-api-golang/models/user"
	"dev-flows-api-golang/models/common"
	"dev-flows-api-golang/models/cluster"
	//apiv1 "k8s.io/client-go/pkg/api/v1"
	"dev-flows-api-golang/util/uuid"
	"time"
	"strings"
	"encoding/json"
	"dev-flows-api-golang/ci/coderepo"
	"k8s.io/client-go/pkg/apis/batch/v1"
	sqlstatus "dev-flows-api-golang/models/sql/status"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"dev-flows-api-golang/models"
	"net/http"
	"dev-flows-api-golang/modules/client"
	"sync"
)

type FlowBuilResp struct {
	Message      string `json:"message,omitempty"`
	FlowBuildId  string `json:"flowBuildId,omitempty"`
	StageBuildId string `json:"stageBuildId,omitempty"`
}

//StartFlowBuild event 是CI的触发条件
func StartFlowBuild(user *user.UserModel, flowId, stageId string, event string, options *models.Option) (interface{}, int) {
	method := "controllers/startFlowBuild"
	var resp FlowBuilResp
	//校验是否存在该flow在数据库
	ciFlow := models.NewCiFlows()
	flow, err := ciFlow.FindFlowById(user.Namespace, flowId)
	if err != nil {
		parseCode, _ := sqlstatus.ParseErrorCode(err)
		if parseCode == sqlstatus.SQLErrNoRowFound {
			resp.Message = "Flow cannot be found"
			glog.Errorf("%s %s %v\n", method, "Failed to find flow "+flowId+" Of "+user.Namespace, err)
			return resp, http.StatusNotFound
		} else {
			resp.Message = "Flow cannot be found"
			glog.Errorf("%s %s %v\n", method, "Failed to find flow "+flowId+" Of "+user.Namespace, err)
			return resp, http.StatusBadRequest
		}
	}

	if flow.Name == "" {
		resp.Message = "Flow cannot be found"
		glog.Errorf("%s %s \n", method, "Failed to find flow "+flowId+" Of "+user.Namespace)
		return resp, http.StatusNotFound
	}

	stageServer := models.NewCiStage()
	flowBuild := models.NewCiFlowBuildLogs()
	// 指定stage时，从指定stage开始构建
	// 未指定stage时，从第一个stage开始构建
	var stage models.CiStages
	//指定stage时,要检查该stage info if exist
	if stageId != "" {
		stage, err = stageServer.FindOneById(stageId)
		if err != nil {
			parseCode, _ := sqlstatus.ParseErrorCode(err)
			if parseCode == sqlstatus.SQLErrNoRowFound {
				resp.Message = "not found the stage!"
				glog.Errorf("%s %s %v\n", method, "Failed to find stage "+stageId+" Of "+user.Namespace, err)
				return resp, http.StatusNotFound
			} else {
				resp.Message = "stage cannot be found"
				glog.Errorf("%s %s %v\n", method, "Failed to find stage "+stageId+" Of "+user.Namespace, err)
				return resp, http.StatusBadRequest
			}

		}
		if stage.StageName == "" {
			resp.Message = "stage cannot be found"
			glog.Errorf("%s %s \n", method, "Failed to find stage "+stageId+" Of "+user.Namespace)
			return resp, http.StatusNotFound
		}

		if stage.FlowId != flowId {
			glog.Errorf("%s %s\n", method, "Stage does not belong to Flow")
			resp.Message = "Stage does not belong to Flow"
			return resp, http.StatusConflict
		}
		// 未指定stage时,查找第一个子任务
	} else {
		stage, err = stageServer.FindFirstOfFlow(flowId)
		if err != nil {
			glog.Errorf("%s FindFirstOfFlow find stage failed from database: %v\n", method, err)
			resp.Message = "not find the stage of flow " + flowId
			return resp, http.StatusNotFound
		}
	}

	//如果是第一个子任务的情况
	if stage.Seq == 1 {
		//查询出三个子任务,判断有多少个子任务在等待执行或者执行状态 记录在flowsbuildingCount变量中
		recentflowsbuilds, total, err := flowBuild.FindAllOfFlow(flowId, 3)
		if err != nil {
			glog.Errorf("%s find stage failed from database:%v\n", method, err)
			resp.Message = "not find the stage of flow " + flowId
			return resp, http.StatusNotFound
		}

		flowsbuildingCount := 0

		if total > 0 {
			for _, recentflowsbuild := range recentflowsbuilds {
				if recentflowsbuild.Status == common.STATUS_BUILDING ||
					common.STATUS_WAITING == recentflowsbuild.Status {
					flowsbuildingCount += 1
				}

				if flowsbuildingCount >= 1 {
					glog.Warningf("%s %v Too many waiting builds of \n", method, flowId)
					resp.Message = "该EnnFlow已有任务在执行,请等待执行完再试"
					return resp, http.StatusForbidden
				}

			}
		}
		//当前执行的子任务不是第一步,这么做的意义就是当前只允许一个flow在执行
	} else {
		recentflowsbuilds, total, err := flowBuild.FindAllOfFlow(flowId, 3)
		if err != nil {
			glog.Errorf("%s find flow info total=%d %v\n", method, err)
			resp.Message = "not find the stage of flow " + flowId
			return resp, http.StatusNotFound
		}
		flowsbuildingCount := 0

		if total > 0 {
			for _, recentflowsbuild := range recentflowsbuilds {
				if recentflowsbuild.Status == common.STATUS_BUILDING {
					flowsbuildingCount += 1
				}

				if flowsbuildingCount >= 1 {
					glog.Warningf("%s %v Too many waiting builds of \n", method, flowId)
					resp.Message = "当前有构建任务在执行,请等待执行完再试"
					return resp, http.StatusForbidden
				}

			}
		}
	}
	//开始执行 把执行日志插入到数据库
	flowBuildId := uuid.NewFlowBuildID()
	glog.Infof("====StartFlowBuild==before==flowBuildId===>%s\n",flowBuildId)
	stageBuildId := uuid.NewStageBuildID()
	codeBranch := ""
	if event != "" {
		//有待修改
		codeBranch = event //webhook 触发的 代码分枝
	}

	var stageBuildRec models.CiStageBuildLogs
	now := time.Now()
	stageBuildRec.FlowBuildId = flowBuildId
	stageBuildRec.StageId=stage.StageId
	stageBuildRec.BuildId = stageBuildId
	stageBuildRec.StageName=stage.StageName
	stageBuildRec.Status=common.STATUS_WAITING
	stageBuildRec.StartTime=now
	stageBuildRec.Namespace=user.Namespace
	stageBuildRec.IsFirst = 1
	stageBuildRec.BranchName = codeBranch
	//如果不是事件推送,就执行DefaultBranch
	if codeBranch == "" {
		stageBuildRec.BranchName = stage.DefaultBranch
	}

	//代码分支，前端触发
	if options.Branch != "" {
		stageBuildRec.BranchName = options.Branch
	}
	stageBuildRec.CreationTime = now
	// 如果 flow 开启了 “统一使用代码库” ，把构建时指定的 branch 保存在 stage 的 option 中
	// 在这个 stage 构建完成后，传递给下一个 stage 统一使用代码仓库UniformRepo
	if flow.UniformRepo == 0 {
		stage.Option = options
	}
	var flowBuildRec models.CiFlowBuildLogs
	flowBuildRec.BuildId = flowBuildId
	flowBuildRec.FlowId = flowId
	flowBuildRec.UserId = user.UserID
	flowBuildRec.CreationTime = now
	//InsertBuildLog will update 执行状态
	err = models.InsertBuildLog(&flowBuildRec, &stageBuildRec, stage.StageId)
	if err != nil {
		resp.Message = "InsertBuildLog to database failed"
		glog.Errorf("%s InsertBuildLog failed:%v", method, err)
		return resp, http.StatusServiceUnavailable
	}

	//开始使用websocket通知前端,开始构建
	NotifyFlowStatus(flowId, flowBuildId, common.STATUS_BUILDING)
	//开始此次构建
	if common.STATUS_BUILDING == stageBuildRec.Status {
		glog.Infof("======get build status============stageBuildRec:%#v\n", stageBuildRec)
		// Only use namespace for teamspace scope
		stageResp, code := StartStageBuild(user, stage, stageBuildRec, flow.Namespace)
		//构建失败
		if code != 200 {
			SetFailedStatus(user, stage, stageBuildRec, flow.Namespace)
			//通知websocket 失败
			NotifyFlowStatus(flowId, flowBuildId, common.STATUS_FAILED)
			resp.Message = "Unexpected error 构建失败 flowid=" + flowId
			return stageResp, code
		}
	}
	resp.Message = "构建成功"
	resp.FlowBuildId = flowBuildId
	resp.StageBuildId = stageBuildId
	return resp, http.StatusOK

}

type StageBuildResp struct {
	Message string `json:"message,omitempty"`
	Setting []models.Setting `json:"setting,omitempty"`
}

//StartStageBuild
func StartStageBuild(user *user.UserModel, stage models.CiStages, ciStagebuildLogs models.CiStageBuildLogs, flowOwer string) (interface{}, int) {
	method := "controller/StartStageBuild"
	var stageBuildResp StageBuildResp
	//project 查询失败
	project := models.NewCiManagedProjects()
	if stage.ProjectId != "" {
		err := project.FindProjectById(user.Namespace, stage.ProjectId)
		if err != nil || project.Id == "" {
			//project不存在，更新构建状态为失败
			glog.Errorf("%s find project failed project:%v  err:%v\n", method, project, err)
			SetFailedStatus(user, stage, ciStagebuildLogs, flowOwer)
			stageBuildResp.Message = "Project is inactive"
			return stageBuildResp, http.StatusForbidden
		}
	}
	glog.Infof("FlowId=%s,StageId=%s,FlowBuildId=%s,BuildId=%s\n",stage.FlowId, stage.StageId,
		ciStagebuildLogs.FlowBuildId, ciStagebuildLogs.BuildId)
	volumeMapping, message, respCode := models.GetVolumeSetting(stage.FlowId, stage.StageId,
		ciStagebuildLogs.FlowBuildId, ciStagebuildLogs.BuildId)
	if respCode != http.StatusOK {
		glog.Errorf("%s get volumeMapping failed: %s\n", method, message.Message)
		//修改状态并判断是否是单独构建 并执行下一个等待的stage
		SetFailedStatus(user, stage, ciStagebuildLogs, flowOwer)
		stageBuildResp.Message = message.Message
		stageBuildResp.Setting = volumeMapping
		return stageBuildResp, respCode
	}

	if common.HarborServerUrl == "" {
		configs := cluster.NewConfigs()
		harborServerUrl, err := configs.GetHarborServer()
		if err != nil {
			glog.Errorf("%s GetHarborServer failed:%v\n", method, err)
			stageBuildResp.Message = "get HarborServerUrl failed"
			return stageBuildResp, http.StatusConflict
		}
		common.HarborServerUrl = harborServerUrl
	}

	var buildInfo models.BuildInfo
	buildInfo.ClusterID = client.ClusterID
	buildInfo.BUILD_INFO_TYPE = 0
	buildInfo.RepoUrl = project.Address
	buildInfo.IsCodeRepo = 1
	if ciStagebuildLogs.BranchName != "" {
		buildInfo.Branch = ciStagebuildLogs.BranchName
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
	// Only build under user namespace or the owner of project(CI case)
	namespace := ""
	if user.Namespace != "" {
		namespace = user.Namespace
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
	buildInfo.FlowBuildId = ciStagebuildLogs.FlowBuildId
	buildInfo.StageBuildId = ciStagebuildLogs.BuildId
	buildInfo.Type = stage.Type
	buildInfo.ImageOwner = strings.ToLower(flowOwer)

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
			//自定义仓库时
			//TODO
		}
		//Dockerfile from where 2 onlinedockerdile code dockerfile
		if ONLINE == buildInfo.TargetImage.DockerfileFrom {
			//获取在线Dockerfile
			dockerfileOL, err := models.NewCiDockerfile().GetDockerfile(user.Namespace, stage.FlowId, stage.StageId)
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
	if ciStagebuildLogs.NodeName != "" {
		buildInfo.NodeName = ciStagebuildLogs.NodeName
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
			MakeScriptEntryEnvForInitContainer(user, containerInfo)
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
		err := repo.FindOneRepo(user.Namespace, models.DepotToRepoType(depot))
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

		if repo.UserInfo == "" {
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
		buildInfo.Git_repo_url = project.Address //TODO 正则 [A-Za-z0-9-_.]+
		buildInfo.PrivateKey = project.PrivateKey
		buildInfo.PublicKey = project.PublicKey
	}

	var buildRec models.CiStageBuildLogs
	buildRec.StartTime = time.Now()

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

	//var buildRec BuildRec
	nextStage, err := models.NewCiStage().FindNextOfFlow(stage.FlowId, stage.Seq)
	if err != nil && strings.Contains(fmt.Sprintf("%s", err), "no row") {
		glog.Errorf("%s %v\n", method, err)
		stageBuildResp.Message = " FindNextOfFlow failed"
		//return stageBuildResp, http.StatusNotFound

	}
	//如果没有下一个构建 BUILD_INFO_TYPE =1 用来是否删除编译好的二进制文件或者 满足构建多个构建镜像
	if nextStage.StageName == "" {
		buildInfo.BUILD_INFO_TYPE = 1 //显示还有下一个stage
	} else {
		buildInfo.BUILD_INFO_TYPE = 2 //显示没有下一个stage
	}

	//buildCluster = "CID-d7d3eb44c1db"

	imageBuilder := models.NewImageBuilder(buildCluster)

	//构建job的参数以及执行job命令
	job, err := imageBuilder.BuildImage(buildInfo, volumeMapping, common.HarborServerUrl)
	if err != nil || job == nil {
		buildRec.Status = common.STATUS_FAILED
		buildRec.EndTime = time.Now()
		glog.Errorf("%s BuildImage failed Err: %v\n", method, err)
		stageBuildResp.Message = "Failed to create job"
		return stageBuildResp, http.StatusInternalServerError
	}

	glog.Infof("%s build images job=%v\n", method, job)

	var options BuildStageOptions
	options.BuildWithDependency = buildWithDependency
	options.FlowOwner = flowOwer
	options.ImageName = buildInfo.TargetImage.Image
	options.UseCustomRegistry = false //不是客户的镜像仓库

	buildRec.JobName = job.ObjectMeta.Name
	buildRec.Namespace = job.ObjectMeta.Namespace

	//等待构建完成
	WaitForBuildToComplete(job, imageBuilder, user, stage, ciStagebuildLogs, options)

	pod, err := imageBuilder.GetPod(job.ObjectMeta.Namespace, job.ObjectMeta.Name)
	if err != nil {
		glog.Errorf("%s get pod info of %s from kubernetes failed:%v\n", method, job.ObjectMeta.Name, err)
		stageBuildResp.Message = "get pod failed from kubernetes"
		return stageBuildResp, http.StatusInternalServerError
	}

	if pod.ObjectMeta.Name != "" {
		buildRec.PodName = pod.ObjectMeta.Name
	}

	updateResult, err := models.NewCiStageBuildLogs().UpdateById(buildRec, ciStagebuildLogs.BuildId)
	if err != nil || updateResult < 1 {
		glog.Errorf("%s update stage build failed updateResult=%d err=:%v\n", method, updateResult, err)
	}

	if ciStagebuildLogs.FlowBuildId != "" {
		endTime := time.Now()
		//如果stage构建为flow构建中的第一步，则更新flow构建的起始时间。
		if 1 == ciStagebuildLogs.IsFirst {
			models.NewCiFlowBuildLogs().UpdateStartTimeAndStatusById(endTime, int(buildRec.Status), ciStagebuildLogs.FlowBuildId)
		}
		//如果stage构建失败，则更新flow结束时间
		if common.STATUS_FAILED == buildRec.Status {
			models.NewCiFlowBuildLogs().UpdateById(endTime, int(buildRec.Status), ciStagebuildLogs.FlowBuildId)
		}
	}

	data, _ := json.Marshal(ciStagebuildLogs)

	return string(data), http.StatusOK
}

func MakeScriptEntryEnvForInitContainer(user *user.UserModel, containerInfo models.Container) {
	scriptID := containerInfo.Scripts_id
	//userName:=user.Username
	//userToken:=""
	//if user.APIToken!=""{
	//	userToken=user.APIToken
	//}
	//TODO 加密解密的问题
	containerInfo.Command = []string{fmt.Sprintf("/app/%s", scriptID)}
	containerInfo.Env = []apiv1.EnvVar{
		{
			Name:  "SCRIPT_ENTRY_INFO",
			Value: scriptID,
		},
		{
			Name:  "SCRIPT_URL",
			Value: common.ScriptUrl,
		},
	}

}
//状态。0-成功 1-失败 2-执行中 3-等待 子任务     flow 状态。0-成功 1-失败 2-执行中
// update build status with 'currentBuild' and start next build of same stage
func UpdateStatusAndHandleWaiting(user *user.UserModel, stage models.CiStages,
	cistagebuildLogs models.CiStageBuildLogs, currentbuildId string, flowower string) {
	method := "updateStatusAndHandleWaiting"
	glog.Infof("%s The CurrentbuildId=%s\n",method, currentbuildId)
	//查询有没有 当前stage 处于 等待状态
	stageBuildLogs, total, err := models.NewCiStageBuildLogs().
		FindAllByIdWithStatus(stage.StageId, common.STATUS_WAITING)
	if err != nil {
		glog.Errorf("%s find stage build log failed :%v\n", method, err)
		return
	}
	glog.Infof("stageBuildLogs=%v\n",stageBuildLogs)
	if total == 0 {
		//如没有等待的构建，则更新当前构建状态
		res, err := models.NewCiStageBuildLogs().UpdateById(cistagebuildLogs, currentbuildId)
		if err != nil {
			glog.Errorf("%s update stage status failed:%d, Err:%v\n", method, res, err)
			return
		}
		return
	}

	//先将下一次构建的状态更新为“执行中”，再更新当前构建的状态
	_, err = models.NewCiStageBuildLogs().UpdateStageBuildStatusById(common.STATUS_BUILDING, stageBuildLogs[0].BuildId)
	if err != nil {
		glog.Errorf("%s UpdateStageBuildStatusById update stagebuild status failed from database  err: %v", method, err)
	}

	_, err = models.NewCiStageBuildLogs().UpdateById(cistagebuildLogs, currentbuildId)
	if err != nil {
		glog.Errorf("%s UpdateById update stagebuild status failed from database  err: %v", method, err)
	}

	StartStageBuild(user, stage, stageBuildLogs[0], flowower)

	return

}

func SetFailedStatus(user *user.UserModel, stage models.CiStages, cistageBuild models.CiStageBuildLogs, flowOwer string) {
	method := "setFailedStatus"
	flowBuildlog := models.CiFlowBuildLogs{}
	now := time.Now()
	flowBuildlog.EndTime = now
	flowBuildlog.Status = common.STATUS_FAILED
	if cistageBuild.FlowBuildId != "" && 1 != cistageBuild.BuildAlone { //是否单独构建
		//不是单独构建
		flowbuild, err := models.NewCiFlowBuildLogs().FindOneById(cistageBuild.FlowBuildId)
		if err != nil {

			glog.Errorf("%s find flow %s build failed from database err:%v \n", method, cistageBuild.FlowBuildId, err)

			return
		}
		if flowbuild.FlowId != "" {
			if fmt.Sprintf("%s", flowbuild.StartTime) == common.Time_NIL {
				flowBuildlog.StartTime = now
			}
		}
		_, err = models.NewCiFlowBuildLogs().UpdateById(flowBuildlog.EndTime, int(flowBuildlog.Status), cistageBuild.FlowBuildId)
		if err != nil {
			glog.Errorf("%s update stagebuild failed: %v\n", method, err)
		}
	}
	//处理下一个构建
	UpdateStatusAndHandleWaiting(user, stage,
		cistageBuild, cistageBuild.BuildId, flowOwer)

}

type BuildStageOptions struct {
	BuildWithDependency bool
	FlowOwner           string
	ImageName           string
	UseCustomRegistry   bool
}

func WaitForBuildToComplete(job *v1.Job, imageBuilder *models.ImageBuilder, user *user.UserModel, stage models.CiStages,
	stageBuild models.CiStageBuildLogs, options BuildStageOptions) {
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
		pod, timeout, err = HandleWaitTimeout(job, imageBuilder)
		if err != nil {
			glog.Infof("%s HandleWaitTimeout get: %v\n", method, err)
		}
		resultChan <- false
		//检查是否超时
		select {
		case <-time.After(3 * time.Minute):
			wg.Done()
			timeout = true
			glog.Infof("Kubernetes Job start timeout:%s\n", timeout)
		case <-resultChan:
			wg.Done()
			timeout = false
			glog.Infof("Kubernetes Job not timeout:%s\n", timeout)
		}

	}()
	wg.Wait()

	statusMessage := imageBuilder.WaitForJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name, options.BuildWithDependency)

	if statusMessage.JobStatus.JobConditionType == models.ConditionUnknown {
		glog.Warningf("%s Waiting for job failed, try again %#v\n", method, statusMessage.JobStatus)
		statusMessage = imageBuilder.WaitForJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name, options.BuildWithDependency)
	}
	statusCode := 1
	//手动停止
	if statusMessage.JobStatus.ForcedStop {
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

	var newBuild models.CiStageBuildLogs
	newBuild.EndTime = time.Now()
	if statusCode == 0 {
		newBuild.Status = common.STATUS_SUCCESS
	} else {
		newBuild.Status = common.STATUS_FAILED
	}

	if pod.ObjectMeta.Name == "" {
		pod, err = imageBuilder.GetPod(job.ObjectMeta.Namespace, job.ObjectMeta.Name)
		if err != nil {
			glog.Errorf("%s get pod from kubernetes cluster failed:%v\n", method, err)
		}
		if pod.ObjectMeta.Name != "" {
			//执行失败时，生成失败原因
			if statusCode == 1 {
				if len(pod.Status.ContainerStatuses) > 0 {
					for _, sontainerStatus := range pod.Status.ContainerStatuses {
						if sontainerStatus.Name == imageBuilder.BuilderName && sontainerStatus.State.Terminated != nil {
							errMsg = fmt.Sprintf(`运行构建的容器异常退出：exit code=%d，退出原因为:%s`, sontainerStatus.State.Terminated.ExitCode,
								sontainerStatus.State.Terminated.Message)
						}
					}
					if errMsg == "" && len(pod.Status.InitContainerStatuses) > 0 {
						for _, scmStatus := range pod.Status.InitContainerStatuses {
							if scmStatus.Name == imageBuilder.ScmName && scmStatus.State.Terminated != nil {
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

	if pod.ObjectMeta.Name != "" {
		newBuild.PodName = pod.ObjectMeta.Name
		if stageBuild.NodeName == "" {
			stageBuild.NodeName = pod.Spec.NodeName
		}
		newBuild.NodeName = stageBuild.NodeName
	}

	//执行失败时要停止相应的job
	if statusCode == 1 {
		glog.Warningf("%s Will Stop job: %s\n", method, job.ObjectMeta.Name)
		//执行失败时，终止job
		if !statusMessage.JobStatus.ForcedStop {
			glog.Infof("stop the run failed job job.ObjectMeta.Name=%s", job.ObjectMeta.Name)
			//不是手动停止
			errMsg = "程序停止构建job"
			_, err = imageBuilder.StopJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name, false, 0)
			if err != nil {
				glog.Errorf("%s Stop the job %s failed: %v\n", method, job.ObjectMeta.Name, err)
			}
		} else {
			errMsg = "构建流程被手动停止"
		}

		if stageBuild.FlowBuildId != "" {
			result, err := models.NewCiFlowBuildLogs().UpdateById(newBuild.EndTime, int(newBuild.Status), stageBuild.FlowBuildId)
			if err != nil || result < 1 {
				glog.Errorf("%s update flow build state failed from database:%v ;result=%d\n", method, err, result)
				models.NewCiFlowBuildLogs().UpdateById(newBuild.EndTime, int(newBuild.Status), stageBuild.FlowBuildId)
			}
			if !statusMessage.JobStatus.ForcedStop {
				NotifyFlowStatus(stage.FlowId, stageBuild.FlowBuildId, int(newBuild.Status))
			}

		}
	}

	//修改状态
	UpdateStatusAndHandleWaiting(user, stage, newBuild, stageBuild.BuildId, options.FlowOwner)

	if newBuild.Status == common.STATUS_SUCCESS {
		if stageBuild.FlowBuildId != "" && stageBuild.BuildAlone != 1 {
			glog.Infof("begin handleNextStageBuild\n")
			//TODO 通知下一个构建流程
			handleNextStageBuild(user, stage, stageBuild, options.FlowOwner)
		}
		//TODO 通知执行成功邮件
		glog.Infof("%s %s\n", method, errMsg)
	} else {
		//TODO 通知失败邮件
		glog.Infof("%s kubernetes run the juo failed:%s\n", method, errMsg)
	}

}

// start build of next stage if exists
func handleNextStageBuild(user *user.UserModel, stage models.CiStages,
	cistagebuildLogs models.CiStageBuildLogs, flowower string) {
	method := "handleNextStageBuild"
	flowBuildId := cistagebuildLogs.FlowBuildId
	nextStage, err := models.NewCiStage().FindNextOfFlow(stage.FlowId, stage.Seq)
	if err != nil {
		glog.Errorf("%s find next stage of flow %s failed from database %v", method, stage.FlowId, err)

	}
	glog.Info("handleNextStageBuild FlowBuildId===%s\n",flowBuildId)
	if nextStage.StageName != "" {
		//存在下一步时
		// 继承上一个 stage 的 options，例如构建时指定 branch
		nextStage.Option = stage.Option
		flowBuild, err := models.NewCiFlowBuildLogs().FindOneById(flowBuildId)
		if err != nil {
			glog.Errorf("%s get flowbuild info failed from database err:%v\n", method, err)
			// 查询出错时，触发下一步构建
			//TODO
			startNextStageBuild(user, nextStage, flowBuildId, cistagebuildLogs.NodeName, flowower)
			return
		}
		if flowBuild.FlowId == "" {
			//flow构建不存在
			return
		}

		if flowBuild.Status < common.STATUS_BUILDING {
			glog.Infof("flowBuild status:%d flowBuild of id:%s\n",flowBuild.Status,flowBuild.BuildId)
			// flow构建已经被stop，此时不再触发下一步构建
			glog.Warningf("%s Flow build is finished, build of next stage stageId:[%s] will not start", method, stage.StageId)
			return
		}

		glog.Infof("will start startNextStageBuild\n")
		//TODO
		startNextStageBuild(user, nextStage, flowBuildId, cistagebuildLogs.NodeName, flowower)

	} else {
		//不存在下一步时，更新flow构建状态为成功
		NotifyFlowStatus(stage.FlowId, flowBuildId, common.STATUS_SUCCESS)
		end_time := time.Now()
		models.NewCiFlowBuildLogs().UpdateById(end_time, common.STATUS_SUCCESS, flowBuildId)
	}

	return

}

func startNextStageBuild(user *user.UserModel, stage models.CiStages,
	flowBuildId, nodeName, flowower string) {

	method := "startNextStageBuild"

	stageBuildId := uuid.NewStageBuildID()
	stageBuildRec := models.CiStageBuildLogs{
		CreationTime:time.Now(),
		BuildId:     stageBuildId,
		FlowBuildId: flowBuildId,
		StageId:     stage.StageId,
		StageName:   stage.StageName,
		Status:      common.STATUS_WAITING,
		Namespace:   user.Namespace,
	}

	if stage.DefaultBranch != "" {
		stageBuildRec.BranchName = stage.DefaultBranch
	}

	if stage.Option.Branch != "" {
		stageBuildRec.BranchName = stage.Option.Branch
	}

	if nodeName != "" {
		stageBuildRec.NodeName = nodeName
	}
	//查询这个stage有没有 正在创建中
	stageBuildLogs, result, err := models.NewCiStageBuildLogs().FindAllByIdWithStatus(stage.StageId, common.STATUS_BUILDING)
	if err != nil || result <= 0 {
		glog.Warningf("%s find next stage of flow %s failed from database %v", method, stage.FlowId, err)
		glog.V(5).Infof("%s stage build logs %v  err:%v\n", method, stageBuildLogs, err)
		//没有执行中的构建记录，则添加“执行中”状态的构建记录
		stageBuildRec.Status = int8(common.STATUS_BUILDING)
	}
	//添加stage构建记录
	models.NewCiStageBuildLogs().InsertOne(stageBuildRec)
	glog.Infof("coming the startNextStageBuild =============>1")
	if stageBuildRec.Status == common.STATUS_BUILDING {
		glog.Infof("coming the startNextStageBuild =============>2")
		StartStageBuild(user, stage, stageBuildRec, flowower)
	}

	return

}

func HandleWaitTimeout(job *v1.Job, imageBuilder *models.ImageBuilder) (pod apiv1.Pod, timeout bool, err error) {
	method := "handleWaitTimeout"

	time.Sleep(3 * time.Second)

	pod, err = imageBuilder.GetPod(job.ObjectMeta.Namespace, job.ObjectMeta.Name)
	if err != nil {
		glog.Errorf("%s get %s pod failed:%v\n", method, pod.ObjectMeta.Name, err)
	}

	glog.Infof("%s - pod=[%s]<<===============>>", method, pod.ObjectMeta.Name)
	if pod.ObjectMeta.Name != "" {

		glog.Infof("%s - Checking if scm container is timeout\n", method)
		if len(pod.Status.InitContainerStatuses) > 0 &&
			IsContainerCreated(imageBuilder.ScmName, pod.Status.InitContainerStatuses) {
			glog.Infof("ContainerStatuses=========imageBuilder.BuilderName=:%s\n", imageBuilder.ScmName)
			timeout = false
			return
		}

		glog.Infof("%s - Checking if build container is timeout\n", method)
		if len(pod.Status.ContainerStatuses) > 0 &&
			IsContainerCreated(imageBuilder.BuilderName, pod.Status.ContainerStatuses) {
			glog.Infof("ContainerStatuses=========imageBuilder.BuilderName=:%s\n", imageBuilder.BuilderName)
			timeout = false
			return
		}

	}

	//终止job
	glog.Infof("%s - stop job=[%s]\n", method, job.ObjectMeta.Name)
	//1 代表手动停止 0表示程序停止
	_, err = imageBuilder.StopJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name, false, 0)
	if err != nil {
		glog.Errorf("%s Stop the job %s failed: %v\n", method, job.ObjectMeta.Name, err)
	}
	timeout = true
	return

}

func IsContainerCreated(ContainerName string, containerStatuses []apiv1.ContainerStatus) bool {

	for _, containerstatus := range containerStatuses {
		if ContainerName == containerstatus.Name {
			glog.Infof("The container %s status:%v\n", ContainerName, containerstatus.State)
			// 判断builder容器是否存在或是否重启过，从而判断是否容器创建成功
			if containerstatus.ContainerID != "" || containerstatus.RestartCount > 0 || containerstatus.State.Waiting != nil {
				glog.Infof("The container %s status:%v\n", ContainerName, containerstatus.State)
				return true
			}
		}

	}
	return false
}
