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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	// "log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	oldcontext "golang.org/x/net/context"

	"google.golang.org/grpc/metadata"

	"github.com/sirupsen/logrus"
	"github.com/coderun-top/coderun/src/core/logging"
	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/rpc"
	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/rpc/proto"
	"github.com/coderun-top/coderun/src/core/pubsub"
	"github.com/coderun-top/coderun/src/core/queue"

	"github.com/coderun-top/coderun/src/model"
	"github.com/coderun-top/coderun/src/remote"
	"github.com/coderun-top/coderun/src/store"

	// "github.com/drone/expr"
	email_util "github.com/coderun-top/coderun/src/email"
	pb "github.com/coderun-top/coderun/src/grpc/user"
)

// This file is a complete disaster because I'm trying to wedge in some
// experimental code. Please pardon our appearance during renovations.

// Config is an evil global configuration that will be used as we transition /
// refactor the codebase to move away from storing these values in the Context.
var Config = struct {
	Services struct {
		Pubsub     pubsub.Publisher
		Queue      queue.Queue
		Logs       logging.Log
		Senders    model.SenderService
		Secrets    model.SecretService
		Registries model.RegistryService
		Environ    model.EnvironService
		Limiter    model.Limiter
		Grpc       pb.GreeterClient
	}
	Storage struct {
		// Users  model.UserStore
		// Repos  model.RepoStore
		// Builds model.BuildStore
		// Logs   model.LogStore
		Config model.ConfigStore
		Files  model.FileStore
		Procs  model.ProcStore
		// Registries model.RegistryStore
		// Secrets model.SecretStore
	}
	Server struct {
		Key            string
		Cert           string
		Host           string
		Port           string
		Pass           string
		RepoConfig     string
		SessionExpires time.Duration
		// Open bool
		// Orgs map[string]struct{}
		// Admins map[string]struct{}
	}
	Prometheus struct {
		AuthToken string
	}
	Pipeline struct {
		Limits     model.ResourceLimit
		Volumes    []string
		Networks   []string
		Privileged []string
	}
}{}

type RPC struct {
	remote remote.Remote
	queue  queue.Queue
	pubsub pubsub.Publisher
	logger logging.Log
	store  store.Store
	host   string
}

type messageDetail struct {
	Repo      string
	Name      string
	User      string
	ImageUrl  string
	CommitUrl string
	Commit    string
	Branche   string
	Time      string
	Url       string
}

type Emails struct {
	Email  string `json:"email"`
	Status bool   `json:"status"`
}

// Next implements the rpc.Next function
func (s *RPC) Next(c context.Context, filter rpc.Filter) (*rpc.Pipeline, error) {
	// logrus.Debugf("rpc next")
	// var hostnames []string
	// metadata, ok := metadata.FromContext(c)
	// if ok {
	// 	hostnames, ok = metadata["hostname"]
	// 	if ok && len(hostnames) != 0 {
	// 		logrus.Debugf("agent connected: %s: polling", hostnames[0])

	//         // public agent for user
	// 		if filter.Labels["type"] = "public" {
	//             _, err := AgentVerify(filter, hostnames[0])
	//             if err != nil {
	//                 msg := "Agent Verify error"
	//             	logrus.Errorf("%s, %s", msg, err.Error())
	//             	return nil, fmt.Errorf("%s, %s", msg, err.Error())
	//             }
	// 		}
	// 	}
	// }

	if filter.Labels["name"] == "" {
		msg := "rpc agent's name invalid"
		logrus.Errorf("%s", msg)
		return nil, fmt.Errorf("%s", msg)
	}

	logrus.Debugf("rpc request next: %v", filter)

	filter.Labels["public"] = "true"
	// coderun agent for user
	if filter.Labels["type"] == "user" {
		filter.Labels["public"] = "false"
	}

	_, err := AgentVerify(s.store, filter)
	if err != nil {
		msg := "Agent Verify error"
		logrus.Errorf("%s, %s", msg, err.Error())
		return nil, fmt.Errorf("%s, %s", msg, err.Error())
	}

	fn, err := createFilterFunc(filter)
	if err != nil {
		return nil, err
	}
	task, err := s.queue.Poll(c, fn, filter.Labels["client_id"])
	if err != nil {
		return nil, err
	} else if task == nil {
		return nil, nil
	}

	pipeline := new(rpc.Pipeline)

	// check if the process was previously cancelled
	// cancelled, _ := s.checkCancelled(pipeline)
	// if cancelled {
	// 	logrus.Debugf("ignore pid %v: cancelled by user", pipeline.ID)
	// 	if derr := s.queue.Done(c, pipeline.ID); derr != nil {
	// 		logrus.Errorf("error: done: cannot ack proc_id %v: %s", pipeline.ID, err)
	// 	}
	// 	return nil, nil
	// }

	err = json.Unmarshal(task.Data, pipeline)
	return pipeline, err
}

