// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"encoding/base32"
	"fmt"
	"net/http"
	"strconv"
	// "os"
	// "path"
	// "encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"

	"github.com/coderun-top/coderun/src/model"
	"github.com/coderun-top/coderun/src/remote"
	"github.com/coderun-top/coderun/src/router/middleware/session"
	"github.com/coderun-top/coderun/src/shared/httputil"
	"github.com/coderun-top/coderun/src/shared/token"
	"github.com/coderun-top/coderun/src/store"
	//"github.com/coderun-top/coderun/src/remote/github"
	//"github.com/coderun-top/coderun/src/remote/gitlab"
	"encoding/base64"
)

func DeleteRepo(c *gin.Context) {
	user := session.User(c)
	repo := session.Repo(c)
	remote, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Fatalf("get remote error. %s", err)
		c.AbortWithError(400, err)
		return
	}

	err = store.ConfigDeleteRepo(c, repo)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = store.DeleteRepo(c, repo)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	remote.Deactivate(user, repo, httputil.GetURL(c.Request))
	c.JSON(200, repo)
}

func PatchRepo(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)

	in := new(model.RepoPatch)
	if err := c.Bind(in); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if (in.IsTrusted != nil || in.Timeout != nil) && !user.Admin {
		c.String(403, "Insufficient privileges")
		return
	}

	if in.AllowPush != nil {
		repo.AllowPush = *in.AllowPush
	}
	if in.AllowPull != nil {
		repo.AllowPull = *in.AllowPull
	}
	if in.AllowDeploy != nil {
		repo.AllowDeploy = *in.AllowDeploy
	}
	if in.AllowTag != nil {
		repo.AllowTag = *in.AllowTag
	}
	if in.IsGated != nil {
		repo.IsGated = *in.IsGated
	}
	if in.IsTrusted != nil {
		repo.IsTrusted = *in.IsTrusted
	}
	if in.Timeout != nil {
		repo.Timeout = *in.Timeout
	}
	if in.Config != nil {
		repo.Config = *in.Config
	}
	if in.Visibility != nil {
		switch *in.Visibility {
		case model.VisibilityInternal, model.VisibilityPrivate, model.VisibilityPublic:
			repo.Visibility = *in.Visibility
		default:
			c.String(400, "Invalid visibility type")
			return
		}
	}
	if in.BuildCounter != nil {
		repo.Counter = *in.BuildCounter
	}

	err := store.UpdateRepo(c, repo)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, repo)
}

func ChownRepo(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	repo.UserID = user.ID

	err := store.UpdateRepo(c, repo)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, repo)
}

func GetRepo2(c *gin.Context) {
	user := session.User(c)

	remote, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Fatalf("get remote error. %s", err)
		c.AbortWithError(400, err)
		return
	}

	repo, err := remote.Repo(user, "", c.Query("name"))
	if err != nil {
		c.String(404, "Cannot find repo. %s", err)
		return
	}
	c.JSON(http.StatusOK, repo)
}

func GetRepo(c *gin.Context) {
	c.JSON(http.StatusOK, session.Repo(c))
}

func RepairRepo(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	remote, err := SetupRemote(user.Provider, user.Host)

	// remote activate hook
	provider := "github"
	if user.Provider == 1 {
		provider = "github"
	} else if user.Provider == 2 {
		provider = "gitlab"
	} else if user.Provider == 3 {
		provider = "coding"
	} else if user.Provider == 4 {
		provider = "bitbucket"
	} else if user.Provider == 5 {
		provider = "gitee"
	}

	// creates the jwt token used to verify the repository
	t := token.New(token.HookToken, strconv.FormatInt(repo.ID, 10))
	sig, err := t.Sign(token.Hash)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	// reconstruct the link
	host := httputil.GetURL(c.Request)
	link := fmt.Sprintf(
		"%s/dougo/hook/%s?access_token=%s",
		host,
		provider,
		sig,
	)

	remote.Deactivate(user, repo, host)
	err = remote.Activate(user, repo, link)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	from, err := remote.Repo(user, "", repo.FullName)
	if err != nil {
		logrus.Errorf("remote repo get error: %s", err)
		c.AbortWithStatus(400)
		return
	}
	repo.Name = from.Name
	repo.Owner = from.Owner
	repo.FullName = from.FullName
	repo.Avatar = from.Avatar
	repo.Link = from.Link
	repo.Clone = from.Clone
	if repo.IsPrivate != from.IsPrivate {
		repo.IsPrivate = from.IsPrivate
		repo.ResetVisibility()
	}
	store.UpdateRepo(c, repo)
	logrus.Errorf("private: %s", from.IsPrivate)

	c.Writer.WriteHeader(http.StatusOK)
}

