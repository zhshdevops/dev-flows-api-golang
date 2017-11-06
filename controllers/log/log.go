/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-18  @author liuyang
 */

package log

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"dev-flows-api-golang/controllers"
	clustermodel "dev-flows-api-golang/models/cluster"
	sqlstatus "dev-flows-api-golang/models/sql/status"
	"dev-flows-api-golang/modules/log"

	"github.com/golang/glog"
)

var (
	PluginNamespace          = "kube-system"
	LoggingService           = "elasticsearch-logging"
	LoggingServicePort       = 9200
	LogTimeInterval    int64 = 60 * 60 * 24 * 5
)

// LogController log controller
type LogController struct {
	controllers.BaseController
}

// QueryLogResponse structure of query log response
type QueryLogResponse struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Log      string `json:"log"`
	ID       string `json:"id"`
	TimeNano string `json:"time_nano"`
	FileName string `json:"filename"`
}

// @Title Post log
// @Description search log
// @Success 200 success
// @router /instances/:instance/logs [post]
func (c *LogController) GetLog() {
	method := "controllers/LogController.GetLog"
	c.GetAuditInfo()
	// parse parameters in url
	namespace := c.Namespace
	clusterID := c.Ctx.Input.Param(":cluster")

	instances := strings.Split(c.Ctx.Input.Param(":instance"), ",")
	if len(instances) == 0 {
		glog.Errorln(method, "no instance requested")
		c.ErrorBadRequest("Invalid Parameter", nil)
		return
	}

	// parse parameters in body
	var params controllers.QueryLogRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &params)
	if err != nil {
		glog.Errorln(method, "failed to unmarshal request body", err)
		c.ErrorBadRequest("Invalid Parameter", nil)
		return
	}

	// get date range
	days, err := getDays(params.DateStart, params.DateEnd)
	if err != nil {
		glog.Errorln(method, "invalid date parameter", err)
		c.ErrorBadRequest("Invalid Parameter", nil)
		return
	}
	// If time_nano is not "", then direction must be "forward" or "backward",
	// server will search around this log record,
	// no default value for direction parameter.
	if params.TimeNano != "" {
		if params.Direction != "forward" && params.Direction != "backward" {
			glog.Errorln(method, "invalid direction parameter", params.Direction)
			c.ErrorBadRequest("Invalid Parameter", nil)
			return
		}
	}

	if params.Kind != "pod" && params.Kind != "service" {
		glog.Warningln(method, "set params.Kind to pod by default")
		params.Kind = "pod"
	}

	// get cluster info
	cluster := &clustermodel.ClusterModel{}
	errno, err := cluster.Get(clusterID)
	if err != nil {
		if errno == sqlstatus.SQLErrNoRowFound {
			glog.Errorln(method, "cluster", clusterID, "not found")
			c.ErrorNotFound(clusterID, "cluster")
		}
		glog.Errorln(method, "Query database failed", err)
		c.ErrorInternalServerError(err)
		return
	}

	if params.Size > 200 {
		params.Size = 200
	}

	if params.From < 0 {
		params.From = 0
	}

	params.From = params.Size * params.From

	// get log client
	logClient := log.NewLoggingClient(cluster.APIProtocol, cluster.APIHost, PluginNamespace, LoggingService+":"+strconv.Itoa(LoggingServicePort), cluster.APIToken)
	response, err := logClient.Get(instances, namespace, params.Kind, params.LogType, params.LogVolume, days, params.From, params.Size, params.Keyword, params.TimeNano, params.Direction, "query", params.FileName, clusterID)
	if err != nil {
		glog.Errorln(method, "Get log failed", err)
		c.ErrorInternalServerError(err)
		return
	}
	// format result
	var logs []QueryLogResponse
	for _, hit := range response.Hits.Hits {
		log := QueryLogResponse{
			Name:     hit.Source.Kubernetes["pod_name"],
			Kind:     "instance",
			Log:      template.HTMLEscapeString(hit.Source.Log),
			ID:       hit.ID,
			TimeNano: hit.Source.TimeNano,
			FileName: hit.Source.FileName,
		}
		logs = append(logs, log)
	}
	logSize := len(logs)

	// reverse log order
	// forward search does not need to reverse order
	if params.Direction != "forward" {
		for i := 0; i*2 < logSize; i++ {
			logs[i], logs[logSize-i-1] = logs[logSize-i-1], logs[i]
		}
	}
	c.ResponseSuccess(logs)
}

