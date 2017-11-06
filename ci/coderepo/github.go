//https://developer.github.com/v3/repos/hooks/
package coderepo

import (
	"os"
	"errors"
	"strconv"
	"encoding/json"
	"io/ioutil"
	"dev-flows-api-golang/models/common"
	"strings"
	"fmt"
)

type GitHubClient struct {
	ClientId       string
	ClientSecret   string
	RedirectUrl    string
	GitHubLoginUrl string
	GitHubApiUrl   string
	Webhook_secret string
	Algorithm      string
	Tocken         string
}

const (
	GitHubLoginUrl = "https://github.com/login/oauth"
	GitHubApiUrl = "https://api.github.com"
	webhook_secret = "ENNCLOUD_GITHUB_CI_SECRET_2017"
	algorithm = "sha1"
)

func NewGitHubClient() *GitHubClient {

	CLIENT_ID := os.Getenv("CLIENT_ID")
	CLIENT_SECRET := os.Getenv("CLIENT_SECRET")
	GITHUB_REDIRECT_URL := os.Getenv("GITHUB_REDIRECT_URL")

	//if CLIENT_ID==""{
	//
	//	CLIENT_ID="43b9c69e79ae49f32919"
	//
	//}
	//
	//if CLIENT_SECRET==""{
	//
	//	CLIENT_SECRET="bf1c5dfd9ae48081073d786e3797cb08a1f0b59e"
	//}
	//
	//if GITHUB_REDIRECT_URL==""{
	//
	//	GITHUB_REDIRECT_URL="https://paasdev.enncloud.cn/api/v2/devops/repos/github/auth-callback"
	//}

	return &GitHubClient{
		ClientId:CLIENT_ID,
		ClientSecret:CLIENT_SECRET,
		RedirectUrl:GITHUB_REDIRECT_URL,
		GitHubLoginUrl:GitHubLoginUrl,
		GitHubApiUrl:GitHubApiUrl,
		Webhook_secret:webhook_secret,
		Algorithm:algorithm,
	}
}

func (gitcli *GitHubClient)GetExchangeTokenUrl(code string) (url string) {
	url = gitcli.GitHubLoginUrl + "/access_token?client_id=" + gitcli.ClientId + "&client_secret=" + gitcli.ClientSecret + "&code=" + code
	return
}

func (gitcli *GitHubClient)GetEndPoint(endpoint string, page int) (string, error) {

	pageSize := ""
	if endpoint == "" {
		return "", errors.New("the endpoint is null")
	}
	if page != 0 {
		pageSize = "?per_page=50&page=" + strconv.Itoa(page)
	}
	return gitcli.GitHubApiUrl + endpoint + pageSize, nil
}

func (gitcli *GitHubClient)ExchangeToken(code string) (error) {

	resp, err := HttpClientRequest("POST", gitcli.GetExchangeTokenUrl(code), nil, nil)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	var access_token_info Access_token_info
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &access_token_info)
	if err != nil {
		return err
	}

	gitcli.Tocken = access_token_info.Access_token

	return nil

}

func (gitcli *GitHubClient)GetUserInfo() (UserInfo, error) {
	var userInfo UserInfo
	Authorization := map[string]string{
		"Authorization":"token " + gitcli.Tocken,
	}
	endpoint, err := gitcli.GetEndPoint("/user", 0)
	if err != nil {
		return userInfo, err
	}
	resp, err := HttpClientRequest("GET", endpoint, nil, Authorization)
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

	return userInfo, nil
}

func (gitcli *GitHubClient)GetUserOrgs() (UserOrgs, error) {
	var userorgs UserOrgs
	header := map[string]string{
		"Authorization":"token " + gitcli.Tocken,
	}

	endpoint, err := gitcli.GetEndPoint("/user/orgs", 0)
	if err != nil {
		return userorgs, err
	}
	resp, err := HttpClientRequest("GET", endpoint, nil, header)
	if err != nil {

		return userorgs, err

	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return userorgs, err
	}
	err = json.Unmarshal(data, &userorgs)
	if err != nil {
		return userorgs, err
	}
	return userorgs, nil

}

