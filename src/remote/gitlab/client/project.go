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

package client

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/coderun-top/coderun/src/model"
)

const (
	projectsUrl       = "/projects"
	projectUrl        = "/projects/:id"
	branchesUrl       = "/projects/:id/repository/branches"
    branchUrl         = "/projects/:id/repository/branches/:branch"
	repoUrlRawFileRef = "/projects/:id/repository/files/:filepath"
	commitStatusUrl   = "/projects/:id/statuses/:sha"
)

type Commit struct {
    SHA         string  `json:"id,omitempty"`
    AuthorName  string  `json:"author_name,omitempty"`
    AuthorEmail string  `json:"author_email,omitempty"`
    Message     string  `json:"message,omitempty"`
    URL         string  `json:"url,omitempty"`
}

type Branch struct {
	Name        string `json:"name"`
    Commit      *Commit
}

// Get a list of all projects owned by the authenticated user.
func (g *Client) AllProjects(hide_archives bool, t int) ([]*Project, error) {
	var per_page = 100
	var projects []*Project

	for i := 1; true; i++ {
		contents, err := g.Projects(i, per_page, hide_archives, t)
		if err != nil {
			return projects, err
		}

		for _, value := range contents {
			projects = append(projects, value)
		}

		if len(projects) == 0 {
			break
		}

		if len(projects)/i < per_page {
			break
		}
	}

	return projects, nil
}

// Get a list of projects owned by the authenticated user.
func (c *Client) Projects(page int, per_page int, hide_archives bool, t int) ([]*Project, error) {
	projectsOptions := QMap{
		"page":       strconv.Itoa(page),
		"per_page":   strconv.Itoa(per_page),
		"membership": "true",
	}

	if hide_archives {
		projectsOptions["archived"] = "false"
	}
	//projectsOptions["search"] = search

	url, opaque := c.ResourceUrl(projectsUrl, nil, projectsOptions)

	var projects []*Project

	contents, err := c.Do("GET", url, opaque, nil, t)
	if err == nil {
		err = json.Unmarshal(contents, &projects)
	}

	return projects, err
}

// Get a project by id
func (c *Client) Project(id string, t int) (*Project, error) {
	url, opaque := c.ResourceUrl(projectUrl, QMap{":id": id}, nil)

	var project *Project

	contents, err := c.Do("GET", url, opaque, nil, t)
	if err == nil {
		err = json.Unmarshal(contents, &project)
	}

	return project, err
}

// Get a project by id
func (c *Client) Branches(id string,t int) ([]*model.Branch, error) {
	url, opaque := c.ResourceUrl(branchesUrl, QMap{":id": id}, nil)
	logrus.Debugf("project id:", id)

	var branches []*model.Branch

	contents, err := c.Do("GET", url, opaque, nil, t)
	if err == nil {
		err = json.Unmarshal(contents, &branches)
	}

	return branches, err
}

func (c *Client) Branch(id string, t int, branchName string) (*model.Branch, error) {
    url, opaque := c.ResourceUrl(branchUrl, QMap{":id": id, ":branch": branchName}, nil)

	var branch *Branch

	contents, err := c.Do("GET", url, opaque, nil, t)
	if err == nil {
		err = json.Unmarshal(contents, &branch)
	}

    var dougoBranch = new(model.Branch)
    dougoBranch.Name   = branch.Name
    var dougoCommit = new(model.Commit)
    dougoCommit.SHA         = branch.Commit.SHA
    dougoCommit.AuthorName  = branch.Commit.AuthorName
    dougoCommit.AuthorEmail = branch.Commit.AuthorEmail
    dougoCommit.Message     = branch.Commit.Message
    dougoCommit.URL         = branch.Commit.URL
    dougoBranch.Commit = dougoCommit

	return dougoBranch, err
}

func (c *Client) RepoRawFileRef(id, ref, filepath string,t int) ([]byte, error) {
	var fileRef FileRef
	url, opaque := c.ResourceUrl(
		repoUrlRawFileRef,
		QMap{
			":id":       id,
			":filepath": filepath,
		},
		QMap{
			"ref": ref,
		},
	)

	contents, err := c.Do("GET", url, opaque, nil, t)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &fileRef)
	if err != nil {
		return nil, err
	}

	fileRawContent, err := base64.StdEncoding.DecodeString(fileRef.Content)
	return fileRawContent, err
}

//
func (c *Client) SetStatus(id, sha, state, desc, ref, link string,t int) error {
	url, opaque := c.ResourceUrl(
		commitStatusUrl,
		QMap{
			":id":  id,
			":sha": sha,
		},
		QMap{
			"state":       state,
			"ref":         ref,
			"target_url":  link,
			"description": desc,
			"context":     "ci/drone",
		},
	)

	_, err := c.Do("POST", url, opaque, nil, t)
	return err
}

// Get a list of projects by query owned by the authenticated user.
func (c *Client) SearchProjectId(namespace string, name string, t int) (id int, err error) {

	url, opaque := c.ResourceUrl(projectsUrl, nil, QMap{
		"query":      strings.ToLower(name),
		"membership": "true",
	})

	var projects []*Project

	contents, err := c.Do("GET", url, opaque, nil, t)
	if err == nil {
		err = json.Unmarshal(contents, &projects)
	} else {
		return id, err
	}

	for _, project := range projects {
		if project.Namespace.Name == namespace && strings.ToLower(project.Name) == strings.ToLower(name) {
			id = project.Id
		}
	}

	return id, err
}
