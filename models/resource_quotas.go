package models

import "github.com/astaxie/beego/orm"

// ResourceQuota is the structure of resource quota definition
type ResourceQuota struct {
	ID           string `orm:"pk;column(id)" json:"id"`
	Namespace    string `orm:"size(64);column(namespace)" json:"namespace"`
	LimitType    uint8  `orm:"column(limit_type)" json:"limit_type"`
	LimitDetails string `orm:"size(400);column(limit_details)" json:"limit_details"`
}
type configQuota struct {
	Group uint32 `json:"group"`
	File  uint32 `json:"file"`
}

type devOps struct {
	CodeRepo     string `json:"code_repo"` // github/gitlab/svn
	Flow         uint32 `json:"flow"`
	Stage        uint32 `json:"stage"`
	HasBaseImage uint8  `json:"HasBaseImage"` // 0 or 1
}

type alerting struct {
	RuleNum uint32 `json:"ruleNum"` // every strategy rule num limit, 5 or 10
}

type ResourceQuotaObject struct {
	App           uint32      `json:"app"`
	Service       uint32      `json:"service"`
	Config        configQuota `json:"config"`
	Memory        uint64      `json:"memory"` //Mi
	Volume        uint32      `json:"volume"` //Mi
	DatabaseCache uint32      `json:"database_cache"`
	Team          uint32      `json:"team"`
	DevOps        devOps      `json:"devops"`
	Image         uint32      `json:"image"`
	Integration   uint8       `json:"integration"` //  0/1
	ComposeFile   uint32      `json:"compose_file"`
	OpenAPI       uint32      `json:"open_api"` // 0/1
	Logging       uint8       `json:"logging"`  //  0/1
	Alert         alerting    `json:"alert"`
}

const (
	ResourceQuotaLimitTypeStandard     = 0 // standard
	ResourceQuotaLimitTypeProfessional = 1 // professional
	ResourceQuotaLimitTypeCustom       = 2 // custom
)

// StandardResourceQuota default resource quota the user should follow for standard edition
var StandardResourceQuota = ResourceQuotaObject{
	App:     100,
	Service: 100,
	Config: configQuota{
		Group: 20,
		File:  20,
	},
	Memory:        5120,    // 5G
	Volume:        1024000, // 1000G
	DatabaseCache: 2,
	Team:          0,
	DevOps: devOps{
		CodeRepo:     "github",
		Flow:         2,
		Stage:        8,
		HasBaseImage: 0,
	},
	Image:       2,
	Integration: 0,
	ComposeFile: 0,
	OpenAPI:     10,
	Logging:     0,
	Alert: alerting{
		RuleNum: 5,
	},
}

// ProfessionalResourceQuota default resource quota the user should follow for professional edition
var ProfessionalResourceQuota = ResourceQuotaObject{
	App:     100,
	Service: 100,
	Config: configQuota{
		Group: 100,
		File:  400,
	},
	Memory:        10240,
	Volume:        1024000, // 1000G
	DatabaseCache: 20,
	Team:          100,
	DevOps: devOps{
		CodeRepo:     "github/gitlab/svn/gogs",
		Flow:         100,
		Stage:        400,
		HasBaseImage: 1,
	},
	Image:       20,
	Integration: 1,
	ComposeFile: 50,
	OpenAPI:     1,
	Logging:     2,
	Alert: alerting{
		RuleNum: 100,
	},
}

// TableName returns the name of table in database
func (r *ResourceQuota) TableName() string {
	return "tenx_resource_quotas"
}

// NewResoureQuota generate a new resource quota object
func NewResoureQuota() *ResourceQuota {
	return &ResourceQuota{}
}

// GetByNamespace return resource quote info by namespace
func (r *ResourceQuota) GetByNamespace(namespace string) error {
	o := orm.NewOrm()
	err := o.QueryTable(r.TableName()).Filter("namespace", namespace).One(r)
	return err
}
