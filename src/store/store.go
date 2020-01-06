package store

import (
	"io"

	"github.com/coderun-top/coderun/src/model"
	"golang.org/x/net/context"
)

type Store interface {
	// GetUser gets a user by unique ID.
	GetUser(int64) (*model.User, error)

	// GetUserLogin gets a user by unique Login name.
	GetUserLogin(string) (*model.User, error)

	GetUserName(string) (*model.User, error)
	GetUserUserName(string) ([]*model.User, error)
	// GetUserUserName2(string,string) (*model.User, error)
	GetUserUserNameOauth(string, bool) ([]*model.User, error)
	GetUserNameToken(string, string, bool) (*model.User, error)
	GetUserNameTokenOauth(string, bool) (*model.User, error)
	GetUserNameTokenOauth2(string, string, bool) (*model.User, error)

	// GetUserList gets a list of all users in the system.
	GetUserList() ([]*model.User, error)

	// GetUserCount gets a count of all users in the system.
	GetUserCount() (int, error)

	// CreateUser creates a new user account.
	CreateUser(*model.User) error

	// UpdateUser updates a user account.
	UpdateUser(*model.User) error

	// DeleteUser deletes a user account.
	DeleteUser(*model.User) error

	GetProject(string) (*model.Project, error)
	GetProjectName(string) (*model.Project, error)

	CreateProject(*model.Project) error

	// DeleteUser deletes a user account.
	DeleteProject(*model.Project) error

	// GetRepo gets a repo by unique ID.
	GetRepo(int64) (*model.Repo, error)

	// GetRepoName gets a repo by user & its full name.
	GetRepoFullName(*model.User, string) (*model.Repo, error)

	// GetRepoName gets a repo by its full name.
	GetRepoName(string) (*model.Repo, error)

	// GetRepoCount gets a count of all repositories in the system.
	GetRepoCount() (int, error)

	GetStarRepo(string, int, int) ([]*model.Repo, error)

	GetStarRepoCount(string) (int, error)
	// CreateRepo creates a new repository.
	CreateRepo(*model.Repo) error
	// CreateRepo creates a new repository.
	CreateRepoConfigs(*model.Repo, *model.Config) error

	// UpdateRepo updates a user repository.
	UpdateRepo(*model.Repo) error

	// DeleteRepo deletes a user repository.
	DeleteRepo(*model.Repo) error

	// GetBuild gets a build by unique ID.
	GetBuild(string) (*model.Build, error)

	// GetBuildNumber gets a build by number.
	GetBuildNumber(*model.Repo, int) (*model.Build, error)
	// GetBuildNumber2(*model.Repo, int) (*model.BuildRepo, error)

	// GetBuildRef gets a build by its ref.
	GetBuildRef(*model.Repo, string) (*model.Build, error)

	// GetBuildCommit gets a build by its commit sha.
	GetBuildCommit(*model.Repo, string, string) (*model.Build, error)

	GetBuildLastByPipeline(string) (*model.Build, error)

	// GetBuildLast gets the last build for the branch.
	// GetBuildLast(*model.Repo, string) (*model.Build, error)
	GetBuildLast(*model.Repo) (*model.Build, error)

	// GetBuildLastBefore gets the last build before build number N.
	GetBuildLastBefore(*model.Repo, string, string) (*model.Build, error)

	GetBuildList(*model.Repo, string, int, int) ([]*model.BuildRepo, error)
	GetBuildListProject(string, string, int, int) ([]*model.BuildRepo, error)
	GetBuildListProjectCountByMonth(string, string) (int, error)

	GetBuildCount2(*model.Repo, string) (int, error)

	// GetBuildQueue gets a list of build in queue.
	GetBuildQueue() ([]*model.Feed, error)

	// GetBuildCount gets a count of all builds in the system.
	GetBuildCount() (int, error)

	// CreateBuild creates a new build and jobs.
	CreateBuild(*model.Build, ...*model.Proc) error

	// UpdateBuild updates a build.
	UpdateBuild(*model.Build) error

	GetBuildTimeProject(string) (int64, error)

	//
	// new functions
	//

	UserFeed(*model.User) ([]*model.Feed, error)

	RepoList(*model.User) ([]*model.Repo, error)
	RepoListProject(string) ([]*model.Repo, error)
	RepoListProjectPage(string, string, int, int) ([]*model.Repo, error)
	RepoListProjectCount(string, string) (int, error)
	RepoListLatest(*model.User) ([]*model.Feed, error)
	RepoBatch([]*model.Repo) error
	GetRepoConfigId(config_id int64) (*model.Repo, error)

	PermFind(user *model.User, repo *model.Repo) (*model.Perm, error)
	PermUpsert(perm *model.Perm) error
	PermBatch(perms []*model.Perm) error
	PermDelete(perm *model.Perm) error
	PermFlush(user *model.User, before int64) error

	ConfigLoad(string) (*model.Config, error)
	ConfigFind(*model.Repo, string) (*model.Config, error)
	ConfigRepoFind(*model.Repo) ([]*model.Config, error)
	ConfigFindApproved(*model.Config) (bool, error)
	ConfigCreate(*model.Config) error
	ConfigUpdate(*model.Config) error
	ConfigDeleteRepo(*model.Repo) error

	SenderFind(*model.Repo, string) (*model.Sender, error)
	SenderList(*model.Repo) ([]*model.Sender, error)
	SenderCreate(*model.Sender) error
	SenderUpdate(*model.Sender) error
	SenderDelete(*model.Sender) error

	SecretFind(*model.Repo, string) (*model.Secret, error)
	SecretList(*model.Repo) ([]*model.Secret, error)
	SecretCreate(*model.Secret) error
	SecretUpdate(*model.Secret) error
	SecretDelete(*model.Secret) error

	StarFind(int64) (*model.Star, error)
	StarFind2(int64, string) (*model.Star, error)
	StarList(string) ([]*model.Star, error)
	StarCreate(*model.Star) error
	StarDelete(*model.Star) error

	RegistryFind(string, string) (*model.Registry, error)
	RegistryList(string) ([]*model.Registry, error)
	RegistryCreate(*model.Registry) error
	RegistryUpdate(*model.Registry) error
	RegistryDelete(string, string) error

	HelmFind(string, string) (*model.Helm, error)
	HelmList(string) ([]*model.Helm, error)
	HelmCreate(*model.Helm) error
	HelmUpdate(*model.Helm) error
	HelmDelete(string, string) error

	ProcLoad(int64) (*model.Proc, error)
	ProcFind(*model.Build, int) (*model.Proc, error)
	ProcChild(*model.Build, int, string) (*model.Proc, error)
	ProcList(*model.Build) ([]*model.Proc, error)
	ProcCreate([]*model.Proc) error
	ProcUpdate(*model.Proc) error
	ProcClear(*model.Build) error

	LogFind(*model.Proc) (io.ReadCloser, error)
	LogSave(*model.Proc, io.Reader) error

	FileList(*model.Build) ([]*model.File, error)
	FileFind(*model.Proc, string) (*model.File, error)
	FileRead(*model.Proc, string) (io.ReadCloser, error)
	FileCreate(*model.File, io.Reader) error

	TaskList() ([]*model.Task, error)
	TaskInsert(*model.Task) error
	TaskDelete(string) error

	// Ping() error

	// Pipeline_Env
	GetPipelineEnvs(string) ([]*model.PipelineEnv, error)
	UpdatePipelineEnvs(string, []*model.PipelineEnv) error

	// Project Env
	GetProjectEnvs(string) ([]*model.ProjectEnv, error)
	UpdateProjectEnvs(string, []*model.ProjectEnv) error

	// K8s cluster
	K8sClusterList(string) ([]*model.K8sCluster, error)
	K8sClusterFind(string, string) (*model.K8sCluster, error)
	K8sClusterCreate(*model.K8sCluster) error
	K8sClusterUpdate(*model.K8sCluster) error
	K8sClusterDelete(string, string) error

	K8sDeployFind(int64) (*model.K8sDeploy, error)
	K8sDeployFindProject(string, string) (*model.K8sDeploy, error)
	K8sDeployList(string) ([]*model.K8sDeploy, error)
	K8sDeployCreate(*model.K8sDeploy) error
	K8sDeployUpdate(*model.K8sDeploy) error
	K8sDeployDelete(int64) error

	GetWebHook(*model.Repo) (*model.WebHook, error)
	CreateWebHook(*model.WebHook) error
	UpdateWebHook(*model.WebHook) error

	GetWebHookProject(string) (*model.WebHookProject, error)
	CreateWebHookProject(*model.WebHookProject) error
	UpdateWebHookProject(*model.WebHookProject) error

	GetEmail(*model.Repo) (*model.Email, error)
	CreateEmail(*model.Email) error
	UpdateEmail(*model.Email) error

	GetEmailProject(string) (*model.EmailProject, error)
	CreateEmailProject(*model.EmailProject) error
	UpdateEmailProject(*model.EmailProject) error

	GetAgentKey(string) (*model.AgentKey, error)
	GetAgentKeyByKey(string) (*model.AgentKey, error)
	CreateAgentKey(*model.AgentKey) error

	GetAgents(string) ([]*model.Agent, error)
	GetAgent(string) (*model.Agent, error)
	GetAgentClient(string, string) (*model.Agent, error)
	GetAgentCount(string) (int, error)
	CreateAgent(*model.Agent) error
	UpdateAgent(*model.Agent) error
}

