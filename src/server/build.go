package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	// "strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/rpc"
	"github.com/coderun-top/coderun/src/core/pubsub"
	"github.com/coderun-top/coderun/src/core/queue"
	"github.com/coderun-top/coderun/src/remote"
	"github.com/coderun-top/coderun/src/shared/httputil"
	"github.com/coderun-top/coderun/src/store"

	pb "github.com/coderun-top/coderun/src/grpc/user"
	"github.com/coderun-top/coderun/src/model"
	"github.com/coderun-top/coderun/src/router/middleware/session"
)

type BuildRepoResult struct {
	Data     []*model.BuildRepo `json:"data"`
	Count    int                `json:"count"`
	PageSize int                `json:"pagesize"`
	Page     int                `json:"page"`
}

func GetProjectBuilds(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pagesize, _ := strconv.Atoi(c.DefaultQuery("pagesize", "10"))
	state := c.Query("state")

	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	builds, err := store.FromContext(c).GetBuildListProject(project.ID, state, page, pagesize)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	count, err := store.FromContext(c).GetBuildListProjectCountByMonth(project.ID, state)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	data := &BuildRepoResult{
		Data:     builds,
		Count:    count,
		Page:     page,
		PageSize: pagesize,
	}
	c.JSON(http.StatusOK, data)
}

