package controllers

import (
	"dev-flows-api-golang/models"
	"dev-flows-api-golang/models/user"
	"encoding/json"

	//"regexp"
	"dev-flows-api-golang/models/common"
	"errors"
	"github.com/golang/glog"
	"dev-flows-api-golang/util/uuid"
	"time"
	gogsClient "github.com/gogits/go-gogs-client"
	gitLabClientv3 "github.com/drone/drone/remote/gitlab3/client"
	sqlstatus "dev-flows-api-golang/models/sql/status"
	"dev-flows-api-golang/ci/coderepo"
	"fmt"
	"strings"
	"strconv"
	"regexp"
	"net/http"
)

const (
	HOOK_EVENT_CREATE        string = "create"
	HOOK_EVENT_PUSH          string = "push"
	HOOK_EVENT_ISSUES        string = "issues"
	HOOK_EVENT_ISSUE_COMMENT string = "issue_comment"
	HOOK_EVENT_PULL_REQUEST  string = "pull_request"
	HOOK_EVENT_RELEASE       string = "release"
	HOOK_EVENT_TAG_PUSH      string = "tag_push"
	HOOK_EVENT_MERGE_REQUEST string = "merge_request"
)

type CiManagedProjectsController struct {
	BaseController
}

// @Title GetManagedProjects
// @Description GetManagedProjects
//@router / [get]
func (cimp *CiManagedProjectsController) GetManagedProjects() {

	method := "Controller.CiManagedProjectsController.GetManagedProjects"

	var i int64

	ciManageProject := &models.CiManagedProjects{}

	listProject, total, err := ciManageProject.ListProjects(cimp.Namespace)
	if err != nil {
		glog.Errorf("%s get code managed project failed: %v\n", method, err)
		cimp.ErrorInternalServerError(errors.New(" select databases error "))
		return
	}

	if total == 0 {
		glog.Warningf("%s =====>> no project active yes <<=======\n", method)
		cimp.ResponseSuccessStatusAndMessageDevops("No project added yet")
		return
	}
	//Remove private info from the result
	for i = 0; i < total; i++ {

		listProject[i].PrivateKey = "undefined"

		if listProject[i].PublicKey != "" {

			listProject[i].PublicKey = string(listProject[i].PublicKey)

		}
		// Add webhook url for svn
		if listProject[i].RepoType == "svn" {

			listProject[i].WebhookUrl = common.WebHookUrlPrefix + listProject[i].Id

		}
	}

	cimp.ResponseSuccessDevops(listProject, total)
}

