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
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"

	pb "github.com/coderun-top/coderun/src/grpc/user"
	"github.com/coderun-top/coderun/src/model"
	tt "github.com/coderun-top/coderun/src/router/middleware/token"
	"github.com/coderun-top/coderun/src/store"
	"github.com/coderun-top/coderun/src/utils"
)

type ProjectCharging struct {
	BuildCount int   `json:"build_num"`
	BuildTime  int64 `json:"build_time"`
	AgentCount int   `json:"agent_num"`
	ConCount   int   `json:"concurrency_num"`
}

func GetProjectCharging(projectName string) (*ProjectCharging, error) {
	r, err := Config.Services.Grpc.GetProjectChargingByName(context.Background(), &pb.ProjectRequest{Token: "", Projectname: projectName})
	if err != nil {
		return nil, fmt.Errorf("grpc error, %s", err.Error())
	}

	projectCharging := new(ProjectCharging)
	err = json.Unmarshal([]byte(r.Message), projectCharging)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error, %s", err.Error())
	}
	return projectCharging, nil
}

func ChargingCheck(c *gin.Context, project *model.Project) error {
	count, err := store.FromContext(c).GetBuildListProjectCountByMonth(project.ID, "")
	if err != nil {
		msg := fmt.Sprintf("Error getting build count for Project: %s. %+v", project.Name, err)
		logrus.Error(msg)
		return fmt.Errorf(msg)
	}

	buildTime, err := store.FromContext(c).GetBuildTimeProject(project.ID)
	if err != nil {
		msg := fmt.Sprintf("Error getting build time for Project: %s. %+v", project.Name, err)
		logrus.Error(msg)
		return fmt.Errorf(msg)
	}

	nums, err := GetProjectCharging(project.Name)
	if err != nil {
		msg := fmt.Sprintf("Error getting charging for Project: %s. %+v", project.Name, err)
		logrus.Error(msg)
		return fmt.Errorf(msg)
	}
	if count > nums.BuildCount {
		msg := fmt.Sprintf("Current Builds %d, The amount has been more than %d/month", count, nums.BuildCount)
		logrus.Error(msg)
		return fmt.Errorf(msg)
	}

	if buildTime > nums.BuildTime {
		msg := fmt.Sprintf("Builds time %d, The amount has been more than %d", buildTime, nums.BuildTime)
		logrus.Error(msg)
		return fmt.Errorf(msg)
	}

	return nil
}

func InitProject(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	projectName := c.Param("projectname")
	userInfo := tt.GetUserInfo(c)

	// create project, if it not exist in dougo
	project, err := store.FromContext(c).GetProjectName(projectName)
	if err != nil {
		r, err := Config.Services.Grpc.GetProject(context.Background(), &pb.ProjectRequest{Token: token, Projectname: projectName})
		if err != nil {
			msg := "Fetch project from douser error"
			logrus.Errorf("%s, %s", msg, err.Error())
			c.JSON(500, msg)
			return
		}
		if r.Message == "null" {
			msg := "Fetch project error, it not exists"
			logrus.Errorf("%s, %s", msg, err.Error())
			c.JSON(400, msg)
			return
		}
		logrus.Debugf("Fetch project: %s", string(r.Message))

		var project model.Project
		err = json.Unmarshal([]byte(r.Message), &project)
		if err != nil {
			msg := "Project unmarshal fail"
			logrus.Errorf("%s, %s", msg, err.Error())
			c.JSON(500, msg)
			return
		}
		// project.Name = projectName

		if err := store.FromContext(c).CreateProject(&project); err != nil {
			errMsg := "Create project error"
			logrus.Errorf("%s, %s", errMsg, err.Error())
			c.String(500, errMsg)
			return
		}
	}

	// create default registry
	_, err = store.FromContext(c).RegistryFind(project.ID, "coderun")
	if err != nil {
		data, err := GetUserToken(token, userInfo.Username)
		if err == nil {
			registry := &model.Registry{
				ProjectID: project.ID,
				Name:      "coderun",
				Type:      "coderun",
				Address:   os.Getenv("REGISTRY_DOMAIN"),
				Username:  userInfo.Username,
				Password:  data.Token,
				Prefix:    projectName,
			}
			if err := store.FromContext(c).RegistryCreate(registry); err != nil {
				logrus.Debugf("registry 生成失败 err: %s", err)
			}
		} else {
			logrus.Debugf(" get user token err: %s", err)
		}
	}

	// create default helm charts
	_, err = store.FromContext(c).HelmFind(project.ID, "coderun")
	if err != nil {
		data, err := GetUserToken(token, userInfo.Username)
		if err == nil {
			helm := &model.Helm{
				ProjectID: project.ID,
				Name:      "coderun",
				Type:      "coderun",
				Address:   os.Getenv("HELM_DOMAIN"),
				Username:  userInfo.Username,
				Password:  data.Token,
				Prefix:    "default",
			}
			if err := store.FromContext(c).HelmCreate(helm); err != nil {
				logrus.Debugf("helm charts 生成失败 err: %s", err)
			}
		} else {
			logrus.Debugf("get user token err: %s", err)
		}
	}

	// create default token
	tokenName := "default"
	uu, err := store.GetUserNameTokenOauth2(c, project.ID, tokenName, true)
	if err != nil {
		data, _ := GetOAuthToken(token)
		if data != nil {
			_, err = AddUser(c, projectName, data.Provider, tokenName, data.Token, 2, "")
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
			}
		}
	} else {
		data, _ := GetOAuthToken(token)
		if data != nil {
			uu.Token = data.Token

			err = store.UpdateUser(c, uu)
			if err != nil {
				c.AbortWithStatus(http.StatusConflict)
				return
			}
		}
	}

	// create default agent key
	_, err = store.FromContext(c).GetAgentKey(project.ID)
	if err != nil {
		newKey := utils.GeneratorId()
		agentKey := &model.AgentKey{
			Key:       newKey,
			ProjectID: project.ID,
		}
		err = store.FromContext(c).CreateAgentKey(agentKey)
		if err != nil {
			c.String(500, err.Error())
			return
		}
	}

	c.JSON(200, project)
}

