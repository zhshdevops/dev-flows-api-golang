package controllers

import (
	"dev-flows-api-golang/models"
	"github.com/golang/glog"
	"dev-flows-api-golang/modules/transaction"
	"encoding/json"
	"time"
	"strings"
	"errors"
	"dev-flows-api-golang/util/uuid"
	sqlstatus "dev-flows-api-golang/models/sql/status"
	"net/http"
	"dev-flows-api-golang/models/common"
	"regexp"
)

// stage类型：1-单元测试，2-代码编译，3-构建镜像，4-集成测试，5-自定义
const BUILD_IMAGE_STAGE_TYPE = 3
const DEFAULT_STAGE_TYPE = BUILD_IMAGE_STAGE_TYPE
const CUSTOM_STAGE_TYPE = 5
const STAGE_TYPE_MIN = 1
const STAGE_TYPE_MAX = CUSTOM_STAGE_TYPE

// Dockerfile来源：1-代码库，2-在线创建的
const FROM_REPO = 1
const ONLINE = 2
const DEFAULT_FROM = FROM_REPO
const FROM_MIN = FROM_REPO
const FROM_MAX = ONLINE

// 镜像仓库类型：1-为本地仓库 2-为DockerHub 3-为自定义
const LOCAL_REGISTRY = 1
const CUSTOM_REGISTRY = 3
const DEFAULT_REGISTRY_TYPE = LOCAL_REGISTRY
const REGISTRY_TYPE_MIN = LOCAL_REGISTRY
const REGISTRY_TYPE_MAX = CUSTOM_REGISTRY

// 镜像tag类型：1-代码分支为tag 2-时间戳为tag 3-自定义tag
const BRANCH_TAG = 1
const CUSTOM_TAG = 3
const DEFAULT_TAG_TYPE = BRANCH_TAG
const TAG_TYPE_MIN = BRANCH_TAG
const TAG_TYPE_MAX = CUSTOM_TAG

type CiStagesController struct {
	BaseController
}

//@router /:stage_id/dockerfile [PUT]
func (stage *CiStagesController) AddOrUpdateDockerfile() {
	method := "CiStagesController.AddOrUpdateDockerfile"
	flowId := stage.Ctx.Input.Param(":flow_id")
	stageId := stage.Ctx.Input.Param(":stage_id")

	stage.Audit.SetOperationType(models.AuditOperationUpdate)
	stage.Audit.SetResourceType(models.AuditResourceCIDockerfiles)

	var dockerRequest models.CiDockerfile
	body := stage.Ctx.Input.RequestBody
	if string(body) == "" {
		glog.Errorf("%s the request body is empty \n", method)
		stage.ResponseErrorAndCode("the request body is empty", http.StatusBadRequest)
		return
	}
	err := json.Unmarshal(body, &dockerRequest)
	if err != nil {
		glog.Errorf("%s json unmarshal failed:%v\n", method, err)
		stage.ResponseErrorAndCode("dokerfile json Unmarshal failed", http.StatusBadRequest)
		return
	}
	// Check if dockerfile already exist before add a new one
	ciDockerfile := models.NewCiDockerfile()
	dockerfile, err := ciDockerfile.GetDockerfile(stage.Namespace, flowId, stageId)
	if err != nil {
		parseNumber, _ := sqlstatus.ParseErrorCode(err)
		if parseNumber != sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s Get dockerfile info by flowId failed from database ERR: %v\n", method, err)
			stage.ResponseErrorAndCode("Lookup dockerfile failed,please try again!", http.StatusForbidden)
			return
		}
	}
	//if exist will update
	if dockerfile.StageId != "" {
		result, err := ciDockerfile.UpdateDockerfile(stage.Namespace, flowId, stageId, stage.User.Username, dockerRequest)
		if err != nil || result < 1 {
			glog.Errorf("%s update dockerfile failed: %v\n", method, err)
			stage.ResponseErrorAndCode("update dockerfile failed", http.StatusConflict)
			return
		}
	} else {
		dockerRequest.FlowId = flowId
		dockerRequest.StageId = stageId
		dockerRequest.Namespace = stage.Namespace
		dockerRequest.ModifiedBy = stage.User.Username
		dockerRequest.CreateTime = time.Now().Format("2006-01-02 15:04:05")
		dockerRequest.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
		_, err = ciDockerfile.AddDockerfile(dockerRequest)
		if err != nil {
			glog.Errorf("%s add dockerfile failed:%v\n", method)
			stage.ResponseErrorAndCode("Dockerfile added failed.", http.StatusConflict)
			return
		}
		stage.ResponseErrorAndCode("Dockerfile added successfully.", http.StatusOK)
		return
	}

	stage.ResponseErrorAndCode("Dockerfile update successfully.", http.StatusOK)
	return

}

