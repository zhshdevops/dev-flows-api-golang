package controllers

import (
	"dev-flows-api-golang/models"
	"net/http"
	"strconv"
	"github.com/golang/glog"
	"dev-flows-api-golang/models/user"
	"dev-flows-api-golang/ci/coderepo"
	gogsClient "github.com/gogits/go-gogs-client"
	gitLabClientv3 "github.com/drone/drone/remote/gitlab3/client"
	"dev-flows-api-golang/modules/client"
	"dev-flows-api-golang/models/common"
	"regexp"
	"fmt"
	"encoding/json"
	"strings"
)

type CiWebhooksController struct {
	ErrorController
}

//@router /  [POST]
func (cimp *CiWebhooksController) InvokeBuildsByWebhook() {

	method := "CiManagedProjectsController.InvokeBuildsByWebhook"
	var event EventHook
	projectId := cimp.Ctx.Input.Param(":project_id")
	if projectId == "" {
		glog.Errorf("%s %s", method, "No projectId in the webhook request.")
		cimp.ResponseErrorAndCode("No projectId in the webhook request.", http.StatusBadRequest)
		return
	}

	body := cimp.Ctx.Input.RequestBody

	//glog.Infof("%s,%s\n", method, string(body))

	project := &models.CiManagedProjects{}
	err := project.FindProjectByIdNew(projectId)
	if err != nil || project.Owner == "" {
		glog.Errorf("%s this project not exist:%v\n", method, err)
		cimp.ResponseErrorAndCode("This project does not exist.", http.StatusNotFound)
		return
	}

	ciStages, total, err := models.NewCiStage().FindByProjectIdAndCI(projectId, 1)
	if err != nil || total < 1 {
		glog.Errorf("%s No stage of CI flow is using this project or CI is disabled. %v", method, err)
		cimp.ResponseErrorAndCode("find project by projectid and ci failed or No stage of CI flow is using this project or CI is disabled.", http.StatusOK)
		return
	}

	glog.Infof("ciStages============>>:%d\n", len(ciStages))

	userModel := &user.UserModel{}
	// use cache for better performance
	_, err = userModel.GetByName(project.Owner)
	if err != nil {
		glog.Errorf(" Gte User '"+project.Owner+" failed:%v\n", err)
	}

	// Use the user/space info of this project
	//var userInfo = {
	//user: project.owner,
	//	name: project.owner,
	//	// Use owner namespace to run the build
	//		namespace: project.namespace,
	//	// Used for query
	//		userNamespace: project.owner
	//}

	if project.RepoType == GOGS || project.RepoType == GITHUB ||
		project.RepoType == SVN || project.RepoType == GITLAB {

		if project.RepoType == GITLAB {
			event, err = getGitlabEventInfo(cimp.Ctx.Request, body, *project)
			if err != nil {
				glog.Errorf("%s gitlab webhook run failed:%v\n", method, err)
				cimp.ResponseErrorAndCode("find project by projectid and ci failed or No stage of CI flow is using this project or CI is disabled.", 501)
				return
			}
		} else if project.RepoType == GITHUB || project.RepoType == GOGS {
			event, err = GetEventInfo(cimp.Ctx.Request, body, *project)
			if err != nil {
				glog.Errorf("%s %v\n", method, err)
				cimp.ResponseErrorAndCode(fmt.Sprintf("%s", err), 501)
				return
			}
		} else if project.RepoType == SVN {
			event, err = GetSvnEventInfo(body, *project)
			if err != nil {
				glog.Errorf("%s %v\n", method, err)
				cimp.ResponseErrorAndCode(fmt.Sprintf("%s", err), 501)
				return
			}
		}

		GitlabProjectId, _ := strconv.Atoi(project.GitlabProjectId)
		if project.RepoType != SVN && event.ScmProjectId != GitlabProjectId {
			glog.Errorf("%s Project id does not match with exiting one\n", method)
			cimp.ResponseErrorAndCode("Project id does not match with exiting one", 404)
			return
		}

		glog.V(1).Infof("Validate CI rule of each stage ...")

		//create builds by ci rules
		err = InvokeCIFlowOfStages(userModel, event, ciStages, project)
		if err != nil {
			cimp.ResponseErrorAndCode("build failed "+fmt.Sprintf("%s", err), 501)
			return
		}
	} else {

		glog.Errorf("%s %s\n", method, "Only gitlab/github/gogs/svn is supported by now")
		cimp.ResponseErrorAndCode("Only gitlab/github/gogs/svn is supported by now", http.StatusBadRequest)
		return
	}

	cimp.ResponseErrorAndCode("Webhook handled normally", http.StatusOK)
	return

}

