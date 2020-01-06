
package server

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/coderun-top/coderun/src/store"
	"github.com/coderun-top/coderun/src/model"
	//"github.com/sirupsen/logrus"
	"github.com/coderun-top/coderun/src/router/middleware/session"
)

// pipeline webhook
func GetPipelineWebHook(c *gin.Context) {
	repo := session.Repo(c)
	webhook, err := store.FromContext(c).GetWebHook(repo)
	if err != nil {
		c.String(209, "Cannot find pipelines. %s", err)
		return
	}
	c.JSON(200, webhook)
}

func PutPipelineWebHook(c *gin.Context) {
	repo := session.Repo(c)

	in := &model.WebHook{}
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	_, err = store.FromContext(c).GetWebHook(repo)
	if err != nil {
		webhook := &model.WebHook{
			RepoID: repo.ID,
			Url:  in.Url,
			State: in.State,
		}
		err = store.FromContext(c).CreateWebHook(webhook)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, webhook)
	} else {
		webhook := &model.WebHook{
			ID:  in.ID,
			RepoID: repo.ID,
			Url:  in.Url,
			State: in.State,
		}
		err = store.FromContext(c).UpdateWebHook(webhook)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, webhook)
	}
}

// project webhook
func GetProjectWebHook(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}
	webhook, err := store.FromContext(c).GetWebHookProject(project.ID)
	if err != nil {
		c.String(404, "Cannot find pipelines. %s", err)
		return
	}
	c.JSON(200, webhook)
}

func PutProjectWebHook(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	in := &model.WebHookProject{}
	err = c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	_, err = store.FromContext(c).GetWebHookProject(project.ID)
	if err != nil {
		webhook := &model.WebHookProject{
			ProjectID: project.ID,
			Url:  in.Url,
			State: in.State,
		}
		err = store.FromContext(c).CreateWebHookProject(webhook)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, webhook)
	} else {
		webhook := &model.WebHookProject{
			ID:  in.ID,
			ProjectID: project.ID,
			Url:  in.Url,
			State: in.State,
		}
		err = store.FromContext(c).UpdateWebHookProject(webhook)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, webhook)
	}
}
