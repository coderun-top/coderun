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
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	// "math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sirupsen/logrus"
	"github.com/coderun-top/coderun/src/model"
	"github.com/coderun-top/coderun/src/remote"
	"github.com/coderun-top/coderun/src/shared/httputil"
	"github.com/coderun-top/coderun/src/shared/token"
	"github.com/coderun-top/coderun/src/store"

	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/rpc"
	"github.com/coderun-top/coderun/src/core/pubsub"
	"github.com/coderun-top/coderun/src/core/queue"

	pb "github.com/coderun-top/coderun/src/grpc/user"
	"github.com/coderun-top/coderun/src/remote/bitbucket"
	"github.com/coderun-top/coderun/src/remote/coding"
	"github.com/coderun-top/coderun/src/remote/gitee"
	"github.com/coderun-top/coderun/src/remote/github"
	"github.com/coderun-top/coderun/src/remote/gitlab"
)

//
// CANARY IMPLEMENTATION
//
// This file is a complete disaster because I'm trying to wedge in some
// experimental code. Please pardon our appearance during renovations.
//

var skipRe = regexp.MustCompile(`\[(?i:ci *skip|skip *ci)\]`)

// func init() {
// 	rand.Seed(time.Now().UnixNano())
// }

func shasum(raw []byte) string {
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum)
}

func SetupRemote(provider int64, host string) (remote.Remote, error) {
	logrus.Debugf("SetupRemote host: %s", host)
	if provider == 2 {
		logrus.Debugf("setup gitlab")
		_url := "https://gitlab.com"
		if host != "" {
			_url = host
		}
		return gitlab.New(gitlab.Opts{
			URL:    _url,
			Client: os.Getenv("DRONE_GITLAB_CLIENT"),
			Secret: os.Getenv("DRONE_GITLAB_SECRET"),
		})
	} else if provider == 1 {
		logrus.Debugf("setup github")
		return github.New(github.Opts{
			URL:     "https://github.com",
			Context: "continuous-integration/coderun",
			Client:  os.Getenv("DRONE_GITHUB_CLIENT"),
			Secret:  os.Getenv("DRONE_GITHUB_SECRET"),
			// Scopes:  []string{"admin:gpg_key", "admin:org", "admin:org_hook", "admin:public_key", "admin:repo_hook", "delete_repo", "gist", "notifications", "repo", "user", "write:discussion"},
			// Scopes:  []string{"user", "repo", "admin:org", "admin:org_hook", "admin:repo_hook"},
		})
	} else if provider == 3 {
		logrus.Debugf("setup coding")
		return coding.New(coding.Opts{
			URL:    "https://coding.net",
			Client: os.Getenv("DRONE_CODING_CLIENT"),
			Secret: os.Getenv("DRONE_CODING_SECRET"),
			Scopes: []string{"user", "project", "project:depot"},
		})
	} else if provider == 4 {
		logrus.Debugf("setup bitbucket")
		return bitbucket.New("6F6fLJGGhydxVdAFfm", "A2DU3FtuYJdPnyy8qRZDacSfjjLDHEk2"), nil
	} else if provider == 5 {
		logrus.Debugf("setup gitee")
		_url := "https://gitee.com"
		return gitee.New(gitee.Opts{
			URL:    _url,
			Client: os.Getenv("DRONE_GITEE_CLIENT"),
			Secret: os.Getenv("DRONE_GITEE_SECRET"),
		})
	} else {
		logrus.Debugf("setup provider does not matched.")
		return nil, errors.New("setup provider does not matched")
	}
}

