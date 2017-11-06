/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-18  @author liuyang
 */

package log

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"
)

var (
	indexPrefix          = "logstash-"
	defaultLogSize       = 200
	defaultLogOffset     = 0
	requiredFields       = []string{"log", "kubernetes.pod_name", "docker.container_id", "time_nano"}
	requiredFieldsString = `["log", "kubernetes.pod_name", "docker.container_id", "time_nano", "filename"]`

	logMatchName      = `{ "match" : {"kubernetes.%s" : "%s"} }`
	logMatchType      = `{"terms": {"stream":["%s", "%s"]}},`
	logMatchAggs      = `,"aggs": {"all_file_name": {"terms": {"field": "%s"}}}`
	logMatchDateRange = `{"range":{"@timestamp":{"gte":"%s", "lte":"%s"}}},`
	logMatchvolume    = `{"term": {"logvolume":"%s"}},`
	logMatchFile      = `{"term": {"filename":"%s"}},`
	logMatchCluster   = `{"term": {"kubernetes.labels.ClusterID":"%s"}},`
	//logMatchKeyword          = `{"wildcard":{"log":"*%s*"}},`
	logMatchKeyword          = `{"match": {"log": {"query": "%s","fuzziness": 3,"prefix_length": 1}}},`
	logMatchTimeNanoBackward = `{ "range" : { "time_nano" : { "lte" : "%s" } } },`
	logMatchTimeNanoForward  = `{ "range" : { "time_nano" : { "gte" : "%s" } } },`
	logGetRequestBody        = `{
    "from": %d,
    "size": %d,
    "_source": %s,
    "sort": [
      {"time_nano"  : {"order" : "%s"}}
    ],
    "query": {
      "bool": {
		"must": [
		   %s
		   %s
		   %s
		   %s
           %s
           %s
           { "match" : {"kubernetes.namespace_name" : "%s"} }
        ],
        "should": [
          %s
        ],
        "minimum_should_match": 1
      }
	}
%s
  }`

	logCiCdRequest =`{
      "from" : 0,
      "sort": [
        {"time_nano"  : {"order" : "desc"}}
      ],
      "query": {
        "bool": {
          "must": [
             {
                 "match": {
                   "kubernetes.pod_name": {
                     "query": "%s",
                     "type": "phrase"
                   }
                 }
             },
             {
                 "match": {
                   "kubernetes.namespace_name": {
                     "query": "%s",
                     "type": "phrase"
                   }
                 }
             },
             {
                 "match": {
                   "kubernetes.labels.ClusterID": {
                     "query": "%s",
                     "type": "phrase"
                   }
                 }
             }
          ],
          "should": [
          	{
                 "match": {
                   "kubernetes.container_name": {
                     "query": "%s",
                     "type": "phrase"
                   }
                 }
             },
             {
                 "match": {
                   "kubernetes.container_name": {
                     "query": "%s",
                     "type": "phrase"
                   }
                 }
             }
          ]
        }
      },
      "_source": ["log", "kubernetes.pod_name", "docker.container_id", "@timestamp"],
      "size": %d
    }`

	reMatchNonSpaces = regexp.MustCompile(`(\S+)\s*`)
)

