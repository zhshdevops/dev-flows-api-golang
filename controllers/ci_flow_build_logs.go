package controllers

import (
	"fmt"
	"dev-flows-api-golang/models"
	"net/http"
	"strconv"
	"github.com/golang/glog"
	"dev-flows-api-golang/models/user"
)

type CiFlowBuildLogsController struct {
	ErrorController
}

//@router /  [POST]
func (cimp *CiFlowBuildLogsController) InvokeBuildsByWebhook() {

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
			event, err = getGitlabEventInfo(cimp.Ctx.Request,body, *project)
			if err != nil {
				glog.Errorf("%s gitlab webhook run failed:%v\n", method, err)
				cimp.ResponseErrorAndCode("find project by projectid and ci failed or No stage of CI flow is using this project or CI is disabled.", 501)
				return
			}
		} else if project.RepoType == GITHUB || project.RepoType == GOGS {
			event, err = GetEventInfo(cimp.Ctx.Request,body, *project)
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
		err = invokeCIFlowOfStages(userModel,body, event, ciStages, project)
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
