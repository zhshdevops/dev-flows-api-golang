/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-10-19  @author liuyang
 */

package models

import (
	//"encoding/json"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
	// shortid "api-server/modules/tenx/id"

	"dev-flows-api-golang/models/cluster"
)

// AuditInfo audit information
type AuditInfo struct {
	// if skip is true, this audit info will not be recorded to db
	Skip bool

	// if UpdateRecord is false, insert new record to db
	UpdateRecord bool

	// request
	StartTime time.Time

	// id
	ID          string
	Method      string
	URL         string
	Operator    string
	Namespace   string
	RequestBody string

	// for controller functions
	operationType  AuditOperation
	resourceType   AuditResource
	resourceID     string
	resourceName   string
	clusterID      string
	customizedName bool

	// response
	HTTPStatusCode int
	ResponseBody   string
	Duration       int
}

// AuditOperation user operation type, create read update delete
type AuditOperation uint

// create get list update delete
// get operation number:
// i=0 && while IFS= read -r line; do echo $line $i; let i++; done <<< $(grep "\sAuditOperation\w\+" audit.go | awk '{print $1}' | sed 's/AuditOperation//')
const (
	AuditOperationUnknown        AuditOperation = iota
	AuditOperationCreate          // 1
	AuditOperationGet
	AuditOperationList
	AuditOperationUpdate
	AuditOperationDelete
	AuditOperationStart
	AuditOperationStop
	AuditOperationRestart
	AuditOperationPause
	AuditOperationResume          // 10
	AuditOperationBatchDelete
	AuditOperationBatchStart
	AuditOperationBatchStop
	AuditOperationBatchRestart
	AuditOperationQuickRestart
	AuditOperationCheckExist
	AuditOperationFormat
	AuditOperationExpand          // 18
	AuditOperationBatchIgnore
	AuditOperationEnablEmail
	AuditOperationDisablEmail
	AuditOperationCreateOrUpdate
	AuditOperationToggleEnable    // 20
	AuditOperationIgnore
	AuditOperationRollBack
	AuditOperationClone
)

func (ao *AuditOperation) String() string {
	ops := []string{
		"Unknown",
		"Create", // 1
		"Get",
		"List",
		"Update",
		"Delete",
		"Start",
		"Stop",
		"Restart",
		"Pause",
		"Resume", // 10
		"BatchDelete",
		"BatchStart",
		"BatchStop",
		"BatchRestart",
		"QuickRestart",
		"CheckExist",
		"Format",
		"Expand", // 18
		"BatchIgnore",
		"EnablEmail", //20
		"DisablEmail",
		"CreateOrUpdate",
		"ToggleEnable",
		"Ignore",
		"RollBack",
		"Clone",
	}

	if int(*ao) >= len(ops) {
		return ops[AuditOperationUnknown]
	}

	return ops[*ao]
}

// AuditResource resources being operated, instace service app ...
type AuditResource uint

// instance service app ...
// get resource number:
// i=0 && while IFS= read -r line; do echo $line $i; let i++; done <<< $(grep "\sAuditResource\w\+" audit.go | awk '{print $1}' | sed 's/AuditResource//')
const (
	AuditResourceUnknown                  AuditResource = iota
	AuditResourceInstance                  // 1
	AuditResourceInstanceEvent
	AuditResourceInstanceLog
	AuditResourceInstanceMetrics
	AuditResourceInstanceContainerMetrics
	AuditResourceService
	AuditResourceServiceInstance
	AuditResourceServiceEvent
	AuditResourceServiceLog
	AuditResourceServiceK8sService         // 10
	AuditResourceServiceRollingUpgrade
	AuditResourceServiceManualScale
	AuditResourceServiceAutoScale
	AuditResourceServiceQuota
	AuditResourceServiceHaOption
	AuditResourceServiceDomain
	AuditResourceApp
	AuditResourceAppService                // app's service
	AuditResourceAppOperationLog
	AuditResourceAppExtraInfo              // icon etc // 20
	AuditResourceAppTopology
	AuditResourceConfigGroup
	AuditResourceConfig
	AuditResourceNode
	AuditResourceNodeMetrics
	AuditResourceThirdPartyRegistry
	AuditResourceVolume
	AuditResourceVolumeConsumption         // 28

	// user
	AuditResourceUser
	AuditResourceUserTeams   // 30
	AuditResourceUserSpaces

	// team
	AuditResourceTeam
	AuditResourceTeamUsers
	AuditResourceTeamSpaces

	// cluster
	AuditResourceCluster

	// ci
	AuditResourceRepos            // 36
	AuditResourceProjects
	AuditResourceFlows
	AuditResourceStages
	AuditResourceLinks            // 40
	AuditResourceBuilds
	AuditResourceCIRules
	AuditResourceCDRules
	AuditResourceCIDockerfiles
	AuditResourceCINotifications
	AuditResourceCDNotifications  // 46

	// instanceexport
	AuditResourceInstanceExport  //47
	// alert
	AuditResourceAlertEmailGroup
	AuditResourceAlertRecord
	AuditResourceAlertStrategy    //50
	AuditResourceAlertRule
	//snapshot
	AuditResourceSnapshot

	// labels
	AuditResourceLabels

	AuditResourceInstanceDelete  //47
	AuditResourceOnlineScript    //47

	AuditResourceCIImages = 1000
)

