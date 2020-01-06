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
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/sirupsen/logrus"
	"github.com/drone/envsubst"
	"github.com/coderun-top/coderun/src/model"
	"github.com/coderun-top/coderun/src/store"

	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/backend"
	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/frontend"
	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/frontend/yaml"
	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/frontend/yaml/compiler"
	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/frontend/yaml/linter"
	"github.com/coderun-top/coderun/src/core/pipeline/pipeline/frontend/yaml/matrix"

	"github.com/drone/expr"
)

type Builder struct {
	Repo  *model.Repo
	User  *model.User
	Curr  *model.Build
	Last  *model.Build
	Proc  *model.Proc
	Netrc *model.Netrc
	Secs  []*model.Secret
	Regs  []*model.Registry
	Helms []*model.Helm
	//Reg  *model.Registry
	K8s   []*model.K8sCluster
	Link  string
	Yaml  string
	Token string
	Envs  map[string]string
	// PipeLineEnvs []*model.PipelineEnvSchema
}

type buildItem struct {
	Proc     *model.Proc
	Platform string
	Labels   map[string]string
	Config   *backend.Config
	Public   bool
	Filter   string
}

func (b *Builder) Build(c *gin.Context) ([]*buildItem, error) {
	if b.Proc == nil {
		logrus.Debugf("start build for buildId: %d", b.Curr.ID)
	} else {
		logrus.Debugf("start build for buildId: %d, procName: %s", b.Curr.ID, b.Proc.Name)
	}

	project, err := store.FromContext(c).GetProject(b.User.ProjectID)
	if err != nil {
		//c.String(404, "Cannot find project. %s", err)
		return nil, err
	}

	// axes(matrix)目前并没有用，它是用于支持多个yaml配置在一个文件中，参考：[Multi-Machine](https://docs.drone.io/user-guide/pipeline/multi-machine/)和[Matrix](https://docs.drone.io/user-guide/pipeline/migrating/#matrix)
	axes, err := matrix.ParseString(b.Yaml)
	if err != nil {
		return nil, err
	}
	if len(axes) == 0 {
		axes = append(axes, matrix.Axis{})
	}

	var items []*buildItem
	// axes(matrix)目前并没有用，它是用于支持多个yaml配置在一个文件中，参考：[Multi-Machine](https://docs.drone.io/user-guide/pipeline/multi-machine/)和[Matrix](https://docs.drone.io/user-guide/pipeline/migrating/#matrix)
	for i, axis := range axes {
		pproc := &model.Proc{
			Name:    "default",
			BuildID: b.Curr.ID,
			PID:     i + 1,
			PGID:    i + 1,
			State:   model.StatusPending,
			// 类型强转
			Environ: model.MapType(axis),
		}

		// 如果指定proc，那么需要更新pproc状态
		if b.Proc != nil {
			pproc, err = store.FromContext(c).ProcFind(b.Curr, b.Proc.PPID)
			pproc.State = model.StatusPending
			// pproc.State = model.StatusRunning
			if err != nil {
				logrus.Errorf("current proc can't find!", err)
				return nil, err
			}
			err = store.FromContext(c).ProcUpdate(pproc)
			if err != nil {
				logrus.Errorf("update proc error!", err)
				return nil, err
			}
		}

		metadata := metadataFromStruct(b.Repo, b.Curr, b.Last, pproc, b.Link)
		environ := metadata.Environ()
		for k, v := range metadata.EnvironDrone() {
			environ[k] = v
		}
		for k, v := range axis {
			environ[k] = v
		}

		// 获取所有pipeline环境变量并添加到env中
		//pipeline_envs,err := store.GetAllPilelineEnv(b,b.Repo)
		//   if err != nil {
		//       return nil, err
		//}

		//for _, v := range b.PipeLineEnvs {
		//	environ[v.EnvKey] = v.EnvValue
		//}

		var secrets []compiler.Secret
		for _, sec := range b.Secs {
			if !sec.Match(b.Curr.Event) {
				continue
			}
			secrets = append(secrets, compiler.Secret{
				Name:  sec.Name,
				Value: sec.Value,
				Match: sec.Images,
			})
		}

		y := b.Yaml
		s, err := envsubst.Eval(y, func(name string) string {
			env := environ[name]
			// 如果不是系统环境变量，则转换用户环境变量
			if env == "" {
				env = b.Envs[name]
			}

			if strings.Contains(env, "\n") {
				env = fmt.Sprintf("%q", env)
			}
			return env
		})
		if err != nil {
			return nil, err
		}
		y = s

		// logrus.Debugf("Build yaml: %s", y)

		parsed, err := yaml.ParseString(y)
		if err != nil {
			logrus.Errorf("yaml parse error: %s", err.Error())
			return nil, err
		}

		if len(parsed.Trust.Include) == 0 {
			parsed.Trust.Include = append(parsed.Trust.Include, "crun/*")
		}

		// 设置解析后的容器是否有效，如果是指定Proc的方式，那么从Proc.Name相同的容器开始截断
		// pproc目前不做控制，如果后续增加多个yaml或者materic那么需要修改这里
		if b.Proc != nil {
			startIndex := -1
			// containers []*yaml.Container
			for i, _ := range parsed.Pipeline.Containers {
				if parsed.Pipeline.Containers[i].Name == b.Proc.Name {
					startIndex = i
				}
			}
			parsed.Pipeline.Containers = parsed.Pipeline.Containers[startIndex:]
		}

		// setting trusted container
		var st *expr.Selector
		for i, _ := range parsed.Pipeline.Containers {
			logrus.Debugf("build container: %s(%s)", parsed.Pipeline.Containers[i].Name, parsed.Pipeline.Containers[i].Image)

			for _, trust := range parsed.Trust.Include {
				st, err = expr.ParseString("image GLOB '" + trust + "'")
				if err != nil {
					logrus.Errorf("expr error: %s", err)
				}
				image := map[string]string{
					"image": parsed.Pipeline.Containers[i].Image,
				}

				if st != nil {
					match, _ := st.Eval(expr.NewRow(image))
					if match == true {
						if parsed.Pipeline.Containers[i].Environment == nil {
							parsed.Pipeline.Containers[i].Environment = map[string]string{}
						}

						parsed.Pipeline.Containers[i].Environment["DOUGO_HOST"] = os.Getenv("DRONE_HOST")
						parsed.Pipeline.Containers[i].Environment["DOUGO_TOKEN"] = b.Token
						parsed.Pipeline.Containers[i].Environment["DOUGO_USERNAME"] = project.Name
						break
					}
				}
			}
		}

		// tmpContainers, _ := json.Marshal(parsed.Pipeline.Containers)
		// logrus.Debugf("containers: %s", tmpContainers)

		metadata.Sys.Arch = parsed.Platform
		if metadata.Sys.Arch == "" {
			metadata.Sys.Arch = "linux/amd64"
		}

		// 如果repo是Trust才运行执行容器的Privileged等权限
		// lerr := linter.New(
		// 	linter.WithTrusted(b.Repo.IsTrusted),
		// ).Lint(parsed)
		// if lerr != nil {
		// 	return nil, lerr
		// }

		var registries []compiler.Registry

		for _, reg := range b.Regs {
			// logrus.Debug("reg address is ", reg.Address)
			// logrus.Debug("reg username is ", reg.Username)
			// logrus.Debug("reg Password is ", reg.Password)
			// logrus.Debug("reg is ", reg)
			registries = append(registries, compiler.Registry{
				Hostname: reg.Address,
				Username: reg.Username,
				Password: reg.Password,
				Email:    reg.Email,
			})
		}

		// logrus.Debugf("Environ1: %s", environ)
		// logrus.Debugf("Environ2: %s", b.Envs)
		// logrus.Debugf("Environ3: %s", proc.Environ)

		ir := compiler.New(
			// 系统默认变量，如：CI_*
			compiler.WithEnviron(environ),
			// 用户自定义变量，任意定义一般不以CI开头
			compiler.WithUserEnviron(b.Envs),
			compiler.WithEscalated(Config.Pipeline.Privileged...),
			compiler.WithResourceLimit(Config.Pipeline.Limits.MemSwapLimit, Config.Pipeline.Limits.MemLimit, Config.Pipeline.Limits.ShmSize, Config.Pipeline.Limits.CPUQuota, Config.Pipeline.Limits.CPUShares, Config.Pipeline.Limits.CPUSet),
			compiler.WithVolumes(Config.Pipeline.Volumes...),
			compiler.WithNetworks(Config.Pipeline.Networks...),
			compiler.WithLocal(false),
			compiler.WithVolumeCacher("/tmp/cache"),
			compiler.WithOption(
				compiler.WithNetrc(
					b.Netrc.Login,
					b.Netrc.Password,
					b.Netrc.Machine,
				),
				b.Repo.IsPrivate,
			),
			compiler.WithRegistry(registries...),
			compiler.WithSecret(secrets...),
			compiler.WithPrefix(
				fmt.Sprintf(
					"%d_%d",
					pproc.ID,
					rand.Int(),
				),
			),
			compiler.WithEnviron(pproc.Environ),
			compiler.WithProxy(),
			// 设置默认git路径：/coderun
			compiler.WithWorkspaceFromURL("/coderun", b.Repo.Link),
			compiler.WithMetadata(metadata),
		).Compile(parsed)

		// tmpIr, _ := json.Marshal(ir)
		// logrus.Debugf("build config: %s", tmpIr)

		// for _, sec := range b.Secs {
		// 	if !sec.MatchEvent(b.Curr.Event) {
		// 		continue
		// 	}
		// 	if b.Curr.Verified || sec.SkipVerify {
		// 		ir.Secrets = append(ir.Secrets, &backend.Secret{
		// 			Mask:  sec.Conceal,
		// 			Name:  sec.Name,
		// 			Value: sec.Value,
		// 		})
		// 	}
		// }
		agent := parsed.Agent
		public := true
		if agent != "" {
			public = false
		}

		item := &buildItem{
			Proc:     pproc,
			Config:   ir,
			Labels:   parsed.Labels,
			Platform: metadata.Sys.Arch,
			Public:   public,
			Filter:   agent,
		}
		if item.Labels == nil {
			item.Labels = map[string]string{}
		}
		items = append(items, item)
	}

	return items, nil
}

