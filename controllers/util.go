package controllers

//import (
//	"fmt"
//	"dev-flows-api-golang/models"
//	"github.com/golang/glog"
//)

//func GetValidStageBuild(flowId, stageId, stageBuildId string) (models.CiStageBuildLogs, error) {
//	method := "SocketLogController.GetValidStageBuild"
//
//	var build models.CiStageBuildLogs
//
//	stage, err := models.NewCiStage().FindOneById(stageId)
//	if err != nil {
//		glog.Errorf("%s find stage by stageId failed or not exist from database: %v\n", method, err)
//		return build, err
//	}
//	if stage.FlowId != flowId {
//
//		return build, fmt.Errorf("Stage is not %s in the flow", stageId)
//
//	}
//
//	build, err = models.NewCiStageBuildLogs().FindOneById(stageBuildId)
//	if err != nil {
//		glog.Errorf("%s find stagebuild by StageBuildId failed or not exist from database: %v\n", method, err)
//		return build, err
//	}
//
//	if stage.StageId != build.StageId {
//
//		return build, fmt.Errorf("Build is not %s one of the stage", build.BuildId)
//
//	}
//
//	return build, nil
//}

