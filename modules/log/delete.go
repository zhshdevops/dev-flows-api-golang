/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2017-05-05  @author Lei
 */

package log

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/golang/glog"
)

var (
	allIndex             = "_all"
	logDeleteRequestbody = `{
    "query": {
      "bool": {
        "must": [
          { "match" : {"kubernetes.namespace_name" : "%s"} }
        ],
        "should": [
          %s
        ],
        "minimum_should_match": 1
      }
    }
  }`
)

// QueryGetLog returns log of certain pod/service
func (lc *LoggingClient) DeleteLog(namespace string, serviceNames []string) (*ESResponse, error) {
	method := "DeleteLog"

	// 1. Build request url
	var url string
	url = fmt.Sprintf("%s/%s/_query?pretty", strings.Trim(lc.URL, "/"), allIndex)

	// 2. Build request body
	serviceMatch := ""
	ruleMatch := `{ "match" : {"kubernetes.container_name" : "%s"} }`
	for index, name := range serviceNames {
		serviceMatch = serviceMatch + fmt.Sprintf(ruleMatch, name)
		if index < len(serviceNames)-1 {
			serviceMatch += ","
		}
	}

	requestBody := fmt.Sprintf(logDeleteRequestbody, namespace, serviceMatch)

	glog.V(2).Infoln(method, "url:", url)
	glog.V(4).Infoln(method, "body:", requestBody)

	// send request
	rawBytes, statusCode, err := lc.httpRequest("DELETE", url, requestBody)
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
		glog.Errorln(method, "Get log failed, namespace:", namespace, ", service name:", serviceNames, "error:", err)
		return nil, err
	}

	// Unmarshal response
	response := &ESResponse{}
	if err := json.Unmarshal(rawBytes, response); err != nil {
		errInfo := strings.Trim(string(rawBytes), "\n \t")
		glog.Errorln(method, "Unmarshal response failed, data:", errInfo, ", error:", err)
		return nil, fmt.Errorf("%s", rawBytes)
	}
	return response, nil
}
