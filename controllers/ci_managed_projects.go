package controllers

import (
	"dev-flows-api-golang/models"

	"encoding/json"

	//"regexp"
	"dev-flows-api-golang/models/common"
	"errors"
	"github.com/golang/glog"
	"dev-flows-api-golang/util/uuid"
	"time"

	sqlstatus "dev-flows-api-golang/models/sql/status"

	"fmt"
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

	cimp.Audit.SetOperationType(models.AuditOperationCreate)
	cimp.Audit.SetResourceType(models.AuditResourceProjects)

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
		project.Username=body.Username
		project.Password=body.Password
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

		glog.Infof("%s Adding a new SVN repository:body\n", method,)
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
	cimp.Audit.SetResourceID(project_id)
	cimp.Audit.SetOperationType(models.AuditOperationDelete)
	cimp.Audit.SetResourceType(models.AuditResourceProjects)
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
	glog.Infof("project_id=%s\n",project_id)
	namespace := cimp.Namespace
	project := &models.CiManagedProjects{}
	err := project.FindProjectById(namespace, project_id)
	if err != nil {
		glog.Errorf("%s err:%v\n", method, err)
		cimp.ResponseErrorAndCode("No project found", http.StatusNotFound)
		return
	}

	// Remove private info from the result
	project.PrivateKey = "undefined"

	cimp.ResponseSuccessDevops(project, 1)

	return

}