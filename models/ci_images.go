package models

import (
	"time"
	"github.com/astaxie/beego/orm"
	"fmt"
	"github.com/golang/glog"
)

type CiImages struct {
	Id              string    `orm:"pk;column(id)" json:"id"`
	ImageName       string `orm:"column(image_name)" json:"image_name"`
	ImageUrl        string `orm:"column(image_url)" json:"image_url"`
	Namespace       string `orm:"column(namespace)" json:"namespace"`
	CategoryId      int8 `orm:"column(category_id)" json:"category_id"`
	CategoryName    string `orm:"column(category_name)" json:"category_name"`
	IsSystem        int8 `orm:"column(is_system)" json:"is_system"`
	Description     string `orm:"column(description)" json:"description"`
	CreateTime      time.Time `orm:"column(create_time)" json:"create_time"`
	IsAllowDeletion int8 `orm:"column(is_allow_deletion)" json:"is_allow_deletion"`
}

type ImageList struct {
	ProjectId int `json:"projectId"`
	ImageName string `json:"imageName"`
}

func (ci *CiImages) TableName() string {
	return "tenx_ci_images"
}

func NewCiImages() *CiImages {
	return &CiImages{}
}

func (ci *CiImages) GetImagesByNamespace(namespace string) (images []CiImages, total int64, err error) {
	o := orm.NewOrm()
	sql := "SELECT id, image_name, image_url, category_id, category_name, is_system, description, is_allow_deletion " +
		"from tenx_ci_images where namespace = ? or is_system = 1 order by category_id"
	total, err = o.Raw(sql, namespace).QueryRows(&images)

	return

}

func (ci *CiImages) CreateNewBaseImage(imageInfo CiImages, orms ...orm.Ormer) (result int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	result, err = o.Insert(&imageInfo)
	return

}

func (ci *CiImages) UpdateBaseImage(id, namespace string, imageInfo CiImages, orms ...orm.Ormer) (result int64, err error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	result, err = o.QueryTable(ci.TableName()).Filter("id", id).Filter("namespace", namespace).Update(orm.Params{
		"category_id":   imageInfo.CategoryId,
		"image_name":    imageInfo.ImageName,
		"image_url":     imageInfo.ImageUrl,
		"category_name": imageInfo.CategoryName,
	})
	return

}

func (ci *CiImages) UpdateBaseImageById(id string, image CiImages, orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	_, err := o.QueryTable(ci.TableName()).Filter("id", id).Update(orm.Params{
		"category_id":   image.CategoryId,
		"image_name":    image.ImageName,
		"image_url":     image.ImageUrl,
		"category_name": image.CategoryName,
	})
	return err

}

func (ci *CiImages) DeleteImageById(id string, orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	res, err := o.Raw(fmt.Sprintf("delete from %s where id=? and is_allow_deletion=?", ci.TableName()), id, 0).
		Exec()
	if num, err := res.RowsAffected(); num >= 1 && err == nil {
		glog.Infof("delete DeleteImageById id=%s Success \n", id)
	}
	return err

}

func (ci *CiImages) DeleteImage(id, namespace string, orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	res, err := o.Raw(fmt.Sprintf("delete from %s where id=? and is_allow_deletion=? and namespace=?",
		ci.TableName()), id, 0, namespace).
		Exec()
	if num, err := res.RowsAffected(); num >= 1 && err == nil {
		glog.Infof("delete DeleteImage id=%s namespace=%s Success \n", id, namespace)
	}
	return err

}

func (ci *CiImages) IsValidImages(isDependent bool, namespace string, images string) bool {
	//method:="CiImages.IsValidImages"
	//sql:="SELECT 1 as count from tenx_ci_images where (namespace = ? or is_system = 1) and image_url in (?) and category_id < 100"
	//
	//if isDependent{
	//	sql="SELECT 1 as count from tenx_ci_images where (namespace = ? or is_system = 1) and image_url in (?) and category_id > 100"
	//}
	//glog.Infof("%s sql= [%s]\n",method,sql)
	cond := orm.NewCondition()
	cond.Or("namespace", namespace).Or("is_system", 1)
	o := orm.NewOrm()
	if isDependent {
		return o.QueryTable(ci.TableName()).Filter("image_url__in", images).
			Filter("category_id__lt", 100).SetCond(cond).Exist()
	}

	return o.QueryTable(ci.TableName()).Filter("image_url__in", images).
		Filter("category_id__gt", 100).SetCond(cond).Exist()

}