// getDays helper function to generate date list from start to end
func getDays(start, end string) ([]time.Time, error) {
	var days []time.Time
	var startDate, endDate time.Time
	var err error
	// defaultDate := "2000-01-01"
	now := time.Now()
	if start != "" {
		//start = start + "T00:00:00Z"
		startDate, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return days, fmt.Errorf("failed to parse start date %s", err.Error())
		}
	} else {
		// func Unix(sec int64, nsec int64) Time
		// func (t Time) Unix() int64
		startSeconds := now.Unix() - LogTimeInterval
		startDate = time.Unix(startSeconds, 0)
	}

	if end != "" {
		//end = end + "T23:59:59Z"
		endDate, err = time.Parse(time.RFC3339, end)
		if err != nil {
			return days, fmt.Errorf("failed to parse end date %s", err.Error())
		}
	} else {
		endDate = now
	}
	for endDate.After(startDate) {
		days = append(days, startDate)
		startDate = startDate.AddDate(0, 0, 1)
	}

	days = append(days, endDate)

	return days, nil
}

//GetServiceLog get Deployment log
// @Success 200 success
// @router /services/:service/logs [post]
func (c *LogController) GetServiceLog() {
	method := "controllers/LogController.GetServiceLog"
	namespace := c.Namespace
	clusterID := c.Ctx.Input.Param(":cluster")
	services := strings.Split(c.Ctx.Input.Param(":service"), ",")
	if len(services) == 0 {
		glog.Errorln(method, "no service requested")
		c.ErrorBadRequest("Invalid Parameter", nil)
		return
	}

	// parse parameters
	var params controllers.QueryLogRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &params)
	if err != nil {
		glog.Errorln(method, "failed to unmarshal request body", nil)
		c.ErrorBadRequest("Invalid Parameter", nil)
		return
	}

	//get date range
	days, err := getDays(params.DateStart, params.DateEnd)
	if err != nil {
		glog.Errorln(method, "invalid data parameter", err)
		c.ErrorBadRequest("Invalid Parameter", nil)
		return
	}

	if params.TimeNano != "" {
		if params.Direction != "forward" && params.Direction != "backward" {
			glog.Errorln("invalid direction paramter", params.Direction)
			c.ErrorBadRequest("Invalid parameter", nil)
			return
		}
	}

	if params.Kind != "service" {
		glog.Warningln(method, "set paarms.Kind to service by default")
		params.Kind = "service"
	}

	cluster := &clustermodel.ClusterModel{}
	errno, err := cluster.Get(clusterID)
	if err != nil {
		if errno == sqlstatus.SQLErrNoRowFound {
			glog.Errorln(method, "cluster", clusterID, "not found")
			c.ErrorNotFound(clusterID, "cluster")
		}
		glog.Errorln(method, "Quer database failed", err)
		c.ErrorInternalServerError(err)
		return
	}

	if params.Size > 200 {
		params.Size = 200
	}

	if params.From < 0 {
		params.From = 0
	}

	params.From = params.Size * params.From

	//get log client
	logClient := log.NewLoggingClient(cluster.APIProtocol, cluster.APIHost, PluginNamespace, LoggingService+":"+strconv.Itoa(LoggingServicePort), cluster.APIToken)
	// get log
	response, err := logClient.Get(services, namespace, params.Kind, params.LogType, params.LogVolume, days, params.From, params.Size, params.Keyword, params.TimeNano, params.Direction, "query", params.FileName, clusterID)
	if err != nil {
		glog.Errorln(method, "Get log failed", err)
		c.ErrorInternalServerError(err)
		return
	}

	// format result
	var logs []QueryLogResponse
	for _, hit := range response.Hits.Hits {
		log := QueryLogResponse{
			Name:     hit.Source.Kubernetes["pod_name"],
			Kind:     "instance",
			Log:      hit.Source.Log,
			ID:       hit.ID,
			TimeNano: hit.Source.TimeNano,
			FileName: hit.Source.FileName,
		}
		logs = append(logs, log)
	}
	logSize := len(logs)

	// reverse log order
	// forward search does not need to reverse order
	if params.Direction != "forward" {
		for i := 0; i*2 < logSize; i++ {
			logs[i], logs[logSize-i-1] = logs[logSize-i-1], logs[i]
		}
	}
	c.ResponseSuccess(logs)

}