func (gitcli *GitHubClient)GetUserAndOrgs() {
	//:=gitcli.GetUserInfo()


}

func (gitcli *GitHubClient)GetAllUsersRepos(userinfos []UserInfo) {

	//header:=map[string]string{
	//	"Authorization":"token "+gitcli.Tocken,
	//}
	//reqUrl:="/user/repos"
	//for _,userinfo:=range userinfos{
	//	if userinfo.Type=="orgs"{
	//		reqUrl="/orgs/"+userinfo.Login+"/repos"
	//	}
	//	endpoint,err:=gitcli.GetEndPoint(reqUrl,1)
	//	if err!=nil{
	//		return
	//	}
	//	resp,err:=HttpClientRequest("GET",endpoint,nil,header)
	//	if err!=nil{
	//		return
	//	}
	//	data,err:=ioutil.ReadAll(resp.Body)
	//	if err!=nil{
	//		return
	//	}
	//
	//	//json.Unmarshal(data,&userinfo)
	//	defer resp.Body.Close()
	//}

	return

}

func (gitcli *GitHubClient)GetRepoAllBranches(repoName string, repoId int) ([]BranchResp, error) {
	var branchs []BranchResp
	branchs = make([]BranchResp, 0)

	header := map[string]string{
		"Authorization":"token " + gitcli.Tocken,
	}
	reqUrl := "/repos/" + repoName + "/branches"
	endpoint, err := gitcli.GetEndPoint(reqUrl, 0)
	if err != nil {
		return branchs, err
	}
	resp, err := HttpClientRequest("GET", endpoint, nil, header)
	if err != nil {
		return branchs, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return branchs, err
	}
	err = json.Unmarshal(data, &branchs)
	if err != nil {
		return branchs, err
	}
	defer resp.Body.Close()
	return branchs, nil

}

func (gitcli *GitHubClient)GetRepoAllTags(repoName string, repoId int) ([]Tag, error) {
	tags := make([]Tag, 0)

	reqUrl := "/repos/" + repoName + "/tags"
	endpoint, err := gitcli.GetEndPoint(reqUrl, 1)
	if err != nil {
		return tags, err
	}
	header := map[string]string{
		"Authorization":"token " + gitcli.Tocken,
	}

	resp, err := HttpClientRequest("GET", endpoint, nil, header)
	if err != nil {
		return tags, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tags, err
	}
	err = json.Unmarshal(data, &tags)
	if err != nil {
		return tags, err
	}
	return tags, nil
}

func (gitcli *GitHubClient)CreateWebhook(projectid string, events []string, repoName string) (WebhookResp, error) {
	var webhookResp WebhookResp
	header := map[string]string{
		"Authorization":"token " + gitcli.Tocken,
	}
	hookUrl := common.WebHookUrlPrefix + projectid
	if len(events) == 0 {
		webhookResp.Url = hookUrl
		return webhookResp, nil
	}
	reqUrl := "/repos/" + repoName + "/hooks"
	giturl, err := gitcli.GetEndPoint(reqUrl, 0)
	if err != nil {
		return webhookResp, err
	}
	reqJsonData := WebhookReq{
		Name:"web",
		Active:true,
		Events:[]string{"push", "pull_request", "release"},
		Config:Config{
			Url:hookUrl,
			Content_type:"json",
			Secret:gitcli.Webhook_secret,
		},
	}
	reqjson, err := json.Marshal(reqJsonData)
	if err != nil {
		return webhookResp, err
	}
	resp, err := HttpClientRequest("POST", giturl, strings.NewReader(string(reqjson)), header)
	if err != nil {
		return webhookResp, err
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return webhookResp, err
	}

	err = json.Unmarshal(respData, &webhookResp)
	if err != nil {
		return webhookResp, err
	}

	return webhookResp, nil
}