// Wait implements the rpc.Wait function
func (s *RPC) Wait(c context.Context, id string) error {
	// logrus.Debugf("wait")
	return s.queue.Wait(c, id)
}

// Extend implements the rpc.Extend function
func (s *RPC) Extend(c context.Context, id string) error {
	// logrus.Debugf("extend")
	return s.queue.Extend(c, id)
}

// Update implements the rpc.Update function
func (s *RPC) Update(c context.Context, id string, state rpc.State) error {
	// logrus.Debugf("update")
	procID, err := strconv.ParseInt(id, 10, 64)
	logrus.Debugf("Proc ID: %d", procID)
	if err != nil {
		return err
	}

	pproc, err := s.store.ProcLoad(procID)
	if err != nil {
		logrus.Errorf("error: rpc.update: cannot find pproc with id %d: %s", procID, err)
		return err
	}

	build, err := s.store.GetBuild(pproc.BuildID)
	if err != nil {
		logrus.Errorf("error: cannot find build with id %d: %s", pproc.BuildID, err)
		return err
	}
	if state.Proc == "rebuild_cache" {
		return nil
	}

	proc, err := s.store.ProcChild(build, pproc.PID, state.Proc)
	if err != nil {
		logrus.Errorf("error: cannot find proc with name %s: %s", state.Proc, err)
		return err
	}

	metadata, ok := metadata.FromContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			proc.Machine = hostname[0]
		}
	}

	repo, err := s.store.GetRepo(build.RepoID)
	if err != nil {
		logrus.Errorf("error: cannot find repo with id %d: %s", build.RepoID, err)
		return err
	}

	if state.Exited {
		proc.Stopped = state.Finished
		proc.ExitCode = state.ExitCode
		proc.Error = state.Error
		proc.State = model.StatusSuccess
		if state.ExitCode != 0 || state.Error != "" {
			proc.State = model.StatusFailure
		}
		if state.ExitCode == 137 {
			proc.State = model.StatusKilled
		}
	} else {
		proc.Started = state.Started
		proc.State = model.StatusRunning
	}

	if proc.Started == 0 && proc.Stopped != 0 {
		proc.Started = build.Started
	}

	if err := s.store.ProcUpdate(proc); err != nil {
		logrus.Errorf("error: rpc.update: cannot update proc: %s", err)
	}

	build.Procs, _ = s.store.ProcList(build)
	build.Procs = model.Tree(build.Procs)

	user2, err := s.store.GetUser(repo.UserID)
	if err != nil {
		logrus.Errorf("get user error:%s", err)
	}

	project, err := s.store.GetProject(user2.ProjectID)
	if err != nil {
		return err
	}

	message := pubsub.Message{
		Labels: map[string]string{
			"test":    "test1",
			"repo":    repo.FullName,
			"project": project.Name,
			"private": strconv.FormatBool(repo.IsPrivate),
		},
	}
	message.Data, _ = json.Marshal(model.Event{
		Repo:  *repo,
		Build: *build,
	})
	logrus.Debugf("update build for buildID: %d", build.ID)
	s.pubsub.Publish(c, "topic/events", message)

	return nil
}