func MoveRepo(c *gin.Context) {
	remote := remote.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)

	to, exists := c.GetQuery("to")
	if !exists {
		err := fmt.Errorf("Missing required to query value")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	owner, name, errParse := model.ParseRepo(to)
	if errParse != nil {
		c.AbortWithError(http.StatusInternalServerError, errParse)
		return
	}

	from, err := remote.Repo(user, owner, name)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !from.Perm.Admin {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	repo.Name = from.Name
	repo.Owner = from.Owner
	repo.FullName = from.FullName
	repo.Avatar = from.Avatar
	repo.Link = from.Link
	repo.Clone = from.Clone
	repo.IsPrivate = from.IsPrivate
	if repo.IsPrivate != from.IsPrivate {
		repo.ResetVisibility()
	}

	errStore := store.UpdateRepo(c, repo)
	if errStore != nil {
		c.AbortWithError(http.StatusInternalServerError, errStore)
		return
	}

	// creates the jwt token used to verify the repository
	t := token.New(token.HookToken, strconv.FormatInt(repo.ID, 10))
	sig, err := t.Sign(token.Hash)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	// reconstruct the link
	host := httputil.GetURL(c.Request)
	link := fmt.Sprintf(
		"%s/hook?access_token=%s",
		host,
		sig,
	)

	remote.Deactivate(user, repo, host)
	err = remote.Activate(user, repo, link)
	if err != nil {
		c.String(500, err.Error())
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}

type RepoPostData struct {
	Branch      string `json:"branch"`
	FileContent string `json:"file_content"`
	FilePath    string `json:"file_path"`
	File        int64  `json:"file"`
	ConfigType  string `json:"config_type"`
	StepBuild   string `json:"step_build"`
}

func PostRepo(c *gin.Context) {
	projectName := c.Param("projectname")
	project, err := store.FromContext(c).GetProjectName(projectName)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	user := session.User(c)

	var postData RepoPostData
	c.BindJSON(&postData)

	repoName := c.Query("name")
	// repoChoiceBranch := postData.Branch
	fileContent := postData.FileContent

	_, err = store.FromContext(c).GetRepoFullName(user, repoName)
	if err == nil {
		logrus.Errorf("repo exiting for name %s.", repoName)
		c.String(400, fmt.Sprintf("repo exiting for name %s.", repoName))
		return
	}

	//provider_string := strconv.FormatInt(user.Provider, 10)
	remote, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Errorf("setup remote error. %s", err)
		c.AbortWithStatus(400)
		return
	}

	repo, err := remote.Repo(user, "", repoName)
	if err != nil {
		logrus.Errorf("remote repo get error: %s", err)
		c.AbortWithStatus(400)
		return
	}

	repo.ProjectID = project.ID
	// repo.Update(from)
	repo.FullName = repoName
	// repo.FullName = fmt.Sprintf("%s_%s", repoName, user.TokenName)

	// 如果项目logo为空，设置为用户头像
	if repo.Avatar == "" {
		userAvatar, ok := c.Get("headimage")
		if !ok {
			c.String(500, "Error fetching feed.")
			return
		}
		//repo.Avatar = user.Avatar
		repo.Avatar = userAvatar.(string)
	}

	if repo.IsActive {
		logrus.Errorf("Repository is already active. %s", err)
		c.AbortWithStatus(409)
		return
	}
	repo.IsActive = true

	if err := Config.Services.Limiter.LimitRepo(user, repo); err != nil {
		logrus.Errorf("Repository activation blocked by limiter. %s", err)
		c.AbortWithStatus(403)
		return
	}

	repo.UserID = user.ID
	repo.IsTrusted = true
	if !repo.AllowPush && !repo.AllowPull && !repo.AllowDeploy && !repo.AllowTag {
		repo.AllowPush = true
		repo.AllowPull = true
	}
	if repo.Visibility == "" {
		repo.Visibility = model.VisibilityPublic
		if repo.IsPrivate {
			repo.Visibility = model.VisibilityPrivate
		}
	}
	if repo.Config == "" {
		repo.Config = Config.Server.RepoConfig
	}
	if repo.Timeout == 0 {
		repo.Timeout = 60 // 1 hour default build time
	}
	if repo.Hash == "" {
		repo.Hash = base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		)
	}

	// create repo & config
	config := &model.Config{}
	confb := []byte(fileContent)
	// yaml文件格式验证
	if postData.File == 1 {
		_, err = yamlParseLinter(fileContent, repo.IsTrusted)
		if err != nil {
			c.String(400, "yaml parse error: %s", err.Error())
			return
		}
	}

	sha := shasum(confb)
	config = &model.Config{
		Data:                fileContent,
		Hash:                sha,
		File:                postData.File,
		FilePath:            postData.FilePath,
		Branche:             postData.Branch,
		AgentPublic:         true,
		AgentFilter:         "",
		ConfigType:          postData.ConfigType,
		StepBuild:           postData.StepBuild,
		StepUnitTest:        "",
		StepIntegrationTest: "",
		StepDeploy:          "",
	}

	err = store.FromContext(c).CreateRepoConfigs(repo, config)
	if err != nil {
		logrus.Errorf("create repo config error. %s", err)
		c.AbortWithStatus(500)
		return
	}

	// remote provider activate hook
	t := token.New(token.HookToken, strconv.FormatInt(repo.ID, 10))
	sig, err := t.Sign(token.Hash)
	if err != nil {
		logrus.Errorf("create jwt token error. %s", err)
		c.AbortWithStatus(500)
		return
	}

	provider := "github"
	if user.Provider == 1 {
		provider = "github"
	} else if user.Provider == 2 {
		provider = "gitlab"
	} else if user.Provider == 3 {
		provider = "coding"
	} else if user.Provider == 4 {
		provider = "bitbucket"
	} else if user.Provider == 5 {
		provider = "gitee"
	}

	link := fmt.Sprintf(
		"%s/dougo/hook/%s?access_token=%s",
		httputil.GetURL(c.Request),
		provider,
		sig,
	)

	err = remote.Activate(user, repo, link)
	if err != nil {
		logrus.Errorf("remote activate error. %s", err)
		c.AbortWithStatus(500)
		return
	}

	// get created repo all info to return
	newrepo, err := store.FromContext(c).GetRepo(repo.ID)
	if err != nil {
		logrus.Errorf("find repo error. %s", err)
		c.AbortWithStatus(400)
		return
	}

	c.JSON(200, newrepo)
}

