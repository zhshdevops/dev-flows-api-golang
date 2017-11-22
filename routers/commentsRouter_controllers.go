package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:AuthController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:AuthController"],
		beego.ControllerComments{
			Method: "AuthByToken",
			Router: `/`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CDRulesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CDRulesController"],
		beego.ControllerComments{
			Method: "GetDeploymentCDRule",
			Router: `/`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CDRulesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CDRulesController"],
		beego.ControllerComments{
			Method: "DeleteDeploymentCDRule",
			Router: `/`,
			AllowHTTPMethods: []string{"DELETE"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiDockerfileController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiDockerfileController"],
		beego.ControllerComments{
			Method: "ListDockerfiles",
			Router: `/`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "GetCIFlows",
			Router: `/`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "CreateCIFlow",
			Router: `/`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "GetCIRules",
			Router: `/:flow_id/ci-rules`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "UpdateCIRules",
			Router: `/:flow_id/ci-rules`,
			AllowHTTPMethods: []string{"PUT"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "GetCIFlowById",
			Router: `/:flow_id`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "RemoveCIFlow",
			Router: `/:flow_id`,
			AllowHTTPMethods: []string{"DELETE"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "UpdateCIFlow",
			Router: `/:flow_id`,
			AllowHTTPMethods: []string{"PUT"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "GetImagesOfFlow",
			Router: `/:flow_id/images`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "ListDeploymentLogsOfFlow",
			Router: `/:flow_id/deployment-logs`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "ListCDRules",
			Router: `/:flow_id/cd-rules`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "CreateCDRule",
			Router: `/:flow_id/cd-rules`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "RemoveCDRule",
			Router: `/:flow_id/cd-rules/:rule_id`,
			AllowHTTPMethods: []string{"DELETE"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "UpdateCDRule",
			Router: `/:flow_id/cd-rules/:rule_id`,
			AllowHTTPMethods: []string{"PUT"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "ListBuilds",
			Router: `/:flow_id/builds`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "CreateFlowBuild",
			Router: `/:flow_id/builds`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "GetLastBuildDetails",
			Router: `/:flow_id/lastbuild`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "ListStagesBuilds",
			Router: `/:flow_id/builds/:flow_build_id`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "StopBuild",
			Router: `/:flow_id/stages/:stage_id/builds/:build_id/stop`,
			AllowHTTPMethods: []string{"PUT"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "ListBuildsOfStage",
			Router: `/:flow_id/stages/:stage_id/builds`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "GetStageBuildLogsFromES",
			Router: `/:flow_id/stages/:stage_id/builds/:stage_build_id/log`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiFlowsController"],
		beego.ControllerComments{
			Method: "GetBuildEvents",
			Router: `/:flow_id/stages/:stage_id/builds/:stage_build_id/events`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiImagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiImagesController"],
		beego.ControllerComments{
			Method: "UpdateBaseImage",
			Router: `/:id`,
			AllowHTTPMethods: []string{"PUT"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiImagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiImagesController"],
		beego.ControllerComments{
			Method: "CreateNewBaseImage",
			Router: `/`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiImagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiImagesController"],
		beego.ControllerComments{
			Method: "GetAvailableImages",
			Router: `/`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiImagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiImagesController"],
		beego.ControllerComments{
			Method: "DeleteBaseImage",
			Router: `/:id`,
			AllowHTTPMethods: []string{"DELETE"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiManagedProjectsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiManagedProjectsController"],
		beego.ControllerComments{
			Method: "GetManagedProjects",
			Router: `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiManagedProjectsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiManagedProjectsController"],
		beego.ControllerComments{
			Method: "CreateManagedProject",
			Router: `/`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiManagedProjectsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiManagedProjectsController"],
		beego.ControllerComments{
			Method: "RemoveManagedProject",
			Router: `/:project_id`,
			AllowHTTPMethods: []string{"DELETE"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiManagedProjectsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiManagedProjectsController"],
		beego.ControllerComments{
			Method: "GetManagedProjectDetail",
			Router: `/:project_id`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"],
		beego.ControllerComments{
			Method: "GetRepositories",
			Router: `/:type`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"],
		beego.ControllerComments{
			Method: "Logout",
			Router: `/:type`,
			AllowHTTPMethods: []string{"DELETE"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"],
		beego.ControllerComments{
			Method: "AddRepository",
			Router: `/:type`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"],
		beego.ControllerComments{
			Method: "SyncRepos",
			Router: `/:type`,
			AllowHTTPMethods: []string{"PUT"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"],
		beego.ControllerComments{
			Method: "GetSupportedRepos",
			Router: `/supported`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"],
		beego.ControllerComments{
			Method: "GetAuthRedirectUrl",
			Router: `/:type/auth`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"],
		beego.ControllerComments{
			Method: "GetUserInfo",
			Router: `/:type/user`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"],
		beego.ControllerComments{
			Method: "GetTags",
			Router: `/:type/tags`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiReposController"],
		beego.ControllerComments{
			Method: "GetBranches",
			Router: `/:type/branches`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiScriptsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiScriptsController"],
		beego.ControllerComments{
			Method: "AddScript",
			Router: `/`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiScriptsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiScriptsController"],
		beego.ControllerComments{
			Method: "GetScriptByID",
			Router: `/:id`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiScriptsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiScriptsController"],
		beego.ControllerComments{
			Method: "DeleteScriptByID",
			Router: `/:id`,
			AllowHTTPMethods: []string{"DELETE"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiScriptsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiScriptsController"],
		beego.ControllerComments{
			Method: "UpdateScriptByID",
			Router: `/:id`,
			AllowHTTPMethods: []string{"PUT"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStageLinksController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStageLinksController"],
		beego.ControllerComments{
			Method: "UpdateLinkDirs",
			Router: `/:target_id`,
			AllowHTTPMethods: []string{"PUT"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"],
		beego.ControllerComments{
			Method: "AddOrUpdateDockerfile",
			Router: `/:stage_id/dockerfile`,
			AllowHTTPMethods: []string{"PUT"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"],
		beego.ControllerComments{
			Method: "AddDockerfile",
			Router: `/:stage_id/dockerfile`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"],
		beego.ControllerComments{
			Method: "RemoveDockerfile",
			Router: `/:stage_id/dockerfile`,
			AllowHTTPMethods: []string{"DELETE"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"],
		beego.ControllerComments{
			Method: "GetDockerfile",
			Router: `/:stage_id/dockerfile`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"],
		beego.ControllerComments{
			Method: "RemoveStage",
			Router: `/:stage_id`,
			AllowHTTPMethods: []string{"DELETE"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"],
		beego.ControllerComments{
			Method: "ListStages",
			Router: `/`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"],
		beego.ControllerComments{
			Method: "CreateStage",
			Router: `/`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"],
		beego.ControllerComments{
			Method: "GetStage",
			Router: `/:stage_id`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiStagesController"],
		beego.ControllerComments{
			Method: "UpdateStage",
			Router: `/:stage_id`,
			AllowHTTPMethods: []string{"PUT"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiWebhooksController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:CiWebhooksController"],
		beego.ControllerComments{
			Method: "InvokeBuildsByWebhook",
			Router: `/`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:InvokeCDController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:InvokeCDController"],
		beego.ControllerComments{
			Method: "NotificationHandler",
			Router: `/notification-handler`,
			AllowHTTPMethods: []string{"POST"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:SocketLogController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:SocketLogController"],
		beego.ControllerComments{
			Method: "CreateAccessToken",
			Router: `/token`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:SocketLogController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:SocketLogController"],
		beego.ControllerComments{
			Method: "CreateWebSocketConn",
			Router: `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers:StatsController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers:StatsController"],
		beego.ControllerComments{
			Method: "CollectServerStats",
			Router: `/`,
			AllowHTTPMethods: []string{"GET"},
			MethodParams: param.Make(),
			Params: nil})

}
