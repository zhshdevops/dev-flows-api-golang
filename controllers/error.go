package controllers

import (
	"encoding/json"
	"net/http"
	"dev-flows-api-golang/modules/tenx/errors"
	"github.com/astaxie/beego"
	"github.com/golang/glog"
	"dev-flows-api-golang/models"
	"github.com/astaxie/beego/context"
)

// ErrorController error controller
type ErrorController struct {
	beego.Controller
	Audit models.AuditInfo
}

// CreateErrorController create an ErrorController with context information
func CreateErrorController(ctx *context.Context) *ErrorController {
	c := &ErrorController{}
	c.Ctx = ctx
	c.Audit = models.AuditInfo{}
	return c
}

func (e *ErrorController) ResponseSuccess(results interface{}) {
	body := errors.NewSuccessStatus(results)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}


func (e *ErrorController) ResponseSupportSuccess(results interface{}) {
	body := errors.NewSuccessStatus(results)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}



func (e *ErrorController) ResponseSuccessDevops(results interface{}, total int64) {
	body := errors.NewSuccessStatusDevops(results, total)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseSuccessStatusAndResultDevops(results interface{}) {
	body := errors.NewSuccessStatusAndResultDevops(results)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseManageProjectDevops(message, project_id, warnings interface{}, status int64) {
	body := errors.NewProjectManagedDevops(message, project_id, warnings, status)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	if http.StatusOK != status {
		e.writeResponseHeader(int(status))
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseSuccessCIRuleDevops(results interface{}) {
	body := errors.NewSuccessStatusCIRuleDevops(results)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseCreateSuccessCDRuleDevops(message,ruleId interface{}) {
	body := errors.NewSuccessStatusCDRuleDevops(message,ruleId)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseSuccessStatusAndMessageDevops(results interface{}) {
	body := errors.NewSuccessStatusAndMessageDevops(results)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseSuccessImageListDevops(results interface{}) {
	body := errors.NewSuccessStatusImageListDevops(results)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseSuccessFLowDevops(message string, flowId string) {
	body := errors.NewSuccessStatusFlowDevops(message, flowId)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) NewSuccessStatusFlowBuildIdDevops(message string, flowId string) {
	body := errors.NewSuccessStatusFlowDevops(message, flowId)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}



func (e *ErrorController) ResponseNotFoundDevops(results string) {
	body := errors.NewSuccessNotFoundDevops(results)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseErrorAndCode(results interface{}, status int) {
	body := errors.NewSuccessMessagesStatusDevops(results, status)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	if http.StatusOK != status {
		e.writeResponseHeader(status)
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseResultAndStatusDevops(results interface{}, status int) {
	body := errors.NewResultStatusDevops(results, status)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	if http.StatusOK != status {
		e.writeResponseHeader(status)
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseScriptAndStatusDevops(results string, status int) {
	body := errors.NewResultIdDevops(results, status)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	if http.StatusOK != status {
		e.writeResponseHeader(status)
	}
	e.writeResponseBody(resp)
}




func (e *ErrorController) ResponseMessageAndResultAndStatusDevops(results interface{}, message interface{}, status int) {
	body := errors.NewResultMessageStatusDevops(results, message, status)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	if http.StatusOK != status {
		e.writeResponseHeader(status)
	}
	e.writeResponseBody(resp)
}

func (e *ErrorController) ResponseError(body *errors.StatusError) {
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error 400 %v failed, error:%s\n", body, err)
	}
	if http.StatusOK != body.Code {
		e.writeResponseHeader(int(body.Code))
	}
	e.writeResponseBody(resp)
}

func (c *ErrorController) ErrorBadJsonBody(missedField string) {
	cause := errors.StatusCause{Message: "Missing field in body", Field: missedField}
	c.ErrorBadRequest("Missing parameter in body",
		&errors.StatusDetails{
			Causes: []errors.StatusCause{cause},
		})
}

// ErrorBadRequest creates an error that indicates that the request is invalid and can not be processed.
// code: 400
func (e *ErrorController) ErrorBadRequest(msg string, details *errors.StatusDetails) {
	body := errors.NewBadRequestErr(msg, details)

	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error 400 %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusBadRequest)
	e.writeResponseBody(resp)
}

// ErrorBadRequestWithField is a wrapper of ErrorBadRequest for convenience
func (e *ErrorController) ErrorBadRequestWithField(msg, field string) {
	details := &errors.StatusDetails{
		Causes: []errors.StatusCause{
			{Field: field},
		},
	}
	e.ErrorBadRequest(msg, details)
}

func (e *ErrorController) ErrorUnprocessable(msg, field string) {
	details := &errors.StatusDetails{
		Causes: []errors.StatusCause{
			{Field: field},
		},
	}
	body := errors.NewUnprocessableErr(msg, details)

	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error 400 %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusUnprocessableEntity)
	e.writeResponseBody(resp)
}

// ErrorUnauthorized defines the response with authorization failure.
// code: 401
func (e *ErrorController) ErrorUnauthorized() {
	body := errors.NewUnauthorizedErr()
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error 401 %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusUnauthorized)
	e.writeResponseBody(resp)
}

// ErrorPaymentRequired defines the response with forbidden failure.
// code: 402
func (e *ErrorController) ErrorPaymentRequired(kind, message string) {
	body := errors.NewPaymentErr(kind, message)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error 402 %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusPaymentRequired)
	e.writeResponseBody(resp)
}

// ErrorForbidden defines the response with forbidden failure.
// code: 403
func (e *ErrorController) ErrorForbidden(kind, name, message string) {
	body := errors.NewForbiddenErr(kind, name, message)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error 403 %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusForbidden)
	e.writeResponseBody(resp)
}

// ErrorForbidden defines the response with forbidden failure.
// code: 403
func (e *ErrorController) ErrorResourceIsNotEnouth(kind, message string, data interface{}) {
	body := errors.NewResourceIsNotEnough(kind, message, data)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error 403 %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusForbidden)
	e.writeResponseBody(resp)
}

// ErrorForbidden defines the response with forbidden failure.
// code: 451
// kind: license, trial, both
func (e *ErrorController) ErrorLicenseExpired(kind string) {
	body := errors.NewLicenseExpireErr(kind)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error 403 %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusUnavailableForLegalReasons)
	e.writeResponseBody(resp)
}

// ErrorNotFound defines the response which indicates that the resource of the kind and the name was not found.
// code: 404
func (e *ErrorController) ErrorNotFound(name, kind string) {
	body := errors.NewNotFoundErr(name, kind)

	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error 404 %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusNotFound)
	e.writeResponseBody(resp)
}

// ErrorNotAcceptable Param defines the response for invalid request params
// code: 406
func (e *ErrorController) ErrorNotAcceptable(name, kind string) {
	body := errors.NewNotAcceptableErr(name, kind)

	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusNotAcceptable)
	e.writeResponseBody(resp)
}

// ErrorAlreadyExist means the resource already exists (create operations)
// code: 409
func (e *ErrorController) ErrorAlreadyExist(name, kind string, err error) {
	body := errors.NewAlreadyExistErr(name, kind, err)

	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusConflict)
	e.writeResponseBody(resp)
}

// ErrorConflict means the resource conflicts with existing ones (for example update with different resource version)
// code: 409
func (e *ErrorController) ErrorConflict(name, kind string, err error) {
	body := errors.NewConflictErr(name, kind, err)

	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusConflict)
	e.writeResponseBody(resp)
}

// ErrorPreconditionFailed defines the response with forbidden failure.
// code: 402
func (e *ErrorController) ErrorPreconditionFailed(kind, level, message string) {
	body := errors.NewPreconditionFailedErr(kind, level, message)
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error 412 %v failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusPreconditionFailed)
	e.writeResponseBody(resp)
}