func (ar *AuditResource) String() string {
	rs := []string{
		"Unknown",
		"Instance", // 1
		"InstanceEvent",
		"InstanceLog",
		"InstanceMetrics",
		"InstanceContainerMetrics",
		"Service",
		"ServiceInstance",
		"ServiceEvent",
		"ServiceLog",
		"ServiceK8sService", // 10
		"ServiceRollingUpgrade",
		"ServiceManualScale",
		"ServiceAutoScale",
		"ServiceQuota",
		"ServiceHaOption",
		"ServiceDomain",
		"App",
		"AppService",
		"AppOperationLog",
		"AppExtraInfo", // 20
		"AppTopology",
		"ConfigGroup",
		"Config",
		"Node",
		"NodeMetrics",
		"ThirdPartyRegistry",
		"Volume",
		"VolumeConsumption", // 28
		"User",
		"UserTeams", // 30
		"UserSpaces",
		"Team",
		"TeamUsers",
		"TeamSpaces",
		"Cluster",
		"Repos",
		"Projects",
		"Flows",
		"Stages",
		"Links", // 40
		"Builds",
		"CIRules",
		"CDRules",
		"CIDockerfiles",
		"CINotifications",
		"CDNotifications", // 46
		"InstanceExport",
		"AlertEmailGroup",
		"AlertRecord",
		"AlertStrategy", //50
		"AlertRule",
		"Snapshot",
		"Labels",
	}

	if int(*ar) >= len(rs) {
		return rs[AuditResourceUnknown]
	}

	return rs[*ar]
}

// SetOperationType set ClusterID member
func (a *AuditInfo) SetOperationType(t AuditOperation) *AuditInfo {
	a.operationType = t
	return a
}

// SetResourceType set ResourceType member
func (a *AuditInfo) SetResourceType(t AuditResource) *AuditInfo {
	a.resourceType = t
	return a
}

// SetResourceID set ResourceID member
func (a *AuditInfo) SetResourceID(r string) *AuditInfo {
	a.resourceID = r
	return a
}

// SetResourceName set ResourceName member
func (a *AuditInfo) SetResourceName(t string) *AuditInfo {
	a.resourceName = t
	return a
}

func (a *AuditInfo) CustomizeResourceName(n string) *AuditInfo {
	a.resourceName = n
	a.customizedName = true
	return a
}

func (to *AuditInfo) CopyCustomizedNameFrom(from *AuditInfo) *AuditInfo {
	if from.customizedName {
		to.resourceName = from.resourceName
		to.customizedName = from.customizedName
	}
	return to
}

// SetClusterID set OperationType member
func (a *AuditInfo) SetClusterID(id string) *AuditInfo {
	a.clusterID = id
	return a
}

func (a *AuditInfo) SetDuration(du int) *AuditInfo {
	a.Duration = du
	return a
}

func (a *AuditInfo) SetHttpCode(code int) *AuditInfo {
	a.HTTPStatusCode = code
	return a
}

// DoWork work routine for workpool
func (a *AuditInfo) DoWork(workRoutine int) {
	method := "controllers.auditRoutine " + strconv.Itoa(workRoutine)
	if a.Skip {
		return
	}

	if a.UpdateRecord {
		err := UpdateAuditRecord(a).Update()
		if err != nil {
			glog.Errorln(method, "failed to update audit record", a.ID, err)
		}
	} else {
		err := NewAuditRecord(a).Insert()
		if err != nil {
			glog.Errorln(method, "failed to insert audit record", err)
		}
	}
}