func GetBuilds(c *gin.Context) {
	repo := session.Repo(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pagesize, _ := strconv.Atoi(c.DefaultQuery("pagesize", "10"))
	state := c.Query("state")
	// if state != "pending"  && state != "skipped" && state != "running" && state != "success" && state != "failure" && state != "killed" && state != "error" && state != "blocked" && state != "declined" {
	// 	state = "all"
	// }

	builds, err := store.GetBuildList(c, repo, state, page, pagesize)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	count, err := store.GetBuildCount2(c, repo, state)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	data := &BuildRepoResult{
		Data:     builds,
		Count:    count,
		Page:     page,
		PageSize: pagesize,
	}
	c.JSON(http.StatusOK, data)
}

func GetBuild(c *gin.Context) {
	buildID := c.Param("number")

	if buildID == "latest" {
		GetBuildLast(c)
		return
	}

	// repo := session.Repo(c)

	build, err := store.GetBuild(c, buildID)
	if err != nil {
		c.String(400, "build cannot find. %s", err.Error())
		return
	}

	files, _ := store.FromContext(c).FileList(build)
	procs, _ := store.FromContext(c).ProcList(build)
	build.Procs = model.Tree(procs)
	build.Files = files

	c.JSON(http.StatusOK, build)
}

func GetBuildLast(c *gin.Context) {
	repo := session.Repo(c)
	// branch := c.DefaultQuery("branch", repo.Branch)

	// build, err := store.GetBuildLast(c, repo, branch)
	build, err := store.GetBuildLast(c, repo)
	if err != nil {
		c.String(400, "build cannot find. %s", err.Error())
		// c.String(http.StatusInternalServerError, err.Error())
		return
	}

	files, _ := store.FromContext(c).FileList(build)
	procs, _ := store.FromContext(c).ProcList(build)
	build.Procs = model.Tree(procs)
	build.Files = files

	c.JSON(http.StatusOK, build)
}

func GetBuildLogs(c *gin.Context) {
	repo := session.Repo(c)

	// parse the build number and job sequence number from
	// the repquest parameter.
	num, _ := strconv.Atoi(c.Params.ByName("number"))
	ppid, _ := strconv.Atoi(c.Params.ByName("pid"))
	name := c.Params.ByName("proc")

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	proc, err := store.FromContext(c).ProcChild(build, ppid, name)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	rc, err := store.FromContext(c).LogFind(proc)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	defer rc.Close()

	c.Header("Content-Type", "application/json")
	io.Copy(c.Writer, rc)
}

func GetProcLogs(c *gin.Context) {
	repo := session.Repo(c)

	// parse the build number and job sequence number from
	// the repquest parameter.
	num, _ := strconv.Atoi(c.Params.ByName("number"))
	pid, _ := strconv.Atoi(c.Params.ByName("pid"))

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	proc, err := store.FromContext(c).ProcFind(build, pid)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	rc, err := store.FromContext(c).LogFind(proc)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	defer rc.Close()

	c.Header("Content-Type", "application/json")
	io.Copy(c.Writer, rc)
}

func DeleteBuild(c *gin.Context) {
	repo := session.Repo(c)

	// parse the build number and job sequence number from
	// the repquest parameter.
	num, _ := strconv.Atoi(c.Params.ByName("number"))
	seq, _ := strconv.Atoi(c.Params.ByName("job"))

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	proc, err := store.FromContext(c).ProcFind(build, seq)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	if proc.State != model.StatusRunning {
		c.String(400, "Cannot cancel a non-running build")
		return
	}

	proc.State = model.StatusKilled
	proc.Stopped = time.Now().Unix()
	if proc.Started == 0 {
		proc.Started = proc.Stopped
	}
	proc.ExitCode = 137
	// TODO cancel child procs
	store.FromContext(c).ProcUpdate(proc)

	Config.Services.Queue.Error(context.Background(), fmt.Sprint(proc.ID), queue.ErrCancel)
	c.String(204, "")
}

// ZombieKill kills zombie processes stuck in an infinite pending
// or running state. This can only be invoked by administrators and
// may have negative effects.
func ZombieKill(c *gin.Context) {
	// repo := session.Repo(c)

	// parse the build number and job sequence number from
	// the repquest parameter.
	// num, _ := strconv.Atoi(c.Params.ByName("number"))

	// build, err := store.GetBuildNumber(c, repo, num)
	// if err != nil {
	// 	c.String(404, err.Error())
	// 	return
	// }

	buildID := c.Param("number")
	build, err := store.GetBuild(c, buildID)
	if err != nil {
		msg := "failure to get build"
		logrus.Errorf("%s, %s", msg, err.Error())
		c.String(404, fmt.Sprintf("%s, %s", msg, err.Error()))
		return
	}

	procs, err := store.FromContext(c).ProcList(build)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	// kill all status
	// if build.Status != model.StatusRunning {
	// 	c.String(400, "Cannot force cancel a non-running build")
	// 	return
	// }

	for _, proc := range procs {
		if proc.Running() {
			proc.State = model.StatusKilled
			proc.ExitCode = 137
			proc.Stopped = time.Now().Unix()
			if proc.Started == 0 {
				proc.Started = proc.Stopped
			}
		}
	}

	for _, proc := range procs {
		store.FromContext(c).ProcUpdate(proc)
		Config.Services.Queue.Error(context.Background(), fmt.Sprint(proc.ID), queue.ErrCancel)
		// if not have running try to cancel pendding
		Config.Services.Queue.Evict(context.Background(), fmt.Sprint(proc.ID))
	}

	build.Status = model.StatusKilled
	build.Finished = time.Now().Unix()
	store.FromContext(c).UpdateBuild(build)

	c.String(204, "")
}

func PostApproval(c *gin.Context) {
	var (
		remote_ = remote.FromContext(c)
		repo    = session.Repo(c)
		user    = session.User(c)
		num, _  = strconv.Atoi(
			c.Params.ByName("number"),
		)
	)

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.String(404, err.Error())
		return
	}
	if build.Status != model.StatusBlocked {
		c.String(500, "cannot decline a build with status %s", build.Status)
		return
	}
	build.Status = model.StatusPending
	build.Reviewed = time.Now().Unix()
	build.Reviewer = user.Login

	//
	//
	// This code is copied pasted until I have a chance
	// to refactor into a proper function. Lots of changes
	// and technical debt. No judgement please!
	//
	//

	// fetch the build file from the database
	conf, err := Config.Storage.Config.ConfigLoad(build.ConfigID)
	if err != nil {
		logrus.Errorf("failure to get build config for %s. %s", repo.FullName, err)
		c.String(404, err.Error())
		return
	}
	if conf.File == 2 {
		confb, err := remote_.FileRef(user, repo, build.Branch, conf.FilePath)
		if err != nil {
			logrus.Errorf("error: %s: cannot find %s in %s: %s")
			c.String(404, err.Error())
			return
		}
		conf.Data = string(confb)
	}

	netrc, err := remote_.Netrc(user, repo)
	if err != nil {
		c.String(500, "Failed to generate netrc file. %s", err)
		return
	}

	if uerr := store.UpdateBuild(c, build); err != nil {
		c.String(500, "error updating build. %s", uerr)
		return
	}

	c.JSON(200, build)

	// get the previous build so that we can send
	// on status change notifications
	last, _ := store.GetBuildLastBefore(c, repo, build.Branch, build.ID)
	secs, err := Config.Services.Secrets.SecretListBuild(repo, build)
	if err != nil {
		logrus.Debugf("Error getting secrets for %s#%d. %s", repo.FullName, build.Number, err)
	}
	regs, err := Config.Services.Registries.RegistryList(repo.ProjectID)
	if err != nil {
		logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	}

	//reg, err := Config.Services.Registries.RegistryFind("projectName", "cfcr")
	//if err != nil {
	//	logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	//}
	helms, err := store.FromContext(c).HelmList(repo.ProjectID)
	if err != nil {
		logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	}
	k8s, err := store.FromContext(c).K8sClusterList(repo.ProjectID)
	if err != nil {
		logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	}
	envs := map[string]string{}
	if Config.Services.Environ != nil {
		globals, _ := Config.Services.Environ.EnvironList(repo)
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}
	// add project envs
	projectEnvs, _ := store.GetProjectEnvs(c, repo.ProjectID)
	for _, projectEnv := range projectEnvs {
		envs[projectEnv.Key] = projectEnv.Value
	}
	// add pipeline envs
	pipelineEnvs, _ := store.GetPipelineEnvs(c, build.ConfigID)
	for _, pipelineEnv := range pipelineEnvs {
		envs[pipelineEnv.Key] = pipelineEnv.Value
	}

	defer func() {
		uri := fmt.Sprintf("%s/%s/%d", httputil.GetURL(c.Request), repo.FullName, build.Number)
		err = remote_.Status(user, repo, build, uri)
		if err != nil {
			logrus.Errorf("01 error setting commit status for %s/%d: %v", repo.FullName, build.Number, err)
		}
	}()

	b := Builder{
		Repo:  repo,
		User:  user,
		Curr:  build,
		Last:  last,
		Netrc: netrc,
		Secs:  secs,
		Regs:  regs,
		Helms: helms,
		K8s:   k8s,
		Link:  httputil.GetURL(c.Request),
		Yaml:  conf.Data,
		Envs:  envs,
	}
	items, err := b.Build(c)
	if err != nil {
		build.Status = model.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = err.Error()
		store.UpdateBuild(c, build)
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
				logrus.Debugf("proc", step.Alias)
				if proc.Name != "rebuild_cache" {
					build.Procs = append(build.Procs, proc)
				}
			}
		}
	}
	store.FromContext(c).ProcCreate(build.Procs)

	//
	// publish topic
	//
	buildCopy := *build
	buildCopy.Procs = model.Tree(buildCopy.Procs)
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsPrivate),
		},
	}
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

	project, err := store.FromContext(c).GetProject(b.Repo.ProjectID)
	if err != nil {
		logrus.Errorf("Cannot find project11. %s", err)
		c.String(404, "Cannot find project11. %s", err)
		return
	}

	for _, item := range items {
		logrus.Debugf("item.Proc.ID", item.Proc.ID)
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

func PostDecline(c *gin.Context) {
	var (
		remote_ = remote.FromContext(c)
		repo    = session.Repo(c)
		user    = session.User(c)
		num, _  = strconv.Atoi(
			c.Params.ByName("number"),
		)
	)

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.String(404, err.Error())
		return
	}
	if build.Status != model.StatusBlocked {
		c.String(500, "cannot decline a build with status %s", build.Status)
		return
	}
	build.Status = model.StatusDeclined
	build.Reviewed = time.Now().Unix()
	build.Reviewer = user.Login

	err = store.UpdateBuild(c, build)
	if err != nil {
		c.String(500, "error updating build. %s", err)
		return
	}

	uri := fmt.Sprintf("%s/%s/%d", httputil.GetURL(c.Request), repo.FullName, build.Number)
	err = remote_.Status(user, repo, build, uri)
	if err != nil {
		logrus.Errorf("02 error setting commit status for %s/%d: %v", repo.FullName, build.Number, err)
	}

	c.JSON(200, build)
}