func InvokeCIFlowOfStages(user *user.UserModel, event EventHook, stageList []models.CiStages, project *models.CiManagedProjects) error {
	method := "CiManagedProjectsController.invokeCIFlowOfStages"
	glog.V(1).Infof("%s Number of stages in the list %d", method, len(stageList))

	for _, stage := range stageList {
		// Check if the CI config matched
		// Convert to object if it's string
		ciConfig := models.CiConfig{}
		eventType := ""
		if stage.CiConfig != "" {
			err := json.Unmarshal([]byte(stage.CiConfig), &ciConfig)
			if err != nil {
				glog.Errorf("%s json marshal failed==>%v\n", method, err)
				continue
			}
		}
		if event.Type != "" {
			eventType = strings.ToLower(event.Type)
		}
		// Check if the rule matched
		var matched bool = false
		if project.RepoType == SVN {
			matched = true
		} else if stage.CiConfig != "" {
			glog.V(1).Infof("%s Event type: :%s\n", method, eventType)
			//merge request
			if ciConfig.MergeRequest && eventType == HOOK_EVENT_MERGE_REQUEST {
				matched = true
				//branch tag
			} else if strings.Contains(stage.CiConfig, eventType) {
				glog.V(1).Infof("%s : [%v] vs [%s]\n", method, ciConfig, event.Name)
				if eventType == "branch" {
					if _, ok := ciConfig.Branch.MatchWay.(bool); ok {
						//the branch same
						if ciConfig.Branch.Name == event.Name {
							matched = true
						}
					} else {
						//检查是否是合法的regexp
						matchWayReg, err := regexp.Compile(ciConfig.Branch.Name)
						if err != nil {
							glog.Errorf("%s 解析正则表达式失败，请检查格式 branch regexp complie failed: %v\n", method, err)
							detail := &EmailDetail{
								Type:    "ci",
								Result:  "failed",
								Subject: fmt.Sprintf(`'%s'构建失败`, stage.StageName),
								Body:    fmt.Sprintf(`解析正则表达式失败，请检查格式: %s`, ciConfig.Branch.Name),
							}
							detail.SendEmailUsingFlowConfig(user.Namespace, stage.FlowId)
							continue
						}
						if matchWayReg.MatchString(event.Name) {
							matched = true

						} else {
							detail := &EmailDetail{
								Type:    "ci",
								Result:  "failed",
								Subject: fmt.Sprintf(`'%s'构建失败`, stage.StageName),
								Body:    fmt.Sprintf(`解析正则表达式失败，请检查格式: %s`, ciConfig.Branch.Name),
							}
							detail.SendEmailUsingFlowConfig(user.Namespace, stage.FlowId)
							continue
						}
					}
				} else if eventType == "tag" {
					if _, ok := ciConfig.Tag.MatchWay.(bool); ok {
						//the branch same
						if ciConfig.Tag.Name == event.Name {
							matched = true
						}
					} else {
						//检查是否是合法的regexp
						matchWayReg, err := regexp.Compile(ciConfig.Tag.Name)
						if err != nil {
							glog.Errorf("%s tag regexp complie failed: %v\n", method, err)
							detail := &EmailDetail{
								Type:    "ci",
								Result:  "failed",
								Subject: fmt.Sprintf(`'%s'构建失败`, stage.StageName),
								Body:    fmt.Sprintf(`解析正则表达式失败，请检查格式: %s`, ciConfig.Tag.Name),
							}
							detail.SendEmailUsingFlowConfig(user.Namespace, stage.FlowId)
							continue
						}
						if matchWayReg.MatchString(event.Name) {
							matched = true

						} else {
							detail := &EmailDetail{
								Type:    "ci",
								Result:  "failed",
								Subject: fmt.Sprintf(`'%s'构建失败`, stage.StageName),
								Body:    fmt.Sprintf(`解析正则表达式失败，请检查格式: %s`, ciConfig.Tag.Name),
							}
							detail.SendEmailUsingFlowConfig(user.Namespace, stage.FlowId)
							continue
						}
					}
				}
			}

		} else {
			glog.Errorf("%s no ci rule \n", method)
			//TODO send email
			matched = false
			detail := &EmailDetail{
				Type:    "ci",
				Result:  "failed",
				Subject: fmt.Sprintf(`'%s'构建失败`, stage.StageName),
				Body:    fmt.Sprintf(`没有配置CI规则`),
			}
			detail.SendEmailUsingFlowConfig(user.Namespace, stage.FlowId)
			continue
		}

		if matched {
			glog.V(1).Infof("%s ---- Add to build queue ----: :%s\n", method, eventType)
			// 开始构建任务
			var ennFlow EnnFlow
			ennFlow.FlowId = stage.FlowId
			ennFlow.StageId = stage.StageId
			ennFlow.CodeBranch = event.Name
			ennFlow.LoginUserName = user.Username
			ennFlow.Namespace = project.Namespace  //用来查询flow
			ennFlow.UserNamespace = user.Namespace //用来构建
			var conn Conn
			imageBuild := models.NewImageBuilder(client.ClusterID)
			stagequeue := NewStageQueueNew(ennFlow, event.Name, ennFlow.Namespace, ennFlow.LoginUserName, stage.FlowId, imageBuild, conn)

			if stagequeue != nil {
				//判断是否该EnnFlow当前有执行中
				err := stagequeue.CheckIfBuiding(stage.FlowId)
				if err != nil {
					glog.Warningf("%s Too many waiting builds of:  %v\n", method, err)
					if strings.Contains(fmt.Sprintf("%s", err), "该EnnFlow已有任务在执行,请等待执行完再试") {
						ennFlow.Message = "该EnnFlow [" + stagequeue.CiFlow.Name + "] 已有任务在执行,请等待执行完再试"
						ennFlow.Status = http.StatusOK
						ennFlow.BuildStatus = common.STATUS_SUCCESS
						ennFlow.FlowBuildId = stagequeue.FlowbuildLog.BuildId
						ennFlow.StageBuildId = stagequeue.StageBuildLog.BuildId
						ennFlow.Flag = 1
						Send(ennFlow, SOCKETS_OF_FLOW_MAPPING_NEW[stage.FlowId])
						continue
					} else {
						ennFlow.Message = "找不到对应的EnnFlow"
						ennFlow.Status = http.StatusOK
						ennFlow.BuildStatus = common.STATUS_SUCCESS
						ennFlow.FlowBuildId = stagequeue.FlowbuildLog.BuildId
						ennFlow.StageBuildId = stagequeue.StageBuildLog.BuildId
						ennFlow.Flag = 1
						Send(ennFlow, SOCKETS_OF_FLOW_MAPPING_NEW[stage.FlowId])
						continue
					}
				}
				//开始执行 把执行日志插入到数据库
				stagequeue.InsertLog()
				go stagequeue.Run()
			}
		}

	}
	return nil

}

