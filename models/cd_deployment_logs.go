package models

import (
	"github.com/astaxie/beego/orm"
	v1beta1 "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"time"
	"github.com/golang/glog"
	"k8s.io/client-go/1.4/pkg/api/v1"
	"strconv"
	"fmt"
	"strings"
)

type CDDeploymentLogs struct {
	Id            string `orm:"pk;column(id)" json:"-"`
	CdRuleId      string `orm:"column(cd_rule_id)" json:"cd_rule_id"`
	TargetVersion string `orm:"column(target_version)" json:"target_version"`
	Result        string `orm:"column(result)" json:"result"`
	CreateTime    time.Time `orm:"column(create_time)" json:"create_time"`
}

type ListLogs struct {
	App_name         string `orm:"column(app_name)" json:"app_name"`
	Image_name       string `orm:"column(image_name)" json:"image_name"`
	Target_version   string `orm:"column(target_version)" json:"target_version"`
	Cluster_name     string `orm:"column(cluster_name)" json:"cluster_name"`
	Upgrade_strategy int `orm:"column(upgrade_strategy)" json:"upgrade_strategy"`
	Result           string  `orm:"column(result)" json:"result"`
	Create_time      time.Time `orm:"column(create_time)" json:"create_time"`
}

type Result struct {
	Status   int `json:"status"`
	Duration int64 `orm:"column(duration)" json:"duration"`
	Error    string
}

type NewDeploymentArray struct {
	BindingDeploymentId string
	Cluster_id          string
	NewTag              string
	Strategy            int8
	Deployment          *v1beta1.Deployment
	Flow_id             string
	Rule_id             string
	Namespace           string
	Match_tag           string
	Start_time          time.Time
}

func (cd *CDDeploymentLogs) TableName() string {

	return "tenx_cd_deployment_logs"

}

func NewCDDeploymentLogs() *CDDeploymentLogs {
	return &CDDeploymentLogs{}
}
func (cd *CDDeploymentLogs) InsertCDLog(log CDDeploymentLogs,orms ...orm.Ormer) (result int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	result, err = o.Insert(&log)
	return
}

func (cd *CDDeploymentLogs) ListLogsByFlowId(namespace, flow_id string, limit int,orms ...orm.Ormer) (logs []ListLogs, total int64, err error) {
	SELECT_FLOW_DEPLOYMENT_LOGS := "SELECT r.binding_deployment_name as app_name," +
		" r.image_name as image_name, l.target_version, " +
		"c.name as cluster_name, r.upgrade_strategy, l.result, " +
		"l.create_time from tenx_cd_deployment_logs l, tenx_cd_rules r, tenx_clusters c " +
		"where r.namespace = ? and r.flow_id = ? and l.cd_rule_id = r.rule_id " +
		"and r.binding_cluster_id = c.id order by l.create_time desc limit ?"
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	total, err = o.Raw(SELECT_FLOW_DEPLOYMENT_LOGS, namespace, flow_id, limit).QueryRows(&logs)
	return
}

func Upgrade(deployment *v1beta1.Deployment, imageName, newTag string, isMatchTag string, strategy int8) bool {
	method := "kubernetes Upgrade"
	matched := false
	ifUpgrade := false
	now := time.Now()
	if deployment.Kind != "Deployment" {
		return ifUpgrade
	}

	if deployment.Spec.Template.ObjectMeta.Labels["tenxcloud.com/cdTimestamp"] != "" {
		cooldownSec := 30
		lastCdTs := deployment.Spec.Template.ObjectMeta.Labels["tenxcloud.com/cdTimestamp"]
		cdTs, _ := strconv.ParseInt(lastCdTs, 10, 64)
		if (time.Now().Unix() - cdTs) < int64(cooldownSec * 1000) {
			glog.Warningf("%s %s\n", method, "Upgrade is rejected because the" +
				" deployment was updated too frequently")
			return ifUpgrade
		}
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		oldImage := parseImageName(container.Image)
		if oldImage.Image == imageName {
			if (isMatchTag == "2") || (isMatchTag == "1"&&newTag == oldImage.Tag) {
				matched = true
				container.Image = oldImage.Host + "/" + oldImage.Image + ":" + newTag
				container.ImagePullPolicy = v1.PullAlways
			}
		}

	}
	volumes := deployment.Spec.Template.Spec.Volumes
	if len(volumes) > 0&&strategy != 1 {
		for _, volume := range volumes {
			if volume.RBD != nil {
				//如果挂载了存储卷，则强制使用重启策略
				strategy = 1
				break
			}
		}
	}

	if strategy == 2 && (deployment.Spec.Strategy.Type != v1beta1.RollingUpdateDeploymentStrategyType ||
		deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.IntVal != 0) {
		deployment.Spec.Strategy.Type = v1beta1.RollingUpdateDeploymentStrategyType
		deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.IntVal = 0

	} else {
		deployment.Spec.Strategy.Type = v1beta1.RecreateDeploymentStrategyType
	}

	if matched {
		deployment.Spec.Template.ObjectMeta.Labels["tenxcloud.com/cdTimestamp"] = fmt.Sprintf("%s", now)
		ifUpgrade = true
		return ifUpgrade
	}
	return ifUpgrade

}

type Image struct {
	Host  string
	Image string
	Tag   string
}

//gcr.io/google_containers/example-dns-backend:v1
//ubuntu:v1
func parseImageName(imageFullName string) (image Image) {
	//var host,image,tag,letter string
	//var separatorNumber int
	count := strings.Count(imageFullName, "/")
	exist := strings.Count(imageFullName, ":")
	if count == 2 &&exist == 1 {
		res := strings.Split(imageFullName, "/")
		image.Host = res[0] + "/" + res[1]
		image.Image = strings.Split(res[2], ":")[0]
		image.Tag = strings.Split(res[2], ":")[1]
	} else if count == 1&&exist == 1 {
		res := strings.Split(imageFullName, "/")
		image.Host = ""
		image.Image = strings.Split(res[1], ":")[0]
		image.Tag = strings.Split(res[1], ":")[1]
	} else if count == 0&&exist == 1 {

		res := strings.Split(imageFullName, ":")
		image.Host = ""
		image.Image = res[0]
		image.Tag = res[1]

	} else if count == 0&&exist == 0 {
		image.Host = ""
		image.Image = imageFullName
		image.Tag = "latest"
	}

	return image
}