//"http://10.39.0.53:9999/svnrepos/xinzhiyuntest"
//is_private
//:
//1
//name
//:
//"dddd"
//password
//:
//"qinzhao"
//repo_type
//:
//"svn"
//username
//:
//"root"
//@router / [POST] 激活
func (cimp *CiManagedProjectsController) CreateManagedProject() {

	method := "Controller/CiManagedProjectsController/CreateManagedProject"

	var project models.CiManagedProjects
	var body models.ActiveRepoReq
	err := json.Unmarshal(cimp.Ctx.Input.RequestBody, &body)
	if err != nil {
		glog.Errorf("%s json Unmarshal failed %v\n", method, err)
		cimp.ErrorInternalServerError(errors.New("RequestBody Json Unmarshal failed"))
		return
	}
	project.Address = body.Address
	project.IsPrivate = int8(body.IsPrivate)
	if body.RepoType != GITLAB {
		project.GitlabProjectId = fmt.Sprintf("%d", body.ProjectId)
	} else if body.RepoType == GITLAB {
		project.GitlabProjectId = fmt.Sprintf("%d", body.GitlabProjectId)
	}

	project.Name = body.Name
	project.RepoType = body.RepoType
	//for gitlab
	if project.IsPrivate == 0 {
		project.IsPrivate = int8(body.Private)
	}
	//校验project的地址
	if regexp.MustCompile(`^(http:|https:|git@|ssh:|svn:)`).
		FindString(project.Address) == "" {
		cimp.ResponseErrorAndCode(`address must begin with "http:", "https:", "git@" "ssh:" or "svn:".`, http.StatusBadRequest)
		return
	}

	project.Id = uuid.NewManagedProjectID()
	project.Owner = cimp.User.Username
	project.Namespace = cimp.Namespace
	project.CreateTime = time.Now()
	project.SourceFullName = body.SourceFullName
	results := &models.CiManagedProjects{}

	results.FindProjectByNameType(cimp.Namespace, project.Name, project.RepoType)
	if err != nil {
		parResultNUmber, _ := sqlstatus.ParseErrorCode(err)
		if sqlstatus.SQLErrNoRowFound != parResultNUmber {
			glog.Errorf("%s find project failed from database: %v\n", method, err)
			cimp.ResponseErrorAndCode("find project failed from database", http.StatusInternalServerError)
			return
		}
	}

	if results.Name != "" {
		cimp.ResponseErrorAndCode("Project (name - "+project.Name+" , type - "+project.RepoType+" ) already exists", http.StatusConflict)
		return
	}

	// Check if the project url alreay exists for svn repo
	if project.RepoType == "svn" {

		err = results.FindProjectByAddressType(cimp.Namespace, project.Address, project.RepoType)
		if err != nil {
			parResultNUmber, _ := sqlstatus.ParseErrorCode(err)
			if sqlstatus.SQLErrNoRowFound != parResultNUmber {
				glog.Errorf("%s find project failed from database: %v\n", method, err)
				cimp.ResponseErrorAndCode("find project failed from database", http.StatusInternalServerError)
				return
			}
		}
		if results.Name != "" {
			cimp.ResponseErrorAndCode("Project (name - "+project.Name+" , type - "+project.RepoType+" ) already exists", 409)
			return
		}
	}
	var scmResult interface{}
	if project.RepoType == "gitlab" || project.RepoType == "github" || project.RepoType == "gogs" {
		//Handle gitlab verified
		verified := true
		err := models.CreateIntegrationParts(cimp.Namespace, verified, &project)
		if err != nil {
			glog.Errorf("%s failed:%v\n", method, err)
			cimp.ResponseErrorAndCode("CreateIntegrationParts failed"+fmt.Sprintf("%s", err), http.StatusInternalServerError)
			return
		}
		scmResult = err
	} else if project.RepoType == "svn" {

		glog.Infof("%s Adding a new SVN repository\n", method)
		// Update user/password if found for each add action
		project.SourceFullName = project.Address
		project.IsPrivate = 0 //公有
		//需要密码的SVN代码库才加入到表repo中s
		if project.Username != "" || project.Password != "" {
			project.IsPrivate = 1
			depot := models.NewCiRepos()
			err = depot.FindOneRepoToken(cimp.Namespace, models.DepotToRepoType(project.RepoType))
			if err != nil {
				res, _ := sqlstatus.ParseErrorCode(err)
				if res == sqlstatus.SQLErrNoRowFound {

				} else {
					glog.Errorf(" %s get repos tocken info from database failed: %v\n", method, err)
					cimp.ResponseErrorAndCode("get svn tocken info from database failed ", http.StatusInternalServerError)
					return
				}
			}
			//depo exist but user_info is empty will delete data from database
			if depot != nil {
				depot.DeleteOneRepo(cimp.Namespace, models.DepotToRepoType(project.RepoType))
			}

			repoInfo := models.CiRepos{}
			repoInfo.UserId = int(cimp.User.UserID)
			repoInfo.RepoType = models.DepotToRepoType(project.RepoType)
			repoInfo.CreateTime = time.Now()
			repoInfo.IsEncrypt = 1
			repoInfo.Namespace = cimp.Namespace
			repoInfo.AccessUserName = body.Username
			repoInfo.AccessToken = body.Password
			repoInfo.GitlabUrl = project.Address

			_, err := depot.CreateOneRepo(repoInfo)
			if err != nil {
				glog.Errorf("CreateOneRepo failed:%s\n", err)
				cimp.ResponseErrorAndCode("insert svn info into database failed ", http.StatusInternalServerError)
				return
			}

		}

	} else {
		cimp.ResponseErrorAndCode("Only support gitlab/svn/github/gogs for now", http.StatusBadRequest)
		return
	}

	err = results.CreateOneProject(project)
	if err != nil {
		glog.Errorf("%s insert project info into  database failed: %v", method, err)
		cimp.ResponseErrorAndCode(" insert project info into  database failed ", http.StatusBadRequest)
		return
	}

	cimp.ResponseManageProjectDevops("Project added successfully", project.Id, scmResult, http.StatusOK)

	return

}

