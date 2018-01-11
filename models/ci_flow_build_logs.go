package models

import (
	"time"
	"github.com/astaxie/beego/orm"
	"fmt"
)

type CiFlowBuildLogs struct {
	BuildId      string `orm:"column(build_id);pk" json:"buildId"`
	FlowId       string `orm:"column(flow_id)" json:"flow_id"`
	CreationTime time.Time `orm:"column(creation_time)" json:"creationTime"`
	StartTime    time.Time `orm:"column(start_time)" json:"startTime"`
	EndTime      time.Time `orm:"column(end_time)" json:"endTime"`
	UserId       int32 `orm:"-" json:"user_id"`
	Status       int8  `orm:"column(status)" json:"status"`    //状态。0-成功 1-失败 2-执行中
	Branch       string  `orm:"column(branch)" json:"branch"`  //状态。0-成功 1-失败 2-执行中
	Creater      string  `orm:"column(creater)" json:"creater"` //状态。0-成功 1-失败 2-执行中
}

func (ci *CiFlowBuildLogs) TableName() string {

	return "tenx_ci_flow_build_logs"
}

func NewCiFlowBuildLogs() *CiFlowBuildLogs {
	return &CiFlowBuildLogs{}
}

func (ci *CiFlowBuildLogs) InsertOne(build CiFlowBuildLogs, orms ...orm.Ormer) (result int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	result, err = o.Insert(&build)
	return
}

func (ci *CiFlowBuildLogs) UpdateById(EndTime time.Time, Status int, buildId string, orms ...orm.Ormer) (int64, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	sql := fmt.Sprintf("update %s set end_time=?,status=? where build_id=?", ci.TableName())
	result, err := o.Raw(sql, EndTime, Status, buildId).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (ci *CiFlowBuildLogs) UpdateStartTimeAndStatusById(startTime time.Time, Status int, buildId string, orms ...orm.Ormer) (int64, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	sql := fmt.Sprintf("update %s set start_time=?,status=? where build_id=?", ci.TableName())
	result, err := o.Raw(sql, startTime, Status).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (ci *CiFlowBuildLogs) FindFlowBuild(flowId, buildId string, orms ...orm.Ormer) (cFbl CiFlowBuildLogs, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	err = o.QueryTable(ci.TableName()).Filter("flow_id", flowId).
		Filter("build_id", buildId).One(&cFbl)
	return
}

func (ci *CiFlowBuildLogs) FindOneById(buildId string, orms ...orm.Ormer) (cFbl CiFlowBuildLogs, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	err = o.QueryTable(ci.TableName()).
		Filter("build_id", buildId).One(&cFbl)
	return
}

func (ci *CiFlowBuildLogs) FindAllOfFlow(flowId string, size int, orms ...orm.Ormer) (cFbl []CiFlowBuildLogs, total int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	total, err = o.QueryTable(ci.TableName()).
		Filter("flow_id", flowId).OrderBy("-creation_time").Limit(size).All(&cFbl)
	return
}

type FlowBuildLogResp struct {
	BuildId      string `json:"buildId"`
	Status       int8 `json:"status"`
	CreationTime string `json:"creationTime"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	Namespace    string `json:"namespace"`
	Branch       string `json:"branch"`
}

func FormatBuild(flowBuildLog CiFlowBuildLogs, fieldPrefix string) (flowbuildResp FlowBuildLogResp) {

	flowbuildResp.BuildId = fieldPrefix + flowBuildLog.BuildId
	flowbuildResp.Status = flowBuildLog.Status
	flowbuildResp.CreationTime = fieldPrefix + flowBuildLog.CreationTime.Format("2006-01-02 15:04:05")
	flowbuildResp.StartTime = fieldPrefix + flowBuildLog.StartTime.Format("2006-01-02 15:04:05")
	flowbuildResp.EndTime = fieldPrefix + flowBuildLog.EndTime.Format("2006-01-02 15:04:05")
	flowbuildResp.Branch = fieldPrefix + flowBuildLog.Branch
	flowbuildResp.Namespace = fieldPrefix + flowBuildLog.Creater
	return
}

type LastBuildInfo struct {
	BuildId                string `orm:"column(build_id)" json:"buildId"`
	FlowId                 string `orm:"column(flow_id)" json:"flow_id"`
	CreationTime           time.Time `orm:"column(creation_time)" json:"creationTime"`
	StartTime              time.Time `orm:"column(start_time)" json:"startTime"`
	EndTime                time.Time `orm:"column(end_time)" json:"endTime"`
	Status                 int `orm:"column(status)" json:"status"` //状态。0-成功 1-失败 2-执行中
	StageId                string `orm:"column(stage_id)" json:"stage_id"`
	StageName              string `orm:"column(stage_name)" json:"stage_name"`
	StageBuildStatus       int `orm:"column(stage_build_status)" json:"stage_build_status"`
	StageBuildCreationTime string `orm:"column(stage_build_creation_time)" json:"stage_build_creation_time"`
	StageBuildStartTime    string `orm:"column(stage_build_start_time)" json:"stage_build_start_time"`
	StageBuildEndTime      string `orm:"column(stage_build_end_time)" json:"stage_build_end_time"`
	StageBuildBuildId      string `orm:"column(stage_build_build_id)" json:"stage_build_build_id"`
}

func (ci *CiFlowBuildLogs) FindLastBuildOfFlowWithStages(flowId string, orms ...orm.Ormer) (cFbl []LastBuildInfo, total int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	SELECT_LAST_BUILD_OF_FLOW_WITH_STAGES_BY_FLOW_ID := "SELECT f.*, stage_id, stage_name, s.status as stage_build_status, s.creation_time as stage_build_creation_time, " +
		"s.start_time as stage_build_start_time, s.end_time as stage_build_end_time, s.build_id as stage_build_build_id " +
		"FROM (SELECT * from tenx_ci_flow_build_logs where flow_id = ? ORDER BY creation_time DESC LIMIT 1) as f " +
		"LEFT JOIN tenx_ci_stage_build_logs as s ON f.build_id = s.flow_build_id " +
		"ORDER BY stage_build_creation_time "
	total, err = o.Raw(SELECT_LAST_BUILD_OF_FLOW_WITH_STAGES_BY_FLOW_ID, flowId).QueryRows(&cFbl)
	return
}

type FlowBuildStats struct {
	Status int `orm:"column(status)" json:"buildId"`
	Count  int `orm:"column(count)" json:"flow_id"`
}

//Query failed/running/success flow builds
func (ci *CiFlowBuildLogs) QueryFlowBuildStats(namespace string, orms ...orm.Ormer) (cFbl []FlowBuildStats, total int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	SELECT_SERVER_FLOW_BUILD_STATS := "SELECT status, count(*) as count FROM tenx_ci_flows f, tenx_ci_flow_build_logs b " +
		"where f.flow_id = b.flow_id and f.namespace = ? group by status"
	total, err = o.Raw(SELECT_SERVER_FLOW_BUILD_STATS, namespace).QueryRows(&cFbl)
	return
}
