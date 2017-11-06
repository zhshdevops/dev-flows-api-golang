package models

import (
	"time"
	"github.com/astaxie/beego/orm"
	"fmt"
)

type CiRepos struct {
	Id                 int `orm:"pk;column(id)" json:"id"`
	UserId             int `orm:"column(user_id)" json:"userid"`
	Namespace          string `orm:"column(namespace)" json:"namespace"`
	RepoType           string `orm:"column(repo_type)" json:"repo_type"`
	AccessToken        string `orm:"column(access_token)" json:"access_token"`
	AccessUserName     string `orm:"column(access_user_name)" json:"access_user_name"`
	CreateTime         time.Time `orm:"column(create_time)" json:"create_time"`
	AccessRefreshToken string `orm:"column(access_refresh_token)" json:"access_refresh_token"`
	AccessTokenSecret  string `orm:"column(access_token_secret)" json:"access_token_secret"`
	UserInfo           string `orm:"column(user_info)" json:"user_info"`
	RepoList           string `orm:"column(repo_list)" json:"repo_list"`
	GitlabUrl          string `orm:"column(gitlab_url)" json:"gitlab_url"`
	IsEncrypt          int8 `orm:"column(is_encrypt)" json:"is_encrypt"`
}

func RepoTypeToDepot(repoType string) string {
	Result := ""
	switch repoType {

	case "1":
		Result = "github"
	case "2":
		Result = "bitbucket"
	case "3":
		Result = "tce"
	case "4":
		Result = "gitcafe"
	case "5":
		Result = "coding"
	case "6":
		Result = "gitlab"
	case "7":
		Result = "svn"
	case "8":
		Result = "gogs"
	default:
		Result = Result

	}
	return Result

}

func DepotToRepoType(depot string) string {
	Result := ""
	switch depot {
	case "github":
		Result = "1"
	case "bitbucket":
		Result = "2"
	case "tce":
		Result = "3"
	case "gitcafe":
		Result = "4"
	case "coding":
		Result = "5"
	case "gitlab":
		Result = "6"
	case "svn":
		Result = "7"
	case "gogs":
		Result = "8"
	default:
		Result = depot
	}

	return Result
}

func (ci *CiRepos) TableName() string {
	return "tenx_ci_repos"
}

func NewCiRepos() *CiRepos {
	return &CiRepos{}
}
func (ci *CiRepos) FindOneRepoToken(namespace, repoType string) error {

	o := orm.NewOrm()

	sql := fmt.Sprintf("select * from %s where namespace=? and repo_type=?", ci.TableName())

	return o.Raw(sql, namespace, repoType).QueryRow(&ci)

}

func (ci *CiRepos) GetGitlabRepo(namespace, repoType string) error {

	repoTypeNumber := DepotToRepoType(repoType)

	if repoTypeNumber == "" {

		return fmt.Errorf("%s", "repoType is null")
	}

	o := orm.NewOrm()

	sql := fmt.Sprintf("select * from %s where namespace=? and repo_type=?", ci.TableName())

	return o.Raw(sql, namespace, repoTypeNumber).QueryRow(&ci)

}

func (ci *CiRepos) FindOneRepo(namespace, repoType string) error {

	o := orm.NewOrm()

	sql := fmt.Sprintf("select * from %s where namespace=? and repo_type=?", ci.TableName())

	return o.Raw(sql, namespace, repoType).QueryRow(&ci)

}
func (ci *CiRepos) DeleteOneRepo(namespace, repoType string, orms ...orm.Ormer)(int64,error)  {

	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	return  o.QueryTable(ci.TableName()).Filter("namespace", namespace).Filter("repo_type", repoType).Delete()

}

func (ci *CiRepos) UpdateOneRepo(namespace, repoType string, repoList string, orms ...orm.Ormer) (int64,error) {

	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}

	return o.QueryTable(ci.TableName()).Filter("namespace", namespace).Filter("repo_type", repoType).Update(
		orm.Params{
			"repo_list": repoList,
		})

}

func (ci *CiRepos) CreateOneRepo(repo CiRepos, orms ...orm.Ormer) (int64, error) {

	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	return o.Insert(&repo)

}
