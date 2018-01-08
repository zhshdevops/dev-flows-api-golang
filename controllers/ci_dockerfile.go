package controllers

import (
	"dev-flows-api-golang/models"
	"github.com/golang/glog"
	"net/http"
	"encoding/json"
)

type CiDockerfileController struct {
	BaseController
}

type Resp struct {
	StatusDevops  int32 `json:"status"`
	TotalDevops   int64 `json:"total"`
	ResultsDevops []string `json:"results"`
}

//@router / [GET]
func (dfile *CiDockerfileController) ListDockerfiles() {
	method := "CiDockerfileController.ListDockerfiles"
	dockerfile := models.CiDockerfile{}
	namespace := dfile.Namespace
	if namespace == "" {
		namespace = dfile.Ctx.Input.Header("username")
	}
	data, total, err := dockerfile.ListDockerfiles(namespace)
	if err != nil {
		glog.Errorf("%s %v \n", method, err)
		dfile.ResponseErrorAndCode("Get Form database failed", http.StatusNotFound)
		return
	}

	//for index,_:=range data{
	//	data[index].Create_time.Format("2006-01-02 15:04:05")
	//	data[index].Update_time.Format("2006-01-02 15:04:05")
	//}

	if total == 0 {
		var body Resp
		body.StatusDevops = http.StatusOK
		resp, _ := json.MarshalIndent(body, "", "  ")
		dfile.writeResponseHeader(http.StatusOK)
		dfile.Ctx.ResponseWriter.Write(resp)
		return
	}

	dfile.ResponseSuccessDevops(data, total)
	return
}
