package models

import (
	"time"
	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
)

type CiStageBuildLogs struct {
	BuildId      string    `orm:"pk;column(build_id)" json:"buildId"`
	FlowBuildId  string `orm:"column(flow_build_id)"`
	StageId      string `orm:"column(stage_id)" json:"stageId"`
	StageName    string `orm:"column(stage_name)" json:"stageName"`
	Status       int8 `orm:"column(status)" json:"status"`
	JobName      string `orm:"column(job_name)"`
	PodName      string `orm:"column(pod_name)"`
	NodeName     string `orm:"column(node_name)"`
	Namespace    string `orm:"column(namespace)"`
	CreationTime time.Time `orm:"column(creation_time)" json:"creationTime"`
	StartTime    time.Time `orm:"column(start_time)" json:"startTime"`
	EndTime      time.Time `orm:"column(end_time)" json:"endTime"`
	BuildAlone   int8 `orm:"column(build_alone)"`
	IsFirst      int8 `orm:"column(is_first)"`
	BranchName   string `orm:"column(branch_name)"`
	IsFetching   bool `orm:"-" json:"isFetching"`
	LogInfo      interface{}  `orm:"-" json:"logInfo"`
}

func (ci *CiStageBuildLogs) TableName() string {

	return "tenx_ci_stage_build_logs"
}

func NewCiStageBuildLogs() *CiStageBuildLogs {
	return &CiStageBuildLogs{}
}

func (ci *CiStageBuildLogs) FindStageBuild(stageId, stageBuildId string) (stagebuild CiStageBuildLogs, err error) {
	o := orm.NewOrm()
	err = o.QueryTable(ci.TableName()).Filter("stage_id", stageId).
		Filter("build_id", stageBuildId).One(&stagebuild)
	return
}

func (ci *CiStageBuildLogs) UpdateById(build CiStageBuildLogs, buildId string, orms ...orm.Ormer) (updateResult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	glog.Infof("UpdateById==============>>build nodeName=%s,buildId=%s\n", build.NodeName, buildId)

	updateResult, err = o.QueryTable(ci.TableName()).
		Filter("build_id", buildId).Update(orm.Params{
		//"flow_build_id":build.FlowBuildId,
		//"stage_id":build.StageId,
		//"stage_name":build.StageName,
		"status":    build.Status,
		"pod_name":  build.PodName,
		"node_name": build.NodeName,
		//"start_time": build.StartTime,
		"end_time": time.Now(),
		//"build_alone": build.BuildAlone,
		//"is_first": build.IsFirst,
		"job_name": build.JobName,
	})
	return
}

func (ci *CiStageBuildLogs) UpdatePodNameAndJobNameByBuildId(build CiStageBuildLogs, buildId string, orms ...orm.Ormer) (updateResult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	glog.Infof("UpdatePodNameAndJobNameByBuildId build NodeName=%s, PodName=%s\n", build.NodeName, build.PodName)
	updateResult, err = o.QueryTable(ci.TableName()).
		Filter("build_id", buildId).Update(orm.Params{
		//"flow_build_id":build.FlowBuildId,
		//"stage_id":build.StageId,
		//"stage_name":build.StageName,
		"status":    build.Status,
		"job_name":  build.JobName,
		"pod_name":  build.PodName,
		"node_name": build.NodeName,
		"namespace": build.Namespace,
		//"start_time": build.StartTime,
		"end_time": time.Now(),
		//"build_alone": build.BuildAlone,
		//"is_first": build.IsFirst,
		//"branch_name": build.BranchName,
	})
	return
}

func (ci *CiStageBuildLogs) UpdateBuildLogById(build CiStageBuildLogs, buildId string, orms ...orm.Ormer) (updateResult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	glog.Infof("UpdateBuildLogById build NodeName=%s\n", build.NodeName)
	updateResult, err = o.QueryTable(ci.TableName()).
		Filter("build_id", buildId).Update(orm.Params{
		//"flow_build_id":build.FlowBuildId,
		//"stage_id":build.StageId,
		//"stage_name":build.StageName,
		"status": build.Status,
		//"job_name":build.JobName,
		"pod_name": build.PodName,
		//"node_name":build.NodeName,
		//"namespace":build.Namespace,
		//"start_time":build.StartTime,
		"end_time": build.EndTime,
		//"build_alone": build.BuildAlone,
		//"is_first": build.IsFirst,
		//"branch_name": build.BranchName,
	})
	return
}

