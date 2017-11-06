/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-10-17  @author mengyuan
 */

package id

import "github.com/renstrom/shortuuid"
import "api-server/util/rand"

const (
	shortIDLen = 12
	longIDlen  = 16
)

const (
	appID            = "AID-"
	permissionID     = "PID-"
	roleID           = "RID-"
	clusterID        = "CID-"
	storagePoolID    = "SPID-"
	teamID           = "TID-"
	templateID       = "TPID-"
	userID           = "UID-"
	teamspaceID      = "TSID-"
	volumeID         = "VID-"
	userPreferenceID = "UPID-"
	appTemplateID    = "ATID-"
	auditID          = "AUDID-"
	integrationID    = "INTID-"
	certificateID    = "CTID-"
	notifyGroupID    = "NGID-"
	strategyID       = "STRAID-"
	snapshotID       = "SSID-"
	proxyGroupID     = "group-"
	proxyHostID     = "host-"
	proxySpaceID     = "space-"
)

// newShortID generate short uuid and add type prefix
func newShortID(prefix string) string {
	uuid := shortuuid.New()
	return prefix + uuid[:shortIDLen]
}

// newLongID generate long uuid and add type prefix
func newLongID(prefix string) string {
	uuid := shortuuid.New()
	return prefix + uuid[:longIDlen]
}
func NewApp() string {
	return newShortID(appID)
}
func NewPermission() string {
	return newShortID(permissionID)
}
func NewRole() string {
	return newShortID(roleID)
}
func NewCluster() string {
	return newShortID(clusterID)
}
func NewStoragePool() string {
	return newShortID(storagePoolID)
}
func NewTeam() string {
	return newShortID(teamID)
}
func NewTemplate() string {
	return newShortID(templateID)
}
func NewUser() string {
	return newShortID(userID)
}
func NewTeamspace() string {
	return newShortID(teamspaceID)
}
func NewVolume() string {
	return newShortID(volumeID)
}
func NewSnapshot() string {
	return newShortID(snapshotID)
}
func NewUserPreference() string {
	return newShortID(userPreferenceID)
}
func NewAppTemplate() string {
	return newLongID(appTemplateID)
}
func NewAudit() string {
	return newShortID(auditID)
}
func NewIntegration() string {
	return newShortID(integrationID)
}
func New16LengthsCode() string {
	return newLongID("")
}
func NewCertificate() string {
	return newShortID(certificateID)
}
func NewNotifyGroup() string {
	return newLongID(notifyGroupID)
}
func NewStrategy() string {
	return newShortID(strategyID)
}
func NewProxyGroup() string {
	return proxyGroupID + rand.RandString(5)
}
func NewProxyHost() string {
	return proxyHostID + rand.RandString(5)
}

func NewProxySpace() string {
	return proxySpaceID + rand.RandString(5)
}