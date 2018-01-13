package coderepo


type Access_token_info struct {
	Access_token string `json:"access_token"`
	Scope        string `json:"scope"`
	Token_type   string `json:"token_type"`
}

//[{"login":"root",
// "type":"user",
// "id":1,
// "url":"http://10.39.0.53:8880/api/v3/root",
// "avatar":"http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
// "email":"admin@example.com","isadmin":true}]
type UserInfo struct {
	Login      string `json:"login,omitempty"`
	Type       string  `json:"type,omitempty"`
	Id         int `json:"id,omitempty"`
	Html_url   string `json:"html_url,omitempty"`
	Avatar_url string `json:"avatar_url,omitempty"`
	Url        string `json:"url"`
	Email      string `json:"email,omitempty"`
	Site_admin bool `json:"site_admin,omitempty"`
	Username   string `json:"username,omitempty"`
	Avatar     string `json:"avatar"`
	Isadmin    bool `json:"isadmin"`
}

type UserOrgs struct {
	Login      string `json:"login"`
	Type       string  `json:"type"`
	Id         int `json:"id"`
	Url        string `json:"url"`
	Avatar_url string `json:"avatar_url"`
	Email      string `json:"email"`
	Site_admin bool `json:"site_admin"`
	Username   string `json:"username"`
}

type BranchResp struct {
	Branch        string `json:"branch"`
	CommitId      string `json:"commit_id"`
	CommitterName string `json:"committer_name"`
	Message       string `json:"message"`
	CommittedDate string `json:"committed_date"`
}

type Branch struct {
	Branch string `json:"name"`
	Commit Commit `json:"commit"`
}
type Commit struct {
	Sha            string `json:"sha,omitempty"`
	Id             string `json:"id,omitempty"`
	Committer_name string `json:"committer_name"`
	Message        string `json:"message"`
	Committed_date string `json:"committed_date"`
}

type Tag struct {
	Tag         string `json:"tag"`
	Commit_id   string `json:"commit_id"`
	Description string `json:"description"`
	CommitterName      string `json:"committer_name"`
	Message      string `json:"message"`
	Committed_date string `json:"committed_date"`
}

type Event struct {
	Push_events      bool `json:"push_events"`
	Tag_push_events  bool `json:"tag_push_events"`
	Pull_request     bool `json:"pull_request"`
	Release_events   bool `json:"release_events"`
	Only_gen_webhook bool `json:"only_gen_webhook"`
}

type EventMap map[string]string

type Config struct {
	Url          string `json:"url"`
	Content_type string `json:"content_type,omitempty"`
	Secret       string `json:"secret,omitempty"`
	Insecure_ssl string `json:"insecure_ssl,omitempty"`
}
type WebhookReq struct {
	Name   string `json:"name,omitempty"`
	Active bool `json:"active,omitempty"`
	Events []string `json:"events,omitempty"`
	Config Config `json:"config"`
	Type   string `json:"type,omitempty"`
}

type WebhookResp struct {
	Id         int `json:"id"`
	Type       string `json:"type,omitempty"`
	Url        string `json:"url"`
	Ping_url   string `json:"ping_url,omitempty"`
	Test_url   string `json:"test_url,omitempty"`
	Name       string `json:"name"`
	Events     []string `json:"events"`
	Active     bool `json:"active"`
	Config     Config `json:"config"`
	Updated_at string `json:"updated_at"`
	Created_at string `json:"created_at"`
}

type AddDeployReq struct {
	Title     string `json:"title"`
	Key       string `json:"key"`
	Read_only bool `json:"read_only"`
}

type AddDeployResp struct {
	Id         int `json:"id"`
	Key        string `json:"key"`
	Url        string `json:"url"`
	Title      string `json:"title"`
	Verified   bool `json:"verified"`
	Created_at string `json:"created_at"`
	Read_only  bool `json:"read_only"`
}

type Owner struct {
	Login      string `json:"login"`
	Id         int `json:"id"`
	State      string `json:"state"`
	Avatar_url string `json:"avatar_url"`
	Web_url    string `json:"html_url"`
}

type Repository struct {
	Full_name   string `json:"full_name"`
	Private     bool `json:"private"`
	Html_url    string `json:"html_url"`
	Ssh_url     string `json:"ssh_url"`
	Clone_url   string `json:"clone_url"`
	Description string `json:"description"`
	Owner       Owner `json:"owner"`
	ProjectId   int `json:"id"`
	Active      int `json:"active"`
}

type ManagedProject struct {
	Active int `json:"active"`
	Id string `json:"id"`
}
type OwnerGitHubGogs struct {
	Name       string `json:"name"`
	Id         int `json:"id"`
	State      string `json:"state"`
	Avatar_url string `json:"avatar_url"`
	Username   string `json:"username"`
	WebUrl string `json:"web_url,omitempty"`
}

type ReposGitHubAndGogs struct {
	CloneUrl    string `json:"clone_url"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Owner OwnerGitHubGogs `json:"owner"`
	Private bool `json:"private"`
	ProjectId int `json:"projectId"`
	SshUrl string `json:"ssh_url"`
	Url string `json:"url"`
	ManagedProject ManagedProject `json:"managed_project"`
}

type RepoGitLab struct {
	Name_with_namespace string `json:"name_with_namespace"`
	Public              string `json:"visibility"`
	Web_url             string `json:"web_url"`
	Ssh_url_to_repo     string `json:"ssh_url_to_repo"`
	Http_url_to_repo    string `json:"http_url_to_repo"`
	Description         string `json:"description"`
	Owner               string `json:"owner"`
	Id                  int `json:"id"`
	Namespace           GitlabNamespace `json:"namespace"`
}

type GitlabOwer struct {
	Name string `json:"name"`
	Username string `json:"username"`
	Id int `json:"id"`
	State string `json:"state"`
	AvatarUrl string `json:"avatar_url"`
	WebUrl string `json:"web_url"`

}

type GitlabNamespace struct {
	Id        int `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Full_path string `json:"full_path"`
}


type SvnHook struct {
	Name       string `json:"name"`
	ClearCache int  `json:"clearCache"`
}
