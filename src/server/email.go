
package server

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/coderun-top/coderun/src/store"
	"github.com/coderun-top/coderun/src/model"
	//"github.com/sirupsen/logrus"
	"github.com/coderun-top/coderun/src/router/middleware/session"
)

func GetPipelineEmail(c *gin.Context) {
	repo := session.Repo(c)
	email, err := store.FromContext(c).GetEmail(repo)
	if err != nil {
		c.String(209, "Cannot find pipelines. %s", err)
		return
	}
	c.JSON(200, email)
}

func PutPipelineEmail(c *gin.Context) {
	repo := session.Repo(c)

	in := &model.Email{}
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	_, err = store.FromContext(c).GetEmail(repo)
	if err != nil {
		email := &model.Email{
			RepoID: repo.ID,
			Email:  in.Email,
			State: in.State,
		}
		err = store.FromContext(c).CreateEmail(email)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, email)
	} else {
		email := &model.Email{
			ID:  in.ID,
			RepoID: repo.ID,
			Email:  in.Email,
			State: in.State,
		}
		err = store.FromContext(c).UpdateEmail(email)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, email)
	}
}


// project email
func GetProjectEmail(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}
	email, err := store.FromContext(c).GetEmailProject(project.ID)
	if err != nil {
		c.String(404, "Cannot find pipelines. %s", err)
		return
	}
	c.JSON(200, email)
}

func PutProjectEmail(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	in := &model.EmailProject{}
	err = c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	_, err = store.FromContext(c).GetEmailProject(project.ID)
	if err != nil {
		email := &model.EmailProject{
			ProjectID: project.ID,
			Email:  in.Email,
			State: in.State,
		}
		err = store.FromContext(c).CreateEmailProject(email)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, email)
	} else {
		email := &model.EmailProject{
			ID:  in.ID,
			ProjectID: project.ID,
			Email:  in.Email,
			State: in.State,
		}
		err = store.FromContext(c).UpdateEmailProject(email)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(200, email)
	}
}