// Upload implements the rpc.Upload function
func (s *RPC) Upload(c context.Context, id string, file *rpc.File) error {
	// logrus.Debugf("upload")
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	pproc, err := s.store.ProcLoad(procID)
	if err != nil {
		logrus.Errorf("error: cannot find parent proc with id %d: %s", procID, err)
		return err
	}

	build, err := s.store.GetBuild(pproc.BuildID)
	if err != nil {
		logrus.Errorf("error: cannot find build with id %d: %s", pproc.BuildID, err)
		return err
	}
	if file.Proc == "rebuild_cache" {
		return nil
	}

	proc, err := s.store.ProcChild(build, pproc.PID, file.Proc)
	if err != nil {
		logrus.Errorf("error: cannot find child proc with name %s: %s", file.Proc, err)
		return err
	}
	logrus.Debugf("Proc Name: %s", proc.Name)

	if file.Mime == "application/json+logs" {
		if proc.Name != "rebuild_cache" {
			return s.store.LogSave(
				proc,
				bytes.NewBuffer(file.Data),
			)
		}
	}

	report := &model.File{
		BuildID: proc.BuildID,
		ProcID:  proc.ID,
		PID:     proc.PID,
		Mime:    file.Mime,
		Name:    file.Name,
		Size:    file.Size,
		Time:    file.Time,
	}
	if d, ok := file.Meta["X-Tests-Passed"]; ok {
		report.Passed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Tests-Failed"]; ok {
		report.Failed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Tests-Skipped"]; ok {
		report.Skipped, _ = strconv.Atoi(d)
	}

	if d, ok := file.Meta["X-Checks-Passed"]; ok {
		report.Passed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Checks-Failed"]; ok {
		report.Failed, _ = strconv.Atoi(d)
	}

	if d, ok := file.Meta["X-Coverage-Lines"]; ok {
		report.Passed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Coverage-Total"]; ok {
		if total, _ := strconv.Atoi(d); total != 0 {
			report.Failed = total - report.Passed
		}
	}

	return Config.Storage.Files.FileCreate(
		report,
		bytes.NewBuffer(file.Data),
	)
}

// Init implements the rpc.Init function
func (s *RPC) Init(c context.Context, id string, state rpc.State) error {
	// logrus.Debugf("init")
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	proc, err := s.store.ProcLoad(procID)
	if err != nil {
		logrus.Errorf("error: cannot find proc with id %d: %s", procID, err)
		return err
	}
	metadata, ok := metadata.FromContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			proc.Machine = hostname[0]
		}
	}

	build, err := s.store.GetBuild(proc.BuildID)
	if err != nil {
		logrus.Errorf("error: cannot find build with id %d: %s", proc.BuildID, err)
		return err
	}

	repo, err := s.store.GetRepo(build.RepoID)
	if err != nil {
		logrus.Errorf("error: cannot find repo with id %d: %s", build.RepoID, err)
		return err
	}

	if build.Status == model.StatusPending {
		build.Status = model.StatusRunning
		build.Started = state.Started
		if err := s.store.UpdateBuild(build); err != nil {
			logrus.Errorf("error: init: cannot update build_id %d state: %s", build.ID, err)
		}
	}

	defer func() {
		build.Procs, _ = s.store.ProcList(build)

		user2, err := s.store.GetUser(repo.UserID)
		if err != nil {
			logrus.Errorf("get user error:%s", err)
		}
		project, err := s.store.GetProject(user2.ProjectID)
		if err != nil {
			logrus.Errorf("get project error:%s", err)
			//return err
		}

		message := pubsub.Message{
			Labels: map[string]string{
				"test":    "test2",
				"repo":    repo.FullName,
				"project": project.Name,
				"private": strconv.FormatBool(repo.IsPrivate),
			},
		}
		message.Data, _ = json.Marshal(model.Event{
			Repo:  *repo,
			Build: *build,
		})
		s.pubsub.Publish(c, "topic/events", message)
	}()

	proc.Started = state.Started
	proc.State = model.StatusRunning
	return s.store.ProcUpdate(proc)
}

