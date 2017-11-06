/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-22  @author mengyuan
 */

package models

import (
	"api-server/util/secure"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
	"github.com/pborman/uuid"
)

const HarborConfigType = "harbor"
const RegistryConfigType = "registry"
const VersionConfigType = "version"
const LDAPConfigType = "ldap"
const RBDConfigType = "rbd"

// DefaultStorageID temporary logic: try to get the newest rdb type configID, if failed, use this one
var DefaultStorageID = "9e6c76aa-ce1a-4e39-98fd-1589855fead1"

type Configs struct {
	ConfigID     string    `orm:"pk;column(config_id)"`
	ConfigType   string    `orm:"column(config_type)"`
	ConfigDetail string    `orm:"column(config_detail)"`
	CreateTime   time.Time `orm:"type(datetime);column(create_time)"`
	Description  string    `orm:"column(description)"`
}

// Config of tenxcloud registry service
type RegistryConfig struct {
	URL          string `json:"url"`
	Protocol     string `json:"protocol"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	V2Server     string `json:"v2Server"`
	V2AuthServer string `json:"v2AuthServer"`
}

// Config of User LDAPConfig
type LDAPConfig struct {
	Addr                  string `json:"addr"`
	TLS                   string `json:"tls"`
	InsecureTLSSkipVerify bool   `json:"insecureTLSSkipVerify"`
	Base                  string `json:"base"`
	BindDN                string `json:"bindDN"`
	BindPassword          string `json:"bindPassword"`
	GroupBaseDN           string `json:"groupBaseDN"`
	GroupFilter           string `json:"groupFilter"`
	UserFilter            string `json:"userFilter"`
	UserProperty          string `json:"userProperty"`
	EmailProperty         string `json:"emailProperty"`
}

type StorageDetailConfig struct {
	Monitors []string `json:"monitors"`
	Pool     string   `json:"pool"`
	User     string   `json:"user"`
	Keyring  string   `json:"keyring"`
	FsType   string   `json:"fsType"`
}

type StorageConfigAgent struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type StorageConfig struct {
	Name        string              `json:"name"`
	URL         string              `json:"url"`
	Agent       StorageConfigAgent  `json:"agent"`
	Config      StorageDetailConfig `json:"config"`
	CalamariUrl string              `json:"calamariUrl"`
}

type MailConfig struct {
	Secure         bool   `json:"secure"`
	SenderMail     string `json:"senderMail"`
	SenderPassword string `json:"senderPassword"`
	MailServer     string `json:"mailServer"`
	ServiceMail    string `json:"service_mail"`
}

type VersionConfig struct {
	Pro struct {
		Version string
		Link    string
	}
	Lite struct {
		Version string
		Link    string
	}
}

func NewConfigs() *Configs {
	return &Configs{}
}

func (t *Configs) TableName() string {
	return "tenx_configs"
}

/*func (t *Configs) GetCICDHost() (string, error) {
	o := orm.NewOrm()
	c := Configs{}
	if err := o.QueryTable(t.TableName()).
		Filter("config_type", "cicd").
		One(&c, "ConfigDetail"); err != nil {
		return "", err
	}
	detail := struct {
		Protocol string
		URL      string
	}{}
	if err := json.Unmarshal([]byte(c.ConfigDetail), &detail); err != nil {
		return "", err
	}
	if detail.Protocol == "" || detail.URL == "" {
		return "", errors.New("protocol/url is empty")
	}
	return detail.Protocol + "://" + detail.URL, nil
}*/
// GetRegistryExtensionServer get registry extension server address
func (t *Configs) GetByIDs(configIDs []string) ([]Configs, error) {
	o := orm.NewOrm()
	var results []orm.Params
	storageConfigs := make([]Configs, len(configIDs))
	if len(configIDs) <= 0 {
		return storageConfigs, nil
	}
	if _, err := o.QueryTable(t.TableName()).
		Filter("config_id__in", configIDs).
		Values(&results); err != nil {
		return storageConfigs, err
	}

	for index, row := range results {
		storageConfigs[index].ConfigID = row["ConfigID"].(string)
		storageConfigs[index].ConfigType = row["ConfigType"].(string)
		storageConfigs[index].ConfigDetail = row["ConfigDetail"].(string)
		if row["Description"] != nil {
			storageConfigs[index].Description = row["Description"].(string)
		}
	}
	return storageConfigs, nil
}

// GetRegistryExtensionServer get registry extension server address
func (t *Configs) GetRegistryExtensionServer() (*RegistryConfig, error) {
	o := orm.NewOrm()
	c := Configs{}
	detail := &RegistryConfig{}
	if err := o.QueryTable(t.TableName()).
		Filter("config_type", RegistryConfigType).
		One(&c, "ConfigDetail"); err != nil {
		return detail, err
	}

	if err := json.Unmarshal([]byte(c.ConfigDetail), &detail); err != nil {
		return detail, err
	}
	if detail.Protocol == "" || detail.Host == "" {
		return detail, errors.New("protocol/host is empty")
	}
	return detail, nil
}

// GetHarborServer get harbor server
func (t *Configs) GetHarborServer() (string, error) {
	o := orm.NewOrm()
	c := Configs{}
	// check parameters
	type serverURL struct {
		URL string `json:"Url"`
	}
	if err := o.QueryTable(t.TableName()).
		Filter("config_type", HarborConfigType).
		One(&c, "ConfigDetail"); err != nil {
		return "", err
	}
	detail := serverURL{}

	if err := json.Unmarshal([]byte(c.ConfigDetail), &detail); err != nil {
		return "", err
	}
	if detail.URL == "" {
		return "", errors.New("url is empty")
	}
	// Only return host name or IP without protocol
	index := strings.Index(detail.URL, "://")
	if index > 0 {
		detail.URL = detail.URL[index+3:]
	}
	return detail.URL, nil
}

// GetRegistryExtensionServer get registry extension server address
func (t *Configs) CheckRegistryInfo() (*RegistryConfig, interface{}) {
	o := orm.NewOrm()
	RowExist := o.QueryTable(t.TableName()).
		Filter("config_type", RegistryConfigType).Exist()
	c := Configs{}
	detail := &RegistryConfig{}
	if RowExist {
		if err := o.QueryTable(t.TableName()).
			Filter("config_type", RegistryConfigType).
			One(&c, "ConfigDetail"); err != nil {
			return detail, errors.New("get registry information error ")
		}
		if err := json.Unmarshal([]byte(c.ConfigDetail), &detail); err != nil {
			return detail, errors.New("unmarshal error")
		}
		if detail.Protocol == "" || detail.Host == "" {
			return detail, "registry protocol or host is empty"
		}
		return detail, nil
	}
	return detail, "has no registry"
}

// CreateRegistryExtensionServer get registry extension server config
func (t *Configs) CreateRegistryExtensionServer() (int64, error) {
	o := orm.NewOrm()
	t.ConfigID = uuid.New()
	t.ConfigType = RegistryConfigType
	t.CreateTime = time.Now()
	t.Description = "TenxCloud Hub"
	count, err := o.Insert(t)
	return count, err
}

// RemoveRegistryExtensionServer remove registry extension server config
func (t *Configs) RemoveRegistryExtensionServer() (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable(t.TableName()).
		Filter("config_type", RegistryConfigType).
		Delete()
	return count, err
}

// UpdateRegistryPassword update the password of registry server admin
func (t *Configs) UpdateRegistryPassword(password string) error {
	o := orm.NewOrm()
	c := Configs{}
	registryConfig := &RegistryConfig{}
	if err := o.QueryTable(t.TableName()).
		Filter("config_type", RegistryConfigType).
		One(&c); err != nil {
		if err == orm.ErrNoRows {
			// Just skip it if no row found
			return nil
		}
		return err
	}

	if err := json.Unmarshal([]byte(c.ConfigDetail), &registryConfig); err != nil {
		return err
	}
	registryConfig.Password = password
	newConfig, err := json.Marshal(registryConfig)
	if err != nil {
		return err
	}
	params := orm.Params{}
	params["config_detail"] = newConfig
	_, err = o.QueryTable(t.TableName()).Filter("config_id", c.ConfigID).Update(params)

	return err
}

// GetVersion get version info
func (t *Configs) GetVersion() (VersionConfig, error) {
	method := "GetVersion"
	o := orm.NewOrm()
	c := Configs{}
	detail := VersionConfig{}
	if err := o.QueryTable(t.TableName()).
		Filter("config_type", VersionConfigType).
		One(&c, "ConfigDetail"); err != nil {
		glog.Errorln(method, "get version config failed.", err)
		return detail, errors.New("internal error")
	}

	if err := json.Unmarshal([]byte(c.ConfigDetail), &detail); err != nil {
		glog.Errorf("%s unmarshal version config failed. raw string: %s, err: %s\n", method, c.ConfigDetail, err)
		return detail, err
	}
	return detail, nil
}

//UpdateStorageConfig
func (t *Configs) UpdateStorageConfigByConfigID(configID string, storageType string, storageConfig StorageConfig) error {
	configDetail, err := json.Marshal(storageConfig)
	if nil != err {
		return err
	}
	o := orm.NewOrm()
	c := Configs{
		ConfigID:     configID,
		ConfigType:   storageType,
		ConfigDetail: string(configDetail[:]),
		CreateTime:   time.Now(),
	}
	_, err = o.Update(c, "config_type", "config_detail", "create_time")
	if nil != err {
		return err
	}
	return nil
}

func (t *Configs) InsertStorageConfig(storageType string, storageConfig StorageConfig) (*Configs, error) {
	t.ConfigID = uuid.New()
	o := orm.NewOrm()
	_, err := o.Insert(t)
	return t, err
}

func (t *Configs) InsertConfig(connect ...orm.Ormer) (*Configs, error) {
	var o orm.Ormer
	if len(connect) > 0 {
		o = connect[0]
	} else {
		o = orm.NewOrm()
	}
	t.ConfigID = uuid.New()
	err := t.encryptPasswordIfNeeded()
	if err != nil {
		return t, err
	}
	_, err = o.Insert(t)
	return t, err
}

func (t *Configs) UpdateConfig() (int64, error) {
	t.CreateTime = time.Now()
	err := t.encryptPasswordIfNeeded()
	if err != nil {
		return 0, err
	}
	o := orm.NewOrm()
	num, err := o.Update(t, "config_detail", "create_time")
	return num, err
}

func (t *Configs) UpdateRBDConfig() (int64, error) {
	t.CreateTime = time.Now()
	err := t.encryptPasswordIfNeeded()
	if err != nil {
		return 0, err
	}
	o := orm.NewOrm()
	num, err := o.QueryTable(t.TableName()).Filter("config_type", "rbd").Update(orm.Params{
		"config_detail": t.ConfigDetail,
		"create_time":   t.CreateTime,
	})
	return num, err
}

// UpdateDescriptionByType update config description colume
func (t *Configs) UpdateDescriptionByType(configType, description string) (int64, error) {
	o := orm.NewOrm()
	params := orm.Params{}
	params["description"] = description
	num, err := o.QueryTable(t.TableName()).Filter("config_type", configType).Update(params)

	return num, err
}

//GetConfig exclude storageConfig
func (t *Configs) GetConfig(orms ...orm.Ormer) ([]*Configs, error) {
	var o orm.Ormer
	if len(orms) == 0 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	//TODO: stroage暂时只有一个rbd类型， 多类型后续支持
	sql := "select * from tenx_configs where config_type in ('cicd', 'registry', 'mail', 'apiServer', 'rbd', 'harbor')"
	allConfigs := make([]*Configs, 0)
	_, err := o.Raw(sql).QueryRows(&allConfigs)
	if err == orm.ErrNoRows {
		return allConfigs, nil
	}
	if err != nil {
		return nil, err
	}
	return allConfigs, nil
}

// GetStorageRBDConfigByClusterID storage config
func (t *Configs) GetStorageRBDConfigByClusterID(clusterID string) (*Configs, error) {
	o := orm.NewOrm()
	sql := "select * from tenx_clusters t1 inner join tenx_configs t2 on t1.storage = t2.config_id where t1.id = ? and t2.config_type = 'rbd'"
	storageConfig := Configs{}
	err := o.Raw(sql, clusterID).QueryRow(&storageConfig)
	if err != nil && err != orm.ErrNoRows {
		return nil, err
	}
	return &storageConfig, nil
}

//GetLDAPConfig user regitry config
func (t *Configs) GetLDAPConfig() (*LDAPConfig, error) {
	ldapConfigs, err := t.GetByType(LDAPConfigType)
	if err != nil {
		return nil, err
	}
	if len(ldapConfigs) > 0 {
		registryConfig := &LDAPConfig{}
		// Use the first one
		if err := json.Unmarshal([]byte(ldapConfigs[0].ConfigDetail), registryConfig); err != nil {
			return registryConfig, err
		}
		return registryConfig, nil
	}
	return nil, nil
}

// RemoveByType remove config by type
func (t *Configs) RemoveByType(o orm.Ormer, configType string) (int64, error) {
	count, err := o.QueryTable(t.TableName()).
		Filter("config_type", configType).
		Delete()
	return count, err
}

func (t *Configs) GetByType(configType string) ([]Configs, error) {
	var result []Configs
	_, err := orm.NewOrm().QueryTable(t.TableName()).
		Filter("config_type", configType).
		OrderBy("-create_time").All(&result)

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *Configs) encryptPasswordIfNeeded() error {
	method := "encryptPasswordIfNeeded"
	if t.ConfigType == LDAPConfigType {
		// Encrypt password before save to database
		ldapConfig := LDAPConfig{}
		err := json.Unmarshal([]byte(t.ConfigDetail), &ldapConfig)
		if err != nil {
			glog.Errorf(method, "Failed to unmarshal LDAP config")
			return err
		}
		ldapConfig.BindPassword, err = secure.EncryptAndBase64(ldapConfig.BindPassword)
		if err != nil {
			glog.Errorf(method, "Failed to encrypt LDAP password")
			return err
		}
		details, err := json.Marshal(ldapConfig)
		if err != nil {
			glog.Errorf(method, "Failed to marshal LDAP config")
			return err
		}
		t.ConfigDetail = string(details)
	}
	return nil
}