// QueryGetLog returns log of certain pod/service
func (lc *LoggingClient) QueryGetLog(indexes, indices, names []string, namespace, kind, lt, lv string, from, size int, kw, nano, direction, action, fileName, clusterID string) (*ESResponse, error) {
	method := "QueryGetLog"
	if size < 1 {
		size = defaultLogSize
	}

	// 1. Build request url
	var url string
	// if indices == nil || len(indices) == 0 {
	// url = fmt.Sprintf("%s/logstash*/_search?pretty&ignore_unavailable=true", strings.Trim(lc.URL, "/"))
	// } else {
	url = fmt.Sprintf("%s/%s/_search?pretty&ignore_unavailable=true", strings.Trim(lc.URL, "/"), strings.Join(indexes, ","))
	// }

	// 2. Build members of request body

	var matchKeyword string
	if kw != "" {
		// matchKeyword = fmt.Sprintf(logMatchKeyword, escape(kw))
		matchKeyword = fmt.Sprintf(logMatchKeyword, kw)
	}

	if clusterID != "" {
		clusterID = fmt.Sprintf(logMatchCluster, clusterID)
	}

	var matchLogType string
	if lt == "file" {
		matchLogType = fmt.Sprintf(logMatchType, lt, "")
	} else {
		matchLogType = fmt.Sprintf(logMatchType, "stderr", "stdout")
	}
	var matchLogVolume string
	if lv != "" {
		matchLogVolume = fmt.Sprintf(logMatchvolume, lv)
	}
	var matchFileName string
	if fileName != "" {
		matchFileName = fmt.Sprintf(logMatchFile, fileName)
	}

	var matchDateRange string
	matchDateRange = fmt.Sprintf(logMatchDateRange, indices[0], indices[1])

	if kind == "pod" {
		kind = "pod_name"
	} else if kind == "service" {
		kind = "container_name"
	} else {
		return nil, fmt.Errorf("Kind not supported")
	}

	var matchNames []string
	for _, name := range names {
		matchNames = append(matchNames, fmt.Sprintf(logMatchName, kind, name))
	}

	var matchTimeNano string
	var searchOrder = "desc"
	if nano != "" {
		if direction == "forward" {
			searchOrder = "asc"
			matchTimeNano = fmt.Sprintf(logMatchTimeNanoForward, nano)
		} else {
			matchTimeNano = fmt.Sprintf(logMatchTimeNanoBackward, nano)
		}
	}
	// TODO check action for aggs or query
	var actions string
	if action == "aggs" {
		actions = fmt.Sprintf(logMatchAggs, "filename")
	}
	// 3. build request body
	requestBody := fmt.Sprintf(logGetRequestBody, from, size, requiredFieldsString, searchOrder, matchLogType, matchLogVolume, matchDateRange, matchFileName, matchKeyword, matchTimeNano, namespace, strings.Join(matchNames, ","), actions)

	glog.V(2).Infoln(method, "url:", url)
	glog.V(4).Infoln(method, "body:", requestBody)

	// send request
	rawBytes, statusCode, err := lc.httpRequest("GET", url, requestBody)
	if err != nil {
		netError, ok := err.(net.Error)
		// if timeout or 404, then return empty result
		if ok && netError.Timeout() {
			glog.Errorln(method, "timeout error", err)
			resp := &ESResponse{}
			return resp, nil
		} else if statusCode == 404 {
			glog.Errorln(method, "404 error", err)
			resp := &ESResponse{}
			return resp, nil
		}
		glog.Errorln(method, "Get log failed, namespace:", namespace, ", name:", names, ", kind:", kind, ", error:", err)
		return nil, err
	}

	glog.V(6).Info(string(rawBytes))
	// unmarshal response
	response := &ESResponse{}
	if err := json.Unmarshal(rawBytes, response); err != nil {
		errInfo := strings.Trim(string(rawBytes), "\n \t")
		glog.Errorln(method, "Unmarshal response failed, data:", errInfo, ", error:", err)
		return nil, fmt.Errorf("%s", rawBytes)
	}
	if response.Status != 0 {
		return nil, fmt.Errorf("invalid keyword: %s", kw)
	}
	return response, nil
}


