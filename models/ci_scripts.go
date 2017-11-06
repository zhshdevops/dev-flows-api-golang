package models

import (
	"github.com/astaxie/beego/orm"
	"fmt"
)

type CiScripts struct {
	ID      string    `orm:"column(id);pk" json:"id"`
	Content string `orm:"column(content)" json:"content"`
}

func NewCiScripts() *CiScripts {
	return &CiScripts{}
}
func (cs *CiScripts) TableName() string {
	return "tenx_ci_scripts"
}

func (cs *CiScripts) AddScript(id, content string, orms ...orm.Ormer) (int64, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	cs.ID = id
	cs.Content = content
	return o.Insert(cs)
}

func (cs *CiScripts) GetScriptByID(id string) error {
	o := orm.NewOrm()
	cs.ID = id
	return o.Read(cs)
}

func (cs *CiScripts) DeleteScriptByID(id string, orms ...orm.Ormer) (int64, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	cs.ID = id
	return o.Delete(cs)
}

func (cs *CiScripts) UpdateScriptByID(id, script string, orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	var ciScripts CiScripts
	ciScripts.ID = id
	ciScripts.Content = script
	num, err := o.Update(&ciScripts, "content")
	if err != nil || num < 1 {
		return fmt.Errorf("UpdateScriptByID failed %d  %v", num, err)
	}
	return nil
}
