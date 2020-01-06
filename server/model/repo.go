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

package model

import (
	"fmt"
	"strings"
)

type RepoLite struct {
	Owner    string `json:"owner"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Avatar   string `json:"avatar_url"`
}

type ReposData struct {
	Data     []*Repo `json:"data"`
	Count    int     `json:"count"`
	PageSize int     `json:"pagesize"`
	Page     int     `json:"page"`
}

// Repo represents a repository.
//
// swagger:model repo
type Repo struct {
	ID          int64    `json:"id,omitempty"             meddler:"repo_id,pk"               gorm:"AUTO_INCREMENT;primary_key;column:repo_id"`
	UserID      int64    `json:"-"                        meddler:"repo_user_id"             gorm:"type:integer;column:repo_user_id;unique_index:idx_repos"`
	ProjectID   string   `json:"-"                        meddler:"repo_project_id"             gorm:"type:varchar(250);column:repo_project_id;unique_index:idx_repos"`
	Label       string   `json:"label"                    meddler:"repo_label"               gorm:"type:varchar(250);column:repo_label;unique_index:idx_repos"`
	Owner       string   `json:"owner"                    meddler:"repo_owner"               gorm:"type:varchar(250);column:repo_owner"`
	Name        string   `json:"name"                     meddler:"repo_name"                gorm:"type:varchar(250);column:repo_name"`
	FullName    string   `json:"full_name"                meddler:"repo_full_name"           gorm:"type:varchar(250);column:repo_full_name"`
	Avatar      string   `json:"avatar_url,omitempty"     meddler:"repo_avatar"              gorm:"type:varchar(500);column:repo_avatar"`
	Link        string   `json:"link_url,omitempty"       meddler:"repo_link"                gorm:"type:varchar(1000);column:repo_link"`
	Kind        string   `json:"scm,omitempty"            meddler:"repo_scm"                 gorm:"type:varchar(50);column:repo_scm"`
	Clone       string   `json:"clone_url,omitempty"      meddler:"repo_clone"               gorm:"type:varchar(1000);column:repo_clone"`
	Branch      string   `json:"default_branch,omitempty" meddler:"repo_branch"              gorm:"type:varchar(500);column:repo_branch"`
	Timeout     int64    `json:"timeout,omitempty"        meddler:"repo_timeout"             gorm:"type:integer;column:repo_timeout"`
	Visibility  string   `json:"visibility"               meddler:"repo_visibility"          gorm:"type:varchar(50);column:repo_visibility"`
	IsPrivate   bool     `json:"private"                  meddler:"repo_private"             gorm:"column:repo_private"`
	IsTrusted   bool     `json:"trusted"                  meddler:"repo_trusted"             gorm:"column:repo_trusted"`
	IsStarred   bool     `json:"starred,omitempty"        meddler:"-"                        gorm:"-"`
	IsGated     bool     `json:"gated"                    meddler:"repo_gated"               gorm:"column:repo_gated"`
	IsActive    bool     `json:"active"                   meddler:"repo_active"              gorm:"column:repo_active"`
	AllowPull   bool     `json:"allow_pr"                 meddler:"repo_allow_pr"            gorm:"column:repo_allow_pr"`
	AllowPush   bool     `json:"allow_push"               meddler:"repo_allow_push"          gorm:"column:repo_allow_push"`
	AllowDeploy bool     `json:"allow_deploys"            meddler:"repo_allow_deploys"       gorm:"column:repo_allow_deploys"`
	AllowTag    bool     `json:"allow_tags"               meddler:"repo_allow_tags"          gorm:"column:repo_allow_tags;default:true"`
	Counter     int      `json:"last_build"               meddler:"repo_counter"             gorm:"column:repo_counter"`
	Config      string   `json:"config_file"              meddler:"repo_config_path"         gorm:"type:varchar(500);column:repo_config_path"`
	Hash        string   `json:"-"                        meddler:"repo_hash"                gorm:"type:varchar(500);column:repo_hash"`
	Perm        *Perm    `json:"-"                        meddler:"-"                        gorm:"-"`
	User        *User    `json:"user"`
	Project     *Project `json:"project"`
	//UserTokenName string
}

func ListCopy(r []*Repo, avatar string) []*Repo {

	data := []*Repo{}
	for _, o := range r {
		avar := o.Avatar
		if avar == "" {
			avar = avatar
		}
		data = append(data, &Repo{
			ID:          o.ID,
			Name:        o.Name,
			UserID:      o.UserID,
			Owner:       o.Owner,
			FullName:    o.FullName,
			Avatar:      avar,
			Link:        o.Link,
			Kind:        o.Kind,
			Clone:       o.Clone,
			Branch:      o.Branch,
			Timeout:     o.Timeout,
			Visibility:  o.Visibility,
			IsPrivate:   o.IsPrivate,
			IsTrusted:   o.IsTrusted,
			IsStarred:   o.IsStarred,
			IsGated:     o.IsGated,
			IsActive:    o.IsActive,
			AllowPull:   o.AllowPull,
			AllowPush:   o.AllowPush,
			AllowDeploy: o.AllowDeploy,
			AllowTag:    o.AllowTag,
			Counter:     o.Counter,
			Config:      o.Config,
			Hash:        o.Hash,
			Perm:        o.Perm,
		})
	}
	return data
}

func (r *Repo) ResetVisibility() {
	r.Visibility = VisibilityPublic
	if r.IsPrivate {
		r.Visibility = VisibilityPrivate
	}
}

// ParseRepo parses the repository owner and name from a string.
func ParseRepo(str string) (user, repo string, err error) {
	var parts = strings.Split(str, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("Error: Invalid or missing repository. eg octocat/hello-world.")
		return
	}
	user = parts[0]
	repo = parts[1]
	return
}

// Update updates the repository with values from the given Repo.
func (r *Repo) Update(from *Repo) {
	r.Avatar = from.Avatar
	r.Link = from.Link
	r.Kind = from.Kind
	r.Clone = from.Clone
	r.Branch = from.Branch
	if from.IsPrivate != r.IsPrivate {
		if from.IsPrivate {
			r.Visibility = VisibilityPrivate
		} else {
			r.Visibility = VisibilityPublic
		}
	}
	r.IsPrivate = from.IsPrivate
}

// RepoPatch represents a repository patch object.
type RepoPatch struct {
	Config       *string `json:"config_file,omitempty"`
	IsTrusted    *bool   `json:"trusted,omitempty"`
	IsGated      *bool   `json:"gated,omitempty"`
	Timeout      *int64  `json:"timeout,omitempty"`
	Visibility   *string `json:"visibility,omitempty"`
	AllowPull    *bool   `json:"allow_pr,omitempty"`
	AllowPush    *bool   `json:"allow_push,omitempty"`
	AllowDeploy  *bool   `json:"allow_deploy,omitempty"`
	AllowTag     *bool   `json:"allow_tag,omitempty"`
	BuildCounter *int    `json:"build_counter,omitempty"`
}

// type Branches struct {
// 	Name        string `json:"name"                     meddler:"repo_name"`
// }

type Commit2 struct {
	Message string `json:"message,omitempty"`
}

type Commit struct {
	SHA         string   `json:"sha,omitempty"`
	AuthorName  string   `json:"author_name,omitempty"`
	AuthorEmail string   `json:"author_email,omitempty"`
	Message     string   `json:"message,omitempty"`
	URL         string   `json:"url,omitempty"`
	Commit      *Commit2 `json:"commit"`
}

type Branch struct {
	Name   string `json:"name"`
	Commit *Commit
}

type Branch2 struct {
	Name string `json:"name"`
	SHA  string `json:"sha,omitempty"`
}
