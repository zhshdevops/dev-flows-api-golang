package common

import "os"

var (
	DEVOPS_PROTOCOL          = os.Getenv("DEVOPS_PROTOCOL")
	DEVOPS_HOST              = os.Getenv("DEVOPS_HOST")
	DEVOPS_EXTERNAL_PROTOCOL = os.Getenv("DEVOPS_EXTERNAL_PROTOCOL")
	DEVOPS_EXTERNAL_HOST     = os.Getenv("DEVOPS_EXTERNAL_HOST")
	WebHookUrlPrefix         = ""
	ScriptUrl                = ""
)

func init() {

	if DEVOPS_EXTERNAL_PROTOCOL == "" {

		DEVOPS_EXTERNAL_PROTOCOL = "http"

	}

	if DEVOPS_EXTERNAL_HOST == "" {

		DEVOPS_EXTERNAL_HOST = "10.39.0.102:8090"
	}

	WebHookUrlPrefix = DEVOPS_EXTERNAL_PROTOCOL + "://" + DEVOPS_EXTERNAL_HOST + "/api/v2/devops/managed-projects/webhooks/"
	ScriptUrl = DEVOPS_EXTERNAL_PROTOCOL + "://" + DEVOPS_EXTERNAL_HOST + "/api/v2/devops/managed-projects/ci-scripts"
}
