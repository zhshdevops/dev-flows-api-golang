/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-18  @author liuyang
 */

package log

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
)

// LoggingClient logging client
type LoggingClient struct {
	URL        string `json:"url"`   // url: "https://192.168.1.93:6443/",
	Token      string `json:"token"` // "bpnxUMd0v10EOOHvAWCwY4A5wtZBXIrz"
	httpClient *http.Client
}

// NewLoggingClient new logging client
func NewLoggingClient(protocol, apiHost, pluginNamespace, loggingService, token string) *LoggingClient {
	// Check env for customized elasticsearch service
	elasticsearchURL := os.Getenv("EXTERNAL_ES_URL")
	elasticsearchURL="http://paasdev.enncloud.cn:9200"
	if elasticsearchURL == "" {
		elasticsearchURL = fmt.Sprintf("%s://%s/api/v1/proxy/namespaces/%s/services/%s", protocol, apiHost, pluginNamespace, loggingService)
	}
	var l = &LoggingClient{
		URL:   elasticsearchURL,
		Token: token,
	}

	l.httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			IdleConnTimeout: time.Second * 30,
		},
		Timeout: 30 * time.Second,
	}
	return l
}

// Get fetch log from elastic search service
func (l *LoggingClient) Get(names []string, namespace, kind, lt, lv string, days []time.Time, from, size int, kw, nano, direction, action, fileName, clusterID string) (*ESResponse, error) {
	method := "modules/log/LoggingClient.Get"

	// convert date list into elasticsearch indices
	indices := days2indices(days)
	indexes := getIndexBydate(days)

	if size < defaultLogSize {
		size = defaultLogSize
	}
	if from < defaultLogOffset {
		from = defaultLogOffset
	}
	logs, err := l.QueryGetLog(indexes, indices, names, namespace, kind, lt, lv, from, size, kw, nano, direction, action, fileName, clusterID)
	if err != nil {
		glog.Errorln(method, "Get logs failed, namespace:", namespace, ", name:", names, ", kind:", kind, ", error:", err)
		return nil, err
	}

	return logs, err
}

func (l *LoggingClient) ClearScroll(scrollId string) error {

	method := "LoggingClient.ClearScroll"
	scrollIdArray := make([]string, 0)
	scrollIdArray = append(scrollIdArray, scrollId)

	// 1. Build request url
	var url string

	url = fmt.Sprintf("%s/_search?scroll", strings.Trim(l.URL, "/"))

	// 2. build request body
	scroll := `{
		"scroll_id":%s
	}`
	requestBody := fmt.Sprintf(scroll, scrollIdArray)

	glog.V(2).Infoln(method, "url:", url)
	glog.V(4).Infoln(method, "body:", requestBody)

	// send request
	rawBytes, _, err := l.httpRequest("DELETE", url, requestBody)
	if err != nil {
		return err
	}
	glog.V(1).Info(string(rawBytes))

	return nil

}

func (l *LoggingClient) RefineLogEnnFlowLog(podName string, esResponse *ESResponse) string {
	method := "modules/log/refineLogEnnFlowLog"

	refinedLogs := ""
	if esResponse!=nil{
		hits := esResponse.Hits.Hits

		if len(hits) != 0 {
			for _, hit := range hits {
				if hit.Source.Kubernetes["pod_name"] == podName {
					refinedLogs += fmt.Sprintf(`<font color="#ffc20e">[%s]</font> %s \n`, hit.Source.Timestamp.Format("2006/01/02 15:04:05"), hit.Source.Log)
				}
			}

			return refinedLogs
		}

		glog.Infof("%s from elasticsearch resp data is empty===>%v\n", method, hits)
	}


	return ""
}

func (l *LoggingClient) GetEnnFlowLog(namespace, podName, containerName string, date time.Time, clusterID string) (*ESResponse, error) {
	method := "modules/log/GetEnnFlowLog.Get"

	//namespace, podName,containerName string, date time.Time,clusterID string
	logs, err := l.QueryGetEnnFLowLog(namespace, podName, containerName, date, clusterID)
	if err != nil {
		glog.Errorln(method, "Get logs failed, namespace:", namespace, ", podName:", podName, ", error:", err)
		return nil, err
	}

	return logs, err
}

// CheckHealth check elasticsearch-logging service health
func (l *LoggingClient) CheckHealth() (map[string]string, error) {
	method := "CheckHealth"

	url := fmt.Sprintf("%s/_cat/health?v", strings.Trim(l.URL, "/"))

	rawBytes, _, err := l.httpRequest("GET", url, "")
	if err != nil {
		glog.Errorln("%s check elastic search health fails, error: %v", method, err)
		return nil, err
	}
	arr := reMatchNonSpaces.FindAll(rawBytes, -1)
	res := make(map[string]string)
	mapLen := len(arr) / 2
	for i := 0; i < mapLen; i++ {
		key := strings.Trim(string(arr[i]), " \n")
		val := strings.Trim(string(arr[i+mapLen]), " \n")
		res[key] = val
	}

	return res, nil
}
