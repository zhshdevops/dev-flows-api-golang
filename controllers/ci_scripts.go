package controllers

import (
	"dev-flows-api-golang/models"

	"dev-flows-api-golang/util/uuid"
	sqlstatus "dev-flows-api-golang/models/sql/status"
	"github.com/golang/glog"
	"encoding/json"
	"net/http"
)

type CiScriptsController struct {
	BaseController
}

//@router / [POST]
func (cs *CiScriptsController) AddScript() {
	method := " CiScriptsController.AddScript"

	cs.Audit.SetOperationType(models.AuditOperationCreate)
	cs.Audit.SetResourceType(models.AuditResourceOnlineScript)

	contet := string(cs.Ctx.Input.RequestBody)
	if contet == "" {
		cs.ResponseErrorAndCode("the script content is empty", http.StatusBadRequest)
		return
	}
	id := uuid.NewScriptID()
	ciScript := &models.CiScripts{}
	num, err := ciScript.AddScript(id, contet)
	if err != nil {
		glog.Errorf("%s AddScript err:%v, num=%d\n", method, err, num)
		cs.ResponseErrorAndCode("add script failed", http.StatusConflict)
		return
	}

	cs.ResponseScriptAndStatusDevops(id, http.StatusOK)
	return

}

//@router /:id [GET]
func (cs *CiScriptsController) GetScriptByID() {

	method := " CiScriptsController.GetScriptByID"

	Id := cs.Ctx.Input.Param(":id")

	ciScript := &models.CiScripts{}
	err := ciScript.GetScriptByID(Id)
	if err != nil {
		parseNumber, _ := sqlstatus.ParseErrorCode(err)
		if parseNumber == sqlstatus.SQLErrNoRowFound {
			glog.Errorf("%s script cannot be found %v\n", method, err)
			cs.ResponseErrorAndCode("script cannot be found", http.StatusNotFound)
			return
		} else {
			glog.Errorf("%s Get script info by flowId %s failed from database ERR: %v\n", method, err)
			cs.ResponseErrorAndCode("Lookup script failed,please try again!", http.StatusForbidden)
			return
		}

	}
	cs.ResponseResultAndStatusDevops(ciScript, http.StatusOK)
	return
}

//@router /:id [DELETE]
func (cs *CiScriptsController) DeleteScriptByID() {
	method := " CiScriptsController.DeleteScriptByID"
	Id := cs.Ctx.Input.Param(":id")

	cs.Audit.SetResourceID(Id)
	cs.Audit.SetOperationType(models.AuditOperationDelete)
	cs.Audit.SetResourceType(models.AuditResourceOnlineScript)

	ciScript := &models.CiScripts{}
	num, err := ciScript.DeleteScriptByID(Id)
	if err != nil {
		glog.Errorf("%s delete script failed: err:%v, num:%d\n", method, err, num)
		cs.ResponseErrorAndCode("delete script failed", http.StatusBadRequest)
		return
	}
	cs.ResponseResultAndStatusDevops("delete Script Success:"+Id, http.StatusOK)
	return

}

//@router /:id [PUT]
func (cs *CiScriptsController) UpdateScriptByID() {
	method := " CiScriptsController.DeleteScriptByID"
	Id := cs.Ctx.Input.Param(":id")

	cs.Audit.SetResourceID(Id)
	cs.Audit.SetOperationType(models.AuditOperationUpdate)
	cs.Audit.SetResourceType(models.AuditResourceOnlineScript)
	contet := string(cs.Ctx.Input.RequestBody)
	if contet == "" {
		cs.ResponseErrorAndCode("the script content is empty", http.StatusBadRequest)
		return
	}
	ciScript := &models.CiScripts{}
	err := json.Unmarshal(cs.Ctx.Input.RequestBody, ciScript)
	if err != nil {
		glog.Errorf("%s %v\n", method, err)
		return
	}

	err = ciScript.UpdateScriptByID(Id, ciScript.Content)
	if err != nil {
		glog.Errorf("%s update online script failed:%v\n", method, err)
		cs.ResponseErrorAndCode("修改在线脚本失败", 401)
		return
	}
	cs.ResponseResultAndStatusDevops("update sctipt Success"+Id, http.StatusOK)
	return

}
