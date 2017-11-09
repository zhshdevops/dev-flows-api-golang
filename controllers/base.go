package controllers

import (
	"dev-flows-api-golang/models"
	"dev-flows-api-golang/models/user"
	//"dev-flows-api-golang/modules/tenx/errors"
	shortid "dev-flows-api-golang/util/uuid"
	"dev-flows-api-golang/modules/workpool"
	"regexp"

	"flag"

	"fmt"
	"net/http"
	"runtime"
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

var workPool *workpool.WorkPool

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

	// init audit channel and audit routine
	workPool = workpool.New(runtime.NumCPU()*3, 100)
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
		c.Audit.RequestBody = string(c.Ctx.Input.RequestBody)
	}

	c.Audit.Method = c.Ctx.Request.Method
	c.Audit.URL = c.Ctx.Request.URL.String()
	c.Audit.RequestBody = string(c.Ctx.Input.RequestBody)
	if isWebSocketConnect(c) == true {
		c.User = &user.UserModel{}
	} else {
		username := c.Ctx.Input.Header("username")
		token := c.Ctx.Input.Header("authorization")
		if len(token) == 0 {
			token = c.Ctx.Input.Header("Authorization")
		}

		if isNoAuth(c) {
			c.User = &user.UserModel{}
		} else {
			var err error
			prefix := "token "
			if !IfCheckTocken {
				if strings.HasPrefix(strings.ToLower(token), prefix) {
					c.User, err = checkToken(username, token[len(prefix):])
					if err != nil {
						glog.Errorln(method, "Check token failed", err)
						c.ErrorUnauthorized()
						return
					}
				} else {
					glog.Errorln(method, "Missing token prefix")
					c.ErrorBadRequest("Invalid authorization header", nil)
					return

				}
			} else {
				c.User, err = checkToken(username, "")
				if err != nil {
					glog.Errorln(method, "Check token failed", err)
					c.ErrorUnauthorized()
					return
				}
			}
		}
		space := c.Ctx.Input.Header("teamspace")
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
			// check whether the user is in this team
			//teamuser := &team2user.TeamUserModel{}
			//if !c.IsUserSuperAdmin() && !teamuser.CheckUserTeamspace(c.User.UserID, space) {
			//	glog.Errorf("%s user %d is not belong to teamspace %s\n", method, c.User.UserID, space)
			//	c.ErrorUnauthorized()
			//	return
			//}
			//c.Namespace = space
			//c.NamespaceType = TeamSpace
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

	// if http status code was not set, then it should be 200
	if c.Audit.HTTPStatusCode == 0 {
		c.Audit.HTTPStatusCode = http.StatusOK
	}

	// only update duration and status code
	err := workPool.PostWork("update record", (&models.AuditInfo{
		ID:             c.Audit.ID,
		UpdateRecord:   true,
		Duration:       int(time.Now().Sub(c.Audit.StartTime) / time.Microsecond),
		HTTPStatusCode: c.Audit.HTTPStatusCode,
	}).CopyCustomizedNameFrom(&c.Audit))
	if err != nil {
		glog.Errorln(method, "update record", c.Audit.ID, "work failed", err)
	}
}

//// check string parameter in url path
//func (c *BaseController) CheckPathParamsOrRespErr(input map[string]StringInputChecker) (map[string]string, bool) {
//	return c.checkStringParamsOrRespErr(input, "Invalid parameter in URL path")
//}
//
//func (c *BaseController) CheckQueryParamsOrRespErr(input map[string]StringInputChecker) (map[string]string, bool) {
//	return c.checkStringParamsOrRespErr(input, "Invalid parameter in query part")
//}
//
//func (c *BaseController) checkStringParamsOrRespErr(input map[string]StringInputChecker, errMsg string) (map[string]string, bool) {
//	var params map[string]string
//	for name, checker := range input {
//		displayed := strings.Trim(name, " :")
//		value, err := checker(displayed, c.Ctx.Input.Query(name))
//		if nil != err {
//			cause := errors.StatusCause{Message: errMsg, Field: displayed}
//			c.ErrorBadRequest(err.Error(),
//				&errors.StatusDetails{
//					Causes: []errors.StatusCause{cause},
//				})
//			return nil, false
//		}
//		if nil == params {
//			params = make(map[string]string)
//		}
//		params[name] = value
//	}
//	return params, true
//}

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
		return nil, fmt.Errorf("User '" + username + "' is not authorized to access TenxCloud API service.")
	}
	if !IfCheckTocken {
		if token != userModel.APIToken {
			glog.Errorln(method, "user", username, "token", token, "is not correct")
			return nil, fmt.Errorf("invalid api token")
		}
	}

	return userModel, nil
}

