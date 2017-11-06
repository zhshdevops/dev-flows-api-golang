package proxy

import (
	"time"
	"api-server/models/common"
	"github.com/astaxie/beego/orm"
	sqlstatus "api-server/models/sql/status"
	"fmt"
	"github.com/golang/glog"
)


type ProxyHost struct {
	ID        string       `orm:"pk;column(id)"`
	Address       string    `orm:"size(64);column(address)"`
	Host     string    `orm:"size(64);column(host)"`
	Instance string    `orm:"size(48);column(instance)"`
	GroupID string    `orm:"size(48);column(group_id)"`
	ClusterID string    `orm:"size(48);column(cluster_id)"`
	CreateAt  time.Time `orm:"column(create_at)"`
}

type ProxyHostDTO struct {
	ID        string        `json:"id,omitempty"`
	Address       string     `json:"address,omitempty"`
	Host     string     `json:"host,omitempty"`
	Instance string      `json:"instance,omitempty"`
	GroupID string     `json:"group_id,omitempty"`
	ClusterID string      `json:"cluster_id,omitempty"`
	CreateAt  time.Time   `json:"create_at,omitempty"`
}

func (*ProxyHost) TableName() string {
	return "tenx_service_proxy_host"
}

func NewProxyHostModel() *ProxyHost {
	return &ProxyHost{}
}

func (p *ProxyHost) ToDTO() *ProxyHostDTO {
	return &ProxyHostDTO{
		ID:p.ID,
		Address:p.Address,
		Host:p.Host,
		Instance:p.Instance,
		GroupID:p.GroupID,
		ClusterID:p.ClusterID,
		CreateAt:p.CreateAt,
	}
}


func (p *ProxyHostDTO) ToORM() *ProxyHost {
	return &ProxyHost{
		ID:p.ID,
		Address:p.Address,
		Host:p.Host,
		Instance:p.Instance,
		GroupID: p.GroupID,
		ClusterID:p.ClusterID,
		CreateAt:p.CreateAt,
	}
}

func (t *ProxyHost) Get() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where id=?;", t.TableName())
	err := o.Raw(sql, t.ID).QueryRow(t)
	return sqlstatus.ParseErrorCode(err)
}


func (u *ProxyHost) List(dataselect *common.DataSelectQuery) ([]ProxyHost, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT * FROM %s where %s %s %s;`, u.TableName(), dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var hostModels []ProxyHost
	_, err := o.Raw(sql).QueryRows(&hostModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}
	return hostModels,sqlstatus.SQLSuccess, nil
}

func (u *ProxyHost) ListByGroupId(groupid string) ([]ProxyHost, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT * FROM %s where group_id=?;`, u.TableName())
	var hostModels []ProxyHost
	_, err := o.Raw(sql,groupid).QueryRows(&hostModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}
	return hostModels,sqlstatus.SQLSuccess, nil
}

func ListHostDTOByGroupId(groupid string) ([]*ProxyHostDTO, error) {
	nodes := []*ProxyHostDTO{}
	hosts,_,err := NewProxyHostModel().ListByGroupId(groupid)
	if err != nil{
		glog.Error("ListDTOByGroupId err",err)
		return nodes,err
	}
	for _,h :=range hosts{
		nodes= append(nodes,h.ToDTO())
	}
	return nodes ,nil
}

func (t *ProxyHost) Insert() (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Insert(t)
	return sqlstatus.ParseErrorCode(err)
}

func (t *ProxyHost) Delete() (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Delete(t)
	return sqlstatus.ParseErrorCode(err)
}
func (t *ProxyHost) Update(cols ...string) error {

	o := orm.NewOrm()

	_, err := o.Update(t, cols...)
	return err
}