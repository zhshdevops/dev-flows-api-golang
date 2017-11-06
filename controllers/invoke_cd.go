package controllers

import (
	"github.com/golang/glog"
	"dev-flows-api-golang/models"
	"k8s.io/apimachinery/pkg/util/json"
	"strings"
	"time"
	"dev-flows-api-golang/modules/client"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InvokeCDController struct {
	BaseController
}
// @Title CreateUser
// @Description create users
// @Param	body		body 	models.User	true		"body for user content"
// @Success 200 {int} models.User.Id
// @Failure 403 body is empty
// @router /notification-handler [POST]
func (ic *InvokeCDController)NotificationHandler() {
	var notification models.Notification
	method := "InvokeCDController.NotificationHandler"
	message := ""
	body := ic.Ctx.Input.RequestBody
	glog.Infof("%s %s\n", method, string(body))
	namespace := ic.Namespace
	if namespace == "" {
		namespace = ic.Ctx.Input.Header("username")
	}
	if string(body) == "" {
		message = " request body is empty."
		ic.ResponseErrorAndCode(message, 400)
		return
	}
	ic.Audit.Skip = true
	err := json.Unmarshal(body, &notification)
	if err != nil {
		message = "json 解析失败"
		glog.Errorf("%s %v\n", method, err)
		ic.ResponseErrorAndCode(message, 400)
		return
	}

	if len(notification.Events) < 1 {
		message = "Invalid request body."
		glog.Errorf("%s %s\n", method, notification)
		ic.ResponseErrorAndCode(message, 400)
		return
	}
	events := notification.Events[0]
	if events.Action != "push" ||
		strings.Index(events.Target.MediaType, "docker.distribution.manifest") < 0 ||
		strings.Index(events.Request.UserAgent, "docker") < 0 {
		message = "Skipped due to: 1) Not a push. 2) Not manifest update. 3. Not from docker client"
		glog.Errorf("%s %v\n", method, "Skipped due to: 1) Not a push. 2) Not manifest update. 3. Not from docker client")
		ic.ResponseErrorAndCode(message, 200)
		return
	}
	var imageInfo models.ImageInfo
	imageInfo.Tag = events.Target.Tag
	imageInfo.Fullname = events.Target.Repository
	imageInfo.Projectname = strings.Split(events.Target.Repository, "/")[0]

	cdrules, result, err := models.NewCdRules().FindEnabledRuleByImage(imageInfo.Fullname)
	if err != nil || len(cdrules) == 0 {
		glog.Infof("%s result=%d err=[%v]\n", method, result, err)
		message = "There is no CD rule that matched this image:" + imageInfo.Fullname
		ic.ResponseErrorAndCode(message, 200)
		return
	}
	//var log models.CDDeploymentLogs
	//cdlog := models.NewCDDeploymentLogs()
	//var cdresult models.Result
	newDeploymentArray := make([]models.NewDeploymentArray, 0)
	newDeployment := models.NewDeploymentArray{}
	for index, cdrule := range cdrules {
		glog.Infof("%s  New image tag =[%s] \n", method, imageInfo.Tag)
		start_time := time.Now()
		k8sClient := client.GetK8sConnection(cdrule.BindingClusterId)
		if k8sClient == nil {
			glog.Errorf(" The specified cluster %s does not exist %s %v \n", cdrule.BindingClusterId, method, err)
			if (len(cdrules) - 1) == index {
				ic.ResponseErrorAndCode("The specified cluster" + cdrule.BindingClusterId + " does not exist", 404)
				return
			}

			continue
		}
		option:=v1.GetOptions{}
		deployment, err := k8sClient.ExtensionsV1beta1Client.Deployments(namespace).Get(cdrule.BindingDeploymentName,option)
		if err != nil {
			glog.Errorf("Exception occurs when validate each CD rule: %s %v \n", method, err)
			//log.CdRuleId = cdrule.RuleId
			//log.TargetVersion = imageInfo.Tag
			//log.CreateTime = time.Now()
			//cdresult.Status = 2
			//cdresult.Duration = time.Now().Unix() - start_time
			//cdresult.Error = fmt.Sprintf("%s", err)
			//data, err := json.Marshal(cdresult)
			//if err != nil {
			//	glog.Errorf("%s %v\n", method, err)
			//	message = "json Marshal failed "+string(data)
			//	ic.ResponseErrorAndCode(message, 401)
			//	return
			//}
			//log.Result=string(data)
			//inertRes,err:=cdlog.InsertCDLog(log)
			//if err!=nil{
			//	glog.Errorf("%s inertRes=%d %v\n", method,inertRes, err)
			//	message = "InsertCDLog failed "+string(data)
			//	ic.ResponseErrorAndCode(message, 401)
			//	return
			//}
			////send mail
			//return
			continue
		}

		newDeployment.Deployment = deployment
		newDeployment.Namespace = cdrule.Namespace
		newDeployment.Cluster_id = cdrule.BindingClusterId
		newDeployment.NewTag = imageInfo.Tag
		newDeployment.Strategy = cdrule.UpgradeStrategy
		newDeployment.Flow_id = cdrule.FlowId
		newDeployment.Rule_id = cdrule.RuleId
		newDeployment.Match_tag = cdrule.MatchTag
		newDeployment.BindingDeploymentId = cdrule.BindingDeploymentId
		newDeployment.Start_time = start_time
		newDeploymentArray = append(newDeploymentArray, newDeployment)

	}
	if len(newDeploymentArray) == 0 {
		glog.Warningf("%s No rule matched to invoke the service deployment. %s\n", method,
			imageInfo.Fullname + " " + imageInfo.Tag)
		message = "No rule matched to invoke the service deployment."
		ic.ResponseErrorAndCode(message, 200)
		return
	}

	for _, dep := range newDeploymentArray {
		if dep.Deployment.Status.AvailableReplicas == 0 ||
			fmt.Sprintf("%s", dep.Deployment.ObjectMeta.UID) !=
				dep.BindingDeploymentId {
			glog.Warningf("%s newDeploymentArray failed. %s\n", method,
				imageInfo.Fullname + " " + imageInfo.Tag)
			continue
		}

		k8sClient := client.GetK8sConnection(dep.Cluster_id)
		if k8sClient == nil {
			glog.Errorf("%s %v \n", method, err)
			//ic.ResponseErrorAndCode("The specified cluster" + cdrule.BindingClusterId + " does not exist", 404)
			continue
		}
		if models.Upgrade(dep.Deployment, imageInfo.Fullname, dep.NewTag, dep.Match_tag, dep.Strategy) {

			dp, err := k8sClient.ExtensionsV1beta1Client.Deployments(namespace).Update(dep.Deployment)
			if err != nil {
				glog.Errorf("%s dp=[%v] %v \n", method, dp, err)
				continue
			}
		}

	}

	glog.Infof("%s %s", method, "Continuous deployment completed successfully")
	ic.ResponseErrorAndCode("Continuous deployment completed successfully", 200)
	return
}
