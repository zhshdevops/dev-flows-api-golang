package controllers

import (
	"encoding/json"
	"github.com/golang/glog"
	"dev-flows-api-golang/models"
	"fmt"
	"net/http"
	"regexp"
)

type CiStageLinksController struct {
	BaseController
}

//@router /:target_id [PUT]
func (link *CiStageLinksController) UpdateLinkDirs() {

	method := "CiStageLinksController.UpdateLinkDirs"
	flowId := link.Ctx.Input.Param(":flow_id")
	stageId := link.Ctx.Input.Param(":stage_id")
	targetId := link.Ctx.Input.Param(":target_id")
	if string(link.Ctx.Input.RequestBody) == "" {
		glog.Errorf("%s request body is empty %s\n", method, "No link specified")
		link.ResponseErrorAndCode("request body is empty,No link specified", http.StatusBadRequest)
		return
	}

	var linkDir models.CiStageLinks
	err := json.Unmarshal(link.Ctx.Input.RequestBody, &linkDir)
	if err != nil {
		glog.Errorf("%s json unmarshal failed:%v\n", method, err)
		link.ResponseErrorAndCode("json 解析失败", http.StatusForbidden)
		return
	}

	glog.Infof("%s request body linkDir=:%#v\n", method, linkDir)
	//校验参数
	if linkDir.Enabled == 1 && (linkDir.TargetDir == "" || linkDir.SourceDir == "") {
		glog.Errorf("%s Can not enable link without source or target directory\n", method)
		link.ResponseErrorAndCode("Can not enable link without source or target directory", http.StatusForbidden)
		return
	}
	if linkDir.Enabled == 1 {

		if regexp.MustCompile(`^/`).FindString(linkDir.SourceDir) == "" || regexp.MustCompile(`^/`).FindString(linkDir.TargetDir) == "" {
			glog.Errorf("%s Absolute path should be specified\n", method)
			link.ResponseErrorAndCode("Absolute path should be specified", http.StatusForbidden)
			return
		}
	}
	//查询一个flow下的所有的Statge
	ids := fmt.Sprintf(`'%s','%s'`, stageId, targetId)

	stages, total, err := models.NewCiStage().FindByIds(flowId, ids)
	if err != nil {
		glog.Errorf("%s find stage failed by id from database:%v\n", method, err)
		link.ResponseErrorAndCode("select stage data failed ", http.StatusMovedPermanently)
		return
	}

	if total < 2 {
		link.ResponseErrorAndCode("Associated stages cannot be found ", http.StatusNotFound)
		return
	}

	var sourceName, targetName, targetDir string

	for _, s := range stages {

		if s.StageId == stageId {

			sourceName = s.StageName
		}

		if s.StageId == targetId {
			targetName = s.StageName
		}
	}

	glog.Infof("sourceName=%s,targetName=%s\n", sourceName, targetName)

	oldStageLinks, result, err := models.NewCiStageLinks().GetAllLinksOfStage(flowId, stageId)
	if result < 1 || err != nil {
		glog.Errorf("%s GetAllLinksOfStage from databse failed: err=%v ,result=%d\n", method, err, result)
		link.ResponseErrorAndCode("No link of the stage", http.StatusNotFound)
		return
	}

	var exist bool = false
	var linkRec models.CiStageLinks
	linkRec.FlowId = flowId
	linkRec.SourceId = stageId
	linkRec.TargetId = targetId

	for _, s := range oldStageLinks {
		if stageId == s.SourceId {
			exist = true
			if targetId != s.TargetId {
				link.ResponseErrorAndCode("Stage does not link to the target", http.StatusNotFound)
				return
			}
			nextLink, err := models.NewCiStageLinks().GetOneBySourceId(targetId)
			if err != nil {
				glog.Errorf("%s get next link faile from database:err=%v\n", method, err)
				link.ResponseErrorAndCode("Stage does not nextLink", http.StatusNotFound)
				return
			}
			// 设置enabled默认值
			if linkDir.Enabled == 0 {
				linkDir.Enabled = 0
			}

			// 设置更新字段
			if s.Enabled != linkDir.Enabled {
				linkRec.Enabled = linkDir.Enabled
			}

			if linkDir.SourceDir != "" && s.SourceDir != linkDir.SourceDir {
				linkRec.SourceDir = linkDir.SourceDir
			}

			if linkDir.TargetDir != "" && s.TargetDir != linkDir.TargetDir {
				//判断是否存在
				if nextLink.SourceId != "" && (nextLink.SourceDir != "" && linkDir.TargetDir == nextLink.SourceDir) {
					glog.Errorf("%s Target directory can not be set because it was used as source of next link\n", method)
					link.ResponseErrorAndCode("Target directory can not be set because it was used as source of next link", http.StatusForbidden)
					return
				}
				linkRec.TargetDir = linkDir.TargetDir

			}

		} else {
			// sourceId === oldLinks[i].target_id
			targetDir = s.TargetDir
		}

	}

	if !exist {
		link.ResponseErrorAndCode("No link of the stage", http.StatusNotFound)
		return
	}

	if linkRec.SourceDir != "" && targetDir == linkRec.SourceDir {
		link.ResponseErrorAndCode("Source directory can not be set because it was used as target of previos link", http.StatusForbidden)
		return
	}
	linkRec.Enabled = linkDir.Enabled
	linkRec.SourceDir = linkDir.SourceDir
	linkRec.TargetDir = linkDir.TargetDir
	glog.Infof("linkRec info:%#v\n", linkRec)
	if linkRec.SourceId != "" {
		_, err := models.NewCiStageLinks().UpdateOneBySrcId(linkRec, stageId)
		if err != nil {
			glog.Errorf("%s update link dir failed:%v\n", method, err)
			link.ResponseErrorAndCode("update failed", http.StatusForbidden)
			return
		}
	}

	link.ResponseErrorAndCode("success", http.StatusOK)
	return

}
