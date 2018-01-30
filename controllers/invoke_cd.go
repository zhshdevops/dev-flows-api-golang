package controllers

import (
	"github.com/golang/glog"
	"dev-flows-api-golang/models"
	"encoding/json"
	"strings"
	"time"
	"dev-flows-api-golang/modules/client"
	"fmt"
	"strconv"
	//"k8s.io/client-go/1.4/pkg/api/v1"
	"sync"
	"net/http"
	"dev-flows-api-golang/util/uuid"
)

var imageMaps ImageMaps

type ImageMaps struct {
	ImageMap        map[string]time.Time
	ImageMapRWMutex sync.RWMutex
}

type InvokeCDController struct {
	ErrorController
}

func init() {

	imageMaps = ImageMaps{
		ImageMap: make(map[string]time.Time, 2048),
	}

	go func() {

		for {
			select {
			case <-time.After(10 * time.Minute):
				imageMaps.ImageMapRWMutex.RLock()
				for key, value := range imageMaps.ImageMap {
					if (time.Now().Sub(value) / time.Second) > 180 {

						delete(imageMaps.ImageMap, key)
					}
				}
				imageMaps.ImageMapRWMutex.RUnlock()
			}

		}

	}()

}

// @router /notification-handler [POST]
func (ic *InvokeCDController) NotificationHandler() {
	ic.Audit.Skip = true
	var notification models.Notification
	method := "InvokeCDController.NotificationHandler"
	message := ""
	body := ic.Ctx.Input.RequestBody
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

	if len(notification.Events) < 1 || string(body) == "" {
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
		glog.Infof("%s %v\n", method, "Skipped due to: 1) Not a push. 2) Not manifest update. 3. Not from docker client")
		ic.ResponseErrorAndCode(message, http.StatusOK)
		return
	}

	var imageInfo models.ImageInfo
	imageInfo.Tag = events.Target.Tag
	imageInfo.Fullname = events.Target.Repository
	imageInfo.Projectname = strings.Split(events.Target.Repository, "/")[0]
	var ImageMapKey string = imageInfo.Fullname + ":" + imageInfo.Tag
	imageMaps.ImageMapRWMutex.RLock()
	_, ok := imageMaps.ImageMap[ImageMapKey]
	if !ok {
		imageMaps.ImageMap[ImageMapKey] = time.Now()
		imageMaps.ImageMapRWMutex.RUnlock()
	} else {
		message = "自动部署触发的次数过多"
		ic.ResponseErrorAndCode(message, http.StatusOK)
		imageMaps.ImageMapRWMutex.RUnlock()
		return
	}

	//查询CD规则
	cdrules, result, err := models.NewCdRules().FindEnabledRuleByImage(imageInfo.Fullname)
	if err != nil || result == 0 {
		glog.Infof("%s There is no CD rule that matched this image:result=%d err=[%v]\n", method, result, err)
		message = "There is no CD rule that matched this image:" + imageInfo.Fullname
		ic.ResponseErrorAndCode(message, http.StatusOK)
		imageMaps.ImageMapRWMutex.RLock()
		delete(imageMaps.ImageMap, ImageMapKey)
		imageMaps.ImageMapRWMutex.RUnlock()
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
				imageMaps.ImageMapRWMutex.RLock()
				delete(imageMaps.ImageMap, ImageMapKey)
				imageMaps.ImageMapRWMutex.RUnlock()
				return
			}
			continue
		}
		deployment, err := k8sClient.ExtensionsClient.Deployments(cdrule.Namespace).Get(cdrule.BindingDeploymentName)
		if err != nil || deployment.Status.Replicas == 0 {

			if _, ok := deployment.Spec.Template.ObjectMeta.Labels["tenxcloud.com/cdTimestamp"]; ok {
				cooldownSec := 30
				lastCdTs := deployment.Spec.Template.ObjectMeta.Labels["tenxcloud.com/cdTimestamp"]
				cdTs, _ := strconv.ParseInt(lastCdTs, 10, 64)
				//当前时间与上一次相差不足冷却间隔时，不进行更新
				if (time.Now().Unix() - cdTs) < int64(cooldownSec) {
					glog.Warningf("%s %s %d\n", method, "Upgrade is rejected because the"+
						" deployment was updated too frequently (time.Now().Unix() - cdTs) < int64(cooldownSec)=",
						(time.Now().Unix()-cdTs)-int64(cooldownSec))
					continue
				}
			}

			glog.Warningf("Exception occurs when validate each CD rule: %s %v \n", method, err)

			log.CdRuleId = cdrule.RuleId
			log.TargetVersion = imageInfo.Tag
			log.CreateTime = time.Now()
			cdresult.Status = 2
			cdresult.Duration = fmt.Sprintf("%d ms", time.Now().Sub(start_time)/time.Millisecond)
			cdresult.Error = fmt.Sprintf("%v", err)
			data, err := json.Marshal(cdresult)
			if err != nil {
				glog.Errorf("%s json marshal failed:%v\n", method, err)
				message = "json Marshal failed " + string(data)
				continue
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
				continue
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
	for index, dep := range newDeploymentArray {
		glog.Infof("第一次：%d 部署", index+1)
		if dep.Deployment.Status.Replicas == 0 ||
			fmt.Sprintf("%s", dep.Deployment.ObjectMeta.UID) !=
				dep.BindingDeploymentId {
			glog.Warningf("%s 该服务已经停止或者没有找到相关服务. %s\n", method,
				imageInfo.Fullname+":"+imageInfo.Tag)

			continue
		}

		k8sClient := client.GetK8sConnection(dep.Cluster_id)
		if k8sClient == nil {
			glog.Errorf("%s get kubernetes clientset failed: %v \n", method, err)
			message = "连不上集群，请稍后再试"
			ic.ResponseErrorAndCode(message, http.StatusInternalServerError)
			imageMaps.ImageMapRWMutex.RLock()
			delete(imageMaps.ImageMap, ImageMapKey)
			imageMaps.ImageMapRWMutex.RUnlock()
			return
		}
		if models.Upgrade(dep.Deployment, imageInfo.Fullname, dep.NewTag, dep.Match_tag, dep.Strategy) {
			glog.Infof("dep.Deployment.Spec.Strategy=%v\n", dep.Deployment.Spec.Strategy)
			dp, err := k8sClient.ExtensionsClient.Deployments(dep.Deployment.ObjectMeta.Namespace).Update(dep.Deployment)
			if err != nil {
				glog.Errorf("%s deployment=[%v], err:%v \n", method, dp.Spec.Strategy, err)
				//失败时插入日志
				log.CdRuleId = dep.Rule_id
				log.TargetVersion = imageInfo.Tag
				log.CreateTime = time.Now()
				cdresult.Status = 2
				cdresult.Duration = fmt.Sprintf("%d ms", time.Now().Sub(start_time)/time.Millisecond)
				cdresult.Error = fmt.Sprintf("%v", err)
				data, err := json.Marshal(cdresult)
				if err != nil {
					glog.Errorf("%s json marshal failed:%v\n", method, err)
					continue
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
					continue
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

			//成功时插入日志
			log.CdRuleId = dep.Rule_id
			log.TargetVersion = imageInfo.Tag
			log.CreateTime = time.Now()
			cdresult.Status = 1
			cdresult.Duration = fmt.Sprintf("%d ms", time.Now().Sub(start_time)/time.Millisecond)
			cdresult.Error = fmt.Sprintf("%v", err)
			data, err := json.Marshal(cdresult)
			if err != nil {
				glog.Errorf("%s json marshal failed:%v\n", method, err)
				continue
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
			continue
		} else {
			log.CdRuleId = dep.Rule_id
			log.TargetVersion = imageInfo.Tag
			log.CreateTime = time.Now()
			cdresult.Status = 2
			cdresult.Duration = fmt.Sprintf("%d ms", time.Now().Sub(start_time)/time.Millisecond)
			cdresult.Error = fmt.Sprintf("%v", err)
			data, err := json.Marshal(cdresult)
			if err != nil {
				glog.Errorf("%s json marshal failed:%v\n", method, err)
				message = "json Marshal failed " + string(data)
				continue
			}
			log.Result = string(data)
			log.Id = uuid.NewCDLogID()
			inertRes, err := cdlog.InsertCDLog(log)
			if err != nil {
				glog.Errorf("%s insert deployment log failed: inertRes=%d, err:%v\n", method, inertRes, err)
			}

			detail := &EmailDetail{
				Type:    "cd",
				Result:  "failed",
				Subject: fmt.Sprintf(`镜像%s持续集成执行失败`, imageInfo.Fullname),
				Body: fmt.Sprintf(`服务[%s]更新版本间隔时间太短`,
					dep.Deployment.ObjectMeta.Name),
			}
			detail.SendEmailUsingFlowConfig(dep.Namespace, dep.Flow_id)
			continue
		}

	}

	glog.Infof("%s %s", method, "Continuous deployment completed successfully")
	imageMaps.ImageMapRWMutex.RLock()
	delete(imageMaps.ImageMap, ImageMapKey)
	imageMaps.ImageMapRWMutex.RUnlock()
	ic.ResponseErrorAndCode("Continuous deployment completed successfully", http.StatusOK)
	return
}
