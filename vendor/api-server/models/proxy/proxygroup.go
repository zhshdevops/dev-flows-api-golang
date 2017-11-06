/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2017-05-04  @author lizhen
 */

package proxy

import (
	"api-server/models/common"
	"github.com/astaxie/beego/orm"
	sqlstatus "api-server/models/sql/status"
	"time"
	"fmt"
	"github.com/golang/glog"
)

//const createTable = `CREATE TABLE IF NOT EXISTS tenx_labels (
//  id INT AUTO_INCREMENT PRIMARY KEY,
//  label VARCHAR(64) NOT NULL COMMENT 'key of label',
//  value VARCHAR(64) NOT NULL COMMENT 'value of label',
//  target VARCHAR(32) NOT NULL COMMENT 'target type, eg. node, pod, service...',
//  cluster_id VARCHAR(48) NOT NULL COMMENT 'tenx_clusters.id',
//  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
//  create_by INT NOT NULL COMMENT 'tenx_users.user_id',
//  CONSTRAINT UNIQUE INDEX USING HASH (label, value, target, cluster_id)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8;`



type ProxyGroup struct {
	ID        string       `orm:"pk;column(id)"`
	Address       string    `orm:"size(64);column(address)"`
	Domain     string    `orm:"size(64);column(domain)"`
	Type    string    `orm:"size(32);column(type)"`
	Name string    `orm:"size(48);column(name)"`
	Lb string    `orm:"size(48);column(lb)"`
	ClusterID string    `orm:"size(48);column(cluster_id)"`
	IsDefaut int    `orm:"size(4);column(is_default)"`
	CreateAt  time.Time `orm:"column(create_at)"`
	Nodes  []*ProxyHost  `orm:"-"`
	Spaces  []*ProxySpace  `orm:"-"`
}



type ProxyGroupDTO struct {
	ID        string       `json:"id,omitempty"`
	Address       string     `json:"address,omitempty"`
	Domain     string     `json:"domain,omitempty"`
	Type    string     `json:"type,omitempty"`
	Name string     `json:"name,omitempty"`
	Lb string     `json:"lb,omitempty"`
	ClusterID string     `json:"cluster_id,omitempty"`
	IsDefaut int   `json:"is_default,omitempty"`
	CreateAt  time.Time `json:"create_at,omitempty"`
	Nodes  []*ProxyHostDTO `json:"nodes,omitempty"`
	Spaces  []*ProxySpaceDTO `json:"spaces,omitempty"`
}


func (*ProxyGroup) TableName() string {
	return "tenx_service_proxy_group"
}

func NewProxyGroupModel() *ProxyGroup {
	return &ProxyGroup{}
}

func (p *ProxyGroupDTO)Equal(x *ProxyGroupDTO) bool {
	if p.ID == x.ID && p.Address == x.Address && p.Domain==x.Domain && p.IsDefaut==x.IsDefaut&& p.Name==x.Name {

		return true
	}
	return false
}
func CompareNode(old ,new *ProxyGroupDTO)(bool,map[string]*ProxyHostDTO,map[string]*ProxyHostDTO,map[string]*ProxyHostDTO){
	oldMap := make(map[string]*ProxyHostDTO,len(old.Nodes))
	newMap := make(map[string]*ProxyHostDTO,len(new.Nodes))
	updateMap := make(map[string]*ProxyHostDTO,len(old.Nodes))
	for _ , o := range  old.Nodes{
		oldMap[o.Host]=o
	}
	for _ , n := range  new.Nodes{
		newMap[n.Host]=n
	}

	for oldk,oldv :=range oldMap{

		if v ,ok := newMap[oldk]; ok{
			glog.Infoln(*oldv,*newMap[oldk])
			if v.Address != oldv.Address || v.Instance != oldv.Instance{
				updateMap[oldk]= v
			}
			delete(oldMap,oldk)
			delete(newMap,oldk)
		}
	}
	if len(oldMap) == 0 && len(newMap)== 0 && len(updateMap) == 0 {
		return true , nil,nil,nil
	}
	return false, newMap, updateMap , oldMap
}


func CompareSpace(old ,new *ProxyGroupDTO)(bool,map[string]*ProxySpaceDTO,map[string]*ProxySpaceDTO){
	oldMap := make(map[string]*ProxySpaceDTO,len(old.Spaces))
	newMap := make(map[string]*ProxySpaceDTO,len(new.Spaces))

	for _ , o := range  old.Spaces{
		oldMap[o.SpaceName]=o
	}
	for _ , n := range  new.Spaces{
		newMap[n.SpaceName]=n
	}

	for oldk,_ :=range oldMap{
		if _ ,ok := newMap[oldk]; ok{
			delete(oldMap,oldk)
			delete(newMap,oldk)
		}
	}
	if len(oldMap) == 0 && len(newMap)== 0{
		return true , nil ,nil
	}
	return false, newMap , oldMap
}
func (p *ProxyGroup) ToDTO() *ProxyGroupDTO {
	 dto := ProxyGroupDTO{
		ID:p.ID,
		Address:p.Address,
		Domain:p.Domain,
		Type:p.Type,
		Name:p.Name,
		Lb:p.Lb,
		ClusterID:p.ClusterID,
		IsDefaut:p.IsDefaut,
		CreateAt:p.CreateAt,
	}
	for _,n :=range p.Nodes{
		dto.Nodes = append(dto.Nodes,n.ToDTO())
	}
	return &dto
}