func getGitlabEventInfo(req *http.Request, body []byte, project models.CiManagedProjects) (EventHook, error) {
	method := "CiManagedProjectsController.getGitlabEventInfo"
	var hookPayload gitLabClientv3.HookPayload
	var event EventHook

	headerEvnt := req.Header.Get("x-gitlab-event")
	err := json.Unmarshal(body, &hookPayload)
	if err != nil {
		glog.Errorf("%s json unmarshal failed:%v\n", method, err)
		return event, err
	}
	glog.Infof("%s gitlab hookPayload info :%v\n", method, hookPayload)

	glog.Infof("%s  gitlab event type in the header: %s\n", method, headerEvnt)

	if hookPayload.ObjectKind != HOOK_EVENT_PUSH && hookPayload.ObjectKind != HOOK_EVENT_MERGE_REQUEST && hookPayload.ObjectKind != HOOK_EVENT_TAG_PUSH {

		glog.V(1).Infof("%s  Skip non-push or merge-request event from : %s\n", method, project.RepoType)

		return event, fmt.Errorf("%s", "Skip non-push or merge-request event from "+project.RepoType)
	}

	projectId := hookPayload.ProjectId

	var pushName, commitId, eventType string

	// Get the project id of merge_request
	if hookPayload.ObjectKind == HOOK_EVENT_MERGE_REQUEST {

		if hookPayload.ObjectAttributes.Action != "merge" {
			glog.V(1).Infof("%s  Skip non-merge merge-request event from gitlab : %s\n", method, project.RepoType)

			return event, fmt.Errorf("Skip non-merge merge-request event from gitlab")
		}

		projectId = hookPayload.ObjectAttributes.SourceProjectId
		eventType = HOOK_EVENT_MERGE_REQUEST
		pushName = hookPayload.ObjectAttributes.TargetBranch

	} else {
		// Push Hook
		// Tag Push Hook

		pushName = strings.SplitN(hookPayload.Ref, "/", 3)[2]

		glog.Infof("%s push payload pushName info :%v\n", method, pushName)

		eventType = strings.SplitN(hookPayload.Ref, "/", 3)[1]

		eventType = formateCIPushType(eventType)
		if len(hookPayload.Commits) != 0 {
			commitId = hookPayload.Commits[len(hookPayload.Commits)-1].Id
		}

		projectId = hookPayload.ProjectId
	}

	if eventType == "Tag" {
		if hookPayload.CheckoutSha == "" {
			glog.V(1).Infof("%s Skip delete tag event from gitlab : %s\n", method, project.RepoType)

			return event, fmt.Errorf("Skip delete tag event from gitlab")
		}
	}

	event.ScmProjectId = projectId
	event.Type = eventType
	event.Name = pushName
	event.CommitId = commitId
	return event, nil

}