func GetBuildQueue(c *gin.Context) {
	out, err := store.GetBuildQueue(c)
	if err != nil {
		c.String(500, "Error getting build queue. %s", err)
		return
	}
	c.JSON(200, out)
}

//
//
//
//
//
//
func CreateBuild(c *gin.Context) {
	v, ok := c.Get("headimage")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	repo := session.Repo(c)
	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		logrus.Errorf("02 failure to find repo owner %s. %s", repo.FullName, err)
		c.String(500, err.Error())
		return
	}

	branchName := c.Query("branch")
	if err != nil {
		logrus.Errorf("failure to get branch")
		c.String(500, err.Error())
		return
	}

	//remote_ := remote.FromContext(c)
	remote_, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Fatal("get remote error. %s", err)
		c.String(400, err.Error())
		return
	}

	branch, err := remote_.Branch(user, "", repo.FullName, branchName)
	if err != nil {
		logrus.Errorf("failure to get branch %d. %s", branchName, err)
		c.String(404, err.Error())
		return
	}

	branchJ, _ := json.Marshal(branch)
	logrus.Debugf("get branch: %s", branchJ)

	// may be stale. Therefore, we should refresh prior to dispatching
	// the job.
	if refresher, ok := remote_.(remote.Refresher); ok {
		ok, _ := refresher.Refresh(user)
		if ok {
			store.UpdateUser(c, user)
		}
	}

	netrc, err := remote_.Netrc(user, repo)
	if err != nil {
		logrus.Errorf("failure to generate netrc for %s. %s", repo.FullName, err)
		c.String(500, err.Error())
		return
	}

	project, err := store.FromContext(c).GetProject(repo.ProjectID)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	r, err := Config.Services.Grpc.GetToken(context.Background(), &pb.TokenRequest{Name: project.Name})
	if err != nil {
		msg := fmt.Sprintf("get token error: %+v", err)
		logrus.Error(msg)
		c.String(500, msg)
		return
	}
	userToken := r.Message

	err = ChargingCheck(c, project)
	if err != nil {
		msg := fmt.Sprintf("Charging Check: %s", err.Error())
		logrus.Error(msg)
		c.String(404, msg)
		return
	}

	var build = new(model.Build)
	build.ID = ""
	build.RepoID = repo.ID
	build.Number = 0
	build.Parent = 0
	build.Status = model.StatusPending
	build.Started = time.Now().Unix()
	build.Finished = build.Started
	build.Enqueued = time.Now().UTC().Unix()
	build.Created = 0
	build.Error = ""
	build.Deploy = ""
	build.Event = model.EventPush
	build.Branch = branchName
	build.Ref = "refs/heads/" + branchName
	build.Refspec = ""

	// build.Title    = ""
	// build.Timestamp int64   `json:"timestamp"     meddler:"build_timestamp"`

	build.Commit = branch.Commit.SHA
	build.Link = branch.Commit.URL
	build.Message = branch.Commit.Message
	build.Author = branch.Commit.AuthorName
	build.Email = branch.Commit.AuthorEmail
	logrus.Debugf("branch: %s", branch)

	if user.Provider == 5 {
		build.Message = branch.Commit.Commit.Message
	}

	//build.Avatar = user.Avatar
	build.Avatar = v.(string)
	build.Sender = user.Login
	build.Remote = repo.Clone

	buildJ, _ := json.Marshal(build)
	logrus.Debugf("create build: %s", buildJ)

	// create pipeline
	confs, err := store.FromContext(c).ConfigRepoFind(repo)
	if err != nil {
		logrus.Errorf("failure to find config for repo: %s. %s", repo.FullName, err)
		c.String(500, err.Error())
		return
	}
	conf := confs[0]
	build.ConfigID = conf.ID
	if conf.File == 2 {
		confb, err := remote_.FileRef(user, repo, build.Branch, conf.FilePath)
		if err != nil {
			logrus.Errorf("error: %s: cannot find %s in %s: %s")
			c.String(404, err.Error())
			return
		}
		conf.Data = string(confb)
	}

	// verify the trigger can be built vs skipped
	parsed, err := yamlParseLinter(conf.Data, repo.IsTrusted)
	if err != nil {
		logrus.Errorf("yaml parse error: %s", err.Error())
		c.String(404, err.Error())
		return
	}
	if !yamlTriggerMatch(parsed, repo, build) {
		c.String(404, "Trigger does not match restrictions defined in yaml")
		return
	}

	err = store.CreateBuild(c, build)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	RunBuild(c, netrc, user, repo, build, nil, conf, project, userToken)
}