//@router /:stage_id/dockerfile [POST]
func (stage *CiStagesController) AddDockerfile() {

	method := "CiStagesController.AddDockerfile"

	flowId := stage.Ctx.Input.Param(":flow_id")

	stageId := stage.Ctx.Input.Param(":stage_id")
	stage.Audit.SetOperationType(models.AuditOperationCreate)
	stage.Audit.SetResourceType(models.AuditResourceCIDockerfiles)
	body := stage.Ctx.Input.RequestBody

	if string(body) == "" {
		glog.Errorf("%s the request body is empty \n", method)
		stage.ResponseErrorAndCode("the request body is empty", http.StatusBadRequest)
		return
	}

	var dockerRequest models.CiDockerfile

	err := json.Unmarshal(body, &dockerRequest)
	if err != nil {
		glog.Errorf("%s json unmarshal failed: %v\n", method, err)
		stage.ResponseErrorAndCode("dockerfile json Unmarshal failed", http.StatusBadRequest)
		return
	}

	ciDockerfile := models.NewCiDockerfile()
	//check stage info if exist
	cistage, err := models.NewCiStage().FindOneById(stageId)
	if err != nil {
		parseNumber, _ := sqlstatus.ParseErrorCode(err)
		if parseNumber == sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s get stage failed from database:%v\n", method, err)
			stage.ResponseErrorAndCode("Stage "+stageId+" does not exist yet.", http.StatusNotFound)
			return
		} else {
			glog.Errorf("%s Get Stage info by flowId failed from database ERR: %v\n", method, err)
			stage.ResponseErrorAndCode("Lookup Stage failed,please try again!", http.StatusForbidden)
			return
		}

	}

	if cistage.FlowId == "" {
		glog.Errorf("%s get stage failed from database:%v\n", method, err)
		stage.ResponseErrorAndCode("Stage "+stageId+" does not exist yet.", http.StatusNotFound)
		return
	}

	// Check if dockerfile already exist before add a new one
	dockerfile, err := ciDockerfile.GetDockerfile(stage.Namespace, flowId, stageId)
	if err != nil {
		parseNumber, _ := sqlstatus.ParseErrorCode(err)
		if parseNumber != sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s Get dockerfile info by flowId failed from database ERR: %v\n", method, err)
			stage.ResponseErrorAndCode("Lookup dockerfile failed,please try again!", http.StatusForbidden)
			return
		}
	}
	if dockerfile.FlowId != "" {
		stage.ResponseErrorAndCode("Dockerfile for this stage already exists", http.StatusConflict)
		return
	}

	dockerRequest.FlowId = flowId
	dockerRequest.StageId = stageId
	dockerRequest.Namespace = stage.Namespace
	dockerRequest.ModifiedBy = stage.User.Username
	dockerRequest.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	dockerRequest.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	_, err = ciDockerfile.AddDockerfile(dockerRequest)
	if err != nil {
		glog.Errorf("%s add dockerfile failed: %v\n", method, err)
		stage.ResponseErrorAndCode("added Dockerfile  failed.", http.StatusConflict)
		return
	}

	stage.ResponseErrorAndCode("Dockerfile added successfully.", http.StatusOK)
	return

}

//@router /:stage_id/dockerfile [DELETE]
func (stage *CiStagesController) RemoveDockerfile() {

	method := "CiStagesController.RemoveDockerfile"

	flowId := stage.Ctx.Input.Param(":flow_id")

	stageId := stage.Ctx.Input.Param(":stage_id")

	stage.Audit.SetResourceID(stageId)
	stage.Audit.SetOperationType(models.AuditOperationDelete)
	stage.Audit.SetResourceType(models.AuditResourceCIDockerfiles)

	ciDockerfile := models.NewCiDockerfile()

	result, err := ciDockerfile.RemoveDockerfile(stage.Namespace, flowId, stageId)
	if err != nil || result < 1 {
		parseNumber, _ := sqlstatus.ParseErrorCode(err)
		if parseNumber != sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s Get dockerfile info by flowId failed from database ERR: %v\n", method, err)
			stage.ResponseErrorAndCode("Lookup dockerfile failed,please try again!", http.StatusForbidden)
			return
		} else {
			glog.Errorf("%s No dockerfile mathcing the flow and stage id %v \n", method, err)
			stage.ResponseErrorAndCode("No dockerfile mathcing the flow and stage id", http.StatusNotFound)
			return
		}

	}

	stage.ResponseNotFoundDevops("Dockerfile removed successfully")
	return

}

//@router /:stage_id/dockerfile [GET]
func (stage *CiStagesController) GetDockerfile() {

	method := "CiStagesController.GetDockerfile"

	flowId := stage.Ctx.Input.Param(":flow_id")

	stageId := stage.Ctx.Input.Param(":stage_id")

	ciDockerfile := models.NewCiDockerfile()
	glog.Infof("stage.Namespace:%s\n",stage.Namespace)
	dockerfileInfo, err := ciDockerfile.GetDockerfile(stage.Namespace, flowId, stageId)
	if err != nil {
		parseNumber, _ := sqlstatus.ParseErrorCode(err)
		if parseNumber == sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s get dockerfile failed from database:%v\n", method, err)
			stage.ResponseErrorAndCode("dockerfile cannot be found", http.StatusNotFound)
			return
		} else {
			glog.Errorf("%s Get dockerfile info by flowId failed from database ERR: %v\n", method, err)
			stage.ResponseErrorAndCode("Lookup dockerfile failed,please try again!", http.StatusForbidden)
			return
		}
	}

	stage.ResponseSuccessStatusAndMessageDevops(dockerfileInfo)

	return

}

