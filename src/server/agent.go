package server

import (
	"net/http"
	// "strconv"

	// "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"

	// "github.com/coderun-top/coderun/src/model"
	// "errors"
	// "github.com/coderun-top/coderun/src/remote"
	// "github.com/coderun-top/coderun/src/router/middleware/session"
	// "github.com/coderun-top/coderun/src/shared/token"
	"github.com/coderun-top/coderun/src/store"
	// "github.com/coderun-top/coderun/src/utils"
)

func GetAgentKey(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	agentKey, err := store.FromContext(c).GetAgentKey(project.ID)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.JSON(200, agentKey)
}

func GetAgents(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	agents, err := store.FromContext(c).GetAgents(project.ID)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.JSON(200, agents)
}

func PutAgent(c *gin.Context) {
	agentID := c.Query("id")

	agent, err := store.FromContext(c).GetAgent(agentID)
	if err != nil {
		c.String(500, err.Error())
		return

	}
	// agent.Label = c.Query("label")
	err = store.FromContext(c).UpdateAgent(agent)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(200, agent)
}

func PutAgentStatus(c *gin.Context) {
	agentID := c.Query("id")

	var status int
	if c.Param("status") == "start" {
		status = 1
	} else if c.Param("status") == "stop" {
		status = 0
	} else {
		c.String(500, "status is 'start' or 'stop' ")
		return
	}

	agent, err := store.FromContext(c).GetAgent(agentID)
	if err != nil {
		c.String(500, err.Error())
		return

	}
	agent.Status = status
	err = store.FromContext(c).UpdateAgent(agent)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(200, agent)
}