// CREATE TABLE `tenx_audit` (
//   `id` char(18) NOT NULL,
//   `namespace` varchar(45) NOT NULL,
//   `cluster_id` varchar(45) DEFAULT NULL,
//   `operation_type` smallint(6) DEFAULT NULL COMMENT '0=Unknown\n1=Create\n2=Get\n3=List\n4=Update\n5=Delete',
//   `resource_type` smallint(6) DEFAULT NULL COMMENT '0=Unknown\n1=Instance\n2=Service\n3=App\n4=ThirdPartyRegistry',
//   `resource_id` varchar(45) DEFAULT NULL,
//   `resource_name` text,
//   `resource_config` text COMMENT 'http request body',
//   `time` datetime NOT NULL COMMENT 'request start time',
//   `duration` int(11) DEFAULT NULL COMMENT 'request duration',
//   `status` smallint(6) DEFAULT NULL COMMENT 'http status code',
//   `remark` varchar(45) DEFAULT NULL COMMENT 'additional status for this operation',
//   `url` text,
//   `http_method` varchar(10) DEFAULT NULL,
//   `operator` varchar(45) DEFAULT NULL,
//   PRIMARY KEY (`id`),
//   UNIQUE KEY `id_UNIQUE` (`id`)
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8

// AuditRecord tenx_audit table
type AuditRecord struct {
	ID             string         `orm:"pk;column(id)" json:"id"`
	URL            string         `orm:"column(url)" json:"-"`
	HTTPMethod     string         `orm:"column(http_method)" json:"-"`
	Namespace      string         `orm:"column(namespace)" json:"namespace"`
	Operator       string         `orm:"column(operator)" json:"operator"`
	ClusterID      string         `orm:"column(cluster_id)" json:"cluster_id"`
	ClusterName    string         `orm:"-" json:"cluster_name"`
	OperationType  AuditOperation `orm:"column(operation_type)" json:"operation_type"`
	ResourceType   AuditResource  `orm:"column(resource_type)" json:"resource_type"`
	ResourceID     string         `orm:"column(resource_id)" json:"resource_id"`
	ResourceName   string         `orm:"column(resource_name)" json:"resource_name"`
	ResourceConfig string         `orm:"column(resource_config)" json:"resource_config"`
	Time           time.Time      `orm:"column(time)" json:"time"`
	Duration       int            `orm:"column(duration)" json:"duration"`
	Status         int            `orm:"column(status)" json:"status"`
	Remark         string         `orm:"column(remark)" json:"-"`
}

// TableName return tenx_audit
func (a *AuditRecord) TableName() string {
	return "tenx_audit"
}

// NewAuditRecord create audit record by http request info
func NewAuditRecord(a *AuditInfo) *AuditRecord {
	//method := "models.NewAuditRecord"
	r := &AuditRecord{}

	r.ID = a.ID
	r.URL = a.URL
	r.HTTPMethod = a.Method
	r.Namespace = a.Namespace
	r.ClusterID = a.clusterID
	r.OperationType = a.operationType
	r.ResourceType = a.resourceType
	r.ResourceID = a.resourceID
	r.ResourceName = a.resourceName
	r.ResourceConfig = a.RequestBody
	r.Status = a.HTTPStatusCode
	r.Time = a.StartTime
	r.Duration = a.Duration
	r.Operator = a.Operator

	//if r.ResourceName == "" && r.ResourceID != "" {
	//	// store resource id in resource name to avoid or operation
	//	r.ResourceName = r.ResourceID
	//}
	//
	//// extract resource name from request body
	//type name struct {
	//	Name     string `json:"name"`
	//	UserName string `json:"userName"` // models/user/usercreate.go UserSpec
	//	TeamName string `json:"teamName"` // models/team/teamlist.go Team
	//}
	//if r.ResourceName == "" && r.ResourceConfig != "" {
	//	var n name
	//	err := json.Unmarshal([]byte(r.ResourceConfig), &n)
	//	if err != nil {
	//		glog.Errorln(method, "failed to extract name from request body", r.ResourceConfig, err)
	//	} else {
	//		if n.Name != "" {
	//			r.ResourceName = n.Name
	//		} else if n.UserName != "" {
	//			r.ResourceName = n.UserName
	//		} else if n.TeamName != "" {
	//			r.ResourceName = n.TeamName
	//		}
	//	}
	//}

	return r
}

