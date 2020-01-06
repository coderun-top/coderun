package server

import (
	"net/http"
	// "strconv"
	// "encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/coderun-top/coderun/src/model"
	"github.com/coderun-top/coderun/src/router/middleware/session"
	"github.com/coderun-top/coderun/src/store"
)

// func PostConf(c *gin.Context) {
// 	in := &model.Config{}
// 	err := c.Bind(in)
// 	if err != nil {
// 		c.String(http.StatusBadRequest, err.Error())
// 		return
// 	}
//
// 	confb := []byte(in.Data)
// 	sha := shasum(confb)
// 	config := &model.Config{
// 		Data:  in.Data,
// 		Hash:   sha,
// 	}
// 	err = store.FromContext(c).ConfigCreate(config)
// 	if err != nil {
// 		c.String(500, err.Error())
// 		return
// 	}
//
// 	c.JSON(200, config)
// }

func GetPipelines(c *gin.Context) {
	repo := session.Repo(c)
	conf, err := store.FromContext(c).ConfigRepoFind(repo)
	if err != nil {
		c.String(404, "Cannot find pipelines. %s", err)
		return
	}
	c.JSON(200, conf)
}

func PutPipeline(c *gin.Context) {
	in := &model.Config{}

	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	repo, err := store.GetRepo(c, in.RepoID)
	if err != nil {
		logrus.Errorf("Cannot find repo by repoID: %s, %s", in.RepoID, err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if in.File == 1 {
		_, err = yamlParseLinter(in.Data, repo.IsTrusted)
		if err != nil {
			c.String(400, "yaml parse error: %s", err.Error())
			return
		}
	}

	confb := []byte(in.Data)
	sha := shasum(confb)
	config := &model.Config{
		ID:                  in.ID,
		RepoID:              in.RepoID,
		Data:                in.Data,
		Hash:                sha,
		FilePath:            in.FilePath,
		File:                in.File,
		AgentPublic:         in.AgentPublic,
		AgentFilter:         in.AgentFilter,
		ConfigType:          in.ConfigType,
		StepBuild:           in.StepBuild,
		StepUnitTest:        in.StepUnitTest,
		StepIntegrationTest: in.StepIntegrationTest,
		StepDeploy:          in.StepDeploy,
	}

	err = store.FromContext(c).ConfigUpdate(config)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.JSON(200, config)
}

// func DeleteConf(c *gin.Context) {
// 	repo := session.Repo(c)
// 	conf, err := store.FromContext(c).ConfigRepoFind(repo)
// 	if err != nil {
// 		c.String(404, "Cannot find conf. %s", err)
// 		return
// 	}
// 	c.JSON(200, conf)
// }
