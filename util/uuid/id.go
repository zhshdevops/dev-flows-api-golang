package uuid

import "github.com/renstrom/shortuuid"
import "dev-flows-api-golang/util/rand"

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
	// ID prefix for managed projects
	MANAGED_PROJECT_ID_PREFIX   = "MPID-"
	STAGE_ID_PREFIX             = "CISID-"
	FLOW_BUILD_ID_PREFIX        = "FBID-"
	STAGE_BUILD_ID_PREFIX       = "SBID-"
	CI_FLOW_ID_PREFIX           = "CIFID-"
	CI_IMAGES                   = "CIMID-"
	CD_RULE_ID_PREFIX           = "CDRID-"
	AUDIT_ID_PREFIX             = "AUDID-"
	CD_DEPLOYMENT_LOG_ID_PREFIX = "CDLID-"
	CI_SCRIPT_PREFIX            = "SCRIPT-"
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

func NewManagedProjectID() string {

	return newShortID(MANAGED_PROJECT_ID_PREFIX)

}
func NewAuditID() string {

	return newShortID(AUDIT_ID_PREFIX)

}

func NewCDLogID() string {

	return newShortID(CD_DEPLOYMENT_LOG_ID_PREFIX)

}

func NewStageID() string {

	return newShortID(STAGE_ID_PREFIX)

}

func NewFlowBuildID() string {

	return newShortID(FLOW_BUILD_ID_PREFIX)

}

func NewStageBuildID() string {

	return newShortID(STAGE_BUILD_ID_PREFIX)

}

func GetCIMID() string {

	return newShortID(CI_IMAGES)

}

func NewCDRuleID() string {

	return newShortID(CD_RULE_ID_PREFIX)

}
func NewScriptID() string {

	return newShortID(CI_SCRIPT_PREFIX)

}

func NewCIFlowID() string {

	return newShortID(CI_FLOW_ID_PREFIX)

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
