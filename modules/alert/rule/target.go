/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-06  @author zhangyongkang
 */
package rule

import (
	"api-server/models/alert"
	"fmt"
)

const (
	ethName = "em.*|eth.*|en.*"

	GBi = 1024 * 1024 * 1024
	MBi = 1024 * 1024
	KBi = 1024
)

var (
	metricNameMapping = map[string]string{
		"cpu/usage_rate":  "container_cpu_usage_seconds_total",
		"memory/usage":    "container_memory_usage_bytes",
		"network/tx_rate": "container_network_transmit_bytes_total",
		"network/rx_rate": "container_network_receive_bytes_total",
	}
)

func FindMetricName(metricName string) (string, error) {
	v, ok := metricNameMapping[metricName]
	if !ok {
		return "", fmt.Errorf("metric %s not supported yet", metricName)
	}
	return v, nil
}

// BuildTarget build metric target
func BuildTarget(targetType int, targetName, metricName, namespace string, rate int64) (string, error) {
	var (
		target     string
		metricItem string
		err        error
	)
	if targetType != alert.StrategyTypeNode && targetType != alert.StrategyTypeService {
		return "", fmt.Errorf("invalid argument targetType %d", targetType)
	}
	err = checkNil(map[string]string{
		"targetName": targetName,
		"metricName": metricName,
	})
	if err != nil {
		return "", err
	}
	if targetType == alert.StrategyTypeService {
		err = checkNil(map[string]string{
			"namespace": namespace,
		})
		if err != nil {
			return "", err
		}
	}
	if metricItem, err = FindMetricName(metricName); err != nil {
		return "", err
	}
	switch targetType {
	case alert.StrategyTypeNode:
		target = buildNodeTarget(targetName, metricItem, rate)
	case alert.StrategyTypeService:
		target = buildPodTarget(namespace, targetName, metricItem, rate)
	}
	if target == "" {
		return "", fmt.Errorf("empty target with targetName: %s, type: %d, metricName: %s, metricItem: %s",
			targetName, targetType, metricName, metricItem)
	}
	target = fmt.Sprintf("ceil(%s * 100)/100", target)
	return target, nil
}

func buildNodeTarget(nodeName, metricItem string, rate int64) string {
	var target string
	switch metricItem {
	case "container_memory_usage_bytes":
		target = fmt.Sprintf(`%s{instance="%s",job="kubernetes-nodes",kubernetes_io_hostname="%s",id="/"} / %d`,
			metricItem, nodeName, nodeName, rate)
	case "container_cpu_usage_seconds_total":
		target = fmt.Sprintf(`avg(rate(%s{instance="%s",job="kubernetes-nodes",kubernetes_io_hostname="%s",id="/"}[5m])) * 100`,
			metricItem, nodeName, nodeName)
	case "container_network_transmit_bytes_total", "container_network_receive_bytes_total":
		target = fmt.Sprintf(`sum(rate(%s{instance="%s",job="kubernetes-nodes",kubernetes_io_hostname="%s",id="/", interface=~"%s"}[5m])) / %d`,
			metricItem, nodeName, nodeName, ethName, rate)
	}
	return target
}

func buildPodTarget(namespace, serviceName, metricItem string, rate int64) string {
	var target string
	switch metricItem {
	case "container_memory_usage_bytes":
		target = fmt.Sprintf(`sum(%s{namespace="%s",pod_name=~"%s"}) / %d`,
			metricItem, namespace, podNameMatcher(serviceName), rate)
	case "container_cpu_usage_seconds_total":
		target = fmt.Sprintf(`sum(rate(%s{namespace="%s",pod_name=~"%s"}[5m])) * 100`,
			metricItem, namespace, podNameMatcher(serviceName))
	case "container_network_transmit_bytes_total", "container_network_receive_bytes_total":
		target = fmt.Sprintf(`sum(rate(%s{namespace="%s",pod_name=~"%s"}[5m])) / %d`,
			metricItem, namespace, podNameMatcher(serviceName), rate)
	}
	return target
}

func podNameMatcher(serviceName string) string {
	return fmt.Sprintf(`^%s-[0-9]{5,15}-[0-9a-zA-Z]{5}$`, serviceName)
}

func checkNil(params map[string]string) error {
	for k, v := range params {
		if v == "" {
			return fmt.Errorf("param %s is empty", k)
		}
	}
	return nil
}