// ErrorInternalServerError defines the response for internal server error operations
// code: 500
func (e *ErrorController) ErrorInternalServerError(err error) {
	body := errors.NewInternalErr(err)

	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error %s failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusInternalServerError)
	e.writeResponseBody(resp)
}

//ResponseNoRegistry defines the response when user has no registry
func (e *ErrorController) ResponseNoRegistry(results string) {
	body := errors.StatusError{
		Status: errors.StatusSuccess,
		Code:   http.StatusNoContent,
		Reason: errors.StatusNotAcceptable,
		Data:   results,
	}
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal %v failed, error:%s\n", body, err)
		e.ErrorInternalServerError(err)
		return
	}
	e.writeResponseHeader(http.StatusNoContent)
	e.writeResponseBody(resp)
}

// ErrorTimeOut defines the response for timeout operations
// code: 500
func (e *ErrorController) ErrorTimeOut(kind, operation string, retryAfterSeconds int) {
	body := errors.NewTimeoutErr(kind, operation, retryAfterSeconds)

	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		glog.Errorf("Marshal error %s failed, error:%s\n", body, err)
	}
	e.writeResponseHeader(http.StatusInternalServerError)
	e.writeResponseBody(resp)
}

// ReplaceErrorResponse replace default error response
func ReplaceErrorResponse() {
	beego.ErrorHandler("404", error404)
}

// error404 default 404 error response
func error404(rw http.ResponseWriter, r *http.Request) {
	body := errors.StatusError{
		Status: errors.StatusFailure,
		Code:   http.StatusNotFound,
		Reason: errors.StatusReasonNotFound,
		Details: &errors.StatusDetails{
			Kind: "url",
		},
		Message: "Requested URL not found.",
	}
	resp, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		beego.Error("Error 404", err)
	}
	rw.WriteHeader(404)
	_, _ = rw.Write(resp)
}

// set audit info and write response header
// response status code is 200 by default, this function is mainly used for errors.
func (e *ErrorController) writeResponseHeader(code int) {
	e.Audit.HTTPStatusCode = code
	e.Ctx.ResponseWriter.WriteHeader(code)
}

// set audit info and write response body, if statusCode is 0, set it to 200
func (e *ErrorController) writeResponseBody(body []byte) {
	if e.Audit.HTTPStatusCode == 0 {
		e.Audit.HTTPStatusCode = http.StatusOK
	}
	e.Audit.ResponseBody = string(body)
	e.Ctx.ResponseWriter.Write(body)
}