//@router /:project_id  [DELETE]
func (cimp *CiManagedProjectsController) RemoveManagedProject() {
	method := "CiManagedProjectsController/RemoveProject"
	project_id := cimp.Ctx.Input.Param(":project_id")
	namespace := cimp.Namespace
	project := &models.CiManagedProjects{}
	//如果找不到projejct TODO 检查sql语句的no rows result set
	err := project.FindProjectById(namespace, project_id)
	if err != nil {
		parseResult, _ := sqlstatus.ParseErrorCode(err)
		if parseResult == sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s %v", method, err)
			cimp.ResponseErrorAndCode("Project removed successfully", http.StatusOK)
			return
		} else {
			glog.Errorf("%s %v", method, err)
			cimp.ResponseErrorAndCode("Project removed Failed", http.StatusConflict)
			return
		}
	}

	// Check if any stage is referring this project
	stage := &models.CiStages{}
	err = stage.FindByProjectId(project_id)
	if err != nil {
		parseResult, _ := sqlstatus.ParseErrorCode(err)
		if parseResult == sqlstatus.SQLErrNoRowFound {
			//not found will go on
		} else {
			glog.Errorf("%s find project by projectId from database: stage=%v  err:=%v", method, stage, err)
			cimp.ResponseErrorAndCode("please try again!", http.StatusInternalServerError)
			return
		}
	}

	if stage.StageName != "" {
		cimp.ResponseErrorAndCode("Stage '"+stage.StageName+"' is using this project, remove the reference from the stage and try again", http.StatusBadRequest)
		return
	}

	if project.RepoType == "gitlab" || project.RepoType == "github" || project.RepoType == "gogs" {
		// Clear deploy keys, webhook, etc... from integrated SCM tools
		err = project.ClearIntegrationParts(namespace)
		if err != nil {
			glog.Errorf("%s clear Integration failed: %v\n", method, err)
			cimp.ResponseErrorAndCode("clear  Stage "+stage.StageName+" Integration failed ", http.StatusInternalServerError)
			return
		}

	} else if project.RepoType == "svn" {
		glog.Info(method, " Removing SVN project \n"+project.Name)
	}
	//remove project from the database
	deleteCount, err := project.RemoveProject(namespace, project_id)
	if err != nil {
		if deleteCount == sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s remove code project failed from database: %d %v\n", method, deleteCount, err)
			cimp.ResponseErrorAndCode("No project found mathcing the project id", http.StatusNotFound)
			return
		} else {
			glog.Errorf("%s remove code project failed from database: %d %v\n", method, deleteCount, err)
			cimp.ResponseErrorAndCode("No project found mathcing the project id", http.StatusInternalServerError)
			return
		}
	}

	glog.Info(method, " Delete project "+project.Name+" Success")
	cimp.ResponseErrorAndCode("Project removed successfully", http.StatusOK)
	return

}

//@router /:project_id  [GET]
func (cimp *CiManagedProjectsController) GetManagedProjectDetail() {
	method := "GetManagedProjectDetail"
	project_id := cimp.Ctx.Input.Param(":project_id")
	glog.Info(project_id)
	namespace := cimp.Namespace
	project := &models.CiManagedProjects{}
	err := project.FindProjectById(namespace, project_id)
	if err != nil {
		glog.Errorf("%s %v", method, err)
		cimp.ResponseErrorAndCode("No project found", 404)
		return
	}

	// Remove private info from the result
	project.PrivateKey = "undefined"

	cimp.ResponseSuccessDevops(project, 1)

	return

}

