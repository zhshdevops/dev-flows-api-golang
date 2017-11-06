package coderepo

import (
	"github.com/golang/glog"
)

type RepoApier interface {
	GetUserInfo() (UserInfo, error)
	CreateWebhook(projectId string, events Event, repoName string) (WebhookResp, error)
	GetUserOrgs() ([]UserInfo, error)
	GetAllUsersRepos(userinfos []UserInfo) ([]ReposGitHubAndGogs, error)
	GetUserAndOrgs() ([]UserInfo, error)
	GetRepoAllBranches(repoName string, repoId int) ([]BranchResp, error)
	GetRepoAllTags(repoName string, repoId int) ([]Tag, error)
	GetOneWebhook(repoFullName, webhookId string) (WebhookResp, error)
	UpdateWebHook(projectId, repoFullName string, events []string) (WebhookResp, error)
	RemoveWebhook(projectId string, hookId int, repoName string) error
	AddDeployKey(projectId, publicKey, repoName string) (AddDeployResp, error)
	RemoveDeployKey(projectId, keyId int, repoName string) error
	GetProjectWebhooks(projectId int) ([]WebhookResp, error)
}

func NewRepoApier(RepoType, git_server_url, Access_token string) RepoApier {
	var repoApier RepoApier
	switch RepoType {
	case "gitlab":
		return NewGitlabClient(git_server_url,Access_token)
	//case "github":
	//	return NewGitlabClient(git_server_url,Access_token)
	//	return NewGitlabClient(git_server_url,Access_token)
	case "gogs":
		repoApier = NewGogsClient(git_server_url, Access_token)
		return repoApier
	default:
		glog.Errorf("Not support %s\n", RepoType)
		return nil
	}
	return nil
}
