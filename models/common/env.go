package common

import "os"

var (
	DEVOPS_PROTOCOL          = os.Getenv("DEVOPS_PROTOCOL")
	DEVOPS_HOST              = os.Getenv("DEVOPS_HOST")
	DEVOPS_EXTERNAL_PROTOCOL = os.Getenv("DEVOPS_EXTERNAL_PROTOCOL")
	DEVOPS_EXTERNAL_HOST     = os.Getenv("DEVOPS_EXTERNAL_HOST")
	USERPORTAL_URL           = os.Getenv("USERPORTAL_URL")
	WebHookUrlPrefix         = ""
	ScriptUrl                = ""
	FlowDetailUrl            = ""
)

func init() {

	if DEVOPS_EXTERNAL_PROTOCOL == "" {

		DEVOPS_EXTERNAL_PROTOCOL = "http"

	}

	if DEVOPS_EXTERNAL_HOST == "" {

		DEVOPS_EXTERNAL_HOST = "10.39.0.102:8090"
	}

	if USERPORTAL_URL == "" {
		USERPORTAL_URL = "https://paasdev.enncloud.cn"
	}
	FlowDetailUrl = USERPORTAL_URL + "/ci_cd/tenx_flow/tenx_flow_build"

	WebHookUrlPrefix = DEVOPS_EXTERNAL_PROTOCOL + "://" + DEVOPS_EXTERNAL_HOST + "/api/v2/devops/managed-projects/webhooks/"
	ScriptUrl = DEVOPS_EXTERNAL_PROTOCOL + "://" + DEVOPS_EXTERNAL_HOST + "/api/v2/devops/managed-projects/ci-scripts"
}
