package models

import (
	"time"
	"github.com/astaxie/beego/orm"
	"dev-flows-api-golang/util/secure"
	"fmt"
	"errors"
	"github.com/golang/glog"
	sqlstatus "api-server/models/sql/status"
	"dev-flows-api-golang/ci/coderepo"
	"strconv"
)

type CiManagedProjects struct {
	Id              string `orm:"pk;column(id)" json:"id"`
	Name            string `orm:"column(name)" json:"name"`
	Owner           string `orm:"column(owner)" json:"owner"`
	Namespace       string `orm:"column(namespace)" json:"namespace"`
	IsPrivate       int8 `orm:"column(is_private)" json:"is_private"`
	RepoType        string `orm:"column(repo_type)" json:"repo_type"`
	SourceFullName  string `orm:"column(source_full_name)" json:"source_full_name"`
	Address         string `orm:"column(address)" json:"address"`
	GitlabProjectId string `orm:"column(gitlab_project_id)" json:"gitlab_project_id"`
	PrivateKey      string `orm:"column(private_key)" json:"private_key"`
	PublicKey       string `orm:"column(public_key)" json:"public_key"`
	DeployKeyId     int `orm:"column(deploy_key_id)" json:"deploy_key_id"`
	WebhookId       int `orm:"column(webhook_id)" json:"webhook_id"`
	WebhookUrl      string `orm:"column(webhook_url)" json:"webhook_url"`
	CreateTime      time.Time `orm:"column(create_time)" json:"create_time"`
	ProjectId       string `orm:"-" json:"projectId"`
	Username        string `orm:"-" json:"username"`
	Password        string `orm:"-" json:"password"`
}

type ActiveRepoReq struct {
	Address         string `json:"address"`
	Description     string `json:"description"`
	IsPrivate       int `json:"is_private"`
	Name            string `json:"name"`
	ProjectId       int `json:"projectId"`
	RepoType        string `json:"repo_type"`
	Private         int `json:"private"`
	GitlabProjectId int `json:"gitlab_project_id"`
	SourceFullName  string `json:"source_full_name"`
	Password        string `json:"password"`
	Username        string `json:"username"`
}

func NewCiManagedProjects() *CiManagedProjects {
	return &CiManagedProjects{}
}
func (ci *CiManagedProjects) TableName() string {

	return "tenx_ci_managed_projects"

}

func (ci *CiManagedProjects) ListProjects(namespace string) ([]*CiManagedProjects, int64, error) {

	var manageproject []*CiManagedProjects

	o := orm.NewOrm()

	total, err := o.QueryTable(ci.TableName()).Filter("namespace", namespace).All(&manageproject)

	return manageproject, total, err
}

func (ci *CiManagedProjects) FindProjectByNameType(namespace, project_name, repo_type string) error {

	o := orm.NewOrm()

	sql := fmt.Sprintf("select * from %s where namespace =? and name =? and repo_type=?", ci.TableName())

	return o.Raw(sql, namespace, project_name, repo_type).QueryRow(&ci)

}

func (ci *CiManagedProjects) FindProjectByAddressType(namespace, address, repo_type string) error {

	o := orm.NewOrm()

	sql := fmt.Sprintf("select * from %s where namespace =? and address =? and repo_type=?", ci.TableName())

	return o.Raw(sql, namespace, address, repo_type).QueryRow(&ci)

}

func (ci *CiManagedProjects) CreateOneProject(project CiManagedProjects, orms ...orm.Ormer) error {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	_, err := o.Insert(&project)
	if err != nil {
		return err
	}
	return nil
}

func CreateIntegrationParts(namespace string, verified bool, project *CiManagedProjects) error {

	method := "models/CreateIntegrationParts"

	if project.RepoType == "gitlab" && project.GitlabProjectId == "" {

		return errors.New("gitlab_project_id is required for gitlab")

	} else if (project.RepoType == "github" || project.RepoType == "gogs") && project.GitlabProjectId == "" {

		return errors.New(" projectId is required for github or gogs")
	}
	// gitlab or github
	if project.GitlabProjectId == "" {
		project.GitlabProjectId = project.ProjectId
	}

	// Generate the project key
	private_key, public_key, err := secure.MakeSSHKeyPair()
	if err != nil {
		glog.Errorf("%s make sshKeyPair failed: %v", method, err)
		return err
	}
	project.PrivateKey = private_key
	project.PublicKey = public_key

	//get repo api token
	repoConfig := &CiRepos{}
	err = repoConfig.FindOneRepoToken(namespace, DepotToRepoType(project.RepoType))
	if err != nil {
		glog.Errorf("%s FindOneRepoToken failed from database: %v", method, err)
		return err
	}

	//get repo api interface
	repoApi := coderepo.NewRepoApier(project.RepoType, repoConfig.GitlabUrl, repoConfig.AccessToken)

	//add deploy key
	resp, err := repoApi.AddDeployKey(project.GitlabProjectId, project.PublicKey, project.Name)
	if err != nil {
		return err
	}
	//add event  push / tag / merge_request
	events := coderepo.Event{
		Push_events:      true,
		Tag_push_events:  true,
		Pull_request:     true,
		Release_events:   true,
		Only_gen_webhook: false,
	}
	glog.Infof("%s AddDeployKey resp:%#v\n", method, resp)
	// verified 用来区别是否只生成webhook地址
	if !verified {
		// Add webhook for this managed project
		events.Only_gen_webhook = true
		createWebHookOnlyGenWebhook, err := repoApi.CreateWebhook(project.Id, events, project.Name)
		if err != nil {
			return err
		}
		project.WebhookUrl = createWebHookOnlyGenWebhook.Url
		return nil

	} else {
		project.DeployKeyId = resp.Id
		repoName := ""
		if project.RepoType == "gitlab" {
			repoName = project.GitlabProjectId
		} else {
			repoName = project.Name
		}
		createWebHook, err := repoApi.CreateWebhook(project.Id, events, repoName)
		if err != nil {
			return err
		}
		project.WebhookUrl = createWebHook.Config.Url

		if createWebHook.Config.Url != "" {
			project.WebhookId = createWebHook.Id
		}
		//gitlab
		if project.WebhookUrl == "" {
			project.WebhookUrl = createWebHook.Url
			project.WebhookId = createWebHook.Id
		}

	}
	return nil

}

