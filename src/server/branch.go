package server

import (
	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"github.com/coderun-top/coderun/src/router/middleware/session"
	"github.com/coderun-top/coderun/src/store"
)

func GetRepoBranches(c *gin.Context) {
	repo := session.Repo(c)
	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		logrus.Errorf("01 failure to find repo owner %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}

	remote, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Errorf("get remote error. %s", err)
		c.AbortWithError(400, err)
		return
	}

	branches, err := remote.Branches(user, "", repo.FullName)

	c.JSON(200, branches)
}

func GetRepoFile(c *gin.Context) {
	repo := session.Repo(c)
	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		logrus.Errorf("01 failure to find repo owner %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}

	branch := c.Query("branch")
	file := c.Query("file")
	//docker_context := c.Query("context")
	//if docker_context == "" {
	//	docker_context = "/"
	//}

	remote, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Errorf("get remote error. %s", err)
		c.AbortWithError(400, err)
		return
	}

	confb, err := remote.FileRef(user, repo, branch, file)
	if err != nil {
		logrus.Errorf("error: %s: cannot find %s in %s: %s")
		c.AbortWithError(404, err)
		return
	}

	c.String(200, string(confb))
}

func RepoEmptyFile(c *gin.Context) {
	repo := session.Repo(c)
	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		logrus.Errorf("01 failure to find repo owner %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}

	branch := c.Query("branch")
	file := c.Query("file")

	remote, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Errorf("get remote error. %s", err)
		c.AbortWithError(400, err)
		return
	}

	_, err = remote.FileRef(user, repo, branch, file)
	if err != nil {
		c.String(405, "Cannot find file. %s", err)
	} else {
		c.String(200, "OK")
	}
}