func GetProjectInfo(c *gin.Context) {
	type RunStatus struct {
		RunCount   int   `json:"running_count"`
		AgentCount int   `json:"agent_count"`
		BuildCount int   `json:"build_count"`
		BuildTime  int64 `json:"build_time"`
	}

	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	agentCount, err := store.FromContext(c).GetAgentCount(project.ID)
	if err != nil {
		msg := "Agent count error"
		logrus.Errorf("%s, %s", msg, err.Error())
		c.String(500, fmt.Sprintf("%s, %s", msg, err.Error()))
		return
	}

	buildCount, err := store.FromContext(c).GetBuildListProjectCountByMonth(project.ID, "")
	if err != nil {
		msg := "Build count error"
		logrus.Errorf("%s, %s", msg, err.Error())
		c.String(500, fmt.Sprintf("%s, %s", msg, err.Error()))
		return
	}

	buildTime, err := store.FromContext(c).GetBuildTimeProject(project.ID)
	if err != nil {
		msg := "Build time count error"
		logrus.Errorf("%s, %s", msg, err.Error())
		c.String(500, fmt.Sprintf("%s, %s", msg, err.Error()))
		return
	}

	rs := RunStatus{
		RunCount:   Config.Services.Queue.GetRuning(c, project.Name),
		AgentCount: agentCount,
		BuildCount: buildCount,
		BuildTime:  buildTime,
	}

	c.JSON(200, rs)
}

// k8s cluster
func GetK8sClusterList(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}
	list, err := store.FromContext(c).K8sClusterList(project.ID)
	if err != nil {
		c.String(500, "Error getting k8s cluster list. %s", err)
		return
	}

	for i, cluster := range list {
		list[i] = cluster.Copy()
	}

	c.JSON(200, list)
}

func GetK8sCluster(c *gin.Context) {
	var (
		name = c.Param("name")
	)

	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	k8sCluster, err := store.FromContext(c).K8sClusterFind(project.ID, name)
	if err != nil {
		c.String(500, "Error find k8s cluster %s. %s", name, err)
		return
	}

	c.JSON(200, k8sCluster.Copy())
}

func GetK8sClusterInside(c *gin.Context) {
	var (
		name = c.Param("name")
	)

	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	k8sCluster, err := store.FromContext(c).K8sClusterFind(project.ID, name)
	if err != nil {
		c.String(500, "Error find k8s cluster %s. %s", name, err)
		return
	}

	c.JSON(200, k8sCluster)
}

func PostK8sCluster(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	in := new(model.K8sCluster)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}

	in.ProjectID = project.ID
	if err := store.FromContext(c).K8sClusterCreate(in); err != nil {
		c.String(500, "Error inserting k8s cluster %s. %s", in.Name, err)
		return
	}

	k8sCluster, err := store.FromContext(c).K8sClusterFind(project.ID, in.Name)
	if err != nil {
		c.String(500, "Error find k8s cluster %s. %s", in.Name, err)
		return
	}

	c.JSON(200, k8sCluster)
}

func PutK8sCluster(c *gin.Context) {
	var (
		name = c.Param("name")
	)

	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	k8sCluster, err := store.FromContext(c).K8sClusterFind(project.ID, name)
	if err != nil {
		c.String(500, "Error find k8s cluster %s. %s", name, err)
		return
	}

	in := new(model.K8sCluster)
	err = c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}

	if in.Address != "" {
		k8sCluster.Address = in.Address
	}
	if in.Cert != "" {
		k8sCluster.Cert = in.Cert
	}
	if in.Token != "" {
		k8sCluster.Token = in.Token
	}
	if in.KubeConfig != "" {
		k8sCluster.KubeConfig = in.KubeConfig
	}

	if err := store.FromContext(c).K8sClusterUpdate(k8sCluster); err != nil {
		c.String(500, "Error updating k8s cluster %q. %s", in.Name, err)
		return
	}
	c.JSON(200, in)
}

func DeleteK8sCluster(c *gin.Context) {
	var (
		name = c.Param("name")
	)

	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	if err := store.FromContext(c).K8sClusterDelete(project.ID, name); err != nil {
		c.String(500, "Error deleting k8s cluster %q. %s", name, err)
		return
	}
	c.String(204, "")
}