// UpdateAuditRecord return audit record for update
func UpdateAuditRecord(a *AuditInfo) *AuditRecord {
	r := &AuditRecord{}

	r.ID = a.ID
	r.Status = a.HTTPStatusCode
	r.Duration = a.Duration
	if a.customizedName {
		r.ResourceName = a.resourceName
	}
	return r
}

// Insert insert audit record
func (a *AuditRecord) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(a)
	return err
}

// Update update audit record support status and druation for now
func (a *AuditRecord) Update() error {
	o := orm.NewOrm()
	var doUpdate = false

	params := make(orm.Params)
	if a.Duration != 0 {
		params["duration"] = a.Duration
		doUpdate = true
	}
	if a.Status != 0 {
		params["status"] = a.Status
		doUpdate = true
	}
	if a.ResourceName != "" {
		params["resource_name"] = a.ResourceName
		doUpdate = true
	}

	if doUpdate {
		_, err := o.QueryTable(a.TableName()).Filter("id", a.ID).Update(params)
		return err
	}
	return nil
}

// List list audit record by namespace, operation type, resource type,
func (a *AuditRecord) List(from, size int, start, end time.Time, status string) (*[]AuditRecord, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(a.TableName())
	if a.Namespace != "" {
		qs = qs.Filter("namespace", a.Namespace)
	}
	if a.OperationType != 0 {
		qs = qs.Filter("operation_type", a.OperationType)
	}
	if a.ResourceType != 0 {
		qs = qs.Filter("resource_type", a.ResourceType)
	}

	if !start.IsZero() {
		qs = qs.Filter("time__gte", start)
	}
	if !end.IsZero() {
		qs = qs.Filter("time__lte", end)
	}

	if status == "failed" { // failed
		qs = qs.Exclude("status", 0).Exclude("status", 200)
	} else if status == "running" { // running
		qs = qs.Filter("status", 0)
	} else if status == "success" { // success
		qs = qs.Filter("status", 200)
	}

	var records []AuditRecord
	_, err := qs.Offset(from).Limit(size).All(&records)
	return &records, err
}

// ListWithCount list audit record by namespace, operation type, resource type,
func (a *AuditRecord) ListWithCount(from, size int, start, end time.Time, status, keyword string, omitempty bool) (int64, *[]AuditRecord, error) {
	clustermodel := cluster.ClusterModel{}
	clusters, err := clustermodel.ListAll()
	if err != nil {
		return 0, nil, err
	}

	clustermap := make(map[string]*cluster.ClusterModel)
	for pos := range clusters {
		clustermap[clusters[pos].ClusterID] = &clusters[pos]
	}

	o := orm.NewOrm()

	qs := o.QueryTable(a.TableName())
	if a.Namespace != "" {
		qs = qs.Filter("namespace", a.Namespace)
	}
	if a.Operator != "" {
		qs = qs.Filter("operator", a.Operator)
	}
	if a.OperationType != 0 {
		qs = qs.Filter("operation_type", a.OperationType)
	}
	if a.ResourceType != 0 {
		qs = qs.Filter("resource_type", a.ResourceType)
	}

	if !start.IsZero() {
		qs = qs.Filter("time__gte", start)
	}
	if !end.IsZero() {
		qs = qs.Filter("time__lte", end)
	}

	if status == "failed" { // failed
		qs = qs.Exclude("status", 0).Exclude("status", 200)
	} else if status == "running" { // running
		qs = qs.Filter("status", 0)
	} else if status == "success" { // success
		qs = qs.Filter("status", 200)
	}

	if omitempty {
		qs = qs.Exclude("operation_type", "")
		qs = qs.Exclude("resource_type", "")
		qs = qs.Exclude("resource_name", "")
	}

	if keyword != "" {
		cond := orm.NewCondition()
		condition := cond.AndCond(cond.Or("namespace__icontains", keyword).Or("operator__icontains", keyword).Or("resource_name__icontains", keyword))
		qs = qs.SetCond(condition)
	}

	count, err := qs.Count()
	if err != nil {
		return 0, nil, err
	}

	var records []AuditRecord
	_, err = qs.OrderBy("-time").Offset(from).Limit(size).All(&records)

	// get cluster name
	for pos := range records {
		if records[pos].ClusterID != "" {
			cluster, ok := clustermap[records[pos].ClusterID]
			if ok {
				records[pos].ClusterName = cluster.ClusterName
			}
		}
	}
	return count, &records, err
}
func (a *AuditInfo) UpdateInfo() (err error) {
	AuditRecord := NewAuditRecord(a)
	err = AuditRecord.Update()
	return
}