func PostBuild(c *gin.Context) {
	v, ok := c.Get("headimage")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	repo := session.Repo(c)
	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		logrus.Errorf("03 failure to find repo owner %s. %s", repo.FullName, err)
		c.String(500, err.Error())
		return
	}
	//remote_ := remote.FromContext(c)
	remote_, err := SetupRemote(user.Provider, "")
	if err != nil {
		logrus.Fatal("get remote error. %s", err)
		c.String(400, err.Error())
		return
	}

	buildID := c.Param("number")
	build, err := store.GetBuild(c, buildID)
	if err != nil {
		msg := "failure to get build"
		logrus.Errorf("%s, %s", msg, err.Error())
		c.String(404, fmt.Sprintf("%s, %s", msg, err.Error()))
		return
	}

	switch build.Status {
	case model.StatusDeclined,
		model.StatusBlocked:
		c.String(500, "cannot restart a build with status %s", build.Status)
		return
	}

	// if the remote has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the job.
	if refresher, ok := remote_.(remote.Refresher); ok {
		ok, _ := refresher.Refresh(user)
		if ok {
			store.UpdateUser(c, user)
		}
	}

	// fetch the .drone.yml file from the database
	conf, err := Config.Storage.Config.ConfigLoad(build.ConfigID)
	if err != nil {
		logrus.Errorf("failure to get build config for %s. %s", repo.FullName, err)
		c.String(404, err.Error())
		return
	}
	if conf.File == 2 {
		confb, err := remote_.FileRef(user, repo, build.Branch, conf.FilePath)
		if err != nil {
			logrus.Errorf("error: %s: cannot find %s in %s: %s")
			c.String(404, err.Error())
			return
		}
		conf.Data = string(confb)
	}

	netrc, err := remote_.Netrc(user, repo)
	if err != nil {
		logrus.Errorf("failure to generate netrc for %s. %s", repo.FullName, err)
		c.String(500, err.Error())
		return
	}

	project, err := store.FromContext(c).GetProject(repo.ProjectID)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	r, err := Config.Services.Grpc.GetToken(context.Background(), &pb.TokenRequest{Name: project.Name})
	if err != nil {
		msg := fmt.Sprintf("get token error: %+v", err)
		logrus.Error(msg)
		c.String(500, msg)
		return
	}
	userToken := r.Message

	err = ChargingCheck(c, project)
	if err != nil {
		msg := fmt.Sprintf("Charging Check: %s", err.Error())
		logrus.Error(msg)
		c.String(404, msg)
		return
	}

	build.ID = ""
	build.Number = 0
	build.Parent = build.Number
	build.Status = model.StatusPending
	build.Started = time.Now().Unix()
	build.Finished = build.Started
	build.Enqueued = time.Now().UTC().Unix()
	build.Error = ""
	build.Deploy = c.DefaultQuery("deploy_to", build.Deploy)
	build.Avatar = v.(string)

	event := c.DefaultQuery("event", build.Event)
	if event == model.EventPush ||
		event == model.EventPull ||
		event == model.EventTag ||
		event == model.EventDeploy {
		build.Event = event
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

	err = store.CreateBuild(c, build)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	RunBuild(c, netrc, user, repo, build, nil, conf, project, userToken)
}

