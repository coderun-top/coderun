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

package internal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"errors"
	"github.com/coderun-top/coderun/src/model"
	"github.com/sirupsen/logrus"
)

type Project struct {
	Owner     string `json:"owner_user_name"`
	Name      string `json:"name"`
	DepotPath string `json:"depot_path"`
	HttpsURL  string `json:"https_url"`
	IsPublic  bool   `json:"is_public"`
	Icon      string `json:"icon"`
	Role      string `json:"current_user_role"`
}

type Depot struct {
	DefaultBranch string `json:"default_branch"`
}

type ProjectListData struct {
	Page      int        `json:"page"`
	PageSize  int        `json:"pageSize"`
	TotalPage int        `json:"totalPage"`
	TotalRow  int        `json:"totalRow"`
	List      []*Project `json:"list"`
}

type BranchesData struct {
	Code int      `json:"code"`
	List      []*model.Branch `json:"data"`
}

func (c *Client) GetProject(globalKey, projectName string, t int) (*Project, error) {
	u := fmt.Sprintf("/user/%s/project/%s", globalKey, projectName)
	resp, err := c.Get(u, nil, t)
	if err != nil {
		return nil, err
	}

	project := &Project{}
	err = json.Unmarshal(resp, project)
	if err != nil {
		return nil, APIClientErr{"fail to parse project data", u, err}
	}
	return project, nil
}

func (c *Client) GetDepot(globalKey, projectName string, t int) (*Depot, error) {
	u := fmt.Sprintf("/user/%s/project/%s/git", globalKey, projectName)
	resp, err := c.Get(u, nil, t)
	if err != nil {
		return nil, err
	}

	depot := &Depot{}
	err = json.Unmarshal(resp, depot)
	if err != nil {
		return nil, APIClientErr{"fail to parse depot data", u, err}
	}
	return depot, nil
}

func (c *Client) GetProjectList(t int) ([]*Project, error) {
	u := "/user/projects"
	resp, err := c.Get(u, nil, t)
	if err != nil {
		return nil, err
	}
	data := &ProjectListData{}
	err = json.Unmarshal(resp, data)
	if err != nil {
		return nil, APIClientErr{"fail to parse project list data", u, err}
	}
	if data.TotalPage == 1 {
		return data.List, nil
	}

	projectList := make([]*Project, 0)
	projectList = append(projectList, data.List...)
	for i := 2; i <= data.TotalPage; i++ {
		params := url.Values{}
		params.Set("page", fmt.Sprintf("%d", i))
		resp, err := c.Get(u, params, t)
		if err != nil {
			return nil, err
		}
		data := &ProjectListData{}
		err = json.Unmarshal(resp, data)
		if err != nil {
			return nil, APIClientErr{"fail to parse project list data", u, err}
		}
		projectList = append(projectList, data.List...)
	}
	return projectList, nil
}

func (c *Client) GetBranches(globalKey, projectName string, t int) ([]*model.Branch, error) {
	u := fmt.Sprintf("/user/%s/project/%s/git/list_branches", globalKey, projectName)
	logrus.Debugf("url: %s", u)
	resp, err := c.Get(u, nil, t)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("resp: %s", resp)
	var branchesdata  []*model.Branch
	err = json.Unmarshal(resp, &branchesdata)
	logrus.Debugf("branches: %s", branchesdata)
	if err != nil {
		return nil, APIClientErr{"fail to parse project list data", u, err}
	}
	return branchesdata,nil
}

func (c *Client) GetBranch(globalKey, projectName string, t int, branchName string) (*model.Branch, error) {
	u := fmt.Sprintf("/user/%s/project/%s/git/list_branches", globalKey, projectName)
	logrus.Debugf("url: %s", u)
	resp, err := c.Get(u, nil, t)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("resp: %s", resp)
	var branchesdata  []*model.Branch2
	err = json.Unmarshal(resp, &branchesdata)
	logrus.Debugf("branches: %s", branchesdata)
	logrus.Debugf("branches: %s", branchesdata)
	if err != nil {
		return nil, APIClientErr{"fail to parse project list data", u, err}
	}
	for _, branch := range branchesdata {
	    if branch.Name == branchName {
		    var dougoBranch = new(model.Branch)
		    dougoBranch.Name   = branch.Name
		    var dougoCommit = new(model.Commit)
		    dougoCommit.SHA         = branch.SHA
		    dougoCommit.AuthorName  = ""
		    dougoCommit.AuthorEmail = ""
		    dougoCommit.Message     = ""
		    dougoCommit.URL         = ""
		    dougoBranch.Commit = dougoCommit
		    return dougoBranch, err
		}
	}
	return nil, errors.New("brance not ")



}