// Done implements the rpc.Done function
func (s *RPC) Done(c context.Context, id string, state rpc.State) error {
	// logrus.Debugf("done")
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	proc, err := s.store.ProcLoad(procID)
	if err != nil {
		logrus.Errorf("error: cannot find proc with id %d: %s", procID, err)
		return err
	}

	build, err := s.store.GetBuild(proc.BuildID)
	if err != nil {
		logrus.Errorf("error: cannot find build with id %d: %s", proc.BuildID, err)
		return err
	}

	repo, err := s.store.GetRepo(build.RepoID)
	if err != nil {
		logrus.Errorf("error: cannot find repo with id %d: %s", build.RepoID, err)
		return err
	}

	proc.Stopped = state.Finished
	proc.Error = state.Error
	proc.ExitCode = state.ExitCode
	proc.State = model.StatusSuccess
	if proc.ExitCode != 0 || proc.Error != "" {
		proc.State = model.StatusFailure
	}
	if err := s.store.ProcUpdate(proc); err != nil {
		logrus.Errorf("error: done: cannot update proc_id %d state: %s", procID, err)
	}

	if err := s.queue.Done(c, id); err != nil {
		logrus.Errorf("error: done: cannot ack proc_id %d: %s", procID, err)
	}

	// TODO handle this error
	procs, _ := s.store.ProcList(build)
	for _, p := range procs {
		if p.Running() && p.PPID == proc.PID {
			p.State = model.StatusSkipped
			// if p.Started != 0 {
			if p.Started != 0 && p.ExitCode == 0 {
				p.State = model.StatusSuccess // for deamons that are killed
				// p.Stopped = proc.Stopped
			}
			if err := s.store.ProcUpdate(p); err != nil {
				logrus.Errorf("error: done: cannot update proc_id %d child state: %s", p.ID, err)
			}
		}
	}

	running := false
	status := model.StatusSuccess
	for _, p := range procs {
		if p.PPID == 0 {
			if p.Running() {
				running = true
			}
			if p.Failing() {
				status = p.State
			}
		}
	}

	user, err := s.store.GetUser(repo.UserID)
	if err != nil {
		logrus.Errorf("get user error:%s", err)
	}

	if !running {
		build.Status = status
		build.Finished = proc.Stopped
		if err := s.store.UpdateBuild(build); err != nil {
			logrus.Errorf("error: done: cannot update build_id %d final state: %s", build.ID, err)
		}

		// update the status
		// user, err := s.store.GetUser(repo.UserID)
		// if err == nil {
		// 	if refresher, ok := s.remote.(remote.Refresher); ok {
		// 		ok, _ := refresher.Refresh(user)
		// 		if ok {
		// 			s.store.UpdateUser(user)
		// 		}
		// 	}
		// 	uri := fmt.Sprintf("%s/%s/%d", s.host, repo.FullName, build.Number)
		// 	err = s.remote.Status(user, repo, build, uri)
		// 	if err != nil {
		// 		logrus.Errorf("error setting commit status for %s/%d: %v", repo.FullName, build.Number, err)
		// 	}
		// }

		webhook, err := s.store.GetWebHook(repo)
		if err == nil {
			if webhook.State == true {
				go httpPost(webhook.Url, build, repo)
			}
		}

		webhook2, err := s.store.GetWebHookProject(repo.ProjectID)
		if err == nil {
			if webhook2.State == true {
				go httpPost(webhook2.Url, build, repo)
			}
		}

		email, err := s.store.GetEmail(repo)
		if err == nil {
			if email.State == true {
				go Send(email.Email, build, repo, user)
			}
		}

		email2, err := s.store.GetEmailProject(repo.ProjectID)
		if err == nil {
			if email2.State == true {
				go Send(email2.Email, build, repo, user)
			}
		}

	}

	if err := s.logger.Close(c, id); err != nil {
		logrus.Errorf("error: done: cannot close build_id %d logger: %s", proc.ID, err)
	}

	project, err := s.store.GetProject(user.ProjectID)
	if err != nil {
		return err
	}

	build.Procs = model.Tree(procs)
	message := pubsub.Message{
		Labels: map[string]string{
			//"test": "test3",
			"repo":    repo.FullName,
			"project": project.Name,
			"private": strconv.FormatBool(repo.IsPrivate),
		},
	}
	message.Data, _ = json.Marshal(model.Event{
		Repo:  *repo,
		Build: *build,
	})
	s.pubsub.Publish(c, "topic/events", message)

	return nil
}

// Log implements the rpc.Log function
func (s *RPC) Log(c context.Context, id string, line *rpc.Line) error {
	// logrus.Debugf("log")
	entry := new(logging.Entry)
	entry.Data, _ = json.Marshal(line)
	s.logger.Write(c, id, entry)
	return nil
}