func RunBuild(c *gin.Context, netrc *model.Netrc, user *model.User, repo *model.Repo, build *model.Build, proc *model.Proc, conf *model.Config, project *model.Project, userToken string) {
	// Read query string parameters into envs, exclude reserved params
	var envs = map[string]string{}
	for key, val := range c.Request.URL.Query() {
		switch key {
		case "fork", "event", "deploy_to":
		default:
			// We only accept string literals, because build parameters will be
			// injected as environment variables
			envs[key] = val[0]
		}
	}

	// get the previous build so that we can send
	// on status change notifications
	last, _ := store.GetBuildLastBefore(c, repo, build.Branch, build.ID)
	secs, err := Config.Services.Secrets.SecretListBuild(repo, build)
	if err != nil {
		msg := fmt.Sprintf("Error getting secrets for %s#%d. %+v", repo.FullName, build.Number, err)
		logrus.Error(msg)
		c.String(404, msg)
		return
	}

	regs, err := Config.Services.Registries.RegistryList(repo.ProjectID)
	if err != nil {
		msg := fmt.Sprintf("Error getting registry credentials for %s#%d. %+v", repo.FullName, build.Number, err)
		logrus.Error(msg)
		c.String(404, msg)
		return
	}

	if Config.Services.Environ != nil {
		globals, _ := Config.Services.Environ.EnvironList(repo)
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}
	// add project envs
	projectEnvs, _ := store.GetProjectEnvs(c, repo.ProjectID)
	for _, projectEnv := range projectEnvs {
		envs[projectEnv.Key] = projectEnv.Value
	}

	// add pipeline envs
	pipelineEnvs, _ := store.GetPipelineEnvs(c, build.ConfigID)
	for _, pipelineEnv := range pipelineEnvs {
		envs[pipelineEnv.Key] = pipelineEnv.Value
	}

	helms, err := store.FromContext(c).HelmList(project.ID)
	if err != nil {
		msg := fmt.Sprintf("Error getting Helm for Project: %s. %+v", project.Name, err)
		logrus.Error(msg)
		c.String(404, msg)
		return
	}

	k8s, err := store.FromContext(c).K8sClusterList(project.ID)
	if err != nil {
		msg := fmt.Sprintf("Error getting kubernetes for Project: %s. %+v", project.Name, err)
		logrus.Error(msg)
		c.String(404, msg)
		return
	}

	b := Builder{
		Repo:  repo,
		User:  user,
		Curr:  build,
		Last:  last,
		Proc:  proc,
		Netrc: netrc,
		Secs:  secs,
		Regs:  regs,
		Helms: helms,
		K8s:   k8s,
		Link:  httputil.GetURL(c.Request),
		Yaml:  conf.Data,
		Envs:  envs,
		Token: userToken,
	}
	items, err := b.Build(c)
	if err != nil {
		build.Status = model.StatusError
		// 如果是Step重新Build不改变Build的时间
		if proc == nil {
			build.Started = time.Now().Unix()
			build.Finished = build.Started
		}
		build.Error = err.Error()
		store.UpdateBuild(c, build)

		logrus.Errorf("build error: %s", err)
		c.String(500, err.Error())
		return
	}

	// 标记proc_pid的第一组为pendding状态
	// pproc, err := store.FromContext(c).ProcFind(build, proc.PPID)
	// if proc != nil {
	// 	pproc.BuildID = build.ID
	// 	pproc.State = model.StatusRunning
	// 	err = store.FromContext(c).ProcUpdate(pproc)
	// 	if err != nil {
	// 		logrus.Errorf("proc update error: %s", err)
	// 		c.AbortWithError(500, err)
	// 		return
	// 	}
	// }

	var pcounter = len(items)
	for _, item := range items {
		// if stepNum == 0 {
		// 	build.Procs = append(build.Procs, item.Proc)
		// 	item.Proc.BuildID = build.ID
		// } else {
		// 	item.Proc.BuildID = build.ID
		// 	item.Proc = pproc
		// 	build.Procs = append(build.Procs, item.Proc)
		// }
		// item.Proc.BuildID = build.ID
		build.Procs = append(build.Procs, item.Proc)

		for _, stage := range item.Config.Stages {
			var gid int

			// logrus.Debugf("Stage Name: %s, Alias: %s", stage.Name, stage.Alias)
			for _, step := range stage.Steps {
				pcounter++
				if gid == 0 {
					gid = pcounter
				}

				logrus.Debugf("Step Name: %s, Alias: %s", step.Name, step.Alias)

				// if proc.Name != "rebuild_cache" {
				if step.Alias != "rebuild_cache" {
					cproc := &model.Proc{
						BuildID: item.Proc.BuildID,
						Name:    step.Alias,
						PID:     pcounter,
						PPID:    item.Proc.PID,
						PGID:    gid,
						State:   model.StatusPending,
					}

					if proc != nil {
						cproc, err = store.FromContext(c).ProcChild(build, item.Proc.PID, step.Alias)
						cproc.State = model.StatusPending
						if err != nil {
							logrus.Errorf("proc find error: %s", err)
							c.String(500, err.Error())
							return
						}
					}

					build.Procs = append(build.Procs, cproc)
				}
			}
		}
	}

	if proc == nil {
		err = store.FromContext(c).ProcCreate(build.Procs)
		if err != nil {
			build.Status = model.StatusError
			build.Started = time.Now().Unix()
			build.Finished = build.Started
			build.Error = err.Error()
			store.UpdateBuild(c, build)

			logrus.Errorf("cannot restart %s#%d: %s", repo.FullName, build.Number, err)
			c.String(500, err.Error())
			return
		}
		build.Procs = model.Tree(build.Procs)
	} else {
		for i := range build.Procs {
			err := store.FromContext(c).ProcUpdate(build.Procs[i])
			if err != nil {
				build.Error = err.Error()
				store.UpdateBuild(c, build)
				logrus.Errorf("proc update error %s#%d: %s", repo.FullName, build.Number, err)
				c.String(500, err.Error())
				return
			}
		}

		// build, err = store.FromContext(c).GetBuildNumber(repo, build.Number)
		// if err != nil {
		// 	c.AbortWithError(500, err)
		// 	return
		// }
		// procs, _ := store.FromContext(c).ProcList(build)
		build.Procs = model.Tree(build.Procs)
	}

	// c.JSON(202, build)
	c.JSON(202, build)

	//
	// publish topic
	//
	var public bool
	public = true
	if conf != nil {
		public = conf.AgentPublic
	}

	//project, err := store.FromContext(c).GetProject(b.User.ProjectID)
	//if err != nil {
	//	c.String(404, "Cannot find project. %s", err)
	//	return
	//}
	// logrus.Debugf("item.Proc.ID 333", items[0].Proc.ID)

	buildCopy := *build
	// buildCopy.Procs = model.Tree(buildCopy.Procs)
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"project": project.Name,
			"private": strconv.FormatBool(repo.IsPrivate),
		},
	}
	// logrus.Debugf("item.Proc.ID 444", items[0].Proc.ID)
	if proc == nil {
		message.Data, _ = json.Marshal(model.Event{
			Type:  model.Enqueued,
			Repo:  *repo,
			Build: buildCopy,
		})
	} else {
		message.Data, _ = json.Marshal(model.Event{
			// Type:  model.Retry,
			Repo:  *repo,
			Build: buildCopy,
		})
	}
	// TODO remove global reference
	Config.Services.Pubsub.Publish(c, "topic/events", message)
	// logrus.Debugf("message: %s", message)
	//
	// end publish topic
	//

	for _, item := range items {
		// logrus.Debugf("item.Proc.ID 555", item.Proc.ID)
		task := new(queue.Task)
		task.ID = fmt.Sprint(item.Proc.ID)
		task.Labels = map[string]string{}
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

		//for k, v := range item.Labels {
		//	task.Labels[k] = v
		//}
		// logrus.Debugf("task.Labels: %v", task.Labels)

		task.Data, _ = json.Marshal(rpc.Pipeline{
			ID:      fmt.Sprint(item.Proc.ID),
			Config:  item.Config,
			Timeout: b.Repo.Timeout,
		})

		Config.Services.Logs.Open(context.Background(), task.ID)
		Config.Services.Queue.Push(context.Background(), task)
	}
}

