package coderepo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"dev-flows-api-golang/models/common"
	client "github.com/gogits/go-gogs-client"
	//"github.com/satori/go.uuid"
	"github.com/golang/glog"
	//"github.com/drone/drone/remote/gogs"
	"dev-flows-api-golang/util/rand"
)

const (
	HOOK_SECRET       = "EnnCloud_GOGS_SECRET_KEY"
	DEFAULT_PAGE_SIZE = 50
	KEY_TITLE         = "qinzhao@ennew.cn"
)

type GogsClient struct {
	Gogs_url     string `json:"gogs_url"`
	Access_token string `json:"access_token"`
	Client   *client.Client

	Header map[string]string
}

var _ RepoApier = &GogsClient{}

/**
gogs api Reference API doc: https://github.com/gogits/go-gogs-client/wiki
*/
func NewGogsClient(Gogs_url, Access_token string) *GogsClient {
	//暂时这样子
	gogsClient := client.NewClient(Gogs_url, Access_token)
	if Gogs_url == "" || Access_token == "" {
		return nil
	}
	if strings.Index(Gogs_url, "/api/v1") < 0 {
		Gogs_url += "/api/v1"
	}

	return &GogsClient{
		Gogs_url:     Gogs_url,
		Access_token: Access_token,
		Client:   gogsClient,
		Header:       map[string]string{"Authorization": "token " + Access_token},
		//Header: map[string]string{"Authorization": "token 5f6b8a97fcd9da50e2581c5648cd09fa9d33fe3e"},
	}

}

func (gogs *GogsClient) GetEndPoint(endpoint string, page int) (string, error) {
	if endpoint == "" {
		return "", errors.New("the endpoint is null")
	}
	pageSize := ""

	if page != 0 {
		pageSize = "?per_page=50&page=" + strconv.Itoa(page)
	}
	return gogs.Gogs_url + endpoint + pageSize, nil
}

func (gogs *GogsClient) GetUserInfo() (UserInfo, error) {
	var userInfo UserInfo
	reqUrl := "/user"
	endpoint, err := gogs.GetEndPoint(reqUrl, 0)
	if err != nil {
		return userInfo, err
	}

	resp, err := HttpClientRequest("GET", endpoint, nil, gogs.Header)
	if err != nil {
		return userInfo, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return userInfo, err
	}

	err = json.Unmarshal(data, &userInfo)
	if err != nil {
		return userInfo, err
	}
	userInfo.Type = "user"
	return userInfo, nil

}

func (gogs *GogsClient) GetUserOrgs() ([]UserInfo, error) {
	var userorgses []UserInfo
	reqUrl := "/user/orgs"
	endpoint, err := gogs.GetEndPoint(reqUrl, 0)
	if err != nil {
		return userorgses, err
	}

	resp, err := HttpClientRequest("GET", endpoint, nil, gogs.Header)
	if err != nil {
		return userorgses, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return userorgses, err
	}
	err = json.Unmarshal(data, &userorgses)
	if err != nil {
		return userorgses, err
	}
	defer resp.Body.Close()
	userorgsesLen := len(userorgses)
	if userorgsesLen != 0 {
		for i := 0; i < userorgsesLen; i++ {
			userorgses[i].Type = "orgs"
		}
	}
	return userorgses, nil
}

func (gogs *GogsClient) GetUserAndOrgs() ([]UserInfo, error) {
	userorgses, err := gogs.GetUserOrgs()
	if err != nil {
		return userorgses, err
	}
	userinfo, err := gogs.GetUserInfo()
	if err != nil {
		return userorgses, err
	}
	userorgses = append(userorgses, userinfo)
	if err != nil {
		return userorgses, err
	}

	return userorgses, nil
}

//获取某个用户的所有的仓库
func (gogs *GogsClient) GetAllUsersRepos(userinfos []UserInfo) ([]ReposGitHubAndGogs, error) {
	method:="GogsClient.GetAllUsersRepos"
	var respRepo ReposGitHubAndGogs
	repos := make([]ReposGitHubAndGogs, 0)
	 reposGogs,err:=gogs.Client.ListMyRepos()
	if err!=nil{
		glog.Errorf("%s get all user repos failed:%v\n",method,err)

	}

	for _,repo :=range reposGogs{
		respRepo.Name=repo.FullName
		respRepo.Private=repo.Private
		respRepo.Url=repo.HTMLURL
		respRepo.SshUrl=repo.SSHURL
		respRepo.CloneUrl=repo.CloneURL
		respRepo.Description=repo.Description
		respRepo.ProjectId=int(repo.ID)
		respRepo.Owner.Name=repo.Owner.UserName
		respRepo.Owner.Username=repo.Owner.UserName
		respRepo.Owner.Id=int(repo.Owner.ID)
		respRepo.Owner.State="active"
		respRepo.Owner.Avatar_url=repo.Owner.AvatarUrl
		respRepo.Owner.WebUrl=repo.Owner.AvatarUrl

		repos=append(repos,respRepo)
	}

	return repos, nil
}