// QueryGetLog returns log of certain
func (lc *LoggingClient) QueryGetEnnFLowLog(namespace, podName,scmContainerName,buildContainerName string, date time.Time,clusterID string) (*ESResponse, error) {
	method := "QueryGetEnnFLowLog"

	// 1. Build request url
	var url string

	url = fmt.Sprintf("%s/%s/_search?scroll=1m&pretty&ignore_unavailable=true", strings.Trim(lc.URL, "/"), indexPrefix+date.Format("2006.01.02"))


	// 2. build request body
	requestBody := fmt.Sprintf(logCiCdRequest, podName, namespace, clusterID,scmContainerName,buildContainerName, defaultLogSize)

	glog.V(2).Infoln(method, "url:", url)
	glog.V(4).Infoln(method, "body:", requestBody)

	// send request
	rawBytes, statusCode, err := lc.httpRequest("GET", url, requestBody)
	if err != nil {
		netError, ok := err.(net.Error)
		// if timeout or 404, then return empty result
		if ok && netError.Timeout() {
			glog.Errorln(method, "timeout error", err)
			resp := &ESResponse{}
			return resp, nil
		} else if statusCode == 404 {
			glog.Errorln(method, "not fround 404 error", err)
			resp := &ESResponse{}
			return resp, nil
		}
		glog.Errorln(method, "Get log failed, namespace:", namespace, ", name:", podName, ", error:", err)
		return nil, err
	}

	//glog.V(1).Info(string(rawBytes))
	// unmarshal response
	response := &ESResponse{}
	if err := json.Unmarshal(rawBytes, response); err != nil {
		errInfo := strings.Trim(string(rawBytes), "\n \t")
		glog.Errorln(method, "Unmarshal response failed, data:", errInfo, ", error:", err)
		return nil, err
	}
	if response.Status != 0 {
		return nil, fmt.Errorf("invalid keyword: %s", err)
	}

	return response, nil
}



// QueryGetLog returns log of certain
func (lc *LoggingClient) ScrollRestLogs(scrollId, podName string) (*ESResponse, error) {
	method := "LoggingClient.ScrollRestLogs"

	// 1. Build request url
	var url string

	url = fmt.Sprintf("%s/_search?scroll", strings.Trim(lc.URL, "/"))

	// 2. build request body
	scrollRest:=`
	{
      "scroll": "1m",
      "scroll_id": %s
    }
	`
	requestBody := fmt.Sprintf(scrollRest, scrollId)

	glog.V(2).Infoln(method, "url:", url)
	glog.V(4).Infoln(method, "body:", requestBody)

	// send request
	rawBytes, statusCode, err := lc.httpRequest("GET", url, requestBody)
	if err != nil {
		netError, ok := err.(net.Error)
		// if timeout or 404, then return empty result
		if ok && netError.Timeout() {
			glog.Errorln(method, "timeout error", err)
			resp := &ESResponse{}
			return resp, nil
		} else if statusCode == 404 {
			glog.Errorln(method, "not fround 404 error", err)
			resp := &ESResponse{}
			return resp, nil
		}
		glog.Errorln(method, "Get log failed", ", name:", podName, ", error:", err)
		return nil, err
	}

	glog.V(1).Info(string(rawBytes))
	// unmarshal response
	response := &ESResponse{}
	if err := json.Unmarshal(rawBytes, response); err != nil {
		errInfo := strings.Trim(string(rawBytes), "\n \t")
		glog.Errorln(method, "Unmarshal response failed, data:", errInfo, ", error:", err)
		return nil, fmt.Errorf("%s", rawBytes)
	}
	if response.Status != 0 {
		return nil, fmt.Errorf("invalid keyword: %s", err)
	}

	return response, nil
}

// days2indices convert date array: 2016-8-8,2016-8-9,2016-8-10
// into elasticsearch indices array:
// "logstash-2016.08.08","logstash-2016.08.09","logstash-2016.08.10"
func days2indices(days []time.Time) (indices []string) {
	// for _, days := range days {
	len_days := len(days)
	indices = append(indices, days[0].Format("2006-01-02T15:04:05+07:00"))
	indices = append(indices, days[len_days-1].Format("2006-01-02T15:04:05+07:00"))
	return
}

// getIndexBydate ...
func getIndexBydate(days []time.Time) (indexs []string) {
	/*for _, day := range days {
		indexs = append(indexs, fmt.Sprintf("%s%4d.%02d.%02d", indexPrefix, day.Year(), day.Month(), day.Day()))
	}*/
	if len(days) == 0 {
		return []string{"logstash-*"}
	}

	startDay := days[0].Format("2006.01.02")
	endDay := days[len(days)-1].Format("2006.01.02")
	if startDay == endDay {
		return []string{fmt.Sprintf("%s%s", indexPrefix, startDay)}
	}

	return []string{"logstash-*"}
}
