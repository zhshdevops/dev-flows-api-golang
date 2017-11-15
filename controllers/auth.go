package controllers

import (
	"github.com/golang/glog"
	"strings"
	"dev-flows-api-golang/models/user"
	"net/http"
)

const ADMIN_ROLE = 2

type AuthController struct {
	BaseController
}

//@router / [GET]
func (auth *AuthController) AuthByToken() {
	method := "AuthController.AuthByToken"
	username := auth.Ctx.Input.Header("username")
	token := auth.Ctx.Input.Header("authorization")
	teamspace := auth.Ctx.Input.Header("teamspace")
	onbehalfuser := auth.Ctx.Input.Header("onbehalfuser")
	if username == "" || token == "" {
		glog.Warningf("%s User is not authorized:%v\n", method, auth.Ctx.Request.Header)
		auth.ResponseErrorAndCode("User is not authorized. Authorization, username are required. ", 400)
		return
	}
	glog.Infof("teamspace=%s,onbehalfuser=%s", teamspace, onbehalfuser)

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

			if userInfo.Role != ADMIN_ROLE {
				if teamspace != "" {
					resultCOunt, err = userInfo.IsHaveAuthor(teamspace)
					if err != nil || resultCOunt < 1 {
						glog.Warningf("%s User is not authorized:%v,err:%v\n", method, auth.Ctx.Request.Header, err)
						auth.ResponseErrorAndCode("Sorry, you cant't switch to this teamspace", http.StatusForbidden)
						return
					}
				}
			}

			currentNS := auth.User.Namespace
			if teamspace != "" {
				currentNS = teamspace

			} else if onbehalfuser != "" {
				currentNS = onbehalfuser
			}

			auth.User.Namespace = currentNS

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

	return

}
