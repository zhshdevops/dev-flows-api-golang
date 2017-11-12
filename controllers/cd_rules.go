package controllers

import (
	"dev-flows-api-golang/models"
	"github.com/golang/glog"
)

type CDRulesController struct {
	BaseController
}

//@router / [GET]
func (cdrule *CDRulesController) GetDeploymentCDRule() {
	method := "CDRulesController.GetDeploymentCDRule"
	cluster := cdrule.GetString("cluster")
	name := cdrule.GetString("name")
	if cluster == "" || name == "" {
		glog.Errorf("%s %#v \n", method, "cluster and name is require")
		cdrule.ResponseErrorAndCode("cluster and name is require", 400)
		return
	}

	namespace := cdrule.Namespace

	cdrules := &models.CDRules{}

	cdrulesData, total, err := cdrules.FindDeploymentCDRule(namespace, cluster, name)
	if err != nil {
		glog.Errorf("%s %#v \n", method, err)
		cdrule.ResponseErrorAndCode("db FindDeploymentCDRule is failed", 500)
		return
	}

	cdrule.ResponseSuccessDevops(cdrulesData, total)

	return

}

//@router / [DELETE]
func (cdrule *CDRulesController) DeleteDeploymentCDRule() {
	method := "CDRulesController.DeleteDeploymentCDRule"
	cluster := cdrule.GetString("cluster")
	name := cdrule.GetString("name")
	if cluster == "" || name == "" {
		glog.Errorf("%s %#v \n", method, "cluster and name is require")
		cdrule.ResponseErrorAndCode("cluster and name is require", 400)
		return
	}

	cdrule.Audit.SetOperationType(models.AuditOperationDelete)
	cdrule.Audit.SetResourceType(models.AuditResourceCDRules)
	namespace := cdrule.Namespace

	cd := &models.CDRules{}
	err := cd.DeleteDeploymentCDRule(namespace, cluster, name)
	if err != nil {
		glog.Errorf("%s %#v \n", method, err)
		cdrule.ResponseErrorAndCode("db DeleteDeploymentCDRule is failed", 500)
		return
	}

	cdrule.ResponseErrorAndCode("CD rule was removed successfully", 200)

	return

}