//GET /repos/:owner/:repo/hooks/:id
//https://developer.github.com/v3/repos/hooks/
func (gitcli *GitHubClient)GetOneWebhook(repo_full_name, webhook_id string) (WebhookResp, error) {

	var webhookResp WebhookResp
	reqUrl := "/repos/" + repo_full_name + "/hooks/" + webhook_id
	header := map[string]string{
		"Authorization":"token " + gitcli.Tocken,
	}
	endpoint, err := gitcli.GetEndPoint(reqUrl, 0)
	if err != nil {
		return webhookResp, err
	}

	resp, err := HttpClientRequest("GET", endpoint, nil, header)
	if err != nil {
		return webhookResp, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return webhookResp, err
	}

	json.Unmarshal(data, &webhookResp)

	return webhookResp, nil

}
//PATCH /repos/:owner/:repo/hooks/:id
func (gitcli *GitHubClient)UpdateWebHook(projectsSourceFullName string, WebhookId int) {

	repo_full_name := projectsSourceFullName
	reqUrl := "/repos/" + repo_full_name + "/hooks/" + strconv.Itoa(WebhookId)
	header := map[string]string{
		"Authorization":"token " + gitcli.Tocken,
	}

	endpoint, err := gitcli.GetEndPoint(reqUrl, 0)
	if err != nil {
		return
	}

	webhookReq := WebhookReq{

	}

	reqdata, err := json.Marshal(webhookReq)
	if err != nil {
		return
	}

	resp, err := HttpClientRequest("PATCH", endpoint, strings.NewReader(string(reqdata)), header)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var webhook WebhookResp
	json.Unmarshal(data, &webhook)

	return
}

//DELETE /repos/:owner/:repo/hooks/:id
func (gitcli *GitHubClient)RemoveWebhook(hook_id int, repoName string) error {
	reqUrl := "/repos/" + repoName + "/:repo/hooks/" + strconv.Itoa(hook_id)
	header := map[string]string{
		"Authorization":"token " + gitcli.Tocken,
	}
	endPoint, err := gitcli.GetEndPoint(reqUrl, 0)
	if err != nil {
		return err
	}

	resp, err := HttpClientRequest("DELETE", endPoint, nil, header)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errors.New("delete failed")
	}
	defer resp.Body.Close()
	return nil

}
func (gitcli *GitHubClient)GetAuthRedirectUrl(state string) string {
	return gitcli.GitHubLoginUrl + "/authorize?client_id=" + gitcli.ClientId + "&redirect_uri=" +
		gitcli.RedirectUrl + "&state=" + state + "&scope=repo, user:email"
}

func (gitcli *GitHubClient)CheckSignature(headers map[string]string, body string) {

}
//POST /repos/:owner/:repo/keys
func (gitcli *GitHubClient)AddDeployKey(projectId, publicKey, repoName string) error {
	header := map[string]string{
		"Authorization":"token " + gitcli.Tocken,
	}
	endPoint := "/repos/" + repoName + "/keys"
	reqUrl, err := gitcli.GetEndPoint(endPoint, 0)
	deployKey := AddDeployReq{
		Title:"qinzhao@ennew.cn",
		Key:publicKey,
		Read_only:true,
	}
	reqData, err := json.Marshal(deployKey)
	if err != nil {
		return err
	}
	resp, err := HttpClientRequest("POST", reqUrl, strings.NewReader(string(reqData)), header)
	if err != nil {
		return err
	}
	var addDeploy AddDeployResp
	respData, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	err = json.Unmarshal(respData, &addDeploy)
	if err != nil {
		return err
	}
	if addDeploy.Verified {
		return nil
	}
	return errors.New("AddDeployKey failed")
}
//DELETE /repos/:owner/:repo/keys/:id
func (gitcli *GitHubClient)RemoveDeployKey(key_id int, repoName string) error {
	reqUrl := fmt.Sprintf("/repos/%s/keys/%d", repoName, key_id)
	endPoint, err := gitcli.GetEndPoint(reqUrl, 0)
	if err != nil {
		return err
	}

	header := map[string]string{
		"Authorization":"token " + gitcli.Tocken,
	}
	resp, err := HttpClientRequest("DELETE", endPoint, nil, header)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if strings.Contains(resp.Status, "204") {
		return nil
	}
	return errors.New("RemoveDeployKey failed")
}