// GetUser gets a user by unique ID.
func GetUser(c context.Context, id int64) (*model.User, error) {
	return FromContext(c).GetUser(id)
}

// GetUserLogin gets a user by unique Login name.
func GetUserLogin(c context.Context, login string) (*model.User, error) {
	return FromContext(c).GetUserLogin(login)
}

// GetUserName gets a user by unique Login name.
func GetUserName(c context.Context, name string) (*model.User, error) {
	return FromContext(c).GetUserName(name)
}

// GetUserName gets a user by unique Login name.
func GetUserNameToken(c context.Context, name, token string, oauth bool) (*model.User, error) {
	return FromContext(c).GetUserNameToken(name, token, oauth)
}

func GetUserNameTokenOauth(c context.Context, token string, oauth bool) (*model.User, error) {
	return FromContext(c).GetUserNameTokenOauth(token, oauth)
}

func GetUserNameTokenOauth2(c context.Context, name, token string, oauth bool) (*model.User, error) {
	return FromContext(c).GetUserNameTokenOauth2(name, token, oauth)
}

// func GetUserUserName2(c context.Context, token, name string) (*model.User, error) {
// 	return FromContext(c).GetUserUserName2(token, name)
// }

// GetUserName gets a user by unique Login name.
func GetUserUserName(c context.Context, name string) ([]*model.User, error) {
	return FromContext(c).GetUserUserName(name)
}

