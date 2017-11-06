/*
* Licensed Materials - Property of tenxcloud.com
* (C) Copyright 2016 TenxCloud. All Rights Reserved.
* 2016-10-24  @author Lei
 */

/*
 Model of application templates
*/

package models

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"

	"api-server/models/common"
	sqlstatus "api-server/models/sql/status"
	"encoding/json"
)

// AppTemplateTableName table name for application templates
const AppTemplateTableName = "tenx_app_templates"
const (
	PluginTemplate int8 = 4
)

// AppTemplate is the structure of application template
/*
Type:
1=user-template
2=db-service
3=app-store
4=plugin

IsPublic:
1=Public
2=Private

Category:
The name of category that the user can define to group templates

*/
type AppTemplate struct {
	ID           string    `orm:"pk;column(id)" json:"id"`
	Name         string    `orm:"size(63);column(name)" json:"name"`
	Type         uint8     `orm:"size(1);column(type)" json:"type"`
	Category     string    `orm:"size(63);column(category)" json:"category"`
	Owner        string    `orm:"size(63);column(owner)" json:"owner"`
	Namespace    string    `orm:"size(63);column(namespace)" json:"namespace"`
	IsPublic     uint8     `orm:"size(1);column(is_public)" json:"is_public"`
	Content      string    `orm:"size(2048);column(content)" json:"content"` // How to use blob for template content???
	ImageURL     string    `orm:"size(100);column(image_url)" json:"image_url"`
	LogId        int       `orm:"size(100);column(log_id)" json:"log_id"` //log id
	Image        string    `orm:"size(255);column(image)" json:"image"`
	ImageVersion string    `orm:"size(255);column(image_version)" json:"image_version"`
	MountPath    string    `orm:"size(255);column(mount_path)" json:"mount_path"`
	Ports        string    `orm:"size(1000);column(ports)" json:"ports"` //模版配置端口
	ServiceInfo  string    `orm:"size(100);column(service_info)" json:"service_info"`
	Description  string    `orm:"size(1024);column(description)" json:"description"`
	CreateTime   time.Time `orm:"type(datetime);column(creation_time)" json:"create_time"`
	ModifyTime   time.Time `orm:"type(datetime);column(modification_time)" json:"modify_time"`
	ImageContent []byte    `orm:"-" json:"image_content"`
	Fromat       string    `orm:"-" json:"fromat"`
	VolumePath   string    `orm:"size(255);column(volume_path)" json:"volume_path"`
	ConfigName   string    `orm:"size(100);column(config_name)" json:"config_name"`
	Action       string    `orm:"size(255);column(action)" json:"action"`
}
type PluinServiceInfo struct {
	EntryPoints []EntryPoint `json:"entryPoints,omitempty"`
	IsSystem    bool         `json:"isSystem"`
	IsDeamonSet bool         `json:"isDeamonSet"` // 以k8s为准
}
type EntryPoint struct {
	Port int    `json:"port"`
	Path string `json:"path"`
}

// TableName returns the name of table in database
func (at *AppTemplate) TableName() string {
	return AppTemplateTableName
}

// NewAppTemplate return an empty AppTemplate
func NewAppTemplate() *AppTemplate {
	return &AppTemplate{}
}

// ServiceInfos umarshal at.Port to struct, return false if at.Port is emtpy
func (at *AppTemplate) ServiceInfos() (*PluinServiceInfo, bool, error) {
	var p PluinServiceInfo
	if at.ServiceInfo == "" {
		return nil, false, nil
	}
	err := json.Unmarshal([]byte(at.ServiceInfo), &p)
	if err != nil {
		return nil, false, err
	}
	return &p, true, nil
}

// GetByType return all templates matched by template type
func (at *AppTemplate) GetByType(templateType int8) ([]AppTemplate, error) {
	o := orm.NewOrm()
	var appTemplates []AppTemplate

	_, err := o.QueryTable(at.TableName()).
		Filter("type", templateType).OrderBy("-creation_time").All(&appTemplates)

	return appTemplates, err
}

// Insert insert template to the database
func (at *AppTemplate) Insert() (int64, error) {
	o := orm.NewOrm()
	at.CreateTime = time.Now()
	at.ModifyTime = time.Now()

	switch common.DType {
	case common.DRMySQL:
		count, err := o.Insert(at)
		return count, err
	case common.DRPostgres:
		sql := `INSERT INTO "tenx_app_templates" ("id", "name", "type", "owner", "namespace", "is_public", "image_url", "content", "description", "creation_time","image","image_version","mount_path","ports","log_id" ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?,?,?,?,?);`
		_, err := o.Raw(sql, at.ID, at.Name, at.Type, at.Owner, at.Namespace, at.IsPublic, at.ImageURL, at.Content, at.Description, at.CreateTime, at.Image, at.ImageVersion, at.MountPath, at.Ports, at.LogId).Exec()
		return 0, err
	}

	return 0, fmt.Errorf("Driver %s not supported", common.DType)
}

// GetByID get app template by id
func (at *AppTemplate) GetByID(id string) (uint32, error) {
	o := orm.NewOrm()

	err := o.QueryTable(at.TableName()).Filter("id", id).One(at)

	return sqlstatus.ParseErrorCode(err)
}

// GetByPluginName get plugin template by pluginName
func (at *AppTemplate) GetByPluginName(pluginName string) (uint32, error) {
	o := orm.NewOrm()

	err := o.QueryTable(at.TableName()).
		Filter("type", PluginTemplate).
		Filter("name", pluginName).
		One(at)

	return sqlstatus.ParseErrorCode(err)
}

