// @APIVersion 1.0.0
// @Title Enncloud DevOps API
// @Description 新智云CICD API
// @Contact qinzhao@ennew.cn
// @TermsOfServiceUrl http://paas.enncloud.cn
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"dev-flows-api-golang/controllers"
	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/api/v2/devops",

		beego.NSNamespace("/ci/images",
			beego.NSInclude(
				&controllers.CiImagesController{},
			),
		),
		beego.NSNamespace("/repos",
			beego.NSInclude(
				&controllers.CiReposController{},
			),
		),
		beego.NSNamespace("/registry",
			beego.NSInclude(
				&controllers.InvokeCDController{},
			),
		),
		beego.NSNamespace("/managed-projects",
			beego.NSInclude(
				&controllers.CiManagedProjectsController{},
			),
		),
		beego.NSNamespace("/cd-rules",
			beego.NSInclude(
				&controllers.CDRulesController{},
			),
		),
		beego.NSNamespace("/dockerfiles",
			beego.NSInclude(
				&controllers.CiDockerfileController{},
			),
		),
		beego.NSNamespace("/ci-flows",
			beego.NSInclude(
				&controllers.CiFlowsController{},
			),
		),
		beego.NSNamespace("/ci-flows/:flow_id/stages",
			beego.NSInclude(
				&controllers.CiStagesController{},
			),
		),
		beego.NSNamespace("/ci-flows/:flow_id/stages/:stage_id/link",
			beego.NSInclude(
				&controllers.CiStageLinksController{},
			),
		),
		beego.NSNamespace("/ci-scripts",
			beego.NSInclude(
				&controllers.CiScriptsController{},
			),
		),
		beego.NSNamespace("/managed-projects/webhooks/:project_id",
			beego.NSInclude(
				&controllers.CiFlowBuildLogsController{},
			),
		),
		beego.NSNamespace("/auth",
			beego.NSInclude(
				&controllers.AuthController{},
			),
		),
		beego.NSNamespace("/stats",
			beego.NSInclude(
				&controllers.StatsController{},
			),
		),


	)

	beego.AddNamespace(ns)
	//beego.Handler("/stagebuild/status/",controllers.NewJobWatcherSocket().Handler)
	beego.Handler("/socket.io/", controllers.SocketId)
	beego.Handler("/stagebuild/log/", controllers.StageBuildLog)
	beego.Handler("/stagebuild/status/", controllers.StageBuildStatus)

}
