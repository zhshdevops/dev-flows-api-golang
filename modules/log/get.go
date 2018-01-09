/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-18  @author liuyang
 */

package log

import "regexp"

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
        {"time_nano"  : {"order" : "asc"}}
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