// Get Find by name and owner
func (at *AppTemplate) Get() (int64, error) {
	o := orm.NewOrm()

	size, err := o.QueryTable(at.TableName()).Filter("name", at.Name).Filter("owner", at.Owner).All(at)

	return size, err
}

// ListMyTemplates list user/space app templates by namespace
func (at *AppTemplate) ListMyTemplates(from, size int) ([]AppTemplate, error) {
	o := orm.NewOrm()
	var records []AppTemplate
	// Match namespace and type, for now type is 1(user template)
	_, err := o.Raw("SELECT id, name, type, category, owner, is_public, image_url, description, creation_time, modification_time, content, image, image_version, mount_path, ports,log_id  FROM " + at.TableName()+
		" WHERE type = ? AND namespace = ? LIMIT ? OFFSET ?", 1, at.Namespace, size, from).QueryRows(&records)
	return records, err
}

// listTemplatesByType
func (at *AppTemplate) listTemplatesByType(tType, isPublic, from, size int) ([]AppTemplate, error) {
	o := orm.NewOrm()
	var records []AppTemplate

	_, err := o.Raw("SELECT id, name, type, category, owner, is_public, image_url, description, creation_time, modification_time, content, image, image_version, mount_path, ports,log_id  FROM " + at.TableName()+
		" WHERE type = ? AND is_public = ? LIMIT ? OFFSET ?", tType, isPublic, size, from).QueryRows(&records)

	return records, err
}

// ListPublicTemplates list all public user templates by
func (at *AppTemplate) ListPublicTemplates(from, size int) ([]AppTemplate, error) {
	// 1 = user template
	return at.listTemplatesByType(1, 1, from, size)
}

// ListDBServiceTemplates list app templates of database & cache
func (at *AppTemplate) ListDBServiceTemplates(from, size int) ([]AppTemplate, error) {
	// 2 = dbservice
	return at.listTemplatesByType(2, 1, from, size)
}

// ListAppStoreTemplates list all public and my appstore templates
func (at *AppTemplate) ListAppStoreTemplates(from, size int) ([]AppTemplate, error) {
	// 3 = appstore
	result, err := at.listTemplatesByType(3, 1, from, size)
	if err != nil {
		return nil, err
	}
	myTemplate, err := at.ListMyAppStoreTemplates(from, size)
	if err != nil {
		return nil, err
	}
	result = append(result, myTemplate...)
	return result, nil
}

// ListAppStoreTemplates list all appstore templates
func (at *AppTemplate) ListAllAppStoreTemplates(from, size int) ([]AppTemplate, error) {
	// 3 = appstore
	result, err := at.listTemplatesByType(3, 1, from, size)
	if err != nil {
		return nil, err
	}
	privateTemplate, err := at.listTemplatesByType(3, 0, from, size)
	if err != nil {
		return nil, err
	}
	result = append(result, privateTemplate...)
	return result, nil
}

// ListMyTemplates list user/space app templates by namespace
func (at *AppTemplate) ListMyAppStoreTemplates(from, size int) ([]AppTemplate, error) {
	o := orm.NewOrm()
	var records []AppTemplate
	// Match namespace and type, for now type is 1(user template)
	_, err := o.Raw("SELECT id, name, type, category, owner, is_public, image_url, description, creation_time, modification_time,content, image, image_version, mount_path, ports,log_id  FROM " + at.TableName()+
		" WHERE type = ? AND namespace = ? LIMIT ? OFFSET ?", 3, at.Namespace, size, from).QueryRows(&records)
	return records, err
}

// DeleteByID delete application template
func (at *AppTemplate) DeleteByID() (int64, error) {
	o := orm.NewOrm()
	num, err := o.QueryTable(at.TableName()).Filter("id", at.ID).Filter("owner", at.Owner).Delete()

	return num, err
}

// UpdateByID update record by id, only support is_public content description
func (at *AppTemplate) UpdateByID() (time.Time, error) {
	o := orm.NewOrm()
	var err error

	var modified bool
	params := orm.Params{}

	if at.IsPublic != 0 {
		params["is_public"] = at.IsPublic
		modified = true
	}

	if at.Content != "" {
		params["content"] = at.Content
		modified = true
	}

	params["description"] = at.Description
	modified = true

	updateTime := time.Now()
	if modified {
		params["modification_time"] = updateTime
		_, err = o.QueryTable(AppTemplateTableName).Filter("id", at.ID).Filter("owner", at.Owner).Update(params)
	}
	return updateTime, err
}

// GetCountByNamespace get the count of template belonging namespaceList
func (at *AppTemplate) GetCountByNamespace(namespaceList []string) (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable(at.TableName()).
		Filter("namespace__in", namespaceList).
		Count()
}

// GetCountDistinctByPermission
func (t *AppTemplate) GetCountByPermission(namespace string) (int, int, error) {
	sql := `SELECT is_public, COUNT(*) as cnt
	FROM ` + t.TableName() + ` WHERE namespace = ? GROUP BY is_public `
	var result []struct {
		IsPublic int
		Cnt      int
	}
	public, private := 0, 0
	o := orm.NewOrm()
	_, err := o.Raw(sql, namespace).QueryRows(&result)
	if err != nil {
		return 0, 0, err
	}
	for _, r := range result {
		if r.IsPublic == 1 {
			public = r.Cnt
		} else {
			private = r.Cnt
		}
	}
	return public, private, nil
}