func PostHook(c *gin.Context) {
	//remote_ := remote.FromContext(c)
	// get the token and verify the hook is authorized
	pToken, err := token.ParseRequest(c.Request, func(t *token.Token) (string, error) {
		return token.Hash, nil
	})
	logrus.Debugf("Hook token parsed text: %s", pToken.Text)

	// pTokenValues := strings.Split(pToken.Text, ":")
	// repoID, err := strconv.ParseInt(pTokenValues[0], 10, 64)
	repoID, err := strconv.ParseInt(pToken.Text, 10, 64)
	if err != nil {
		logrus.Errorf("Hook RepoID parse failed. %s", err)
		c.AbortWithError(400, err)
		return
	}

	repo, err := store.GetRepo(c, repoID)
	if err != nil {
		logrus.Errorf("Cannot find repo by repoID: %s, %s", repoID, err)
		c.AbortWithError(400, err)
		return
	}
	if !repo.IsActive {
		logrus.Errorf("ignoring hook. %s/%s(%s) is inactive.", repo.Owner, repo.Name, repo.FullName)
		c.AbortWithError(204, err)
		return
	}

	// set remote
	providerName := c.Param("provider")
	var provider int64 = 0
	if providerName == "github" {
		provider = 1
	} else if providerName == "gitlab" {
		provider = 2
	} else if providerName == "coding" {
		provider = 3
	} else if providerName == "bitbucket" {
		provider = 4
	} else if providerName == "gitee" {
		provider = 5
	}
	// 原则上这里应该获取的是user表的host才对
	providerUrl, err := url.Parse(repo.Link)
	if err != nil {
		logrus.Errorf("repo_link url error", err)
		c.AbortWithError(400, err)
		return
	}

	remote_, err := SetupRemote(provider, providerUrl.Scheme+"://"+providerUrl.Host)
	if err != nil {
		logrus.Fatal("setup remote error. %s", err)
		c.AbortWithError(400, err)
		return
	}

	rRepo, build, err := remote_.Hook(c.Request)
	if err != nil {
		logrus.Errorf("failure to parse hook. %s", err)
		c.AbortWithError(400, err)
		return
	}

	if rRepo == nil {
		logrus.Errorf("failure to ascertain repo from hook.")
		c.Writer.WriteHeader(400)
		return
	}
	if build == nil {
		logrus.Errorf("failure to ascertain build from hook.")
		c.Writer.WriteHeader(200)
		return
	}

	// tmprepoJ, _ := json.Marshal(tmprepo)
	// buildJ, _ := json.Marshal(build)
	// logrus.Debugf("Hook from repo: %s, build: %s", tmprepoJ, buildJ)
	logrus.Debugf("Hook from repo: %s/%s(%s), branch: %s, message: %s", rRepo.Owner, rRepo.Name, rRepo.FullName, build.Branch, build.Message)

	// skip the build if any case-insensitive combination of the words "skip" and "ci"
	// wrapped in square brackets appear in the commit message
	skipMatch := skipRe.FindString(build.Message)
	if len(skipMatch) > 0 {
		logrus.Infof("ignoring hook. %s found in %s", skipMatch, build.Commit)
		c.Writer.WriteHeader(204)
		return
	}

	// 使用hook中的repo代替使用全名称查找到的repo
	// repo, err := store.GetRepoName(c, rRepo.FullName)
	// if err != nil {
	// 	logrus.Errorf("failure to find repo %s/%s from hook. %s", tmprepo.Owner, tmprepo.Name, err)
	// 	c.AbortWithError(404, err)
	// 	return
	// }
	// if !repo.IsActive {
	// 	logrus.Errorf("ignoring hook. %s/%s is inactive.", tmprepo.Owner, tmprepo.Name)
	// 	c.AbortWithError(204, err)
	// 	return
	// }

	// // get the token and verify the hook is authorized
	// parsed, err := token.ParseRequest(c.Request, func(t *token.Token) (string, error) {
	// 	return repo.Hash, nil
	// })
	// if err != nil {
	// 	logrus.Errorf("failure to parse token from hook for %s. %s", repo.FullName, err)
	// 	c.AbortWithError(400, err)
	// 	return
	// }
	// if parsed.Text != repo.FullName {
	// 	logrus.Errorf("failure to verify token from hook. Expected %s, got %s", repo.FullName, parsed.Text)
	// 	c.AbortWithStatus(403)
	// 	return
	// }

	if repo.UserID == 0 {
		logrus.Warnf("ignoring hook. repo %s has no owner.", repo.FullName)
		c.Writer.WriteHeader(204)
		return
	}
	var skipped = true
	if (build.Event == model.EventPush && repo.AllowPush) ||
		(build.Event == model.EventPull && repo.AllowPull) ||
		(build.Event == model.EventDeploy && repo.AllowDeploy) ||
		(build.Event == model.EventTag && repo.AllowTag) {
		skipped = false
	}

	if skipped {
		logrus.Infof("ignoring hook. repo %s is disabled for %s events.", repo.FullName, build.Event)
		c.Writer.WriteHeader(204)
		return
	}

	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		logrus.Errorf("04 failure to find repo owner %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}

	// build 图片设置为用户头像
	// build.Avatar = user.Avatar

	// if the remote has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the build.
	if refresher, ok := remote_.(remote.Refresher); ok {
		ok, _ := refresher.Refresh(user)
		if ok {
			store.UpdateUser(c, user)
		}
	}

	// load config 取其中一个
	// conf_id_str := c.Query("conf_id")
	// conf_id, _ := strconv.ParseInt(conf_id_str, 10, 64)
	// conf, _ := store.FromContext(c).ConfigLoad(conf_id)
	confs, err := store.FromContext(c).ConfigRepoFind(repo)
	if err != nil {
		logrus.Errorf("failure to find config for repo: %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}
	conf := confs[0]
	build.ConfigID = conf.ID
	if conf.File == 2 {
		confb, err := remote_.FileRef(user, repo, build.Branch, conf.FilePath)
		if err != nil {
			logrus.Errorf("error: %s: cannot find %s in %s: %s")
			c.AbortWithError(404, err)
			return
		}
		conf.Data = string(confb)
	}

	netrc, err := remote_.Netrc(user, repo)
	if err != nil {
		c.String(500, "Failed to generate netrc file. %s", err)
		return
	}

	// update some build fields
	build.RepoID = repo.ID
	build.Verified = true
	build.Status = model.StatusPending

	if repo.IsGated {
		allowed, _ := Config.Services.Senders.SenderAllowed(user, repo, build, conf)
		if !allowed {
			build.Status = model.StatusBlocked
		}
	}

	if err = Config.Services.Limiter.LimitBuild(user, repo, build); err != nil {
		c.String(403, "Build blocked by limiter")
		return
	}

	// verify the trigger can be built vs skipped
	parsed, err := yamlParseLinter(conf.Data, repo.IsTrusted)
	if err != nil {
		logrus.Errorf("yaml parse error: %s", err.Error())
		c.String(400, err.Error())
		return
	}
	if !yamlTriggerMatch(parsed, repo, build) {
		c.String(400, "Trigger does not match restrictions defined in yaml")
		return
	}

	build.Trim()
	err = store.CreateBuild(c, build, build.Procs...)
	if err != nil {
		logrus.Errorf("failure to save commit for %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, build)

	if build.Status == model.StatusBlocked {
		return
	}

	envs := map[string]string{}
	if Config.Services.Environ != nil {
		globals, _ := Config.Services.Environ.EnvironList(repo)
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}
	// add project envs
	projectEnvs, _ := store.GetProjectEnvs(c, user.ProjectID)
	for _, projectEnv := range projectEnvs {
		envs[projectEnv.Key] = projectEnv.Value
	}
	// add pipeline envs
	pipelineEnvs, _ := store.GetPipelineEnvs(c, build.ConfigID)
	for _, pipelineEnv := range pipelineEnvs {
		envs[pipelineEnv.Key] = pipelineEnv.Value
	}

	secs, err := Config.Services.Secrets.SecretListBuild(repo, build)
	if err != nil {
		logrus.Debugf("Error getting secrets for %s#%d. %s", repo.FullName, build.Number, err)
	}

	regs, err := Config.Services.Registries.RegistryList(user.ProjectID)
	if err != nil {
		logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	}
	helms, err := store.FromContext(c).HelmList(user.ProjectID)
	if err != nil {
		logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	}
	//reg, err := Config.Services.Registries.RegistryFind(user.Name, "cfcr")
	//if err != nil {
	//	logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	//}
	k8s, err := store.FromContext(c).K8sClusterList(user.ProjectID)
	if err != nil {
		logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	}

	// get the previous build so that we can send
	// on status change notifications
	last, _ := store.GetBuildLastBefore(c, repo, build.Branch, build.ID)

	//
	// BELOW: NEW
	//

	defer func() {
		uri := fmt.Sprintf("%s/%s/%d", httputil.GetURL(c.Request), repo.FullName, build.Number)
		err = remote_.Status(user, repo, build, uri)
		if err != nil {
			logrus.Errorf("03 error setting commit status for %s/%d: %v", repo.FullName, build.Number, err)
		}
	}()

	// pipeline_envs,_ := store.GetAllPilelineEnv(c,repo)
	count, err := store.FromContext(c).GetBuildListProjectCountByMonth(user.ProjectID, "")
	if err != nil {
		logrus.Debugf("count err:%s", err)
	}

	project, err := store.FromContext(c).GetProject(user.ProjectID)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	buildTime, err := store.FromContext(c).GetBuildTimeProject(project.ID)
	if err != nil {
		logrus.Debugf("count err:%s", err)
	}

	nums, err := GetProjectCharging(project.Name)
	if err != nil {
		logrus.Errorf("get num error: %s", err)
		build.Status = model.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = "get num error"
		store.UpdateBuild(c, build)
		return
	}
	if count > nums.BuildCount {
		build.Status = model.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = "Current Builds " + strconv.Itoa(count) + ", The amount has been more than " + strconv.Itoa(nums.BuildCount) + "/month"
		store.UpdateBuild(c, build)
		return
	}
	if buildTime > nums.BuildTime {
		build.Status = model.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = "Builds time " + strconv.FormatInt(buildTime, 10) + " , The amount has been more than " + strconv.FormatInt(nums.BuildTime, 10)
		store.UpdateBuild(c, build)
		return
	}

	r, err := Config.Services.Grpc.GetToken(context.Background(), &pb.TokenRequest{Name: project.Name})
	if err != nil {
		build.Status = model.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = strings.Split(err.Error(), "desc = ")[1]
		store.UpdateBuild(c, build)

		logrus.Errorf("get token error: %s", err.Error())
		return
	}
	logrus.Debugf("GetToken: %s", r.Message)

	b := Builder{
		Repo:  repo,
		User:  user,
		Curr:  build,
		Last:  last,
		Proc:  nil,
		Netrc: netrc,
		Secs:  secs,
		Regs:  regs,
		Helms: helms,
		K8s:   k8s,
		Envs:  envs,
		Link:  httputil.GetURL(c.Request),
		Yaml:  conf.Data,
		Token: r.Message,
	}
	items, err := b.Build(c)
	if err != nil {
		build.Status = model.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = err.Error()
		store.UpdateBuild(c, build)

		logrus.Errorf("build error: %s", err)
		return
	}

	var pcounter = len(items)
	for _, item := range items {
		build.Procs = append(build.Procs, item.Proc)
		item.Proc.BuildID = build.ID

		for _, stage := range item.Config.Stages {
			var gid int
			for _, step := range stage.Steps {
				pcounter++
				if gid == 0 {
					gid = pcounter
				}
				proc := &model.Proc{
					BuildID: build.ID,
					Name:    step.Alias,
					PID:     pcounter,
					PPID:    item.Proc.PID,
					PGID:    gid,
					State:   model.StatusPending,
				}
				logrus.Errorf("proc", step.Alias)
				if proc.Name != "rebuild_cache" {
					build.Procs = append(build.Procs, proc)
				}
			}
		}
	}
	err = store.FromContext(c).ProcCreate(build.Procs)
	if err != nil {
		logrus.Errorf("error persisting procs %s/%d: %s", repo.FullName, build.Number, err)
	}

	//
	// publish topic
	//
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"project": project.Name,
			"private": strconv.FormatBool(repo.IsPrivate),
			"repo_id": strconv.FormatInt(repo.ID, 10),
		},
	}
	// logrus.Errorf("Hook message Labels %s", message.Labels)
	buildCopy := *build
	buildCopy.Procs = model.Tree(buildCopy.Procs)
	message.Data, _ = json.Marshal(model.Event{
		Type:  model.Enqueued,
		Repo:  *repo,
		Build: buildCopy,
	})
	// TODO remove global reference
	Config.Services.Pubsub.Publish(c, "topic/events", message)
	//
	// end publish topic
	//
	var public bool
	public = true
	if conf != nil {
		public = conf.AgentPublic
	}

	for _, item := range items {
		task := new(queue.Task)
		task.ID = fmt.Sprint(item.Proc.ID)
		task.Labels = map[string]string{}
		for k, v := range item.Labels {
			task.Labels[k] = v
		}
		task.Labels["platform"] = item.Platform
		task.Labels["project"] = project.Name
		task.Labels["repo"] = b.Repo.FullName
		task.Labels["public"] = strconv.FormatBool(public)
		if item.Public == false {
			task.Labels["filter"] = item.Filter
			task.Labels["public"] = strconv.FormatBool(item.Public)
		} else {
			if public == false {
				task.Labels["filter"] = conf.AgentFilter
			}
		}

		task.Data, _ = json.Marshal(rpc.Pipeline{
			ID:      fmt.Sprint(item.Proc.ID),
			Config:  item.Config,
			Timeout: b.Repo.Timeout,
		})

		Config.Services.Logs.Open(context.Background(), task.ID)
		Config.Services.Queue.Push(context.Background(), task)
	}
}