func GetUserUserNameOauth(c context.Context, name string, oauth bool) ([]*model.User, error) {
	return FromContext(c).GetUserUserNameOauth(name, oauth)
}

// GetUserList gets a list of all users in the system.
func GetUserList(c context.Context) ([]*model.User, error) {
	return FromContext(c).GetUserList()
}

// GetUserCount gets a count of all users in the system.
func GetUserCount(c context.Context) (int, error) {
	return FromContext(c).GetUserCount()
}

func CreateUser(c context.Context, user *model.User) error {
	return FromContext(c).CreateUser(user)
}

func UpdateUser(c context.Context, user *model.User) error {
	return FromContext(c).UpdateUser(user)
}

func DeleteUser(c context.Context, user *model.User) error {
	return FromContext(c).DeleteUser(user)
}

func GetRepo(c context.Context, id int64) (*model.Repo, error) {
	return FromContext(c).GetRepo(id)
}

func GetRepoName(c context.Context, name string) (*model.Repo, error) {
	return FromContext(c).GetRepoName(name)
}

func GetRepoOwnerName(c context.Context, funll_name string) (*model.Repo, error) {
	return FromContext(c).GetRepoName(funll_name)
}
func GetRepoConfigId(c context.Context, config_id int64) (*model.Repo, error) {
	return FromContext(c).GetRepoConfigId(config_id)
}

