package server

import (
	"net/http"
	"strconv"

	//"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"github.com/coderun-top/coderun/src/model"
	"github.com/coderun-top/coderun/src/store"
	"github.com/coderun-top/coderun/src/router/middleware/session"
	//"github.com/coderun-top/coderun/src/utils"
	//"github.com/coderun-top/coderun/src/core/pipeline/pipeline/frontend/yaml"
	"gopkg.in/yaml.v2"

)


func CreateK8sDeploy(c *gin.Context) {
	in := &model.K8sDeploy{}
	if err := c.Bind(in); err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	if in.ClusterName == "" {
			c.String(500, "cluster not null")
			return
	}
	if in.Name == ""{
			c.String(500, "label not null")
			return
	}
	if in.Type == 1 {
		if in.Namespace == ""{
			c.String(500, "not deployment namespace")
			return
		}
		if in.Replicas < 1{
			c.String(500, "not deployment namespace")
			return
		}
		if in.Image == ""{
			c.String(500, "not deployment image")
			return
		}
		if in.InternalPort < 1{
			c.String(500, "not deployment internalport < 0")
			return
		}
		if in.ExposePort < 0{
			c.String(500, "not deployment exposeport > 0")
			return
		}
	} else if in.Type == 2{
		if in.DeployContent == ""{
			c.String(500, "not file deploy content")
			return
		}
		if in.ServiceContent == ""{
			c.String(500, "not file service content")
			return
		}
	} else {
			c.String(500, "type error 1,2")
			return
	}
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}
	if in.ID > 0 {

		k8s_deploy := &model.K8sDeploy{
			ID:  in.ID,
			ClusterName: in.ClusterName,
			ProjectID:  project.ID,
			Namespace: in.Namespace,
			Name: in.Name,
			Image: in.Image,
			ExposePort: in.ExposePort,
			Replicas: in.Replicas,
			InternalPort: in.InternalPort,
			DeployContent: in.DeployContent,
			ServiceContent: in.ServiceContent,
			Type: in.Type,
		}
		err := store.FromContext(c).K8sDeployUpdate(k8s_deploy)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, k8s_deploy)
	}  else {
		k8s_deploy := &model.K8sDeploy{
			ClusterName: in.ClusterName,
			ProjectID:  project.ID,
			Namespace: in.Namespace,
			Name: in.Name,
			Image: in.Image,
			ExposePort: in.ExposePort,
			Replicas: in.Replicas,
			InternalPort: in.InternalPort,
			DeployContent: in.DeployContent,
			ServiceContent: in.ServiceContent,
			Type: in.Type,
		}
		err := store.FromContext(c).K8sDeployCreate(k8s_deploy)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, k8s_deploy)
	}
}

func GetK8sDeployList(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}
	k8s_deploy_list, err := store.FromContext(c).K8sDeployList(project.ID)
	if err != nil {
		c.String(500, err.Error())
		return
	}
	c.JSON(200, k8s_deploy_list)
}
func GetK8sDeploy(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}
	k8s_deploy, err := store.FromContext(c).K8sDeployFindProject(project.ID, c.Param("name"))
	if err != nil {
		c.String(500, err.Error())
		return
	}
	c.JSON(200, k8s_deploy)
}


func K8sDeployBind(c *gin.Context) {
	repo := session.Repo(c)
	k8s_deploy_id, _ := strconv.ParseInt(c.Query("id"), 10, 64)

	conf, err := store.FromContext(c).ConfigRepoFind(repo)
	if err != nil {
		c.String(404, "Cannot find pipelines. %s", err)
		return
	}
	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(conf[0].Data), m)
	if err != nil {
		c.String(404, "Cannot find pipelines. %s", err)
		return
	}

	k8s_deploy, err := store.FromContext(c).K8sDeployFind(k8s_deploy_id)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	namespace := "default"
	if k8s_deploy.Namespace != "" {
		namespace = k8s_deploy.Namespace
	}
	m2 := map[interface{}]interface{}{"deploy_name": k8s_deploy.Name,
				     "cluster": k8s_deploy.ClusterName,
				     "tags": "latest",
				     "image":  "crun/kube",
				     "namespace": namespace,
				     "template": "deployment.yaml",
				     "template_image": k8s_deploy.Image,
				      }

	mm, ok := m["steps"].(map[interface{}]interface{})
	if ok {
	    mm[k8s_deploy.Name] = m2
	}

	//m["steps"] = map[interface{}]interface{}{c.Query("label"): m2,}
	d, err := yaml.Marshal(m)
	if err != nil {
		c.String(500, err.Error())
	}
	sha := shasum(d)
	config := &model.Config{
		ID:          conf[0].ID,
		RepoID:      conf[0].RepoID,
		Data:        string(d),
		Hash:        sha,
		FilePath:    conf[0].FilePath,
		File:        conf[0].File,
		AgentPublic: conf[0].AgentPublic,
		AgentFilter: conf[0].AgentFilter,
	}

	err = store.FromContext(c).ConfigUpdate(config)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.JSON(200, "ok")
}

func DeleteK8sDeploy(c *gin.Context) {
	k8s_deploy_id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	err := store.FromContext(c).K8sDeployDelete(k8s_deploy_id)
	if err != nil {
		c.String(404, "Cannot find pipelines. %s", err)
		return
	}
	return
}