//// GetAuditInfo get audit info
//func (c *BaseController) GetAuditInfo() {
//	method := "controllers/BaseController.GetAuditInfo"
//
//	if c.Audit.Skip {
//		return
//	}
//
//	// get caller name
//	pc, file, line, ok := runtime.Caller(1)
//	if ok {
//		callerName := runtime.FuncForPC(pc).Name()
//		glog.V(7).Infoln(method, "caller info:", pc, file, line, ok)
//		glog.V(7).Infoln(method, "caller name:", callerName)
//
//		// example callerName: "api-server/controllers/app.(*Controller).List"
//		i, ok := apiAuditInfos[strings.TrimPrefix(callerName, "api-server/controllers/")]
//		if ok {
//			c.Audit.SetResourceType(i.res)
//			c.Audit.SetOperationType(i.op)
//			c.Audit.SetResourceName(c.getParameter(i.resNameKey))
//			c.Audit.SetResourceID(c.getParameter(i.resIDKey))
//
//			if i.resNameInRequestBody {
//				// resource name is in request body
//				// store in both resource_name and resource_config
//				c.Audit.SetResourceName(c.Audit.RequestBody)
//			}
//
//			if i.res == models.AuditResourceUser && i.op == models.AuditOperationUpdate {
//				userID, _ := c.GetInt32(":user")
//				um := &user.UserModel{UserID: userID}
//				um.Get()
//				c.Audit.SetResourceName(um.Username)
//			}
//
//		} else {
//			glog.Errorln(method, "Failed to get audit info for api", callerName)
//			glog.Errorln(method, c.Audit.Namespace, c.Audit.Method, c.Audit.URL)
//			c.Audit.Skip = true
//			return
//		}
//	} else {
//		glog.Errorln(method, "Failed to get caller pc")
//		glog.Errorln(method, c.Audit.Namespace, c.Audit.Method, c.Audit.URL)
//	}
//
//	// make sure use ":cluster" as cluster id in routers/router.go
//	c.Audit.SetClusterID(c.getParameter(":cluster"))
//
//	// Insert audit record
//	err := workPool.PostWork("insert record", &c.Audit)
//	if err != nil {
//		glog.Errorln(method, "insert record", c.Audit.ID, "work failed", err)
//	}
//}

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

//IsUserCurrentTeamAdmin validates whether current user is the admin of specified team
func (c *BaseController) IsUserCurrentTeamAdmin(teamID string) bool {
	//method := "IsUserCurrentTeamAdmin"
	//
	//// Super admin can be team admin
	//if c.IsUserSuperAdmin() {
	//	return true
	//}
	// check user is a member of team and is administrator role
	//userID := c.User.UserID
	//role, err := team2user.NewTeamUserModel().GetRole(teamID, userID)
	//if err == orm.ErrNoRows {
	//	glog.Infof("user %v is not a mermber of team %v", userID, teamID)
	//	c.ErrorUnauthorized()
	//	return false
	//}
	//if err != nil {
	//	glog.Errorln(method, "get team info failed.", err)
	//	c.ErrorInternalServerError(err)
	//	return false
	//}
	//if role == team2user.TeamNormalUser {
	//	glog.Infof("user %v is a normal user of team %v, can't get team detail info", userID, teamID)
	//	c.ErrorUnauthorized()
	//	return false
	//}
	return true
}

// CanOperateCurrentUser check if the request can update specified user(view, update)
// teamID: current team the request user is in (current contenxt team)
// userID: the user that request user is trying to view
// For /users router by now
func (c *BaseController) CanOperateCurrentUser(teamID string, operatedUser int32) bool {
	//method := "CanOperateCurrentUser"
	//
	//userID := c.User.UserID
	//// 1. operatedUser and current user is the same
	//if userID == operatedUser {
	//	return true
	//}
	//// No team context
	//if teamID == "" {
	//	return false
	//}
	//// Super admin can operate
	//if c.IsUserSuperAdmin() {
	//	return true
	//}
	// 2. Check if request user is admin role of the team
	// Only admin role has the posibility to view other member
	//role, err := team2user.NewTeamUserModel().GetRole(teamID, userID)
	//if err == orm.ErrNoRows {
	//	glog.Errorf("user %v is not a mermber of team %v\n", userID, teamID)
	//	return false
	//}
	//if err != nil {
	//	glog.Errorln(method, "get team info of current user failed.", err)
	//	return false
	//}
	//if role == team2user.TeamNormalUser {
	//	glog.Errorf("user %v is a normal user of team %v, can't get team detail info\n", userID, teamID)
	//	return false
	//}
	//// 3. Check if operatedUser is a member of the team
	//// Don't allow to operate user outside current team context
	//_, err = team2user.NewTeamUserModel().GetRole(teamID, operatedUser)
	//if err == orm.ErrNoRows {
	//	glog.Errorf("user %v is not a mermber of team %v\n", operatedUser, teamID)
	//	return false
	//}
	//if err != nil {
	//	glog.Errorln(method, "get team info of operator user failed.", err)
	//	return false
	//}

	return true
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
	//if user.Role == team2user.SuperAdminUser {
	//	return true
	//}
	return false
}

func isNoAuth(c *BaseController) bool {
	url := c.Audit.URL
	method := c.Audit.Method
	if url == "/api/v2/users/login" && method == "POST" {
		return true
	} else if strings.HasPrefix(url, "/spi/v2/teams/invitations?code=") && method == "GET" {
		return true
	} else if url == "/spi/v2/users/jointeam" && method == "POST" {
		return true
	} else if url == "/spi/v2/users" && method == "POST" {
		return true
	} else if url == "/spi/v2/users/vsettan" && method == "POST" {
		return true
	} else if strings.HasPrefix(c.Ctx.Request.URL.Path, "/spi/v2/oem") && method == "GET" {
		return true
	}
	return false
}

func isWebSocketConnect(c *BaseController) bool {
	var connectionUpgradeRegex = regexp.MustCompile("(^|.*,\\s*)upgrade($|\\s*,)")
	return connectionUpgradeRegex.MatchString(strings.ToLower(c.Ctx.Input.Header("Connection"))) && strings.ToLower(c.Ctx.Input.Header("Upgrade")) == "websocket"
}
