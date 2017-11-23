package controllers

import (
	"github.com/golang/glog"
	"dev-flows-api-golang/models"
	"k8s.io/apimachinery/pkg/util/json"
	"strings"
	"time"
	"dev-flows-api-golang/modules/client"
	"fmt"

	//"k8s.io/client-go/1.4/pkg/api/v1"

	"net/http"
	"dev-flows-api-golang/util/uuid"
)

type InvokeCDController struct {
	ErrorController
}

//"events": [
//{
//"id": "6d71d21f-e378-4e0a-84ad-f56dda34d794",
//"timestamp": "2017-11-23T12:45:48.618301753+08:00",
//"action": "push",
//"target": {
//"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
//"size": 1779,
//"digest": "sha256:5d2f8a1b3e58fdbbdd698e5e7615fca758afea765b32cd93e0854226dc97e5f6",
//"length": 1779,
//"repository": "qinzhao-harbor/gogsimage",
//"url": "http://10.39.0.102/v2/qinzhao-harbor/gogsimage/manifests/sha256:5d2f8a1b3e58fdbbdd698e5e7615fca758afea765b32cd93e0854226dc97e5f6",
//"tag": "20170919.155947.88"
//},
//"request": {
//"id": "69ddb608-6483-49c2-bf72-36570b220a0c",
//"addr": "10.39.0.102",
//"host": "10.39.0.102",
//"method": "PUT",
//"useragent": "docker/1.13.0 go/go1.7.3 git-commit/49bf474 kernel/3.10.0-327.4.5.el7.x86_64 os/linux arch/amd64 UpstreamClient(Docker-Client/1.13.0 \\(linux\\))"
//},
//"actor": {
//"name": "qinzhao"
//},
//"source": {
//"addr": "61f24253d6ad:5000",
//"instanceID": "e4a873b8-3af7-4867-a061-4a81d4d4a926"
//}
//}
//]
//}

