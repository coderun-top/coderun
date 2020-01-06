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
	"strconv"
	"github.com/coderun-top/coderun/src/model"
	"github.com/gin-gonic/gin"
	"github.com/coderun-top/coderun/src/store"
)

// GetSecret gets the named secret from the database and writes
// to the response in json format.
func GetStarList(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	stars, err := store.FromContext(c).StarList(username.(string))
	if err != nil {
		c.String(404, "Error getting secret %q. %s", username.(string), err)
		return
	}
	c.JSON(200, stars)
}

func GetStarList2(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pagesize, _ := strconv.Atoi(c.DefaultQuery("pagesize", "10"))
	username, ok := c.Get("username")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	repos, err := store.FromContext(c).GetStarRepo(username.(string), page, pagesize)
	if err != nil {
		c.String(404, "Error getting secret %q. %s", username.(string), err)
		return
	}

	count, err := store.FromContext(c).GetStarRepoCount(username.(string))
	if err != nil {
		c.String(404, "Error getting secret %q. %s", username.(string), err)
		return
	}

	data := &model.ReposData{
		Data: repos,
		Page:     page,
		PageSize: pagesize,
		Count:    count,
	}
	c.JSON(200, data)
}

func PostStar(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	repo_id, err := strconv.ParseInt(c.Query("repo_id"), 10, 64)
	if err != nil {
		c.String(500, "repo_id args err  %s",  err)
		return
	}

	_, err = store.FromContext(c).StarFind2(repo_id, username.(string))
	if err == nil {
		c.String(404, "exiting star %s. %s. %s", repo_id, username.(string), err)
		return
	}

	if err := store.FromContext(c).StarCreate(&model.Star{
		UserName:   username.(string),
		RepoID:     repo_id,
	}); err != nil {
		c.String(500, "Error inserting star %s. %s", username.(string), err)
		return
	}
	c.JSON(200, "OK")
}

func DeleteStar(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	repo_id, err := strconv.ParseInt(c.Query("repo_id"), 10, 64)
	star, err := store.FromContext(c).StarFind2(repo_id, username.(string))
	if err != nil {
		c.String(404, "Error repo_id err %s. %s", repo_id, err)
		return
	}
	if err := store.FromContext(c).StarDelete(star); err != nil {
		c.String(500, "Error deleting star_id %q. %s", repo_id, err)
		return
	}
	c.String(204, "")
}
