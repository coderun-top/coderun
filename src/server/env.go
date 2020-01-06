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
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/coderun-top/coderun/src/model"
	"github.com/coderun-top/coderun/src/store"
)

// Project Env
func GetProjectEnv(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	data, err := store.FromContext(c).GetProjectEnvs(project.ID)
	if err != nil {
		c.String(500, "Error getting project_env list. %s", err)
		return
	}
	c.JSON(200, data)
}

func PutProjectEnv(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	var envs []*model.ProjectEnv

	err = c.BindJSON(&envs)
	if err != nil {
		logrus.Errorf("Error. %s", err)
		c.String(500, "内部错误")
		return
	}

	err = store.FromContext(c).UpdateProjectEnvs(project.ID, envs)
	if err != nil {
		logrus.Errorf("Error. %s", err)
		c.String(500, "内部错误")
		return
	}
}

// pipeline Env
func GetPipelineEnv(c *gin.Context) {
	// repo := session.Repo(c)

	// pipelineID, err := strconv.ParseInt(c.Param("pipeline_id"), 10, 64)
	pipelineID := c.Param("pipeline_id")

	data, err := store.GetPipelineEnvs(c, pipelineID)
	if err != nil {
		c.String(500, "Error getting pipeline_env list. %s", err)
		return
	}
	c.JSON(200, data)
}

func PutPipelineEnv(c *gin.Context) {
	// repo := session.Repo(c)

	// pipelineID, err := strconv.ParseInt(c.Param("pipeline_id"), 10, 64)
	// if err != nil{
	// 	c.String(500, "pipeline_id值为数字类型")
	// 	return
	// }
	pipelineID := c.Param("pipeline_id")

	var envs []*model.PipelineEnv

	err := c.BindJSON(&envs)
	if err != nil {
		logrus.Errorf("Error. %s", err)
		c.String(500, "内部错误")
		return
	}

	err = store.UpdatePipelineEnvs(c, pipelineID, envs)
	if err != nil {
		logrus.Errorf("Error. %s", err)
		c.String(500, "内部错误")
		return
	}
}
