package controllers

import (
	"dev-flows-api-golang/models"
	"dev-flows-api-golang/util/uuid"
	"encoding/json"
	"github.com/golang/glog"
)

type CiImagesController struct {
	BaseController
}

//@router /:id [PUT]
func (ciImage *CiImagesController) UpdateBaseImage() {

	method := "CiImagesController.UpdateBaseImage"

	var image models.CiImages
	Id := ciImage.Ctx.Input.Param(":id")

	ciImage.Audit.SetResourceID(Id)
	ciImage.Audit.SetResourceName(Id)
	ciImage.Audit.SetOperationType(models.AuditOperationUpdate)
	ciImage.Audit.SetResourceType(models.AuditResourceCIImages)

	contet := string(ciImage.Ctx.Input.RequestBody)

	if contet == "" {
		ciImage.ResponseErrorAndCode("the script content is empty", 402)
		return
	}
	err := json.Unmarshal(ciImage.Ctx.Input.RequestBody, &image)
	if err != nil {
		glog.Errorf("%s %v\n", method, err)
		ciImage.ResponseErrorAndCode("the request body is empty", 501)
		return
	}

	if ciImage.User.Role == 2 {
		image.IsSystem = 1
	} else {
		image.IsSystem = 0
	}

	image.CategoryName = GetCategoryName(int(image.CategoryId))

	ciImageModel := models.NewCiImages()

	if ciImage.User.Role == 2 {

		ciImageModel.UpdateBaseImageById(Id, image)

	} else {

		ciImageModel.UpdateBaseImage(Id, ciImage.Namespace, image)

	}

	image.Id = Id
	ciImage.ResponseSuccessCIRuleDevops(image)
	return

}

//@router / [POST]
func (ciImage *CiImagesController) CreateNewBaseImage() {

	method := "CiImagesController.CreateNewBaseImage"
	image := models.CiImages{}
	image.Id = uuid.GetCIMID()
	ciImage.Audit.SetResourceID(image.Id)
	ciImage.Audit.SetResourceName(image.Id)
	ciImage.Audit.SetOperationType(models.AuditOperationCreate)
	ciImage.Audit.SetResourceType(models.AuditResourceCIImages)

	contet := string(ciImage.Ctx.Input.RequestBody)
	if contet == "" {
		ciImage.ResponseErrorAndCode("the request body is empty", 402)
		return
	}

	err := json.Unmarshal(ciImage.Ctx.Input.RequestBody, &image)
	if err != nil {
		glog.Errorf("%s %v\n", method, err)
		ciImage.ResponseErrorAndCode("the request body is empty", 501)
		return
	}

	image.Namespace = ciImage.Namespace

	if ciImage.User.Role == 2 {
		image.IsSystem = 1
	} else {
		image.IsSystem = 0
	}

	image.CategoryName = GetCategoryName(int(image.CategoryId))
	image.IsAllowDeletion = 0

	_, err = models.NewCiImages().CreateNewBaseImage(image)
	if err != nil {
		glog.Errorf("%s %v\n", method, err)
		ciImage.ResponseErrorAndCode("the request body is empty", 502)
		return
	}

	ciImage.ResponseSuccessCIRuleDevops(image)
	return

}

//@router / [GET]
func (ciImage *CiImagesController) GetAvailableImages() {
	method := "CiImagesController.GetAvailableImages"

	images, total, err := models.NewCiImages().GetImagesByNamespace(ciImage.Namespace)
	if err != nil {
		glog.Errorf("%s %v\n", method, err)
		ciImage.ResponseErrorAndCode("GetAvailableImages failed ", 502)
		return
	}

	ciImage.ResponseSuccessDevops(images, total)
	return

}

//@router /:id [DELETE]
func (ciImage *CiImagesController) DeleteBaseImage() {
	method := "CiImagesController.DeleteBaseImage"
	Id := ciImage.Ctx.Input.Param(":id")
	err := models.NewCiImages().DeleteImage(Id, ciImage.Namespace)
	ciImage.Audit.SetResourceID(Id)
	ciImage.Audit.SetResourceName(Id)
	ciImage.Audit.SetOperationType(models.AuditOperationDelete)
	ciImage.Audit.SetResourceType(models.AuditResourceCIImages)

	if err != nil {
		glog.Errorf("%s %v\n", method, err)
		ciImage.ResponseErrorAndCode(method+" DeleteBaseImage failed ", 502)
		return
	}

	ciImage.ResponseSuccessCIRuleDevops("")
	return

}
func GetCategoryName(category_id int) string {
	switch category_id {
	case 1:
		return "单元测试"
	case 2:
		return "代码编译"
	case 3:
		return "构建镜像"
	case 4:
		return "集成测试"
	}
	return ""
}