func (p *ProxyGroupDTO) ToORM() *ProxyGroup {
	 orm := ProxyGroup{
		ID:p.ID,
		Address:p.Address,
		Domain:p.Domain,
		Type:p.Type,
		Name:p.Name,
		Lb:p.Lb,
		ClusterID:p.ClusterID,
		IsDefaut:p.IsDefaut,
		CreateAt:p.CreateAt,
	}

	for _,n :=range p.Nodes{
		orm.Nodes = append(orm.Nodes,n.ToORM())
	}
	for _,s :=range p.Spaces{
		orm.Spaces = append(orm.Spaces,s.ToORM())
	}

	return &orm
}



func NodeSliceEqual(a, b []*ProxyHostDTO) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for i, v := range a {
		if *v != *b[i] {
			return false
		}
	}

	return true
}


func (t *ProxyGroup) Get() (uint32, error) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where id=?;", t.TableName())
	err := o.Raw(sql, t.ID).QueryRow(t)
	return sqlstatus.ParseErrorCode(err)
}


func (u *ProxyGroup) List(dataselect *common.DataSelectQuery) ([]*ProxyGroupDTO, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT * FROM %s where %s %s %s;`, u.TableName(), dataselect.FilterQuery, dataselect.SortQuery, dataselect.PaginationQuery)
	var groupModels []ProxyGroup
	_, err := o.Raw(sql).QueryRows(&groupModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}
	var groups []*ProxyGroupDTO
	for _, g := range groupModels{
		gdto := g.ToDTO()
		nodes,err := ListHostDTOByGroupId(g.ID)
		if err !=nil{
			return nil,sqlstatus.SQLErrInternalErr,err
		}
		gdto.Nodes= nodes

		spaces,err := ListSpaceDTOByGroupId(g.ID)
		if err !=nil{
			return nil,sqlstatus.SQLErrInternalErr,err
		}
		gdto.Spaces= spaces

		groups = append(groups,gdto)
	}


	return groups,sqlstatus.SQLSuccess, nil
}

func (u *ProxyGroup) ListByClusterId(ClusterId string) ([]*ProxyGroupDTO, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`SELECT * FROM %s where cluster_id=?;`, u.TableName())
	var groupModels []ProxyGroup
	_, err := o.Raw(sql,ClusterId).QueryRows(&groupModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}
	var groups []*ProxyGroupDTO
	for _, g := range groupModels{
		gdto := g.ToDTO()
		nodes,err := ListHostDTOByGroupId(g.ID)
		if err !=nil{
			return nil,sqlstatus.SQLErrInternalErr,err
		}
		gdto.Nodes= nodes
		groups = append(groups,gdto)
	}

	return groups,sqlstatus.SQLSuccess, nil
}

func (u *ProxyGroup) ListByNamespace(ClusterId ,namespace string) ([]*ProxyGroupDTO, uint32, error) {
	o := orm.NewOrm()

	sql := fmt.Sprintf(`select g.id,g.is_default, g.address,g.cluster_id,g.domain,g.name,g.type from %s g
left JOIN tenx_service_proxy_space s ON g.id=s.group_id
where s.cluster_id =? and s.space_name=? or g.is_default=1 and g.cluster_id=? `, u.TableName())
	var groupModels []ProxyGroup
	_, err := o.Raw(sql,ClusterId,namespace,ClusterId).QueryRows(&groupModels)
	if err != nil {
		errCode, err := sqlstatus.ParseErrorCode(err)
		return nil, errCode, err
	}
	var groups []*ProxyGroupDTO
	for _, g := range groupModels{
		gdto := g.ToDTO()
		nodes,err := ListHostDTOByGroupId(g.ID)
		if err !=nil{
			return nil,sqlstatus.SQLErrInternalErr,err
		}
		gdto.Nodes= nodes
		groups = append(groups,gdto)
	}

	return groups,sqlstatus.SQLSuccess, nil
}

func (t *ProxyGroup) Insert() (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Insert(t)
	return sqlstatus.ParseErrorCode(err)
}

func (t *ProxyGroup) Delete() (uint32, error) {
	o := orm.NewOrm()
	_, err := o.Delete(t)
	for _,n := range t.Nodes{
		n.Delete()
	}
	for _,s := range t.Spaces{
		s.Delete()
	}
	return sqlstatus.ParseErrorCode(err)
}

func (t *ProxyGroup) Update(cols ...string) error {

	o := orm.NewOrm()

	_, err := o.Update(t, cols...)
	return err
}