// Get log file list in instance
// @Title GetLogFileList
// @Description get Deployment log file
// @Param   cluster  path string true "the id of cluster"
// @Param   service  path string true "the name of service"
// @Success 200 {string} the log of service
// @Failure 500 internal server error
// @router /instances/:instance/logfiles [post]
func (c *LogController) GetLogFileList() {
	method := "controllers/LogController.GetLogFileList"
	c.GetAuditInfo()
	// parse parameters in url
	namespace := c.Namespace
	clusterID := c.Ctx.Input.Param(":cluster")

	instances := strings.Split(c.Ctx.Input.Param(":instance"), ",")
	if len(instances) == 0 {
		glog.Errorln(method, "no instance requested")
		c.ErrorBadRequest("Invalid Parameter", nil)
		return
	}
	glog.V(5).Info(string(c.Ctx.Input.RequestBody))
	// parse parameters in body
	var params controllers.QueryLogRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &params)
	if err != nil {
		glog.Errorln(method, "failed to unmarshal request body", err)
		c.ErrorBadRequest("Invalid Parameter", nil)
		return
	}
	days, err := getDays(params.DateStart, params.DateEnd)
	if err != nil {
		glog.Errorln(method, "invalid data parameter", err)
		c.ErrorBadRequest("Invalid Parameter", nil)
		return
	}
	if params.Kind != "pod" && params.Kind != "service" {
		glog.Warningln(method, "set params.Kind to pod by default")
		params.Kind = "pod"
	}
	cluster := &clustermodel.ClusterModel{}
	errno, err := cluster.Get(clusterID)
	if err != nil {
		if errno == sqlstatus.SQLErrNoRowFound {
			glog.Errorln(method, "cluster", clusterID, "not found")
			c.ErrorNotFound(clusterID, "cluster")
		}
		glog.Errorln(method, "Query database failed", err)
		c.ErrorInternalServerError(err)

	}

	if params.Size > 200 {
		params.Size = 200
	}

	if params.From < 0 {
		params.From = 0
	}

	params.From = params.Size * params.From
	// get log client
	logClient := log.NewLoggingClient(cluster.APIProtocol, cluster.APIHost, PluginNamespace, LoggingService+":"+strconv.Itoa(LoggingServicePort), cluster.APIToken)
	// get log
	response, err := logClient.Get(instances, namespace, params.Kind, params.LogType, params.LogVolume, days, params.From, params.Size, params.Keyword, params.TimeNano, params.Direction, "aggs", params.FileName, clusterID)
	if err != nil {
		glog.Errorln(method, "Get log failed", err)
		c.ErrorInternalServerError(err)
		return
	}
	glog.V(5).Info(response.Aggs.AllFileNameKey.Buckets)
	// format result
	var fileList []string
	for _, bkt := range response.Aggs.AllFileNameKey.Buckets {
		fileList = append(fileList, bkt.Key)
	}
	c.ResponseSuccess(fileList)

}