func GetBranches(c *gin.Context) {
	user := session.User(c)
	remote, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Errorf("get remote error. %s", err)
		c.AbortWithError(400, err)
		return
	}
	branches, err := remote.Branches(user, "", c.Query("name"))

	c.JSON(http.StatusOK, branches)
}

func EmptyFile(c *gin.Context) {
	user := session.User(c)

	branch := c.Query("branch")
	file := c.Query("file")

	remote, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Errorf("get remote error. %s", err)
		c.AbortWithError(400, err)
		return
	}

	repo := &model.Repo{FullName: c.Query("name")}
	_, err = remote.FileRef(user, repo, branch, file)
	if err != nil {
		c.String(405, "Cannot find file. %s", err)
	} else {
		c.String(200, "OK")
	}
}

func GetFile(c *gin.Context) {
	user := session.User(c)

	branch := c.Query("branch")
	file := c.Query("file")
	docker_context := c.Query("context")
	// name := c.Query("name")
	if docker_context == "" {
		docker_context = "/"
	}
	// if name == ""{
	// 	name = c.Query("name")
	// }

	remote, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Errorf("get remote error. %s", err)
		c.AbortWithError(400, err)
		return
	}

	repo := &model.Repo{FullName: c.Query("name")}
	confb, err := remote.FileRef(user, repo, branch, file)
	if err != nil {
		logrus.Errorf("error: %s: cannot find %s in %s: %s")
		c.AbortWithError(404, err)
		return
	}

	// 修改成由前端生成文件
	// filename := path.Base(file)

	// if filename == "Dockerfile" {
	// 	// imagename := user.Name+"/"+name
	// 	imagename := name
	// 	confb = []byte("steps:\n  docker:\n    image: crun/docker\n    registry_name: coderun\n    repo_name: " + imagename + "\n    dockerfile: " + file + "\n    context: " + docker_context + "\n    tags: latest\n\nbranches: [" + branch + "]")
	// }
	// logrus.Errorf("error: %s", string(confb))

	c.String(200, string(confb))
}

func Dig(c *gin.Context) {
	token := c.Query("token")
	provider_str := c.Query("provider")
	host := c.Query("host")
	login := c.Query("login")
	var provider int64 = 0

	if provider_str == "github" {
		provider = 1
	} else if provider_str == "gitlab" {
		provider = 2
	} else if provider_str == "coding" {
		provider = 3
	} else if provider_str == "bitbucket" {
		provider = 4
		username_password := login + ":" + token
		token = base64.StdEncoding.EncodeToString([]byte(username_password))
	} else if provider_str == "gitee" {
		provider = 5
	}

	remote, _ := SetupRemote(provider, host)
	_, err := remote.Auth(token, "")
	if err != nil {
		logrus.Error(err)
		c.String(400, "token invalid")
		return
	}
	c.String(200, "OK")
	return
}

func Test(c *gin.Context) {
	c.String(200, "OK")
	return
}