func (ci *CiStageBuildLogs) UpdatePodNameById(podName string, buildId string, orms ...orm.Ormer) (updateResult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	glog.Infof("UpdatePodNameById build podName=%s\n", podName)
	updateResult, err = o.QueryTable(ci.TableName()).
		Filter("build_id", buildId).Update(orm.Params{
		"pod_name": podName,
	})
	return
}

func (ci *CiStageBuildLogs) UpdateStageBuildStatusById(status int, buildId string, orms ...orm.Ormer) (updateResult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	updateResult, err = o.QueryTable(ci.TableName()).
		Filter("build_id", buildId).Update(orm.Params{
		"status": status,
	})
	return
}

func (ci *CiStageBuildLogs) UpdateStageBuildNodeById(nodeName, buildId string, orms ...orm.Ormer) (updateResult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	updateResult, err = o.QueryTable(ci.TableName()).
		Filter("build_id", buildId).Update(orm.Params{
		"node_name": nodeName,
	})
	return
}

func (ci *CiStageBuildLogs) UpdateStageBuildNodeNameAndPodNameById(nodeName, podName, buildId string, orms ...orm.Ormer) (updateResult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	updateResult, err = o.QueryTable(ci.TableName()).
		Filter("build_id", buildId).Update(orm.Params{
		"node_name": nodeName,
		"pod_name":  podName,
	})
	return
}

func (ci *CiStageBuildLogs) DeleteById(buildId string, orms ...orm.Ormer) (result int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	result, err = o.QueryTable(ci.TableName()).Filter("build_id", buildId).
		Delete()
	return
}

func (ci *CiStageBuildLogs) InsertOne(build CiStageBuildLogs, orms ...orm.Ormer) (result int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	result, err = o.Insert(&build)
	return
}

func (ci *CiStageBuildLogs) FindUnfinishedByFlowBuildId(flowBuildId string) (stageBuildLog []CiStageBuildLogs, result int64, err error) {
	o := orm.NewOrm()
	result, err = o.QueryTable(ci.TableName()).
		Filter("flow_build_id", flowBuildId).
		Filter("status__gt", 1).All(&stageBuildLog)
	return
}

func (ci *CiStageBuildLogs) FindOneOfStageByFlowBuildId(flowBuildId, stageId string) (stageBuildLog CiStageBuildLogs, err error) {
	o := orm.NewOrm()
	err = o.QueryTable(ci.TableName()).
		Filter("flow_build_id", flowBuildId).
		Filter("stage_id", stageId).One(&stageBuildLog)
	return
}

func (ci *CiStageBuildLogs) FindOneById(buildId string) (stageBuildLog CiStageBuildLogs, err error) {
	o := orm.NewOrm()
	err = o.QueryTable(ci.TableName()).
		Filter("build_id", buildId).
		One(&stageBuildLog)
	return
}

func (ci *CiStageBuildLogs) FindAllOfStage(stageId string, size int) (stageBuildLog []CiStageBuildLogs, reslut int64, err error) {
	o := orm.NewOrm()
	reslut, err = o.QueryTable(ci.TableName()).
		Filter("stage_id", stageId).OrderBy("-creation_time").Limit(size).
		All(&stageBuildLog)
	return
}

func (ci *CiStageBuildLogs) FindAllOfFlowBuild(flowBuildId string) (stageBuildLog []CiStageBuildLogs, reslut int64, err error) {
	o := orm.NewOrm()
	reslut, err = o.QueryTable(ci.TableName()).
		Filter("flow_build_id", flowBuildId).OrderBy("creation_time").
		All(&stageBuildLog)
	return
}

func (ci *CiStageBuildLogs) FindAllByIdWithStatus(stageId string, status int) (stageBuildLog []CiStageBuildLogs, reslut int64, err error) {
	o := orm.NewOrm()
	reslut, err = o.QueryTable(ci.TableName()).
		Filter("stage_id", stageId).Filter("status", status).OrderBy("creation_time").
		All(&stageBuildLog)
	return
}