// return the metadata from the cli context.
func metadataFromStruct(repo *model.Repo, build, last *model.Build, proc *model.Proc, link string) frontend.Metadata {
	host := link
	uri, err := url.Parse(link)
	if err == nil {
		host = uri.Host
	}
	return frontend.Metadata{
		Repo: frontend.Repo{
			Name:    repo.FullName,
			Link:    repo.Link,
			Remote:  repo.Clone,
			Private: repo.IsPrivate,
			Branch:  repo.Branch,
		},
		Curr: frontend.Build{
			Number:   build.Number,
			Parent:   build.Parent,
			Created:  build.Created,
			Started:  build.Started,
			Finished: build.Finished,
			Status:   build.Status,
			Event:    build.Event,
			Link:     build.Link,
			Target:   build.Deploy,
			Commit: frontend.Commit{
				Sha:     build.Commit,
				Ref:     build.Ref,
				Refspec: build.Refspec,
				Branch:  build.Branch,
				Message: build.Message,
				Author: frontend.Author{
					Name:   build.Author,
					Email:  build.Email,
					Avatar: build.Avatar,
				},
			},
		},
		Prev: frontend.Build{
			Number:   last.Number,
			Created:  last.Created,
			Started:  last.Started,
			Finished: last.Finished,
			Status:   last.Status,
			Event:    last.Event,
			Link:     last.Link,
			Target:   last.Deploy,
			Commit: frontend.Commit{
				Sha:     last.Commit,
				Ref:     last.Ref,
				Refspec: last.Refspec,
				Branch:  last.Branch,
				Message: last.Message,
				Author: frontend.Author{
					Name:   last.Author,
					Email:  last.Email,
					Avatar: last.Avatar,
				},
			},
		},
		Job: frontend.Job{
			Number: proc.PID,
			Matrix: proc.Environ,
		},
		Sys: frontend.System{
			Name: "drone",
			Link: link,
			Host: host,
			Arch: "linux/amd64",
		},
	}
}

func yamlParseLinter(s string, trusted bool) (*yaml.Config, error) {
	parsed, err := yaml.ParseBytes([]byte(s))
	if err != nil {
		return nil, err
	}

	// if len(parsed.Vargs) != 0 {
	// 	return nil, errors.New("yaml exit some keys is invalid.")
	// }

	lerr := linter.New(
		linter.WithTrusted(trusted),
	).Lint(parsed)
	if lerr != nil {
		return nil, lerr
	}

	return parsed, nil
}

func yamlTriggerMatch(y *yaml.Config, repo *model.Repo, build *model.Build) bool {
	logrus.Debugf("Yaml Trigger Match: %v vs %s, %s, %s, %s", y.Trigger, repo.Name, build.Event, build.Branch, build.Ref)

	if build.Event != model.EventTag &&
		build.Event != model.EventDeploy {
		if !y.Branches.Match(build.Branch) {
			return false
		}
	}

	if !(y.Trigger.Repo.Match(repo.Name) &&
		y.Trigger.Event.Match(build.Event) &&
		y.Trigger.Branch.Match(build.Branch) &&
		y.Trigger.Ref.Match(build.Ref)) {
		return false
	}

	return true
}