//@router /:stage_id [DELETE]
func (stage *CiStagesController) RemoveStage() {

	method := "CiStagesController.RemoveStage"

	flowId := stage.Ctx.Input.Param(":flow_id")

	stageId := stage.Ctx.Input.Param(":stage_id")

	stage.Audit.SetResourceID(stageId)
	stage.Audit.SetOperationType(models.AuditOperationDelete)
	stage.Audit.SetResourceType(models.AuditResourceStages)
	//检查是否为flow的最后一个stage
	stages, err := models.NewCiStage().FindExpectedLast(flowId, stageId)
	if err != nil || stages.StageName == "" {
		glog.Errorf("%s FindExpectedLast failed:%v\n", method, err)
		stage.ResponseErrorAndCode("Only the last stage of the flow can be removed", http.StatusForbidden)
		return
	}

	var containerInfo models.Container
	if strings.Contains(stages.ContainerInfo, "scripts_id") {

		err = json.Unmarshal([]byte(stages.ContainerInfo), &containerInfo)
		if err != nil {
			glog.Errorf("%s json unmarshal container info failed: %v\n", method, err)
			stage.ResponseErrorAndCode("json unmarshal container info failed:", http.StatusForbidden)
			return
		}

		_, err = models.NewCiScripts().DeleteScriptByID(containerInfo.Scripts_id)
		if err != nil {
			glog.Errorf("%s DeleteScriptByID failed:%v\n", method, err)

		}

	}

	_, err = models.NewCiStage().DeleteById(stageId)
	if err != nil {
		glog.Errorf("%s delete stage by id %s from database failed:%v\n", method, stageId, err)
		stage.ResponseErrorAndCode("delete stage"+stageId+" failed", http.StatusInternalServerError)
		return
	}
	type Resp struct {
		StageId string  `json:"stageId"`
	}

	var resp Resp

	resp.StageId = stageId

	stage.ResponseSuccessCIRuleDevops(resp)

	return

}

//@router / [GET]
func (stage *CiStagesController) ListStages() {

	method := "CiStagesController.ListStages"

	flowId := stage.Ctx.Input.Param(":flow_id")

	cistage := models.NewCiStage()

	stages, total, err := cistage.FindWithLinksByFlowId(flowId)
	if err != nil {

		glog.Errorf("%s %v", method, err)

		stage.ResponseErrorAndCode("FindWithLinksByFlowId failed from database ", http.StatusForbidden)
		return
	}
	stageInfo := make([]models.Stage_info, 0)
	for _, sta := range stages {
		stageInfo = append(stageInfo, models.FormatStage(sta))
	}

	stage.ResponseSuccessDevops(stageInfo, total)
	return

}

