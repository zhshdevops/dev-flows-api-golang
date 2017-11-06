/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-10-24  @author liuyang
 */

package controllers

import (
	// "time"
	// "container/list"

	"dev-flows-api-golang/models"
	// "github.com/golang/glog"
)

type apiAuditInfo struct {
	res                  models.AuditResource
	op                   models.AuditOperation
	resNameKey           string
	resIDKey             string
	resNameInRequestBody bool // set it to true if pass resource name in request body. e.g. batch operation and quick restart
}

// get map of resource type to operation type
// grep "{models\." audit.go | awk -F"models." '{print $2, $3}' | awk -F"," '{print $1, $2}' | sed 's/\bAuditResource\|\bAuditOperation//g' | LANG=C sort

// golang const value does not support struct
var apiAuditInfos = map[string]apiAuditInfo{
	// "function name" : { "resource type", "operation type", "resource name key", "resource id key", "cluster id key" }
	// function name format: <package_name>.([*]<controller_name>).<function_name>

	// controllers/app
	"app.(*Controller).Create":         {models.AuditResourceApp, models.AuditOperationCreate, "", "", false},
	"app.(*Controller).CreateService":  {models.AuditResourceAppService, models.AuditOperationCreate, ":appName", "", false},
	"app.(*Controller).Delete":         {models.AuditResourceApp, models.AuditOperationDelete, "", "", true},
	"app.(*Controller).StopApps":       {models.AuditResourceApp, models.AuditOperationBatchStop, "", "", true},
	"app.(*Controller).StartApps":      {models.AuditResourceApp, models.AuditOperationBatchStart, "", "", true},
	"app.(*Controller).RestartApps":    {models.AuditResourceApp, models.AuditOperationBatchRestart, "", "", true},
	"app.(*SPIController).CheckExist":  {models.AuditResourceApp, models.AuditOperationCheckExist, ":appName", "", false},
	"app.(*SPIController).SetTopology": {models.AuditResourceAppTopology, models.AuditOperationUpdate, ":app", "", false},

	// controllers/volume
	"volume.(*Controller).Create":           {models.AuditResourceVolume, models.AuditOperationCreate, "", "", false},
	"volume.(*Controller).BatchDelete":      {models.AuditResourceVolume, models.AuditOperationBatchDelete, "", "", true},
	"volume.(*Controller).Format":           {models.AuditResourceVolume, models.AuditOperationFormat, ":volume", "", false},
	"volume.(*Controller).Expand":           {models.AuditResourceVolume, models.AuditOperationExpand, ":volume", "", false},
	"volume.(*Controller).CreateSnapshot":   {models.AuditResourceSnapshot, models.AuditOperationCreate, ":volume", "", false},
	"volume.(*Controller).DeleteSnapshot":   {models.AuditResourceSnapshot, models.AuditOperationBatchDelete, "", "", true},
	"volume.(*Controller).RollbackSnapshot": {models.AuditResourceSnapshot, models.AuditOperationRollBack, ":volume", "", false},
	"volume.(*Controller).CloneSnapshot":    {models.AuditResourceSnapshot, models.AuditOperationClone, ":volume", "", false},
	"volume.(*Controller).SetCalamariUrl":   {models.AuditResourceVolume, models.AuditOperationCreate, "", "", false},
	// controllers/service
	"service.(*ServiceController).DeleteServices":  {models.AuditResourceService, models.AuditOperationBatchDelete, "", "", true},
	"service.(*ServiceController).StartServices":   {models.AuditResourceService, models.AuditOperationBatchStart, "", "", true},
	"service.(*ServiceController).StopServices":    {models.AuditResourceService, models.AuditOperationBatchStop, "", "", true},
	"service.(*ServiceController).RestartServices": {models.AuditResourceService, models.AuditOperationBatchRestart, "", "", true},
	"service.(*ServiceController).QuickRestart":    {models.AuditResourceService, models.AuditOperationQuickRestart, ":service", "", true},
	"service.(*ServiceController).ManualScale":     {models.AuditResourceServiceManualScale, models.AuditOperationUpdate, ":service", "", false},
	"service.(*ServiceController).AutoScale":       {models.AuditResourceServiceAutoScale, models.AuditOperationUpdate, ":service", "", false},
	"service.(*ServiceController).DeleteAutoScale": {models.AuditResourceServiceAutoScale, models.AuditOperationDelete, ":service", "", false},
	"service.(*ServiceController).ModifyQuota":     {models.AuditResourceServiceQuota, models.AuditOperationUpdate, ":svcName", "", false},
	"service.(*ServiceController).ModifyHaOption":  {models.AuditResourceServiceHaOption, models.AuditOperationUpdate, ":svcName", "", false},
	"service.(*ServiceSPIController).BindDomain":   {models.AuditResourceServiceDomain, models.AuditOperationCreate, ":service", "", false},
	"service.(*ServiceSPIController).UnbindDomain": {models.AuditResourceServiceDomain, models.AuditOperationDelete, ":service", "", false},
	"service.(*ServiceSPIController).CheckExist":   {models.AuditResourceServiceDomain, models.AuditOperationCheckExist, ":service", "", false},

	"service.(*UpgradeController).RollingUpgrade":  {models.AuditResourceServiceRollingUpgrade, models.AuditOperationUpdate, ":service", "", false},
	"service.(*UpgradeController).PauseUpgrading":  {models.AuditResourceServiceRollingUpgrade, models.AuditOperationPause, ":service", "", false},
	"service.(*UpgradeController).ResumeUpgrading": {models.AuditResourceServiceRollingUpgrade, models.AuditOperationResume, ":service", "", false},

	// controllers/configgroup
	"configgroup.(*ConfigController).AddConfigGroup":     {models.AuditResourceConfigGroup, models.AuditOperationCreate, ":groupname", "", false},
	"configgroup.(*ConfigController).DeleteConfigGroups": {models.AuditResourceConfigGroup, models.AuditOperationBatchDelete, "", "", true},
	"configgroup.(*ConfigController).AddConfig":          {models.AuditResourceConfig, models.AuditOperationCreate, ":configname", "", false},
	"configgroup.(*ConfigController).EditConfig":         {models.AuditResourceConfig, models.AuditOperationUpdate, ":configname", "", false},
	"configgroup.(*ConfigController).DeleteConfigs":      {models.AuditResourceConfig, models.AuditOperationBatchDelete, ":groupname", "", true},

	// controllers/registry
	"registry.(*RegistryController).Add":    {models.AuditResourceThirdPartyRegistry, models.AuditOperationCreate, ":registry", "", false},
	"registry.(*RegistryController).Delete": {models.AuditResourceThirdPartyRegistry, models.AuditOperationDelete, "", ":id", false},

	// controllers/user
	"user.(*UserController).CreateUser": {models.AuditResourceUser, models.AuditOperationCreate, "", "", false},
	"user.(*UserController).UpdateUser": {models.AuditResourceUser, models.AuditOperationUpdate, ":user", "", false},
	"user.(*UserController).DeleteUser": {models.AuditResourceUser, models.AuditOperationDelete, ":user", "", false},

	// controllers/team
	"team.(*TeamController).CreateTeam": {models.AuditResourceTeam, models.AuditOperationCreate, "", "", false},
	"team.(*TeamController).UpdateTeam": {models.AuditResourceTeam, models.AuditOperationDelete, ":team", "", false},
	"team.(*TeamController).DeleteTeam": {models.AuditResourceTeam, models.AuditOperationDelete, ":team", "", false},
	// controllers/team/:team/clusters
	"team.(*TeamController).AddTeamClusters":    {models.AuditResourceTeam, models.AuditOperationCreate, ":team", "", false},
	"team.(*TeamController).DeleteTeamClusters": {models.AuditResourceTeam, models.AuditOperationBatchDelete, ":team", "", false},
	// controller/team/:team/users
	"team.(*TeamController).AddTeamUsers":    {models.AuditResourceTeamUsers, models.AuditOperationCreate, ":team", "", false},
	"team.(*TeamController).DeleteTeamUsers": {models.AuditResourceTeamUsers, models.AuditOperationBatchDelete, ":team", "", false},
	// controller/team/:team/spaces
	"team.(*TeamController).CreateTeamSpace": {models.AuditResourceTeamSpaces, models.AuditOperationCreate, ":team", "", false},
	"team.(*TeamController).DeleteTeamSpace": {models.AuditResourceTeamSpaces, models.AuditOperationDelete, ":team", "", false},

	// controller/team/:team/spaces/:space/users
	"team.(*TeamController).AddTeamSpaceUsers":    {models.AuditResourceTeamSpaces, models.AuditOperationCreate, ":team", "", false},
	"team.(*TeamController).DeleteTeamSpaceUsers": {models.AuditResourceTeamSpaces, models.AuditOperationBatchDelete, ":team", "", false},

	// controller/instance
	"instance.(*InstanceController).ExportInstance": {models.AuditResourceInstanceExport, models.AuditOperationCreate, "", "", true},
	"instance.(*InstanceController).Delete":         {models.AuditResourceInstanceDelete, models.AuditOperationCreate, "", "", true},

	// controller/alerts/groups
	"alert.(*Controller).CreateGroup": {models.AuditResourceAlertEmailGroup, models.AuditOperationCreate, "", "", true},
	// controller/alerts/groups/:groupid
	"alert.(*Controller).ModifyGroup": {models.AuditResourceAlertEmailGroup, models.AuditOperationUpdate, "", ":groupid", true},
	// controller/alerts/groups/batch-delete
	"alert.(*Controller).BatchDeleteGroup": {models.AuditResourceAlertEmailGroup, models.AuditOperationBatchDelete, "", "", true},
	// controller/alerts/records
	"alert.(*Controller).DeleteRecordList": {models.AuditResourceAlertRecord, models.AuditOperationDelete, "", "strategyID", false},
	// controller/alerts/strategy
	"alert.(*SPIController).SaveStrategy": {models.AuditResourceAlertRecord, models.AuditOperationCreateOrUpdate, "", "", true},
	// controller/alerts/strategy
	"alert.(*SPIController).Delete": {models.AuditResourceAlertStrategy, models.AuditOperationBatchDelete, "", "strategyIDs", false},
	// controller/alerts/strategy/toggle_enable
	"alert.(*SPIController).ToggleEnable": {models.AuditResourceAlertStrategy, models.AuditOperationToggleEnable, "", "", true},
	// controller/alerts/strategy/toggle_send_email
	"alert.(*SPIController).ToggleSendEmail": {models.AuditResourceAlertStrategy, models.AuditOperationToggleEnable, "", "", true},
	// controller/alerts/strategy/ingore
	"alert.(*SPIController).Ignore": {models.AuditResourceAlertStrategy, models.AuditOperationIgnore, "", "", true},
	// controller/alerts/rule
	"alert.(*SPIController).DeleteRules": {models.AuditResourceAlertStrategy, models.AuditOperationBatchDelete, "", "", false},
}
