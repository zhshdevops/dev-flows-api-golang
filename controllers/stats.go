package controllers

import (
	"github.com/golang/glog"
	"strings"
	"dev-flows-api-golang/models/user"
	"dev-flows-api-golang/models"
)

type StatsController struct {
	BaseController
}

//@router / [GET]
func (auth *StatsController) CollectServerStats() {
	method := "AuthController.collectServerStats"
	username := auth.Ctx.Input.Header("username")
	token := auth.Ctx.Input.Header("authorization")
	namespace := auth.Ctx.Input.Header("teamspace")
	onbehalfuser := auth.Ctx.Input.Header("onbehalfuser")

	if username == "" || token == "" {
		glog.Warningf("%s User is not authorized:%v\n", method, auth.Ctx.Request.Header)
		auth.ResponseErrorAndCode("User is not authorized. Authorization, username are required. ", 400)
		return
	}
	glog.Infof("teamspace=%s,onbehalfuser=%s", namespace, onbehalfuser)

	prefix := "token "
	if strings.HasPrefix(strings.ToLower(token), prefix) {
		if token[len(prefix):] == auth.User.APIToken {
			userInfo := user.NewUserModel()
			resultCOunt, err := userInfo.FindByToken()
			if err != nil || resultCOunt < 1 {
				glog.Warningf("%s User is not authorized:%v,err:%v\n", method, auth.Ctx.Request.Header, err)
				auth.ResponseErrorAndCode("User is not authorized. Authorization, username are required. ", 400)
				return
			}

			if namespace == "" {
				namespace = userInfo.Namespace
			}

			if userInfo == nil {
				namespace = ""
			} else {
				if namespace == userInfo.Namespace {
					if userInfo.Role != ADMIN_ROLE {
						resultCOunt, err = userInfo.IsHaveAuthor(namespace)
						if err != nil || resultCOunt < 1 {
							glog.Warningf("%s User is not authorized:%v,err:%v\n", method, auth.Ctx.Request.Header, err)
							namespace = ""
						}
					}
				}
			}

			if onbehalfuser != "" && auth.User.Role == ADMIN_ROLE {
				namespace = onbehalfuser
			}

		} else {
			glog.Errorln(method, "Missing token prefix")
			auth.ErrorBadRequest("Invalid authorization header", nil)
			return
		}

	} else {
		glog.Errorln(method, "Missing token prefix")
		auth.ErrorBadRequest("Invalid authorization header", nil)
		return
	}
	CollectServerStats(namespace)
	return

}

type FlowBuild struct {
	SucceedNumber int `json:"succeed_number"`
	RunningNumber int `json:"running_number"`
	FailedNumber  int `json:"failed_number"`
}
type ResultStats struct {
	FlowBuild FlowBuild  `json:"flow_build"`
}

func CollectServerStats(namespace string) ResultStats {
	var stats ResultStats
	method := "CollectServerStats"
	if namespace != "" {
		flowBuilds, total, err := models.NewCiFlowBuildLogs().QueryFlowBuildStats(namespace)
		if err != nil {
			glog.Errorf("%s, total:%d, err:%v\n", method, total, err)
			return stats
		}

		for _, flowBuid := range flowBuilds {
			glog.Infof("%s, total:%d, err:%v\n", method, total, flowBuid)
			if flowBuid.Status == 0 {
				stats.FlowBuild.SucceedNumber = flowBuid.Count
			} else if flowBuid.Status == 1 {
				stats.FlowBuild.FailedNumber = flowBuid.Count
			} else if flowBuid.Status == 2 {
				stats.FlowBuild.RunningNumber = flowBuid.Count
			}
		}

	}

	return stats
}
