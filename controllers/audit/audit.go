/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-10-28  @author liuyang
 */

package audit

import (
	"encoding/json"
	"time"

	"github.com/golang/glog"

	"dev-flows-api-golang/controllers"
	"dev-flows-api-golang/models"
)

// AuditController audit controller
type AuditController struct {
	controllers.BaseController
}

// List list audit records by namespace, operation type, resource type, start and end time
// @Success 200 success
// @router /logs [post]
func (c *AuditController) List() {
	method := "controllers/audit.AuditController.List"
	timeLayout := "2006-1-2 15:4:5"

	param := struct {
		From      int                   `json:"from"`
		Size      int                   `json:"size"`
		Namespace string                `json:"namespace"`
		Operation models.AuditOperation `json:"operation"`
		Resource  models.AuditResource  `json:"resource"`
		StartTime string                `json:"start_time"`
		EndTime   string                `json:"end_time"`
		Status    string                `json:"status"`
		Keyword   string                `json:"keyword"`
		Omitempty bool                  `json:"omitempty"`

		startTime time.Time
		endTime   time.Time
	}{}

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &param)
	if err != nil {
		glog.Errorln(method, "failed to parse request body", err)
		c.ErrorBadRequest("Invalid Request Body", nil)
		return
	}

	now := time.Now()

	if param.StartTime != "" {
		param.startTime, err = time.ParseInLocation(timeLayout, param.StartTime, now.Location())
		if err != nil {
			glog.Errorln(method, "failed to parse start time", err)
			c.ErrorBadRequestWithField("Failed to parse start time", "StartTime")
			return
		}
	}

	if param.EndTime != "" {
		param.endTime, err = time.ParseInLocation(timeLayout, param.EndTime, now.Location())
		if err != nil {
			glog.Errorln(method, "failed to parse end time", err)
			c.ErrorBadRequestWithField("Failed to parse end time", "EndTime")
			return
		}
	}

	if param.Size < 0 {
		param.Size = 50
	}

	audit := &models.AuditRecord{
		Namespace:     c.Namespace,
		OperationType: param.Operation,
		ResourceType:  param.Resource,
	}
	count, records, err := audit.ListWithCount(param.From, param.Size, param.startTime, param.endTime, param.Status, param.Keyword, param.Omitempty)
	if err != nil {
		glog.Errorln(method, "failed to query db", err)
		c.ErrorInternalServerError(err)
		return
	}

	response := &struct {
		Count   int64 `json:"count"`
		Records *[]models.AuditRecord
	}{
		Count:   count,
		Records: records,
	}

	c.ResponseSuccess(response)
}
