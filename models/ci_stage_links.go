package models

import (
	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
	"dev-flows-api-golang/models/common"
	"dev-flows-api-golang/util/rand"
	"net/http"
	"fmt"
	"database/sql"
)

type CiStageLinks struct {
	SourceId  string `orm:"pk;column(source_id)" json:"source_id"`
	TargetId  string `orm:"column(target_id)" json:"target_id"`
	FlowId    string `orm:"column(flow_id)" json:"flow_id"`
	SourceDir string `orm:"column(source_dir)" json:"sourceDir"`
	TargetDir string `orm:"column(target_dir)" json:"targetDir"`
	Enabled   int8 `orm:"column(enabled)" json:"enabled"`
}

func (ci *CiStageLinks) TableName() string {
	return "tenx_ci_stage_links"
}

func NewCiStageLinks() *CiStageLinks {
	return &CiStageLinks{}
}

func (ci *CiStageLinks) InsertOneLink(cisrageLink CiStageLinks, orms ...orm.Ormer) (reult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	sql := fmt.Sprintf("INSERT INTO %s (flow_id, source_id) VALUES (?,?);", ci.TableName())
	raw, err := o.Raw(sql, cisrageLink.FlowId, cisrageLink.SourceId).Exec()

	reult, err = raw.RowsAffected()
	//reult, err = o.InsertOrUpdate(&cisrageLink)

	return
}

func (ci *CiStageLinks) InsertLink(cisrageLink CiStageLinks, orms ...orm.Ormer) (err error) {
	var o orm.Ormer

	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	sql := fmt.Sprintf("INSERT INTO %s (flow_id, source_id,target_id,source_dir,target_dir,enabled) VALUES (?,?,?,?,?,?);", ci.TableName())

	_, err = o.Raw(sql, cisrageLink.FlowId, cisrageLink.SourceId, cisrageLink.TargetId,
		cisrageLink.SourceDir, cisrageLink.TargetDir, cisrageLink.Enabled).Exec()

	return err

}

func (ci *CiStageLinks) Insert(cisrageLink CiStageLinks, orms ...orm.Ormer) (err error) {
	var o orm.Ormer

	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	sql := fmt.Sprintf("INSERT INTO %s (flow_id, source_id,source_dir,target_dir,enabled) VALUES (?,?,?,?,?);", ci.TableName())

	_, err = o.Raw(sql, cisrageLink.FlowId, cisrageLink.SourceId,
		cisrageLink.SourceDir, cisrageLink.TargetDir, cisrageLink.Enabled).Exec()

	return err

}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func (ci *CiStageLinks) UpdateOneBySrcId(link CiStageLinks, srcId string, orms ...orm.Ormer) (reult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	reult, err = o.QueryTable(ci.TableName()).
		Filter("source_id", srcId).Update(orm.Params{
		"target_id":  link.TargetId,
		"flow_id":    link.FlowId,
		"source_dir": link.SourceDir,
		"target_dir": link.TargetDir,
		"enabled":    link.Enabled,
	})
	return
}

func (ci *CiStageLinks) UpdateOneBySrcIdNew(link CiStageLinks, srcId string, orms ...orm.Ormer) (reult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	reult, err = o.QueryTable(ci.TableName()).
		Filter("source_id", srcId).Update(orm.Params{
		"target_id": link.TargetId,
	})
	return
}

func (ci *CiStageLinks) UpdateOneByTargetId(targetId, srcId string, orms ...orm.Ormer) (reult int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	reult, err = o.QueryTable(ci.TableName()).
		Filter("source_id", srcId).Update(orm.Params{
		"target_id": targetId,
	})
	return
}
func (ci *CiStageLinks) GetAllLinksOfStage(flowId, stageId string) (link []CiStageLinks, reult int64, err error) {
	o := orm.NewOrm()
	cond := orm.NewCondition()
	cond.Or("source_id", stageId).Or("target_id", stageId)
	reult, err = o.QueryTable(ci.TableName()).
		Filter("flow_id", flowId).SetCond(cond).All(&link)
	return
}

func (ci *CiStageLinks) GetNilTargets(flowId string) (link []CiStageLinks, reult int64, err error) {
	o := orm.NewOrm()
	cond := orm.NewCondition()
	cond1 := cond.And("target_id__isnull", true)
	reult, err = o.QueryTable(ci.TableName()).SetCond(cond1).
		Filter("flow_id", flowId).All(&link)
	return
}

func (ci *CiStageLinks) GetOneBySourceId(source_id string) (link CiStageLinks, err error) {
	o := orm.NewOrm()
	err = o.QueryTable(ci.TableName()).
		Filter("source_id", source_id).One(&link)
	return
}

type Setting struct {
	Type          string
	ContainerPath string
	VolumePath    string
	Name          string
}

type GetSettingResp struct {
	Message string `json:"message,omitempty"`
}

func GetVolumeSetting(flowId, stageId, flowBuildId, stageBuildId string) ([]Setting, GetSettingResp, int) {
	var settings []Setting
	settings = make([]Setting, 0)
	var setting Setting
	var getSettingResp GetSettingResp
	method := "getVolumeSetting"
	links, result, err := NewCiStageLinks().GetAllLinksOfStage(flowId, stageId)
	glog.Infof("links======len==>%d\n", len(links))
	if err != nil || result < 1 {
		glog.Errorf("%s get volume failed from database:%v\n", method, err)
		getSettingResp.Message = "No link exists of the stage"
		return settings, getSettingResp, http.StatusInternalServerError
	}
	for _, link := range links {
		if stageId == link.SourceId && link.Enabled == 1 && link.TargetDir != "" &&
			link.TargetId != "" && link.SourceDir != "" {
			setting.Type = "source"

			setting.ContainerPath = link.SourceDir

			setting.VolumePath = common.BUILD_DIR + rand.FormatTime("20060102") + "/" + flowId + link.SourceDir

			settings = append(settings, setting)
		} else if stageId == link.TargetId && link.Enabled == 1 && link.TargetDir != "" &&
			link.SourceDir != "" {
			glog.Info("GetVolumeSetting==========flowBuildId=%s,SourceId=%s\n", flowBuildId, link.SourceId)
			//获取上一步stage对应的build
			stagebuildLog, err := NewCiStageBuildLogs().FindOneOfStageByFlowBuildId(flowBuildId,
				link.SourceId)
			if err != nil || stagebuildLog.Namespace == "" {
				glog.Errorf("%s %v", method, err)
				getSettingResp.Message = "Cannot find build of source stage which should be built before"
				return settings, getSettingResp, http.StatusNotFound
			}

			setting.Type = "target"

			setting.ContainerPath = link.TargetDir

			setting.VolumePath = common.BUILD_DIR + rand.FormatTime("20060102") + "/" + flowId + link.SourceDir

			settings = append(settings, setting)

		}

	}

	return settings, getSettingResp, http.StatusOK
}