//@router /webhooks/:project_id  [POST]
func (cimp *CiManagedProjectsController) InvokeBuildsByWebhook() {

	method := "CiManagedProjectsController.InvokeBuildsByWebhook"
	var event EventHook
	projectId := cimp.Ctx.Input.Param(":project_id")
	if projectId == "" {
		glog.Errorf("%s %s", method, "No projectId in the webhook request.")
		cimp.ResponseErrorAndCode("No projectId in the webhook request.", http.StatusBadRequest)
		return
	}

	body := cimp.Ctx.Input.RequestBody

	glog.Infof("%s,%s\n", method, string(body))

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
			event, err = cimp.getGitlabEventInfo(body, *project)
			if err != nil {
				glog.Errorf("%s gitlab webhook run failed:%v\n", method, err)
				cimp.ResponseErrorAndCode("find project by projectid and ci failed or No stage of CI flow is using this project or CI is disabled.", 501)
				return
			}
		} else if project.RepoType == GITHUB || project.RepoType == GOGS {
			event, err = cimp.GetEventInfo(body, *project)
			if err != nil {
				glog.Errorf("%s %v\n", method, err)
				cimp.ResponseErrorAndCode(fmt.Sprintf("%s", err), 501)
				return
			}
		} else if project.RepoType == SVN {
			event, err = cimp.GetSvnEventInfo(body, *project)
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
		err = cimp.invokeCIFlowOfStages(userModel,body, event, ciStages, project)
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

func (cimp *CiManagedProjectsController) invokeCIFlowOfStages(user *user.UserModel,body []byte, event EventHook, stageList []models.CiStages, project *models.CiManagedProjects) error {
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
				return err
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
					//if ciConfig.Branch.MatchWay != "RegExp" {
					if RegExp, ok := ciConfig.Branch.MatchWay.(string); ok && RegExp != "RegExp" {
						//the branch same
						if ciConfig.Branch.Name == event.Name {
							matched = true
						}
					} else {
						//检查是否是合法的regexp
						matchWayReg, err := regexp.Compile(ciConfig.Branch.Name)
						if err != nil {
							glog.Errorf("%s branch regexp complie failed: %v\n", method, err)
							//TODO send email
							return err
						}
						if matchWayReg.MatchString(event.Name) {
							matched = true
						} else {
							//TODO send email
							return fmt.Errorf("解析正则表达式失败，请检查格式 %s", ciConfig.Tag.Name)
						}
					}
				} else if eventType == "tag" {
					if RegExp, ok := ciConfig.Tag.MatchWay.(string); ok && RegExp != "RegExp" {
						//the branch same
						if ciConfig.Branch.Name == event.Name {
							matched = true
						}
					} else {
						//检查是否是合法的regexp
						matchWayReg, err := regexp.Compile(ciConfig.Tag.Name)
						if err != nil {
							glog.Errorf("%s tag regexp complie failed: %v\n", method, err)
							//TODO send email
							return err
						}
						if matchWayReg.MatchString(event.Name) {
							matched = true
						} else {
							//TODO send email
							return fmt.Errorf("解析正则表达式失败，请检查格式 %s", ciConfig.Tag.Name)
						}
					}
				}
			}

		} else {
			glog.Errorf("%s no ci rule \n", method)
			//TODO send email
			matched = false
			return fmt.Errorf("no ci rule")
		}

		if matched {
			glog.V(1).Infof("%s ---- Add to build queue ----: :%s\n", method, eventType)
			//TODO 开始构建任务
			//go StartFlowBuild(cimp.User, stage.FlowId, stage.StageId, event.Name, &models.Option{})
			buildBody := models.BuildReqbody{
				StageId: stage.StageId,
				Options: &models.Option{Branch: event.Name},
			}
			imageBuild := models.NewImageBuilder()
			stagequeue,result,httpStatusCode:=NewStageQueue(user, buildBody, event.Name, cimp.Namespace, stage.FlowId, imageBuild)
			if httpStatusCode == http.StatusOK {
				go func(){
					result,httpStatusCode=stagequeue.Run()
					glog.Infof("invokeCIFlowOfStages %s %d",result,httpStatusCode)
				}()
			}

		}

	}

	//TODO 通知repo 构建情况 打算用协程做处理

	return nil

}

func (cimp *CiManagedProjectsController) getGitlabEventInfo(body []byte, project models.CiManagedProjects) (EventHook, error) {
	method := "CiManagedProjectsController.getGitlabEventInfo"
	var hookPayload gitLabClientv3.HookPayload
	var event EventHook
	headerEvnt := cimp.Ctx.Input.Header("x-gitlab-event")
	err := json.Unmarshal(body, &hookPayload)
	if err != nil {
		glog.Errorf("%s json unmarshal failed:%v\n", method, err)
		return event, err
	}
	glog.V(1).Infof("%s gitlab hookPayload info :%v\n", method, hookPayload)

	glog.V(1).Infof("%s  gitlab event type in the header: %s\n", method, headerEvnt)

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

		glog.V(1).Infof("%s push payload pushName info :%v\n", method, pushName)

		eventType = strings.SplitN(hookPayload.Ref, "/", 3)[1]

		eventType = cimp.formateCIPushType(eventType)
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

func (cimp *CiManagedProjectsController) GetEventInfo(body []byte, project models.CiManagedProjects) (EventHook, error) {
	var event EventHook
	method := "CiManagedProjectsController.getEventInfo"
	headerEvnt := cimp.Ctx.Input.Header("x-github-event")
	if headerEvnt == "" {
		headerEvnt = cimp.Ctx.Input.Header("x-gogs-event")
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
		headerEvnt = cimp.formateCIPushType(pushType)
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
		headerEvnt = cimp.formateCIPushType(headerEvnt)
		glog.V(1).Infof("%s create or release payload pushName info :%v\n", method, pushName)
	}

	event.ScmProjectId = projectId
	event.Type = headerEvnt
	event.Name = pushName
	event.CommitId = commitId
	return event, nil
}

func (cimp *CiManagedProjectsController) formateCIPushType(pushType string) string {
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
func (cimp *CiManagedProjectsController) checkSignature() error {

	return errors.New("get error failed")

}

type EventHook struct {
	Name         string `json:"name,omitempty"`
	ScmProjectId int `json:"scmProjectId"`
	Type         string `json:"type"`
	CommitId     string `json:"commitId"`
}

func (cimp *CiManagedProjectsController) GetSvnEventInfo(body []byte, project models.CiManagedProjects) (EventHook, error) {
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
