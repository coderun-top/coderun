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

package main

import (
	"fmt"
	// "time"

	"github.com/urfave/cli"
	"github.com/dimfeld/httptreemux"

	"github.com/coderun-top/coderun/server/queue"
	"github.com/coderun-top/coderun/server/model"
	"github.com/coderun-top/coderun/server/plugins/registry"
	"github.com/coderun-top/coderun/server/plugins/secrets"
	"github.com/coderun-top/coderun/server/remote"
	"github.com/coderun-top/coderun/server/remote/coding"
	"github.com/coderun-top/coderun/server/remote/gitea"
	"github.com/coderun-top/coderun/server/remote/github"
	"github.com/coderun-top/coderun/server/remote/gitlab"
	"github.com/coderun-top/coderun/server/remote/gitlab3"
	"github.com/coderun-top/coderun/server/remote/gogs"
	"github.com/coderun-top/coderun/server/store"
	"github.com/coderun-top/coderun/server/store/datastore"
)

func setupStore(c *cli.Context) store.Store {
	return datastore.New(
		c.String("driver"),
		c.String("datasource"),
	)
}

func setupQueue(c *cli.Context, s store.Store) queue.Queue {
	return model.WithTaskStore(queue.New(), s)
}

func setupSecretService(c *cli.Context, s store.Store) model.SecretService {
	return secrets.New(s)
}

func setupRegistryService(c *cli.Context, s store.Store) model.RegistryService {
	return registry.New(s)
}

func setupEnvironService(c *cli.Context, s store.Store) model.EnvironService {
	return nil
}

func setupLimiter(c *cli.Context, s store.Store) model.Limiter {
	return new(model.NoLimit)
}

func setupPubsub(c *cli.Context)        {}
func setupStream(c *cli.Context)        {}
func setupGatingService(c *cli.Context) {}

// helper function to setup the remote from the CLI arguments.
func SetupRemote(c *cli.Context) (remote.Remote, error) {
	switch {
	case c.Bool("github"):
		return setupGithub(c)
	case c.Bool("gitlab"):
		return setupGitlab(c)
	case c.Bool("gitea"):
		return setupGitea(c)
	case c.Bool("coding"):
		return setupCoding(c)
	default:
		return nil, fmt.Errorf("version control system not configured")
	}
}

// helper function to setup the Gitea remote from the CLI arguments.
func setupGitea(c *cli.Context) (remote.Remote, error) {
	return gitea.New(gitea.Opts{
		URL:         c.String("gitea-server"),
		Context:     c.String("gitea-context"),
		Username:    c.String("gitea-git-username"),
		Password:    c.String("gitea-git-password"),
		PrivateMode: c.Bool("gitea-private-mode"),
		SkipVerify:  c.Bool("gitea-skip-verify"),
	})
}

// helper function to setup the Gitlab remote from the CLI arguments.
func setupGitlab(c *cli.Context) (remote.Remote, error) {
	if c.Bool("gitlab-v3-api") {
		return gitlab3.New(gitlab3.Opts{
			URL:         c.String("gitlab-server"),
			Client:      c.String("gitlab-client"),
			Secret:      c.String("gitlab-secret"),
			Username:    c.String("gitlab-git-username"),
			Password:    c.String("gitlab-git-password"),
			PrivateMode: c.Bool("gitlab-private-mode"),
			SkipVerify:  c.Bool("gitlab-skip-verify"),
		})
	}
	return gitlab.New(gitlab.Opts{
		URL:         c.String("gitlab-server"),
		Client:      c.String("gitlab-client"),
		Secret:      c.String("gitlab-secret"),
		Username:    c.String("gitlab-git-username"),
		Password:    c.String("gitlab-git-password"),
		PrivateMode: c.Bool("gitlab-private-mode"),
		SkipVerify:  c.Bool("gitlab-skip-verify"),
	})
}

// helper function to setup the GitHub remote from the CLI arguments.
func setupGithub(c *cli.Context) (remote.Remote, error) {
	return github.New(github.Opts{
		URL:         c.String("github-server"),
		Context:     c.String("github-context"),
		Client:      c.String("github-client"),
		Secret:      c.String("github-secret"),
		// Scopes:      c.StringSlice("github-scope"),
		Username:    c.String("github-git-username"),
		Password:    c.String("github-git-password"),
		PrivateMode: c.Bool("github-private-mode"),
		SkipVerify:  c.Bool("github-skip-verify"),
		MergeRef:    c.BoolT("github-merge-ref"),
	})
}

// helper function to setup the Coding remote from the CLI arguments.
func setupCoding(c *cli.Context) (remote.Remote, error) {
	return coding.New(coding.Opts{
		URL:        c.String("coding-server"),
		Client:     c.String("coding-client"),
		Secret:     c.String("coding-secret"),
		Scopes:     c.StringSlice("coding-scope"),
		Machine:    c.String("coding-git-machine"),
		Username:   c.String("coding-git-username"),
		Password:   c.String("coding-git-password"),
		SkipVerify: c.Bool("coding-skip-verify"),
	})
}

func setupTree(c *cli.Context) *httptreemux.ContextMux {
	tree := httptreemux.NewContextMux()
	// web.New(
	// 	web.WithDir(c.String("www")),
	// 	web.WithSync(time.Hour*72),
	// ).Register(tree)
	return tree
}

func before(c *cli.Context) error { return nil }
