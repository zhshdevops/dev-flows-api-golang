package controllers

import (
	"dev-flows-api-golang/models"
	"dev-flows-api-golang/models/user"
	shortid "dev-flows-api-golang/util/uuid"
	"dev-flows-api-golang/models/team2user"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"
	"dev-flows-api-golang/modules/client"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

const (
	UserSpace    spaceType = iota
	TeamSpace
	OnBehalfUser
)

type spaceType int32

type BaseController struct {
	ErrorController
	User          *user.UserModel
	Namespace     string
	NamespaceType spaceType
}

// QueryLogRequest structure of query log request
type QueryLogRequest struct {
	DateStart string `json:"date_start"`
	DateEnd   string `json:"date_end"`
	Kind      string `json:"kind"`
	From      int    `json:"from"`
	Size      int    `json:"size"`
	Keyword   string `json:"keyword"`
	TimeNano  string `json:"time_nano"`
	Direction string `json:"direction"`
	LogType   string `json:"log_type"`
	LogVolume string `json:"log_volume"`
	FileName  string `json:"filename"`
}

// response
type response struct {
	Message string `json:"message"`
}

func init() {
	// allow specify config file path
	configPath := flag.String("appconfig", "conf/app.conf", "config path")

	flag.Parse()
	flag.Set("logtostderr", "true")
	client.Initk8sClient()
	client.GetHarborServer()
	err := beego.LoadAppConfig("ini", *configPath)
	if err != nil {
		panic("Load config file " + *configPath + " failed: " + err.Error())
	}
}

var IfCheckTocken bool
// Prepare validation all requests
func (c *BaseController) Prepare() {
	method := "controllers/BaseController.Prepare"
	glog.Infof("request url:%s\n", c.Ctx.Request.URL.String())
	if strings.Contains(c.Ctx.Request.URL.String(), "managed-projects/webhooks") {
		IfCheckTocken = true
	}
	// set audit id
	if strings.ToUpper(c.Ctx.Request.Method) == "GET" {
		c.Audit.Skip = true
	} else {
		c.Audit.ID = shortid.NewAudit()
		c.Audit.StartTime = time.Now()
		c.Audit.Operator = c.Ctx.Input.Header("username")
		c.Audit.Method = strings.ToUpper(c.Ctx.Request.Method)
		c.Audit.URL = c.Ctx.Request.URL.String()
		c.Audit.RequestBody = string(c.Ctx.Input.RequestBody)
	}

	c.Audit.Method = c.Ctx.Request.Method
	c.Audit.URL = c.Ctx.Request.URL.String()
	c.Audit.RequestBody = string(c.Ctx.Input.RequestBody)
	if IfCheckTocken {
		c.User = &user.UserModel{}
	} else {

		username := c.Ctx.Input.Header("username")

		token := c.Ctx.Input.Header("authorization")
		if len(token) == 0 {
			token = c.Ctx.Input.Header("Authorization")
		}

		//团队空间
		space := c.Ctx.Input.Header("teamspace")

		var err error
		prefix := "token "
		if strings.HasPrefix(strings.ToLower(token), prefix) {
			c.User, err = checkToken(username, token[len(prefix):])
			if err != nil {
				glog.Errorln(method, "Check token failed", err)
				c.ErrorUnauthorized()
				return
			}
			c.Namespace = c.User.Namespace
		} else {
			glog.Errorln(method, "Missing token prefix")
			c.ErrorBadRequest("Invalid authorization header", nil)
			return
		}

		// If no teamspace defined, then it's the user space
		if "" == space {
			// Check if system admin is managing user space
			onbehalfuser := c.Ctx.Input.Header("onbehalfuser")
			if onbehalfuser == "" {
				c.Namespace = c.User.Namespace
				c.NamespaceType = UserSpace
			} else {
				// Manage user space only for admin user
				if c.IsUserSuperAdmin() {
					c.Namespace = onbehalfuser
					c.NamespaceType = OnBehalfUser
				} else {
					c.Namespace = c.User.Namespace
					c.NamespaceType = UserSpace
				}
			}
		} else {
			//check whether the user is in this team
			teamuser := &team2user.TeamUserModel{}
			if !c.IsUserSuperAdmin() && !teamuser.CheckUserTeamspace(c.User.UserID, space) {
				glog.Errorf("%s user %d is not belong to teamspace %s\n", method, c.User.UserID, space)
				c.ErrorUnauthorized()
				return
			}
			c.Namespace = space
			c.NamespaceType = TeamSpace
		}
		glog.V(3).Infof("Namespace is %s and type is %v\n", c.Namespace, c.NamespaceType)
		c.Audit.Namespace = c.Namespace
	}
}

// Finish operation log
func (c *BaseController) Finish() {
	// record operation result
	method := "controller/BaseController.Finish"

	if c.Audit.Skip {
		return
	}

	if c.Audit.HTTPStatusCode == 0 {
		c.Audit.HTTPStatusCode = http.StatusOK
	}

	c.Audit.Namespace = c.Namespace

	c.Audit.Duration = int(time.Now().Sub(c.Audit.StartTime) / time.Microsecond)
	c.Audit.UpdateRecord = true
	err := models.NewAuditRecord(&c.Audit).Insert()
	if err != nil {
		glog.Errorf("%s insert into AuditRecord to database failed:%v\n", method, err)
	}

}

// checkToken check user's token
func checkToken(username, token string) (*user.UserModel, error) {
	method := "controllers/checkToken"
	if username == "" {
		glog.Errorf("%s, username cannot be empty\n", method)
		return nil, fmt.Errorf("Bad user" + username + ".")
	}
	userModel := &user.UserModel{}
	// use cache for better performance
	_, err := userModel.GetByName(username)
	if err != nil {
		return nil, fmt.Errorf("User '" + username + "' is not authorized to access EnnCloud Devops API service.")
	}
	if token != userModel.APIToken {
		glog.Errorln(method, "user", username, "token", token, "is not correct")
		return nil, fmt.Errorf("invalid api token")
	}

	return userModel, nil
}

func (c *BaseController) SetAuditInfo() *models.AuditInfo {
	return &models.AuditInfo{
		ID:           c.Audit.ID,
		UpdateRecord: true,
		Method:       c.Audit.Method,
		Namespace:    c.Namespace,
		Operator:     c.Ctx.Input.Header("username"),
		RequestBody:  string(c.Ctx.Input.RequestBody),
		StartTime:    c.Audit.StartTime,
		URL:          c.Ctx.Request.URL.String(),
	}
}

func (c *BaseController) getParameter(param string) string {
	if param == "" {
		return ""
	}
	return c.GetString(param, "")
}

// IsUserBelongsToTeam check whether the current user belows to the team
// For teams router by now
func (c *BaseController) IsUserBelongsToTeam(userID int32, teamID string) bool {
	//method := "controllers/BaseController.IsUserBelongsToTeam"
	// Super admin can access all teams
	if c.IsUserSuperAdmin() {
		return true
	}
	//_, err := team2user.NewTeamUserModel().GetRole(teamID, userID)
	//if err == orm.ErrNoRows {
	//	glog.Errorln(method, "User does not belong to this team.", err)
	//	return false
	//}
	//if err != nil {
	//	glog.Errorln(method, "get user/team info failed.", err)
	//	return false
	//}
	return true
}

// IsUserGlobalTeamAdmin return whether user can create team, role = 1 or 2
func (c *BaseController) IsUserGlobalTeamAdmin() bool {
	method := "IsUserGlobalTeamAdmin"
	user := user.NewUserModel()
	user.UserID = c.User.UserID
	_, err := user.Get()
	if err == orm.ErrMultiRows {
		glog.Errorln(method, "get user role failed", err)
		return false
	}
	//if user.Role == team2user.TeamAdmin || user.Role == team2user.SuperAdminUser {
	//	return true
	//}
	return false
}

// IsUserSuperAdmin return whether user is super admin, role = 2
func (c *BaseController) IsUserSuperAdmin() bool {
	method := "IsUserSuperAdmin"
	user := user.NewUserModel()
	user.UserID = c.User.UserID
	_, err := user.Get()
	if err == orm.ErrMultiRows {
		glog.Errorln(method, "get user role failed", err)
		return false
	}
	if user.Role == team2user.SuperAdminUser {
		return true
	}
	return false
}