func (s *RPC) checkCancelled(pipeline *rpc.Pipeline) (bool, error) {
	// logrus.Debugf("checkCancelled")
	pid, err := strconv.ParseInt(pipeline.ID, 10, 64)
	if err != nil {
		return false, err
	}
	proc, err := s.store.ProcLoad(pid)
	if err != nil {
		return false, err
	}
	if proc.State == model.StatusKilled {
		return true, nil
	}
	return false, err
}

func AgentVerify(s store.Store, filter rpc.Filter) (*model.Agent, error) {
	projectID := "dougo"
	projectName := ""

	// if filter.Labels["type"] == "dougo" {
	//     return nil, nil
	// }

	maxProcs, err := strconv.Atoi(filter.Labels["max_procs"])
	if err != nil {
		return nil, fmt.Errorf("convert maxProcs error, %s", err.Error())
	}

	if filter.Labels["type"] == "user" {
		agentKey, err := s.GetAgentKeyByKey(filter.Labels["key"])
		if err != nil {
			return nil, fmt.Errorf("Agent key not find, %s", err)
		}

		project, err := s.GetProject(agentKey.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("Project not find, %s", err)
		}

		projectID = project.ID
		projectName = project.Name
	}

	agentStatus := 1
	// check agent count
	if filter.Labels["type"] == "user" {
		// agent count check
		agentStatus = 0
		agentCount, err := s.GetAgentCount(projectID)
		if err != nil {
			return nil, fmt.Errorf("Get agent count error, %s", err.Error())
		}

		projectCharging, err := GetProjectCharging(projectName)
		if err != nil {
			return nil, fmt.Errorf("Get project charging error, %s", err.Error())
		}

		if agentCount < projectCharging.AgentCount {
			agentStatus = 1
		}
	}

	agent, err := s.GetAgentClient(projectID, filter.Labels["client_id"])
	if err != nil {
		agent := &model.Agent{
			ProjectID: projectID,
			Type:      filter.Labels["type"],
			ClientID:  filter.Labels["client_id"],
			Platform:  filter.Labels["platform"],
			Name:      filter.Labels["name"],
			Addr:      filter.Labels["addr"],
			Tags:      filter.Labels["tags"],
			MaxProcs:  maxProcs,
			Status:    agentStatus,
		}
		err = s.CreateAgent(agent)
		if err != nil {
			return nil, fmt.Errorf("Create Agent error, %s", err.Error())
		}

		return agent, nil
	}

	agent.Type = filter.Labels["type"]
	agent.Platform = filter.Labels["platform"]
	agent.Name = filter.Labels["name"]
	agent.Addr = filter.Labels["addr"]
	agent.Tags = filter.Labels["tags"]
	agent.MaxProcs = maxProcs
	agent.Status = agentStatus

	err = s.UpdateAgent(agent)
	if err != nil {
		return nil, fmt.Errorf("Update Agent error, %s", err.Error())
	}

	return agent, nil
}

func createFilterFunc(filter rpc.Filter) (queue.Filter, error) {
	return func(task *queue.Task) bool {
		// var st *expr.Selector
		// var err error
		// logrus.Debugf("task.Labels: %v", task.Labels)
		// logrus.Debugf("filter.Labels: %v", filter.Labels)
		// vv := map[string]string{}
		if filter.Labels["public"] == task.Labels["public"] {
			if filter.Labels["public"] == "true" {
				return true
			} else {
				// if task.Labels["filter"] != "" {
				// 	if filter.Labels["tags_str"] != "" {
				// 		task_tags := strings.Split(task.Labels["filter"], ",")
				// 		for i := 0; i <= len(task_tags)-1; i++ {
				// 			expr_str := "'" + task_tags[i] + "' IN " + filter.Labels["tags_str"]
				// 			// logrus.Debugf("expr_str: %s", expr_str)
				// 			st, err = expr.ParseString(expr_str)
				// 			if err != nil {
				// 				logrus.Errorf("expr parse error: %v", err)
				// 				return false
				// 			}
				// 			if st != nil {
				// 				//if task.Labels["tags"] != ""{
				// 				//	 err = json.Unmarshal(task.Labels["tags"], &zzz)
				// 				//	    if err != nil {
				// 				//		logrus.Debugf("err2: %v", err)
				// 				//		    false
				// 				//	    }
				// 				//}
				// 				match, _ := st.Eval(expr.NewRow(vv))
				// 				if match == true {
				// 					return true
				// 				} else if i == len(task_tags)-1 {
				// 					return false
				// 				}
				// 			}
				// 		}
				// 	} else {
				// 		return false
				// 	}
				// } else {
				if task.Labels["filter"] != "" && filter.Labels["tags"] != "" {
					task_tags := strings.Split(task.Labels["filter"], ",")
					agent_tags := strings.Split(filter.Labels["tags"], ",")
					for _, t := range task_tags {
						matched := false
						for _, a := range agent_tags {
							if t == a {
								matched = true
								break
							}
						}
						if !matched {
							return false
						}
					}
					return true
					// 如果task的filter没有内容则匹配失败
				} else {
					return false
				}
			}
		} else {
			return false
		}
		return true
	}, nil
}