func (ci *CiManagedProjects) RemoveProject(namespace, projectId string, orms ...orm.Ormer) (uint32, error) {
	var o orm.Ormer
	if len(orms) != 1 {
		o = orm.NewOrm()
	} else {
		o = orms[0]
	}
	//return o.Delete(&CiManagedProjects{Namespace: namespace, Id: projectId})
	sql := fmt.Sprintf("delete from %s where id=? and namespace=?;", ci.TableName())
	res, err := o.Raw(sql, projectId, namespace).Exec()
	if err != nil {
		return sqlstatus.ParseErrorCode(err)
	}
	if rowNumber, err := res.RowsAffected(); err != nil {
		return sqlstatus.ParseErrorCode(err)
	} else if rowNumber == 0 {
		return sqlstatus.SQLErrNoRowFound, fmt.Errorf("project %s does not exist", projectId)
	}
	return sqlstatus.SQLSuccess, nil

}

func (ci *CiManagedProjects) FindProjectById(namespace, projectId string) error {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where namespace=? and id=?", ci.TableName())
	return o.Raw(sql, namespace, projectId).QueryRow(&ci)

}

func (ci *CiManagedProjects) FindProjectByIdNew(projectId string) error {
	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where namespace=? and id=?", ci.TableName())
	return o.Raw(sql, projectId).QueryRow(&ci)

}
func (ci *CiManagedProjects) FindProjectOnlyById(projectId string) error {

	o := orm.NewOrm()
	sql := fmt.Sprintf("select * from %s where id=?", ci.TableName())
	return o.Raw(sql, projectId).QueryRow(&ci)
}

func (ci *CiManagedProjects) ClearIntegrationParts(namespace string) error {
	method := "models/ClearIntegrationParts"

	if ci.RepoType == "gitlab" || ci.RepoType == "github" || ci.RepoType == "gogs" {
		glog.Infof("will delete repoType %s the IntegrationParts\n", ci.RepoType)
		repo := &CiRepos{}
		err := repo.FindOneRepoToken(namespace, DepotToRepoType(ci.RepoType))
		if err != nil {
			glog.Errorf("%s find repo token failed: %v \n", method, err)
			return err
		}
		//get repo interface
		repoApi := coderepo.NewRepoApier(ci.RepoType, repo.GitlabUrl, repo.AccessToken)
		if repoApi == nil {
			return fmt.Errorf("not support this  %s", ci.RepoType)
		}
		gitlabProjectId, _ := strconv.Atoi(ci.GitlabProjectId)

		if ci.DeployKeyId != 0 {
			// Remove key
			err = repoApi.RemoveDeployKey(gitlabProjectId, ci.DeployKeyId, ci.Name)
			if err != nil {
				return fmt.Errorf(" RemoveDeployKey %v "+method, err)
			}
			glog.Infof(method, "Remove deploy key Success => %s\n"+ci.Name)
		}

		if ci.WebhookId != 0 {
			// Remove Webhook
			err = repoApi.RemoveWebhook(ci.GitlabProjectId, ci.WebhookId, ci.Name)
			if err != nil {
				return fmt.Errorf(" RemoveWebhook %v "+method, err)
			}
			glog.Infof(method, "RemoveWebhook Success => %s\n"+ci.Name)
		}

	} else if ci.RepoType == "svn" {
		glog.Infof(method, "Only need to remove SVN project %s \n"+ci.Name)
	}
	return nil
}
func (ci *CiManagedProjects) ListProjectsByType(namespace, repo_type string) (ciManagedProjects []CiManagedProjects, total int64, err error) {

	o := orm.NewOrm()
	total, err = o.QueryTable(ci.TableName()).Filter("namespace", namespace).Filter("repo_type", repo_type).
		OrderBy("-create_time").All(&ciManagedProjects)
	return
}