func (gogs *GogsClient) GetRepoAllBranches(repoName string, repoId int) ([]BranchResp, error) {
	branchs := make([]BranchResp, 0)

	reqUrl := "/repos/" + repoName + "/branches"

	endpoint, err := gogs.GetEndPoint(reqUrl, 1)
	if err != nil {
		return branchs, err
	}
	resp, err := HttpClientRequest("GET", endpoint, nil, gogs.Header)
	if err != nil {
		return branchs, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return branchs, err
	}
	var gogsBranch []client.Branch
	if string(data) != "" {
		err = json.Unmarshal(data, &gogsBranch)
		if err != nil {
			return branchs, err
		}
	}
	var branch BranchResp
	for _, br := range gogsBranch {
		branch.Branch = br.Name
		branch.CommitId = br.Commit.ID
		branch.CommitterName = br.Commit.Committer.Name
		branch.Message = br.Commit.Message
		branch.CommittedDate = br.Commit.Timestamp.Format("2006-01-02 15:04:05")
		branchs = append(branchs, branch)
	}

	return branchs, nil
}

func (gogs *GogsClient) GetRepoAllTags(repoName string, repoId int) ([]Tag, error) {
	tags := make([]Tag, 0)
	reqUrl := "/repos/" + repoName + "/tags"
	endpoint, err := gogs.GetEndPoint(reqUrl, 1)
	if err != nil {
		return tags, err
	}
	resp, err := HttpClientRequest("GET", endpoint, nil, gogs.Header)
	if err != nil {
		return tags, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tags, err
	}

	if string(data) != "" {
		err = json.Unmarshal(data, &tags)
		if err != nil {
			glog.Errorf("GogsClient get tags failed ===>%v", err)
			return tags, nil
		}
	}

	return tags, nil
}

func (gogs *GogsClient) CreateWebhook(projectId string, events Event, repoName string) (WebhookResp, error) {
	method:="GogsClient/CreateWebhook"
	var webhook WebhookResp

	hookUrl := common.WebHookUrlPrefix + projectId
	//如果只生成webhook
	if events.Only_gen_webhook {
		webhook.Url = hookUrl
		return webhook, nil
	}

	eventsArray := make([]string, 0)
	//push event
	if events.Push_events {
		eventsArray = append(eventsArray, "push")
	}
	//pull_request event
	if events.Tag_push_events {
		eventsArray = append(eventsArray, "create", "pull_request")
	}

	if events.Release_events {
		eventsArray = append(eventsArray, "release")
	}

	reqUrl := "/repos/" + repoName + "/hooks"
	reqdata := WebhookReq{
		Type:   "gogs", // Required The type of webhook, either gogs or slack
		Active: true,
		Events: eventsArray,
		Config: Config{
			Url:          hookUrl,
			Content_type: "json",
			Secret:       HOOK_SECRET, // Don't support for now
		},
	}

	endpoint, err := gogs.GetEndPoint(reqUrl, 0)
	if err != nil {
		return webhook, err
	}

	reqWebhook, err := json.Marshal(reqdata)
	if err != nil {
		return webhook, err
	}

	resp, err := HttpClientRequest("POST", endpoint, strings.NewReader(string(reqWebhook)), gogs.Header)
	if err != nil {
		return webhook, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return webhook, err
	}
	glog.Infof("%s webhook info:%s\n",method,string(data))
	err = json.Unmarshal(data, &webhook)
	if err != nil {
		return webhook, err
	}
	return webhook, nil

}

func (gogs *GogsClient) GetOneWebhook(repo_full_name, webhook_id string) (WebhookResp, error) {

	var webhook WebhookResp

	reqUrl := "/repos/" + repo_full_name + "/hooks/" + webhook_id
	endpoint, err := gogs.GetEndPoint(reqUrl, 0)
	if err != nil {
		return webhook, err
	}
	resp, err := HttpClientRequest("GET", endpoint, nil, gogs.Header)
	if err != nil {
		return webhook, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return webhook, err
	}

	err = json.Unmarshal(data, &webhook)
	if err != nil {
		return webhook, err
	}

	return webhook, nil

}