func GetEventInfo(req *http.Request, body []byte, project models.CiManagedProjects) (EventHook, error) {
	var event EventHook
	method := "CiManagedProjectsController.getEventInfo"
	headerEvnt := req.Header.Get("x-github-event")
	if headerEvnt == "" {
		headerEvnt = req.Header.Get("x-gogs-event")
	}

	glog.V(1).Infof("%s  event type in the header: %s\n", method, headerEvnt)
	// Gogs release will be 'release'/UI and 'create'/command
	if headerEvnt != HOOK_EVENT_PUSH && headerEvnt != HOOK_EVENT_PULL_REQUEST && headerEvnt != HOOK_EVENT_RELEASE && headerEvnt != HOOK_EVENT_CREATE {

		glog.V(1).Infof("%s  Skip non-push or merge-request event from : %s\n", method, project.RepoType)

		return event, fmt.Errorf("%s", "Skip non-push or merge-request event from "+project.RepoType)
	}

	var projectId int
	var pushType, pushName, commitId string

	if headerEvnt == HOOK_EVENT_PULL_REQUEST {
		var requestPayload gogsClient.PullRequestPayload

		err := json.Unmarshal(body, &requestPayload)
		if err != nil {
			glog.Errorf("%s json unmarshal failed:%v\n", method, err)
			return event, err
		}
		glog.V(1).Infof("%s pull request payload info :%v\n", method, requestPayload)

		if requestPayload.PullRequest == nil || requestPayload.PullRequest.Merged == nil {
			glog.V(1).Infof("%s skip non-merged pull-request event from :%v\n", method, project.RepoType)
			return event, fmt.Errorf("%s skip non-merged pull-request event from %s", method, project.RepoType)
		}

		headerEvnt = "merge_request"

		if project.RepoType == GOGS {
			pushName = requestPayload.PullRequest.BaseBranch
		} else {
			//github
			//pushName = eventBody.pull_request.head.ref
		}
		projectId = int(requestPayload.Repository.ID)

	} else if headerEvnt == HOOK_EVENT_PUSH {
		var pushPayload gogsClient.PushPayload

		err := json.Unmarshal(body, &pushPayload)
		if err != nil {
			glog.Errorf("%s push payload json unmarshal failed:%v\n", method, err)
			return event, err
		}
		glog.V(1).Infof("%s push payload info :%v\n", method, pushPayload.Ref)

		pushName = strings.SplitN(pushPayload.Ref, "/", 3)[2]

		glog.V(1).Infof("%s push payload pushName info :%v\n", method, pushName)

		pushType = strings.SplitN(pushPayload.Ref, "/", 3)[1]
		headerEvnt = formateCIPushType(pushType)
		commitId = pushPayload.Commits[len(pushPayload.Commits)-1].ID
		projectId = int(pushPayload.Repo.ID)
	}

	if project.RepoType == GOGS && (headerEvnt == HOOK_EVENT_CREATE || headerEvnt == HOOK_EVENT_RELEASE) {
		var createPayload gogsClient.CreatePayload
		err := json.Unmarshal(body, &createPayload)
		if err != nil {
			glog.Errorf("%s create or release payload json unmarshal failed:%v\n", method, err)
			return event, err
		}
		glog.V(1).Infof("%s create or release payload info :%v\n", method, createPayload.Ref)

		projectId = int(createPayload.Repo.ID)

		if headerEvnt == HOOK_EVENT_RELEASE {
			pushName = createPayload.Ref
		} else {
			if createPayload.RefType == "tag" {
				pushName = createPayload.Ref
			}
		}
		headerEvnt = formateCIPushType(headerEvnt)
		glog.V(1).Infof("%s create or release payload pushName info :%v\n", method, pushName)
	}

	event.ScmProjectId = projectId
	event.Type = headerEvnt
	event.Name = pushName
	event.CommitId = commitId
	return event, nil
}

func formateCIPushType(pushType string) string {
	switch pushType {
	case "heads":
		return "Branch"
	case "tags":
		return "Tag"
	case "release": // Release will be used as tag
	case "create":
		return "Tag"

	}
	return pushType

}

type EventHook struct {
	Name         string `json:"name,omitempty"`
	ScmProjectId int `json:"scmProjectId"`
	Type         string `json:"type"`
	CommitId     string `json:"commitId"`
}

func GetSvnEventInfo(body []byte, project models.CiManagedProjects) (EventHook, error) {
	var event EventHook
	method := "CiManagedProjectsController.GetSvnEventInfo"
	svnHook := coderepo.SvnHook{}
	err := json.Unmarshal(body, &svnHook)
	if err != nil {
		glog.Errorf("%s %v\n", method, err)
		return event, err
	}
	if svnHook.Name == "" {
		glog.Warningf("%s %v\n", method, "Skip, name is required.")
		return event, fmt.Errorf("%s", "Skip, name is required.")
	}
	event.Name = svnHook.Name

	return event, nil

}
