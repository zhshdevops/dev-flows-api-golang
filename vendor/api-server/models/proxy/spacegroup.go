package proxy

import (
	"github.com/astaxie/beego/orm"
	"fmt"
	sqlstatus "api-server/models/sql/status"
	"github.com/golang/glog"
)

type ProxySpace struct {
	ID        string       `orm:"pk;column(id)"`
	SpaceName       string    `orm:"size(64);column(space_name)"`
	GroupId     string    `orm:"size(64);column(group_id)"`
	ClusterID string    `orm:"size(48);column(cluster_id)"`

}

type ProxySpaceDTO struct {
	ID        string       `json:"id,omitempty"`
	SpaceName       string   `json:"space_name,omitempty"`
	GroupId     string    `json:"group_id,omitempty"`
	ClusterID string    `json:"cluster_id,omitempty"`

}


func (p *ProxySpace) ToDTO() *ProxySpaceDTO {
	return &ProxySpaceDTO{
		ID:p.ID,
		SpaceName:p.SpaceName,
		GroupId:p.GroupId,
		ClusterID:p.ClusterID,
	}
}


func (p *ProxySpaceDTO) ToORM() *ProxySpace {
	return &ProxySpace{
		ID:p.ID,
		SpaceName:p.SpaceName,
		GroupId:p.GroupId,
		ClusterID:p.ClusterID,
	}
}

func (ProxySpace) TableName() string {
	return "tenx_service_proxy_space"
}

func NewProxySpaceModel() *ProxySpace {
	return &ProxySpace{}
}

func (t *ProxySpace) Get() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where id=?;", t.TableName())
	err := o.Raw(sql, t.ID).QueryRow(t)
	return sqlstatus.ParseErrorCode(err)
}

func (t *ProxySpace) Insert() (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Insert(t)
	return sqlstatus.ParseErrorCode(err)
}

func (t *ProxySpace) Delete() (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Delete(t)
	return sqlstatus.ParseErrorCode(err)
}
func (t *ProxySpace) Update(cols ...string) error {

	o := orm.NewOrm()

	_, err := o.Update(t, cols...)
	return err
}

func  ListByGroupId(groupid string) ([]ProxySpace, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT * FROM %s where group_id=?;`, NewProxySpaceModel().TableName())
	var spaceModels []ProxySpace
	_, err := o.Raw(sql,groupid).QueryRows(&spaceModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}
	return spaceModels,sqlstatus.SQLSuccess, nil
}

func ListSpaceDTOByGroupId(groupid string) ([]*ProxySpaceDTO, error) {
	spaces := []*ProxySpaceDTO{}
	ss,_,err := ListByGroupId(groupid)
	if err != nil{
		glog.Error("ListDTOByGroupId err",err)
		return spaces,err
	}
	for _,h :=range ss{
		spaces = append(spaces,h.ToDTO())
	}
	return spaces,nil
}