func (gogs *GogsClient) UpdateWebHook(projectId, repo_full_name string, events []string) (WebhookResp, error) {
	var webhook WebhookResp
	hookUrl := common.WebHookUrlPrefix + projectId

	reqUrl := "/repos/" + repo_full_name + "/hooks/" + projectId
	endpoint, err := gogs.GetEndPoint(reqUrl, 0)
	if err != nil {
		return webhook, err
	}
	webhookReq := WebhookReq{
		Active: true,
		Events: events,
		Config: Config{
			Url:          hookUrl,
			Content_type: "json",
		},
	}

	data, err := json.Marshal(webhookReq)

	resp, err := HttpClientRequest("PATCH", endpoint, strings.NewReader(string(data)), gogs.Header)
	if err != nil {
		return webhook, err
	}
	defer resp.Body.Close()

	resp_data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return webhook, err
	}

	err = json.Unmarshal(resp_data, &webhook)
	if err != nil {
		return webhook, err
	}

	return webhook, nil

}

func (gogs *GogsClient) RemoveWebhook(projectId string, hook_id int, repoName string) error {

	reqUrl := "/repos/" + repoName + "/hooks/" + fmt.Sprintf("%d", hook_id)

	endpoint, err := gogs.GetEndPoint(reqUrl, 0)
	if err != nil {
		return err
	}

	resp, err := HttpClientRequest("DELETE", endpoint, nil, gogs.Header)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	glog.Infof("RemoveWebhook resp.Status:%s\n",resp.Status)
	if strings.Contains(resp.Status, "Status: 204 No Content") {
		return nil
	}

	//return errors.New("DELETE Webhook Failed")
	return nil

}

func (gogs *GogsClient) AddDeployKey(projectId, publicKey, repoName string) (AddDeployResp, error) {

	var deployResp AddDeployResp
	//gogs.Client.CreateDeployKey()
	reqUrl := "/repos/" + repoName + "/keys"

	endpoint, err := gogs.GetEndPoint(reqUrl, 0)
	if err != nil {
		return deployResp, err
	}
	deployReqData := AddDeployReq{
		Title:     fmt.Sprintf("qinzhao@ennew.cn-%s",rand.RandString(5)),
		Key:       publicKey,
		Read_only: true,
	}

	deployKey, err := json.Marshal(deployReqData)
	if err != nil {
		return deployResp, err
	}

	resp, err := HttpClientRequest("POST", endpoint, strings.NewReader(string(deployKey)), gogs.Header)
	if err != nil {
		return deployResp, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return deployResp, err
	}
	glog.Infof("generate AddDeployReq resp body:%#v\n",string(data))
	err = json.Unmarshal(data, &deployResp)
	if err != nil {
		return deployResp, err
	}

	return deployResp, nil

}

func (gogs *GogsClient) RemoveDeployKey(project_id, key_id int, repoName string) error {

	reqUrl := "/repos/" + repoName + "/keys/" + fmt.Sprintf("%d", key_id)

	endpoint, err := gogs.GetEndPoint(reqUrl, 0)
	if err != nil {

		return err

	}

	resp, err := HttpClientRequest("DELETE", endpoint, nil, gogs.Header)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	glog.Infof("RemoveDeployKey response Status info:[%s]",resp.Status)
	if strings.Contains(resp.Status, "204") {
		return nil
	}

	//return errors.New("gogs RemoveDeployKey failed")
	return nil
}

func (gogs *GogsClient) CheckSignature(headers map[string]string, body string) {

}

///per_page=\d+&page=\d+>; rel="last"/
func GetTotalPage(link string) int {
	reg := regexp.MustCompile(`/per_page=\d+&page=\d+>; rel="last"/`)

	result := reg.FindAllString(link, -1)
	glog.Infof("GetTotalPage=%s", result)
	if len(result) != 0 {
		totalPage := regexp.MustCompile(`/\d+>/`).FindAllString(result[0], -1)

		if totalPage[0] == "" {
			page, err := strconv.Atoi(totalPage[0])
			if err != nil {
				return 0
			}
			return page
		}
	}
	return 0
}
func (g *GogsClient) GetProjectWebhooks(projectId int) ([]WebhookResp, error) {
	return nil, nil
}