//
//
//

// DroneServer is a grpc server implementation.
type DroneServer struct {
	Remote remote.Remote
	Queue  queue.Queue
	Pubsub pubsub.Publisher
	Logger logging.Log
	Store  store.Store
	Host   string
}

func (s *DroneServer) Next(c oldcontext.Context, req *proto.NextRequest) (*proto.NextReply, error) {
	// logrus.Debugf("next drone")
	peer := RPC{
		remote: s.Remote,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	filter := rpc.Filter{
		Labels: req.GetFilter().GetLabels(),
		Expr:   req.GetFilter().GetExpr(),
	}

	res := new(proto.NextReply)
	pipeline, err := peer.Next(c, filter)
	if err != nil {
		return res, err
	}
	if pipeline == nil {
		return res, err
	}

	res.Pipeline = new(proto.Pipeline)
	res.Pipeline.Id = pipeline.ID
	res.Pipeline.Timeout = pipeline.Timeout
	res.Pipeline.Payload, _ = json.Marshal(pipeline.Config)

	return res, err

	// fn := func(task *queue.Task) bool {
	// 	for k, v := range req.GetFilter().Labels {
	// 		if task.Labels[k] != v {
	// 			return false
	// 		}
	// 	}
	// 	return true
	// }
	// task, err := s.Queue.Poll(c, fn)
	// if err != nil {
	// 	return nil, err
	// } else if task == nil {
	// 	return nil, nil
	// }
	//
	// pipeline := new(rpc.Pipeline)
	// json.Unmarshal(task.Data, pipeline)
	//
	// res := new(proto.NextReply)
	// res.Pipeline = new(proto.Pipeline)
	// res.Pipeline.Id = pipeline.ID
	// res.Pipeline.Timeout = pipeline.Timeout
	// res.Pipeline.Payload, _ = json.Marshal(pipeline.Config)
	//
	// // check if the process was previously cancelled
	// // cancelled, _ := s.checkCancelled(pipeline)
	// // if cancelled {
	// // 	logrus.Debugf("ignore pid %v: cancelled by user", pipeline.ID)
	// // 	if derr := s.queue.Done(c, pipeline.ID); derr != nil {
	// // 		logrus.Errorf("error: done: cannot ack proc_id %v: %s", pipeline.ID, err)
	// // 	}
	// // 	return nil, nil
	// // }
	//
	// return res, nil
}

func (s *DroneServer) Init(c oldcontext.Context, req *proto.InitRequest) (*proto.Empty, error) {
	// logrus.Debugf("init drone")
	peer := RPC{
		remote: s.Remote,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	state := rpc.State{
		Error:    req.GetState().GetError(),
		ExitCode: int(req.GetState().GetExitCode()),
		Finished: req.GetState().GetFinished(),
		Started:  req.GetState().GetStarted(),
		Proc:     req.GetState().GetName(),
		Exited:   req.GetState().GetExited(),
	}
	res := new(proto.Empty)
	err := peer.Init(c, req.GetId(), state)
	return res, err
}

func (s *DroneServer) Update(c oldcontext.Context, req *proto.UpdateRequest) (*proto.Empty, error) {
	// logrus.Debugf("update drone")
	peer := RPC{
		remote: s.Remote,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	state := rpc.State{
		Error:    req.GetState().GetError(),
		ExitCode: int(req.GetState().GetExitCode()),
		Finished: req.GetState().GetFinished(),
		Started:  req.GetState().GetStarted(),
		Proc:     req.GetState().GetName(),
		Exited:   req.GetState().GetExited(),
	}
	res := new(proto.Empty)
	err := peer.Update(c, req.GetId(), state)
	return res, err
}

func (s *DroneServer) Upload(c oldcontext.Context, req *proto.UploadRequest) (*proto.Empty, error) {
	// logrus.Debugf("upload drone")
	peer := RPC{
		remote: s.Remote,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	file := &rpc.File{
		Data: req.GetFile().GetData(),
		Mime: req.GetFile().GetMime(),
		Name: req.GetFile().GetName(),
		Proc: req.GetFile().GetProc(),
		Size: int(req.GetFile().GetSize()),
		Time: req.GetFile().GetTime(),
		Meta: req.GetFile().GetMeta(),
	}

	res := new(proto.Empty)
	err := peer.Upload(c, req.GetId(), file)
	return res, err
}

func (s *DroneServer) Done(c oldcontext.Context, req *proto.DoneRequest) (*proto.Empty, error) {
	// logrus.Debugf("done drone")
	peer := RPC{
		remote: s.Remote,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	state := rpc.State{
		Error:    req.GetState().GetError(),
		ExitCode: int(req.GetState().GetExitCode()),
		Finished: req.GetState().GetFinished(),
		Started:  req.GetState().GetStarted(),
		Proc:     req.GetState().GetName(),
		Exited:   req.GetState().GetExited(),
	}
	res := new(proto.Empty)
	err := peer.Done(c, req.GetId(), state)
	return res, err
}

func (s *DroneServer) Wait(c oldcontext.Context, req *proto.WaitRequest) (*proto.Empty, error) {
	// logrus.Debugf("wait drone")
	peer := RPC{
		remote: s.Remote,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	res := new(proto.Empty)
	err := peer.Wait(c, req.GetId())
	return res, err
}

func (s *DroneServer) Extend(c oldcontext.Context, req *proto.ExtendRequest) (*proto.Empty, error) {
	// logrus.Debugf("extend drone")
	peer := RPC{
		remote: s.Remote,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	res := new(proto.Empty)
	err := peer.Extend(c, req.GetId())
	return res, err
}

func (s *DroneServer) Log(c oldcontext.Context, req *proto.LogRequest) (*proto.Empty, error) {
	// logrus.Debugf("log drone")
	peer := RPC{
		remote: s.Remote,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	line := &rpc.Line{
		Out:  req.GetLine().GetOut(),
		Pos:  int(req.GetLine().GetPos()),
		Time: req.GetLine().GetTime(),
		Proc: req.GetLine().GetProc(),
	}
	res := new(proto.Empty)
	err := peer.Log(c, req.GetId(), line)
	return res, err
}

func Send(email string, build *model.Build, repo *model.Repo, user *model.User) {
	valid, err := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, email)
	if err != nil {
		fmt.Printf("failed to match regexp: %v", err)
	}

	if !valid {
		fmt.Printf("error")
		return
	}

	messageTemplate, err := template.ParseFiles("/views/error.tpl")
	if err != nil {
		fmt.Printf("error")
		return
	}
	if build.Status == "success" {
		messageTemplate, err = template.ParseFiles("/views/success.tpl")
		if err != nil {
			fmt.Printf("error")
			return
		}
	}

	message := new(bytes.Buffer)

	commit_url := ""
	//"https://gitlab.com/douwa/registry/dougo/commit/785007b1936686df75fddb9c357d97d584295c93"
	if user.Provider == 1 {
		commit_url = "https://github.com/" + repo.FullName + "/commit/" + build.Commit
	} else if user.Provider == 2 {
		commit_url = "https://gitlab.com/" + repo.FullName + "/commit/" + build.Commit
	}
	build_time := strconv.FormatInt(build.Finished-build.Started, 10)

	shortCommit := build.Commit
	if len(build.Commit) >= 7 {
		shortCommit = build.Commit[0:7]
	}

	err = messageTemplate.Execute(message, messageDetail{
		Repo:      repo.FullName,
		Name:      build.Message,
		User:      user.TokenName,
		ImageUrl:  user.Avatar,
		CommitUrl: commit_url,
		Commit:    shortCommit,
		Branche:   build.Branch,
		Time:      build_time,
		Url:       "https://dev.crun.top/repos",
	})

	if err != nil {
		fmt.Printf("error")
		return
		//log.Errorf("Message template error: %v", err)
		//cc.CustomAbort(http.StatusInternalServerError, "internal_error")
	}

	addr := net.JoinHostPort(os.Getenv("EMAIL_HOST"), os.Getenv("EMAIL_PORT"))
	_ssl, _ := strconv.ParseBool(os.Getenv("EMAIL_SSL"))

	err = email_util.Send(addr,
		"",
		os.Getenv("EMAIL_USR"),
		os.Getenv("EMAIL_PWD"),
		60, _ssl,
		true,
		os.Getenv("EMAIL_FROM"),
		[]string{email},
		"repo build",
		message.String())
	if err != nil {
		fmt.Printf("error")
		return
		//log.Errorf("Send email failed: %v", err)
		//cc.CustomAbort(http.StatusInternalServerError, "send_email_failed")
	}
}

//func (e *EmailAPI) Ping() {
//	var host, username, password, identity string
//	var port int
//	var ssl, insecure bool
//	body := e.Ctx.Input.CopyBody(1 << 32)
//	if body == nil || len(body) == 0 {
//		cfg, err := config.Email()
//		if err != nil {
//			log.Errorf("failed to get email configurations: %v", err)
//			e.CustomAbort(http.StatusInternalServerError,
//				http.StatusText(http.StatusInternalServerError))
//		}
//		host = cfg.Host
//		port = cfg.Port
//		username = cfg.Username
//		password = cfg.Password
//		identity = cfg.Identity
//		ssl = cfg.SSL
//		insecure = cfg.Insecure
//	} else {
//		settings := &struct {
//			Host     string  `json:"email_host"`
//			Port     *int    `json:"email_port"`
//			Username string  `json:"email_username"`
//			Password *string `json:"email_password"`
//			SSL      bool    `json:"email_ssl"`
//			Identity string  `json:"email_identity"`
//			Insecure bool    `json:"email_insecure"`
//		}{}
//		e.DecodeJSONReq(&settings)
//
//		if len(settings.Host) == 0 || settings.Port == nil {
//			e.CustomAbort(http.StatusBadRequest, "empty host or port")
//		}
//
//		if settings.Password == nil {
//			cfg, err := config.Email()
//			if err != nil {
//				log.Errorf("failed to get email configurations: %v", err)
//				e.CustomAbort(http.StatusInternalServerError,
//					http.StatusText(http.StatusInternalServerError))
//			}
//
//			settings.Password = &cfg.Password
//		}
//
//		host = settings.Host
//		port = *settings.Port
//		username = settings.Username
//		password = *settings.Password
//		identity = settings.Identity
//		ssl = settings.SSL
//		insecure = settings.Insecure
//	}
//
//	addr := net.JoinHostPort(host, strconv.Itoa(port))
//	if err := email.Ping(addr, identity, username,
//		password, pingEmailTimeout, ssl, insecure); err != nil {
//		log.Errorf("failed to ping email server: %v", err)
//		// do not return any detail information of the error, or may cause SSRF security issue #3755
//		e.RenderError(http.StatusBadRequest, "failed to ping email server")
//		return
//	}
//}

func httpPost(url string, build *model.Build, repo *model.Repo) {
	data := &model.WebHookData{
		Build: build,
		Repo:  repo,
	}

	bytesData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST",
		url,
		reader)
	if err != nil {
		fmt.Println(err)
		return
	}

	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func GetEmail2(project_name string) ([]*Emails, error) {
	r, err := Config.Services.Grpc.GetProjectEmail(context.Background(), &pb.ProjectRequest{Token: "", Projectname: project_name})
	if err != nil {
		logrus.Errorf("GetProjectEmail error: %v", err)
		return nil, err
	}

	data := []*Emails{}
	err = json.Unmarshal([]byte(r.Message), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