//

func DeleteBuildLogs(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	num, _ := strconv.Atoi(c.Params.ByName("number"))

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	procs, err := store.FromContext(c).ProcList(build)
	if err != nil {
		c.String(404, err.Error())
		return
	}

	switch build.Status {
	case model.StatusRunning, model.StatusPending:
		c.String(400, "Cannot delete logs for a pending or running build")
		return
	}

	for _, proc := range procs {
		t := time.Now().UTC()
		logrus.Debugf("proc Name", proc.Name)
		if proc.Name == "rebuild_cache" {
			continue
		}
		buf := bytes.NewBufferString(fmt.Sprintf(deleteStr, proc.Name, user.Login, t.Format(time.UnixDate)))
		lerr := store.FromContext(c).LogSave(proc, buf)
		if lerr != nil {
			err = lerr
		}
	}
	if err != nil {
		logrus.Errorf("There was a problem deleting your logs. %s", err)
		c.String(400, err.Error())
		return
	}

	c.String(204, "")
}

func PostStepBuild(c *gin.Context) {
	// v, ok := c.Get("headimage")
	// if !ok {
	// 	c.String(500, "Error fetching feed.")
	// 	return
	// }
	repo := session.Repo(c)
	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		logrus.Errorf("03 failure to find repo owner %s. %s", repo.FullName, err)
		c.String(500, err.Error())
		return
	}
	//remote_ := remote.FromContext(c)
	remote_, err := SetupRemote(user.Provider, "")
	if err != nil {
		logrus.Fatal("get remote error. %s", err)
		c.String(400, err.Error())
		return
	}

	buildID := c.Param("number")
	build, err := store.GetBuild(c, buildID)
	if err != nil {
		msg := "failure to get build"
		logrus.Errorf("%s, %s", msg, err.Error())
		c.String(404, fmt.Sprintf("%s, %s", msg, err.Error()))
		return
	}

	switch build.Status {
	case model.StatusDeclined,
		model.StatusBlocked,
		model.StatusPending,
		model.StatusRunning:
		c.String(500, "cannot restart a step build with it's build status %s", build.Status)
		return
	}

	pid, err := strconv.Atoi(c.Param("pid"))
	if err != nil {
		c.String(400, err.Error())
		return
	}

	proc, err := store.FromContext(c).ProcFind(build, pid)
	if err != nil {
		c.String(400, err.Error())
		return
	}

	// if the remote has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the job.
	if refresher, ok := remote_.(remote.Refresher); ok {
		ok, _ := refresher.Refresh(user)
		if ok {
			store.UpdateUser(c, user)
		}
	}

	// fetch the .drone.yml file from the database
	conf, err := Config.Storage.Config.ConfigLoad(build.ConfigID)
	if err != nil {
		logrus.Errorf("failure to get build config for %s. %s", repo.FullName, err)
		c.String(404, err.Error())
		return
	}
	if conf.File == 2 {
		confb, err := remote_.FileRef(user, repo, build.Branch, conf.FilePath)
		if err != nil {
			logrus.Errorf("error: %s: cannot find %s in %s: %s")
			c.String(404, err.Error())
			return
		}
		conf.Data = string(confb)
	}

	netrc, err := remote_.Netrc(user, repo)
	if err != nil {
		logrus.Errorf("failure to generate netrc for %s. %s", repo.FullName, err)
		c.String(500, err.Error())
		return
	}

	project, err := store.FromContext(c).GetProject(repo.ProjectID)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	r, err := Config.Services.Grpc.GetToken(context.Background(), &pb.TokenRequest{Name: project.Name})
	if err != nil {
		msg := fmt.Sprintf("get token error: %+v", err)
		logrus.Error(msg)
		c.String(500, msg)
		return
	}
	userToken := r.Message

	err = ChargingCheck(c, project)
	if err != nil {
		msg := fmt.Sprintf("Charging Check: %s", err.Error())
		logrus.Error(msg)
		c.String(404, msg)
		return
	}
	// build.ID = 0
	// build.Number = 0
	// build.Parent = num
	build.Status = model.StatusPending
	// build.Started = time.Now().Unix()
	// build.Finished = build.Started
	// build.Enqueued = time.Now().UTC().Unix()
	build.Error = ""
	build.Deploy = c.DefaultQuery("deploy_to", build.Deploy)
	// build.Avatar = v.(string)

	event := c.DefaultQuery("event", build.Event)
	if event == model.EventPush ||
		event == model.EventPull ||
		event == model.EventTag ||
		event == model.EventDeploy {
		build.Event = event
	}

	err = store.UpdateBuild(c, build)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	RunBuild(c, netrc, user, repo, build, proc, conf, project, userToken)
}

var deleteStr = `[
	{
	  "proc": %q,
	  "pos": 0,
	  "out": "logs purged by %s on %s\n"
	}
]`