// @router /notification-handler [POST]
func (ic *InvokeCDController) NotificationHandler() {
	ic.Audit.Skip = true
	var notification models.Notification
	method := "InvokeCDController.NotificationHandler"
	message := ""
	body := ic.Ctx.Input.RequestBody
	glog.Infof("%s %s\n", method, string(body))
	if string(body) == "" {
		message = " request body is empty or Invalid request body."
		ic.ResponseErrorAndCode(message, http.StatusBadRequest)
		return
	}
	err := json.Unmarshal(body, &notification)
	if err != nil {
		message = "json 解析失败"
		glog.Errorf("%s json 解析失败:%v\n", method, err)
		ic.ResponseErrorAndCode(message, http.StatusBadRequest)
		return
	}

	glog.Infof("response:======>>%#v\n", notification)

	if len(notification.Events) < 1 {
		message = "Invalid request body."
		glog.Errorf("%s Invalid request body:%s\n", method, notification)
		ic.ResponseErrorAndCode(message, http.StatusBadRequest)
		return
	}
	events := notification.Events[0]
	if events.Action != "push" ||
		strings.Index(events.Target.MediaType, "docker.distribution.manifest") < 0 ||
		strings.Index(events.Request.UserAgent, "docker") < 0 {
		message = "Skipped due to: 1) Not a push. 2) Not manifest update. 3. Not from docker client"
		glog.Errorf("%s %v\n", method, "Skipped due to: 1) Not a push. 2) Not manifest update. 3. Not from docker client")
		ic.ResponseErrorAndCode(message, http.StatusOK)
		return
	}

	var imageInfo models.ImageInfo
	imageInfo.Tag = events.Target.Tag
	imageInfo.Fullname = events.Target.Repository
	imageInfo.Projectname = strings.Split(events.Target.Repository, "/")[0]

	glog.Infof("imageInfo====>%v\n", imageInfo)

	//查询CD规则
	cdrules, result, err := models.NewCdRules().FindEnabledRuleByImage(imageInfo.Fullname)
	if err != nil || result == 0 {
		glog.Infof("%s There is no CD rule that matched this image:result=%d err=[%v]\n", method, result, err)
		message = "There is no CD rule that matched this image:" + imageInfo.Fullname
		ic.ResponseErrorAndCode(message, http.StatusOK)
		return
	}

	var log models.CDDeploymentLogs
	cdlog := models.NewCDDeploymentLogs()
	var cdresult models.Result
	start_time := time.Now()

	newDeploymentArray := make([]models.NewDeploymentArray, 0)
	newDeployment := models.NewDeploymentArray{}
	for index, cdrule := range cdrules {
		glog.Infof("%s  New image tag =[%s] \n", method, imageInfo.Tag)
		k8sClient := client.GetK8sConnection(cdrule.BindingClusterId)
		if k8sClient == nil {
			glog.Errorf(" The specified cluster %s does not exist %s %v \n", cdrule.BindingClusterId, method, err)
			if (len(cdrules) - 1) == index {
				ic.ResponseErrorAndCode("The specified cluster"+cdrule.BindingClusterId+" does not exist", http.StatusNotFound)
				return
			}
			continue
		}
		deployment, err := k8sClient.ExtensionsClient.Deployments(cdrule.Namespace).Get(cdrule.BindingDeploymentName)
		if err != nil || deployment.Status.AvailableReplicas == 0 {
			glog.Errorf("Exception occurs when validate each CD rule: %s %v \n", method, err)

			log.CdRuleId = cdrule.RuleId
			log.TargetVersion = imageInfo.Tag
			log.CreateTime = time.Now()
			cdresult.Status = 2
			cdresult.Duration = int64(time.Now().Sub(start_time) / time.Microsecond)
			cdresult.Error = fmt.Sprintf("%s", err)
			data, err := json.Marshal(cdresult)
			if err != nil {
				glog.Errorf("%s json marshal failed:%v\n", method, err)
				message = "json Marshal failed " + string(data)
				ic.ResponseErrorAndCode(message, 401)
				return
			}
			log.Result = string(data)
			log.Id = uuid.NewCDLogID()
			inertRes, err := cdlog.InsertCDLog(log)
			if err != nil {
				detail := &EmailDetail{
					Type:    "cd",
					Result:  "failed",
					Subject: fmt.Sprintf(`镜像%s的持续集成执行失败`, cdrule.ImageName),
					Body:    fmt.Sprintf(`校验持续集成规则时发生异常或者该服务已经停止或删除`),
				}
				detail.SendEmailUsingFlowConfig(cdrule.Namespace, cdrule.FlowId)
				glog.Errorf("%s inertRes=%d %v\n", method, inertRes, err)
				message = "InsertCDLog failed " + string(data)
				ic.ResponseErrorAndCode(message, http.StatusConflict)
				return
			}
			////send mail
			detail := &EmailDetail{
				Type:    "cd",
				Result:  "failed",
				Subject: fmt.Sprintf(`镜像%s的持续集成执行失败`, cdrule.ImageName),
				Body:    fmt.Sprintf(`校验持续集成规则时发生异常或者该服务已经停止`),
			}
			detail.SendEmailUsingFlowConfig(cdrule.Namespace, cdrule.FlowId)
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
			imageInfo.Fullname+" "+imageInfo.Tag)
		message = "No rule matched to invoke the service deployment."
		ic.ResponseErrorAndCode(message, http.StatusOK)
		return
	}

	//开始升级
	for _, dep := range newDeploymentArray {

		if dep.Deployment.Status.AvailableReplicas == 0 ||
			fmt.Sprintf("%s", dep.Deployment.ObjectMeta.UID) !=
				dep.BindingDeploymentId {
			glog.Warningf("%s 该服务已经停止或者没有找到相关服务. %s\n", method,
				imageInfo.Fullname+":"+imageInfo.Tag)
			continue
		}

		k8sClient := client.GetK8sConnection(dep.Cluster_id)
		if k8sClient == nil {
			glog.Errorf("%s get kubernetes clientset failed: %v \n", method, err)
			continue
		}
		if models.Upgrade(dep.Deployment, imageInfo.Fullname, dep.NewTag, dep.Match_tag, dep.Strategy) {

			dp, err := k8sClient.ExtensionsClient.Deployments(dep.Deployment.ObjectMeta.Namespace).Update(dep.Deployment)
			if err != nil {
				glog.Errorf("%s deployment=[%v], err:%v \n", method, dp, err)
				//失败时插入日志
				log.CdRuleId = dep.Rule_id
				log.TargetVersion = imageInfo.Tag
				log.CreateTime = time.Now()
				cdresult.Status = 2
				cdresult.Duration = int64(time.Now().Sub(start_time) / time.Microsecond)
				cdresult.Error = fmt.Sprintf("%s", err)
				data, err := json.Marshal(cdresult)
				if err != nil {
					glog.Errorf("%s json marshal failed:%v\n", method, err)
					message = "json Marshal failed " + string(data)
					ic.ResponseErrorAndCode(message, 401)
					return
				}
				log.Result = string(data)
				log.Id = uuid.NewCDLogID()
				inertRes, err := cdlog.InsertCDLog(log)
				if err != nil {
					detail := &EmailDetail{
						Type:    "cd",
						Result:  "failed",
						Subject: fmt.Sprintf(`镜像%s的持续集成执行失败`, imageInfo.Fullname),
						Body:    fmt.Sprintf(`校验持续集成规则时发生异常或者该服务已经停止或删除`),
					}
					detail.SendEmailUsingFlowConfig(dep.Namespace, dep.Flow_id)
					glog.Errorf("%s insert deployment log failed: inertRes=%d, err:%v\n", method, inertRes, err)
					message = "InsertCDLog failed " + string(data)
					ic.ResponseErrorAndCode(message, http.StatusConflict)
					return
				}

				detail := &EmailDetail{
					Type:    "cd",
					Result:  "failed",
					Subject: fmt.Sprintf(`镜像%s的持续集成执行失败`, imageInfo.Fullname),
					Body:    fmt.Sprintf(`更新服务时发生异常:%v`, err),
				}
				detail.SendEmailUsingFlowConfig(dep.Namespace, dep.Flow_id)
				continue
			}

			glog.Infof("kubernetes CD Success, deployment :%v\n", dp)

		}

		//成功时插入日志
		log.CdRuleId = dep.Rule_id
		log.TargetVersion = imageInfo.Tag
		log.CreateTime = time.Now()
		cdresult.Status = 1
		cdresult.Duration = int64(time.Now().Sub(start_time) / time.Microsecond)
		cdresult.Error = fmt.Sprintf("%s", err)
		data, err := json.Marshal(cdresult)
		if err != nil {
			glog.Errorf("%s json marshal failed:%v\n", method, err)
			message = "json Marshal failed " + string(data)
			ic.ResponseErrorAndCode(message, 401)
			return
		}
		log.Result = string(data)
		log.Id = uuid.NewCDLogID()
		inertRes, err := cdlog.InsertCDLog(log)
		if err != nil {
			glog.Errorf("%s insert deployment log failed: inertRes=%d, err:%v\n", method, inertRes, err)
		}

		detail := &EmailDetail{
			Type:    "cd",
			Result:  "success",
			Subject: fmt.Sprintf(`持续集成执行成功，镜像%s已更新`, imageInfo.Fullname),
			Body: fmt.Sprintf(`已将服务%s使用的镜像更新为%s:%s的最新版本`,
				dep.Deployment.ObjectMeta.Name, imageInfo.Fullname, imageInfo.Tag),
		}
		detail.SendEmailUsingFlowConfig(dep.Namespace, dep.Flow_id)

	}

	glog.Infof("%s %s", method, "Continuous deployment completed successfully")
	ic.ResponseErrorAndCode("Continuous deployment completed successfully", http.StatusOK)
	return
}
