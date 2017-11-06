package models

import (
	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
	"time"
)

type CiDockerfile struct {
	FlowId     string `orm:"column(flow_id)" json:"flow_id"`
	StageId    string `orm:"pk;column(stage_id)" json:"stage_id"`
	Namespace  string `orm:"column(namespace)" json:"namespace"`
	Content    string `orm:"size(2048);column(content)" json:"content"`
	ModifiedBy string `orm:"column(modified_by)" json:"modified_by"`
	CreateTime string `orm:"column(create_time)" json:"create_time"`
	UpdateTime string `orm:"column(update_time)" json:"update_time"`
	Type       int8 `orm:"column(type)" json:"type"` //0: text\n1: visual(可视化)
}

type Dockerfiles struct {
	Flow_id     string `orm:"column(flow_id);pk" json:"flow_id"`
	Stage_id    string `orm:"column(stage_id)" json:"stage_id"`
	Type        int `orm:"column(type)" json:"type"`
	Stage_name  string `orm:"column(stage_name)" json:"stage_name"`
	Name        string `orm:"column(name)" json:"name"`
	Create_time string `orm:"column(create_time)" json:"create_time"`
	Update_time string `orm:"column(update_time)" json:"update_time"`
}

func NewCiDockerfile() *CiDockerfile {
	return &CiDockerfile{}
}
func (cd *CiDockerfile) TableName() string {

	return "tenx_ci_dockerfiles"
}

func (cd *CiDockerfile) AddDockerfile(file CiDockerfile,orms ...orm.Ormer) (total int64, err error) {

	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	total, err = o.Insert(&file)
	if err != nil {
		glog.Errorf(" AddDockerfile failed %v\n", err)
		return
	}
	return
}

func (cd *CiDockerfile) ListDockerfiles(namespace string,orms ...orm.Ormer) (cidockerfile []Dockerfiles, total int64, err error) {
	SELECT_CI_DOCKERFILES := "SELECT "
	SELECT_CI_DOCKERFILES += "d.flow_id as flow_id, d.stage_id as stage_id,"
	SELECT_CI_DOCKERFILES += "d.type as type, s.stage_name as stage_name,"
	SELECT_CI_DOCKERFILES += "f.name as name, d.create_time as create_time,"
	SELECT_CI_DOCKERFILES += "d.update_time as update_time FROM tenx_ci_flows f, tenx_ci_stages s, tenx_ci_dockerfiles d "
	SELECT_CI_DOCKERFILES += "where f.namespace = ?  "
	SELECT_CI_DOCKERFILES += "and f.flow_id = d.flow_id "
	SELECT_CI_DOCKERFILES += "and s.stage_id = d.stage_id "
	SELECT_CI_DOCKERFILES += "order by d.update_time desc"

	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	total, err = o.Raw(SELECT_CI_DOCKERFILES, namespace).QueryRows(&cidockerfile)
	if err != nil {
		glog.Errorf("get ListDockerfiles failed %v\n", err)

		return
	}

	return
}

func (cd *CiDockerfile) GetDockerfile(namespace, flow_id, stage_id string,orms ...orm.Ormer) (cidockerfile CiDockerfile, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	glog.Infof("namespace=%s,flow_id=%s,stage_id=%s\n",namespace,flow_id,stage_id)
	err = o.QueryTable(cd.TableName()).Filter("namespace", namespace).
		Filter("flow_id", flow_id).
		Filter("stage_id", stage_id).One(&cidockerfile)
	if err != nil {
		glog.Errorf("CiDockerfile GetDockerfile failed %v\n", err)
		return
	}

	return
}

func (cd *CiDockerfile) GetAllByFlowId(namespace, flow_id string,orms ...orm.Ormer) (cidockerfile []Dockerfiles, total int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	total, err = o.QueryTable(cd.TableName()).Filter("namespace", namespace).
		Filter("flow_id", flow_id).All(&cidockerfile)
	if err != nil {
		glog.Errorf("CiDockerfile GetAllByFlowId failed %v\n", err)
		return
	}

	return
}

func (cd *CiDockerfile) UpdateDockerfile(namespace, flow_id, stage_id, modified_by string, ciDockerfile CiDockerfile,orms ...orm.Ormer) (total int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	total, err = o.QueryTable(cd.TableName()).Filter("namespace", namespace).
		Filter("stage_id", stage_id).Update(orm.Params{
		"content":     ciDockerfile.Content,
		"modified_by": modified_by,
		"update_time": time.Now().Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		glog.Errorf("update dockerfile failed:%v\n", err)
		return
	}

	return
}
func (cd *CiDockerfile) RemoveByFlowId(namespace string, flow_id string,orms ...orm.Ormer) (total int64, err error) {

	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	if cd.CheckDockerfileExist(namespace, flow_id) {
		total, err = o.QueryTable(cd.TableName()).Filter("namespace", namespace).
			Filter("flow_id", flow_id).Delete()
		if err != nil {
			glog.Errorf("CiDockerfile.RemoveByFlowId failed  %v\n", err)
			return
		}

	}
	glog.Infof("no dockerfile in this flowid=%s\n", flow_id)
	return
}

func (cd *CiDockerfile) CheckDockerfileExist(namespace string, flow_id string,orms ...orm.Ormer) bool {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	return o.QueryTable(cd.TableName()).Filter("namespace", namespace).
		Filter("flow_id", flow_id).Exist()
}

func (cd *CiDockerfile) RemoveDockerfile(namespace, flow_id, stage_id string,orms ...orm.Ormer) (total int64, err error) {

	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	total, err = o.QueryTable(cd.TableName()).Filter("namespace", namespace).
		Filter("flow_id", flow_id).Filter("stage_id", stage_id).Delete()
	if err != nil {
		glog.Errorf("CiDockerfile.RemoveDockerfile failed  %v\n", err)
		return
	}

	return
}