func CreateRepo(c context.Context, repo *model.Repo) error {
	return FromContext(c).CreateRepo(repo)
}

func UpdateRepo(c context.Context, repo *model.Repo) error {
	return FromContext(c).UpdateRepo(repo)
}

func DeleteRepo(c context.Context, repo *model.Repo) error {
	return FromContext(c).DeleteRepo(repo)
}

func GetBuild(c context.Context, id string) (*model.Build, error) {
	return FromContext(c).GetBuild(id)
}

func GetBuildNumber(c context.Context, repo *model.Repo, num int) (*model.Build, error) {
	return FromContext(c).GetBuildNumber(repo, num)
}

func GetBuildRef(c context.Context, repo *model.Repo, ref string) (*model.Build, error) {
	return FromContext(c).GetBuildRef(repo, ref)
}

func GetBuildCommit(c context.Context, repo *model.Repo, sha, branch string) (*model.Build, error) {
	return FromContext(c).GetBuildCommit(repo, sha, branch)
}

func GetBuildLastByPipeline(c context.Context, pipelineID string) (*model.Build, error) {
	return FromContext(c).GetBuildLastByPipeline(pipelineID)
}

func GetBuildLast(c context.Context, repo *model.Repo) (*model.Build, error) {
	return FromContext(c).GetBuildLast(repo)
}

func GetBuildLastBefore(c context.Context, repo *model.Repo, branch string, buildID string) (*model.Build, error) {
	return FromContext(c).GetBuildLastBefore(repo, branch, buildID)
}

func GetBuildList(c context.Context, repo *model.Repo, state string, page, pageSize int) ([]*model.BuildRepo, error) {
	return FromContext(c).GetBuildList(repo, state, page, pageSize)
}

func GetBuildCount2(c context.Context, repo *model.Repo, state string) (int, error) {
	return FromContext(c).GetBuildCount2(repo, state)
}

func GetBuildQueue(c context.Context) ([]*model.Feed, error) {
	return FromContext(c).GetBuildQueue()
}

func CreateBuild(c context.Context, build *model.Build, procs ...*model.Proc) error {
	return FromContext(c).CreateBuild(build, procs...)
}

func UpdateBuild(c context.Context, build *model.Build) error {
	return FromContext(c).UpdateBuild(build)
}

//pipeline 相关
func GetPipelineEnvs(c context.Context, ConfigID string) ([]*model.PipelineEnv, error) {
	return FromContext(c).GetPipelineEnvs(ConfigID)
}

func UpdatePipelineEnvs(c context.Context, ConfigID string, pipelineEnvs []*model.PipelineEnv) error {
	return FromContext(c).UpdatePipelineEnvs(ConfigID, pipelineEnvs)
}

//project env
func GetProjectEnvs(c context.Context, project string) ([]*model.ProjectEnv, error) {
	return FromContext(c).GetProjectEnvs(project)
}

func UpdateProjectEnvs(c context.Context, project string, projectEnvs []*model.ProjectEnv) error {
	return FromContext(c).UpdateProjectEnvs(project, projectEnvs)
}

// 根据Repo删除Config
func ConfigDeleteRepo(c context.Context, repo *model.Repo) error {
	return FromContext(c).ConfigDeleteRepo(repo)
}