//{"DockerfileFrom":1,"registryType":1,"imageTagType":2,
//"noCache":false,"image":"xinzhiyuntest","project":"qinzhao-harbor","projectId":7,
//"DockerfileName":"Dockerfile","DockerfilePath":"/"}
//{"DockerfileFrom":1,"DockerfileName":"Dockerfile","DockerfilePath":"/","image":"dsa","imageTagType":2,
//"project":"qinzhao-harbor","projectId":7,"registryType":1}
//@router / [POST]
func (stage *CiStagesController) CreateStage() {
	method := "controllers/CiStagesController.CreateStage"

	flowId := stage.Ctx.Input.Param(":flow_id")

	stage.Audit.SetOperationType(models.AuditOperationCreate)
	stage.Audit.SetResourceType(models.AuditResourceStages)

	if string(stage.Ctx.Input.RequestBody) == "" {
		glog.Errorf("%s %s\n", method, "the request body is empty")
		stage.ResponseErrorAndCode("the request body is empty", http.StatusBadRequest)
		return
	}
	//获取flow检查是否存在相应的数据
	ciFlow, err := models.NewCiFlows().FindFlowById(stage.Namespace, flowId)
	if err != nil {
		parnumber, _ := sqlstatus.ParseErrorCode(err)
		if parnumber == sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s get flow info failed:%v\n", method, err)
			stage.ResponseErrorAndCode("CI flow cannot be found", http.StatusNotFound)
			return
		}
		glog.Errorf("%s get flow info from database failed:%v\n", method, err)
		stage.ResponseErrorAndCode("CI flow cannot be found", http.StatusInternalServerError)
		return
	}

	if ciFlow.Name == "" {
		glog.Errorf("%s get flow info failed:%v\n", method, err)
		stage.ResponseErrorAndCode("CI flow cannot be found", http.StatusNotFound)
		return
	}
	//解析stage——info
	var stageInfo models.Stage_info
	err = json.Unmarshal(stage.Ctx.Input.RequestBody, &stageInfo)
	if err != nil {
		glog.Errorf("%s RequestBody json unmarshal failed:%v\n", method, err)
		stage.ResponseErrorAndCode("RequestBody json 解析失败", http.StatusBadRequest)
		return
	}

	//检查提交的stage中必须的字段
	//检查提交的stage中必须的字段 //_checkRequired  ==========begin
	if stageInfo.Metadata.Name == "" {
		stage.ResponseErrorAndCode(".metadata.name is required", http.StatusBadRequest)
		return
	}

	if &stageInfo.Spec == nil {
		stage.ResponseErrorAndCode(".spec is required", http.StatusBadRequest)
		return
	}
	//如果是创建镜像，检查镜像
	if stageInfo.Metadata.Type != common.BUILD_IMAGE_STAGE_TYPE {

		if stageInfo.Spec.Container.Image == "" {

			stage.ResponseErrorAndCode(".spec.container.image is required", http.StatusBadRequest)
			return
		}
	}
	//_checkRequired  ==========end

	if !models.NewCiImages().IsValidImages(false, stage.Namespace, stageInfo.Spec.Container.Image) {
		stage.ResponseErrorAndCode("Unknown build image:"+stageInfo.Spec.Container.Image, http.StatusBadRequest)
		return
	}

	if stageInfo.Spec.Project.Id != "" {
		project := models.NewCiManagedProjects()
		err = project.FindProjectOnlyById(stageInfo.Spec.Project.Id)
		if err != nil || project.Id == "" {
			parseNumber, _ := sqlstatus.ParseErrorCode(err)
			if parseNumber == sqlstatus.SQLErrNoRowFound {
				glog.Errorf("%s not found the project: err=%v\n", method, err)
				stage.ResponseErrorAndCode("Project does not exist: "+stageInfo.Spec.Project.Id, http.StatusForbidden)
				return
			} else {
				glog.Errorf("%s get the project info failed from database:%v\n", method, err)
				stage.ResponseErrorAndCode("not found this project "+stageInfo.Spec.Project.Id, http.StatusBadRequest)
				return
			}
		}

	} else {

		stageInfo.Spec.Project.Id = ""
		stageInfo.Spec.Project.Branch = ""

	}

	if len(stageInfo.Spec.Container.Dependencies) > 0 {
		images := make([]string, 0)
		for _, depen := range stageInfo.Spec.Container.Dependencies {
			images = append(images, depen.Service)
		}
		imageInfo := strings.Join(images, ",")
		if !models.NewCiImages().IsValidImages(true, stage.Namespace, imageInfo) {
			stage.ResponseErrorAndCode("Unknown budependency service image:"+imageInfo, http.StatusBadRequest)
			return
		}
	}

	//检查type范围 默认就是构建镜像
	buildType := stageInfo.Metadata.Type
	if buildType == 0 {
		buildType = common.BUILD_IMAGE_STAGE_TYPE
	}
	if buildType < STAGE_TYPE_MIN || buildType > STAGE_TYPE_MAX {
		stage.ResponseErrorAndCode("Invalid .metadata.type", http.StatusForbidden)
		return
	}

	// 自定义类型时，检查是否设置了自定义类型的文本
	if CUSTOM_STAGE_TYPE == buildType && ( stageInfo.Metadata.CustomType == "" ||
		strings.TrimSpace(stageInfo.Metadata.CustomType) == "") {
		stage.ResponseErrorAndCode(".metadata.customType is required with custom stage type", http.StatusForbidden)
		return
	}
	//===============客户自定义的构建类型
	stageInfo.Metadata.Type = buildType
	if buildType == CUSTOM_STAGE_TYPE {
		stageInfo.Metadata.CustomType = strings.TrimSpace(stageInfo.Metadata.CustomType)
	}
	//如果是构建镜像
	if buildType == BUILD_IMAGE_STAGE_TYPE {
		// stage为构建镜像类型时，需要.spec.build.image
		if stageInfo.Spec.Build.Image == "" {
			stage.ResponseErrorAndCode(".spec.build.image is required with this stage type", http.StatusForbidden)
			return
		}

		parts := strings.Split(stageInfo.Spec.Build.Image, "/")
		stageInfo.Spec.Build.Image = strings.ToLower(parts[len(parts)-1])
		//校验镜像名称
		if regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*$`).FindString(stageInfo.Spec.Build.Image) == "" {
			stage.ResponseErrorAndCode(".spec.build.image contains invalid character(s)", http.StatusForbidden)
			return
		}

		//1 表示代码仓库repo_dockerfile 2 表示在线Dockerfile
		if stageInfo.Spec.Build.DockerfileFrom == 0 {
			stageInfo.Spec.Build.DockerfileFrom = DEFAULT_FROM
		}
		//检查.spec.build.DockerfileFrom范围
		if stageInfo.Spec.Build.DockerfileFrom < FROM_MIN || stageInfo.Spec.Build.DockerfileFrom > FROM_MAX {

			stage.ResponseErrorAndCode("Invalid .spec.build.DockerfileFrom", http.StatusForbidden)
			return
		}

		if stageInfo.Spec.Build.DockerfilePath == "" {
			stageInfo.Spec.Build.DockerfilePath = "/"
		}

		if stageInfo.Spec.Build.DockerfileFrom == ONLINE {
			stageInfo.Spec.Build.DockerfileName = "Dockerfile"
		}

		//检查.spec.build.registryType
		if stageInfo.Spec.Build.RegistryType == 0 {
			stageInfo.Spec.Build.RegistryType = DEFAULT_REGISTRY_TYPE
		}
		if stageInfo.Spec.Build.RegistryType > REGISTRY_TYPE_MAX || stageInfo.Spec.Build.RegistryType < REGISTRY_TYPE_MIN {
			stage.ResponseErrorAndCode("Invalid .spec.build.registryType", http.StatusForbidden)
			return
		}

		//自定义registry时，检查customRegistry是否存在

		if CUSTOM_REGISTRY == stageInfo.Spec.Build.RegistryType {
			if stageInfo.Spec.Build.CustomRegistry == "" ||
				strings.TrimSpace(stageInfo.Spec.Build.CustomRegistry) == "" {
				stage.ResponseErrorAndCode(".spec.build.customRegistry is required with custom registry type", http.StatusForbidden)
				return
			}
			//TODO 暂时不考虑第三方的镜像

		}

		//检查.spec.build.imageTagType范围
		if stageInfo.Spec.Build.ImageTagType == 0 {
			stageInfo.Spec.Build.ImageTagType = DEFAULT_TAG_TYPE
		}
		if stageInfo.Spec.Build.ImageTagType < TAG_TYPE_MIN || stageInfo.Spec.Build.ImageTagType > TAG_TYPE_MAX {
			stage.ResponseErrorAndCode("Invalid .spec.build.imageTagType", http.StatusForbidden)
			return
		}

		//自定义tag类型时，检查.spec.build.customTag是否存在
		if stageInfo.Spec.Build.ImageTagType == CUSTOM_TAG &&
			(stageInfo.Spec.Build.CustomTag == "" || strings.TrimSpace(stageInfo.Spec.Build.CustomTag) == "") {
			stage.ResponseErrorAndCode(".spec.build.customTag is required with custom tag", http.StatusForbidden)
			return
		}
		//默认无缓存

	}
	// ============>_checkAndSetDefaults end
	glog.Infof("%s stageInfo:%#v \n", method, stageInfo)
	glog.Infof("%s stageInfo Build:%#v \n", method, stageInfo.Spec.Build)

	//var stageRec models.CiStages
	// 判断name是否唯一 begin
	checkStateInfo, err := models.NewCiStage().FindOneByName(flowId, stageInfo.Metadata.Name)
	if err != nil {
		parseNumber, _ := sqlstatus.ParseErrorCode(err)
		if parseNumber != sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s get stage from stage failed:%v\n", method, err)
			stage.ResponseErrorAndCode("look up stage name failed", http.StatusInternalServerError)
			return
		}
	}
	if checkStateInfo.StageId != "" {
		stage.ResponseErrorAndCode("Stage already exists", http.StatusForbidden)
		return
	}
	// 判断name是否唯一 ================ end
	// 当用户勾选 “当前流水线所有任务（包括新建任务），统一使用该代码库” 时，更新其他 state 以及 flow
	stage.updateForUniformRepo(stage.Namespace, flowId, stageInfo)
	//构建stage 信息
	maxSeq, err := models.NewCiStage().FindFlowMaxSeq(flowId)
	if err != nil {
		glog.Errorf("%s get max seq failed:%v\n", method, err)
		stage.ResponseErrorAndCode("get max seq failed", http.StatusForbidden)
		return
	}

	if maxSeq == 0 {
		maxSeq = 1
	} else {
		maxSeq = maxSeq + 1
	}

	stageDb := models.CiStages{}
	stageDb.StageId = uuid.NewStageID()
	stageDb.Seq = maxSeq
	stageDb.StageName = stageInfo.Metadata.Name
	stageDb.ProjectId = stageInfo.Spec.Project.Id
	stageDb.DefaultBranch = stageInfo.Spec.Project.Branch
	stageDb.Type = stageInfo.Metadata.Type
	stageDb.CustomType = stageInfo.Metadata.CustomType
	stageDb.Image = stageInfo.Spec.Container.Image
	stageDb.CiEnabled = stageInfo.Spec.Ci.Enabled
	stageDb.FlowId = flowId
	//ci config
	ciConfig, err := json.Marshal(stageInfo.Spec.Ci.CiConfig)
	if err != nil {
		glog.Errorf("%s json marshal ci config failed: %v\n", method, err)
	}
	stageDb.CiConfig = string(ciConfig)
	//container info
	containerInfo, err := json.Marshal(stageInfo.Spec.Container)
	if err != nil {
		glog.Errorf("%s json marshal container info failed:%v\n", method, err)
	}
	stageDb.ContainerInfo = string(containerInfo)
	stageDb.CreationTime = time.Now()
	//build info
	if stageInfo.Spec.Build != nil {
		buildInfo, err := json.Marshal(stageInfo.Spec.Build)
		if err != nil {
			glog.Errorf("%s json marshal build info failed %v\n", method, err)
		}
		stageDb.BuildInfo = string(buildInfo)
	}
	glog.Infof("stageDb.BuildInfo:%v\n", stageDb.BuildInfo)
	//事物处理
	recordDB := func() bool {
		trans := transaction.New()
		trans.Do(func() {
			_, err := models.NewCiStage().InsertOneStage(stageDb, trans.O())
			if err != nil {
				trans.Rollback(method, "insert stage info to database failed", err)
			}

			links, _, err := models.NewCiStageLinks().GetNilTargets(flowId)
			if err != nil {
				trans.Rollback(method, "get nil targets from database failed", err)
			}
			glog.Info("links info :%#v\n", links)
			ciLink := models.CiStageLinks{
				SourceId: stageDb.StageId,
				FlowId:   flowId,
			}
			if len(links) < 2 {
				//没有或只有一个tail时，将当前stage作为tail插入flow
				linkNumber, err := models.NewCiStageLinks().InsertOneLink(ciLink, trans.O())
				if err != nil || linkNumber < 1 {
					trans.Rollback(method, " insert stage link to database failed:", err)
				}
				//存在tail时，更新旧tail的target为新建的stage
				if len(links) == 1 {
					oldTail := models.CiStageLinks{
						TargetId: stageDb.StageId,
					}
					//如果指定了旧tail与当前stage的link directories，则一并更新
					//更新link
					_, err = models.NewCiStageLinks().UpdateOneBySrcIdNew(oldTail, links[0].SourceId, trans.O())
					if err != nil {
						trans.Rollback(method, "update target id in database failed", err)
					}
				}

			}

		}).Done()

		return trans.IsCommit()
	}

	if !recordDB() {

		stage.ResponseErrorAndCode("the database happend error", http.StatusInternalServerError)
		return
	}

	resp := struct {
		StageId string `json:"stageId"`
	}{
		StageId: stageDb.StageId,
	}

	stage.ResponseSuccessCIRuleDevops(resp)

	return

}

//@router /:stage_id [GET]
func (stage *CiStagesController) GetStage() {

	method := "CiStagesController.GetStage"

	flowId := stage.Ctx.Input.Param(":flow_id")

	stageId := stage.Ctx.Input.Param(":stage_id")

	cistage := models.NewCiStage()

	stageInfo, err := cistage.FindOneById(stageId)
	if err != nil || stageInfo.StageName == "" {
		glog.Errorf("%s %v\n", method, err)
		stage.ResponseErrorAndCode("not found this stage "+stageId, 404)
		return
	}

	if stageInfo.FlowId != flowId {
		stage.ResponseErrorAndCode("Stage does not belong to Flow", 501)
		return
	}

	stage.ResponseSuccess(stageInfo)

	return

}

func (stage *CiStagesController) checkUnique(flowId, name string) error {

	method := "CiStagesController.checkUnique"

	stageInfo, err := models.NewCiStage().FindOneByName(flowId, name)
	if err != nil {
		glog.Errorf("%s Stage name conflict:  %v\n", method, err)
		return err
	}

	if stageInfo.StageName == "" {
		return nil
	}

	return errors.New("Stage already exists")

}

//0：统一使用一个代码库
//1：不统一使用一个代码库
//// 当用户勾选 “当前流水线所有任务（包括新建任务），统一使用该代码库” 时，更新其他 state 以及 flow
func (stage *CiStagesController) updateForUniformRepo(namespace, flowId string, stageInfo models.Stage_info) {

	method := "CiStagesController.updateForUniformRepo"

	ciFlow := models.NewCiFlows()

	ciFlowInfo, err := ciFlow.FindFlowById(namespace, flowId)
	if err != nil {
		glog.Errorf("%s ciflow find flow by id failed:  %v\n", method, err)
		return
	}
	//相等则不用修改了
	if ciFlowInfo.UniformRepo == stageInfo.Spec.UniformRepo {
		return
	}

	updateResult, err := ciFlow.Update(namespace, flowId, stageInfo.Spec.UniformRepo)
	if updateResult != 1 || err != nil {
		glog.Errorf("%s update failed: %v\n", method, err)
		return
	}

	if stageInfo.Spec.UniformRepo == 0 {
		err = models.NewCiStage().Update(flowId, stageInfo.Spec.Project.Id, stageInfo.Spec.Project.Branch)
		if err != nil {
			glog.Errorf("%s update stage info failed: %v\n", method, err)
			return
		}
	}
	return

}

//@router /:stage_id [PUT]
func (stage *CiStagesController) UpdateStage() {

	method := "CiStagesController.UpdateStage"

	flowId := stage.Ctx.Input.Param(":flow_id")

	stageId := stage.Ctx.Input.Param(":stage_id")

	stage.Audit.SetResourceID(stageId)
	stage.Audit.SetOperationType(models.AuditOperationUpdate)
	stage.Audit.SetResourceType(models.AuditResourceStages)

	if string(stage.Ctx.Input.RequestBody) == "" {
		glog.Errorf("%s %s\n", method, "the request body is empty")
		stage.ResponseErrorAndCode("the request body is empty", http.StatusBadRequest)
		return
	}
	//获取stage并检查stage是否属于flow

	oldStageInfo, err := models.NewCiStage().FindOneById(stageId)
	if err != nil {
		parseNumber, _ := sqlstatus.ParseErrorCode(err)
		if parseNumber == sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s not found the stage: err=%v\n", method, err)
			stage.ResponseErrorAndCode("not found this stage "+stageId, http.StatusNotFound)
			return
		} else {
			glog.Errorf("%s get the stage info failed from database:%v\n", method, err)
			stage.ResponseErrorAndCode("not found this stage "+stageId, http.StatusBadRequest)
			return
		}
	}

	if oldStageInfo.StageName == "" {
		glog.Errorf("%s not found the stage: err=%v\n", method, err)
		stage.ResponseErrorAndCode("not found this stage "+stageId, http.StatusNotFound)
		return
	}

	if oldStageInfo.FlowId != flowId {
		stage.ResponseErrorAndCode("Stage does not belong to Flow", http.StatusBadRequest)
		return
	}

	//检查提交的stage字段有效性

	var stageInfo models.Stage_info
	body := stage.Ctx.Input.RequestBody
	if string(body) == "" {
		glog.Errorf("%s RequestBody is empty\n", method)
		stage.ResponseErrorAndCode("RequestBody is empty", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &stageInfo)
	if err != nil {
		glog.Errorf("%s json unmarshal failed: %v\n", method, err)
		stage.ResponseErrorAndCode("json 解析失败", http.StatusBadRequest)
		return
	}

	//检查提交的stage中必须的字段 //_checkRequired  ==========begin
	if stageInfo.Metadata.Name == "" {
		stage.ResponseErrorAndCode(".metadata.name is required", http.StatusBadRequest)
		return
	}

	if &stageInfo.Spec == nil {
		stage.ResponseErrorAndCode(".spec is required", http.StatusBadRequest)
		return
	}
	//如果是创建镜像，检查镜像
	if stageInfo.Metadata.Type != common.BUILD_IMAGE_STAGE_TYPE {

		if stageInfo.Spec.Container.Image == "" {

			stage.ResponseErrorAndCode(".spec.container.image is required", http.StatusBadRequest)
			return
		}
	}
	//_checkRequired  ==========end

	if !models.NewCiImages().IsValidImages(false, stage.Namespace, stageInfo.Spec.Container.Image) {
		stage.ResponseErrorAndCode("Unknown build image:"+stageInfo.Spec.Container.Image, http.StatusBadRequest)
		return
	}

	if stageInfo.Spec.Project.Id != "" {
		project := models.NewCiManagedProjects()
		err = project.FindProjectOnlyById(stageInfo.Spec.Project.Id)
		if err != nil || project.Id == "" {
			parseNumber, _ := sqlstatus.ParseErrorCode(err)
			if parseNumber == sqlstatus.SQLErrNoRowFound {
				glog.Errorf("%s not found the project: err=%v\n", method, err)
				stage.ResponseErrorAndCode("Project does not exist: "+stageInfo.Spec.Project.Id, http.StatusForbidden)
				return
			} else {
				glog.Errorf("%s get the project info failed from database:%v\n", method, err)
				stage.ResponseErrorAndCode("not found this project "+stageInfo.Spec.Project.Id, http.StatusBadRequest)
				return
			}
		}

	}

	if len(stageInfo.Spec.Container.Dependencies) > 0 {
		images := make([]string, 0)
		for _, depen := range stageInfo.Spec.Container.Dependencies {
			images = append(images, depen.Service)
		}
		imageInfo := strings.Join(images, ",")
		if !models.NewCiImages().IsValidImages(true, stage.Namespace, imageInfo) {
			stage.ResponseErrorAndCode("Unknown budependency service image:"+imageInfo, http.StatusBadRequest)
			return
		}
	}

	//检查type范围 默认就是构建镜像
	buildType := stageInfo.Metadata.Type
	if buildType == 0 {
		buildType = common.BUILD_IMAGE_STAGE_TYPE
	}
	if buildType < STAGE_TYPE_MIN || buildType > STAGE_TYPE_MAX {
		stage.ResponseErrorAndCode("Invalid .metadata.type", http.StatusForbidden)
		return
	}

	// 自定义类型时，检查是否设置了自定义类型的文本
	if CUSTOM_STAGE_TYPE == buildType && ( stageInfo.Metadata.CustomType == "" ||
		strings.TrimSpace(stageInfo.Metadata.CustomType) == "" ) {
		stage.ResponseErrorAndCode(".metadata.customType is required with custom stage type", http.StatusForbidden)
		return
	}
	//===============客户自定义的构建类型
	stageInfo.Metadata.Type = buildType
	if buildType == CUSTOM_STAGE_TYPE {
		stageInfo.Metadata.CustomType = strings.TrimSpace(stageInfo.Metadata.CustomType)
	}
	//如果是构建镜像
	if buildType == BUILD_IMAGE_STAGE_TYPE {
		// stage为构建镜像类型时，需要.spec.build.image
		if stageInfo.Spec.Build.Image == "" {
			stage.ResponseErrorAndCode(".spec.build.image is required with this stage type", http.StatusForbidden)
			return
		}

		parts := strings.Split(stageInfo.Spec.Build.Image, "/")
		stageInfo.Spec.Build.Image = strings.ToLower(parts[len(parts)-1])
		//校验镜像名称
		if regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*$`).FindString(stageInfo.Spec.Build.Image) == "" {
			stage.ResponseErrorAndCode(".spec.build.image contains invalid character(s)", http.StatusForbidden)
			return
		}

		//1 表示代码仓库repo_dockerfile 2 表示在线Dockerfile
		if stageInfo.Spec.Build.DockerfileFrom == 0 {
			stageInfo.Spec.Build.DockerfileFrom = DEFAULT_FROM
		}
		//检查.spec.build.DockerfileFrom范围
		if stageInfo.Spec.Build.DockerfileFrom < FROM_MIN || stageInfo.Spec.Build.DockerfileFrom > FROM_MAX {

			stage.ResponseErrorAndCode("Invalid .spec.build.DockerfileFrom", http.StatusForbidden)
			return
		}

		if stageInfo.Spec.Build.DockerfilePath == "" {
			stageInfo.Spec.Build.DockerfilePath = "/"
		}

		if stageInfo.Spec.Build.DockerfileFrom == ONLINE {
			stageInfo.Spec.Build.DockerfileName = "Dockerfile"
		}

		//检查.spec.build.registryType
		if stageInfo.Spec.Build.RegistryType == 0 {
			stageInfo.Spec.Build.RegistryType = DEFAULT_REGISTRY_TYPE
		}
		if stageInfo.Spec.Build.RegistryType > REGISTRY_TYPE_MAX || stageInfo.Spec.Build.RegistryType < REGISTRY_TYPE_MIN {
			stage.ResponseErrorAndCode("Invalid .spec.build.registryType", http.StatusForbidden)
			return
		}

		//自定义registry时，检查customRegistry是否存在

		if CUSTOM_REGISTRY == stageInfo.Spec.Build.RegistryType {
			if stageInfo.Spec.Build.CustomRegistry == "" ||
				strings.TrimSpace(stageInfo.Spec.Build.CustomRegistry) == "" {
				stage.ResponseErrorAndCode(".spec.build.customRegistry is required with custom registry type", http.StatusForbidden)
				return
			}
			//TODO 暂时不考虑第三方的镜像

		}

		//检查.spec.build.imageTagType范围
		if stageInfo.Spec.Build.ImageTagType == 0 {
			stageInfo.Spec.Build.ImageTagType = DEFAULT_TAG_TYPE
		}
		if stageInfo.Spec.Build.ImageTagType < TAG_TYPE_MIN || stageInfo.Spec.Build.ImageTagType > TAG_TYPE_MAX {
			stage.ResponseErrorAndCode("Invalid .spec.build.imageTagType", http.StatusForbidden)
			return
		}

		//自定义tag类型时，检查.spec.build.customTag是否存在
		if stageInfo.Spec.Build.ImageTagType == CUSTOM_TAG &&
			(stageInfo.Spec.Build.CustomTag == "" || strings.TrimSpace(stageInfo.Spec.Build.CustomTag) == "") {
			stage.ResponseErrorAndCode(".spec.build.customTag is required with custom tag", http.StatusForbidden)
			return
		}
		//默认无缓存

	}
	// ============>_checkAndSetDefaults end
	glog.Infof("%s stageInfo:%#v \n", method, stageInfo)
	var stageRec models.CiStages
	// 用户修改name时要判断是否唯一
	if oldStageInfo.StageName != stageInfo.Metadata.Name {

		checkStateInfo, err := models.NewCiStage().FindOneByName(flowId, stageInfo.Metadata.Name)
		if err != nil {
			parseNumber, _ := sqlstatus.ParseErrorCode(err)
			if parseNumber != sqlstatus.SQLErrNoRowFound {
				glog.Errorf("%s get stage from stage failed:%v\n", method, err)
				stage.ResponseErrorAndCode("look up stage name failed", http.StatusInternalServerError)
				return
			}
		}
		if checkStateInfo.StageId != "" {
			stage.ResponseErrorAndCode("Stage already exists", http.StatusForbidden)
			return
		}
		//不存在则可以修改
		stageRec.StageName = stageInfo.Metadata.Name
	}

	// 当用户勾选 “当前流水线所有任务（包括新建任务），统一使用该代码库” 时，更新其他 state 以及 flow
	stage.updateForUniformRepo(stage.Namespace, flowId, stageInfo)

	//更新一些字段
	stageRec.StageName = stageInfo.Metadata.Name
	stageRec.ProjectId = stageInfo.Spec.Project.Id
	stageRec.DefaultBranch = stageInfo.Spec.Project.Branch
	stageRec.Type = stageInfo.Metadata.Type
	stageRec.CustomType = stageInfo.Metadata.CustomType
	stageRec.Image = stageInfo.Spec.Container.Image
	stageRec.CiEnabled = stageInfo.Spec.Ci.Enabled
	//更新Ci config字段
	CIConfig, err := json.Marshal(stageInfo.Spec.Ci.CiConfig)
	if err != nil {
		glog.Errorf("%s json Marshal CiConfig failed:%v\n", method, err)
		stage.ResponseErrorAndCode("json Marshal CiConfig failed", http.StatusForbidden)
		return
	}
	stageRec.CiConfig = string(CIConfig)

	//更新container_info字段
	container_info, err := json.Marshal(stageInfo.Spec.Container)
	if err != nil {
		glog.Errorf("%s stageInfo.Spec.Container:%v\n", method, err)
		stage.ResponseErrorAndCode("json Marshal stageInfo.Spec.Container failed", http.StatusForbidden)
		return
	}

	stageRec.ContainerInfo = string(container_info)
	//修改old stage 时 ，如果新的stage不存在scripts_id 则删除
	if !strings.Contains(string(container_info), "scripts_id") &&
		strings.Contains(oldStageInfo.ContainerInfo, "scripts_id") {
		//解析获取脚本id
		var container models.Container
		err = json.Unmarshal([]byte(oldStageInfo.ContainerInfo), &container)
		if err != nil {
			glog.Errorf("%s ContainerInfo json unmarshal failed: %v\n", method, err)
			stage.ResponseErrorAndCode("json Marshal stageInfo.Spec.Container failed", http.StatusForbidden)
			return
		}

		_, err = models.NewCiScripts().DeleteScriptByID(container.Scripts_id)
		if err != nil {
			glog.Errorf("%s delete script failed:%v\n", method, err)
		}

	}

	//更新build_info字段 if exist //数据库记录有build_info，请求中没有.spec.build时，清空build_info
	if stageInfo.Spec.Build != nil {
		buildInfo, err := json.Marshal(stageInfo.Spec.Build)
		if err != nil {
			glog.Errorf("%s %v\n", method, err)
			stage.ResponseErrorAndCode("json Marshal stageInfo.Spec.Build failed", http.StatusForbidden)
			return
		}
		buidInfoStr := string(buildInfo)
		stageRec.BuildInfo = buidInfoStr
		glog.Infof("%s stage rec: %v\n", method, buidInfoStr)
	}
	glog.Infof("===stageRec===%#v\n", stageRec)
	err = models.NewCiStage().UpdateById(stageId, stageRec)
	if err != nil {
		glog.Errorf("%s update failed: %v\n", method, err)
		stage.ResponseErrorAndCode("update stage failed", http.StatusForbidden)
		return
	}

	resp := struct {
		StageId string `json:"stageId"`
	}{
		StageId: stageId,
	}
	stage.ResponseResultAndStatusDevops(resp, http.StatusOK)

	return

}
