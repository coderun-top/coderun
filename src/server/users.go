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
	"encoding/base32"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"

	"github.com/coderun-top/coderun/src/model"
	"github.com/coderun-top/coderun/src/remote"
	"github.com/coderun-top/coderun/src/router/middleware/token"
	"github.com/coderun-top/coderun/src/store"
)

func GetUsers(c *gin.Context) {
	userInfo := token.GetUserInfo(c)

	project, err := store.FromContext(c).GetProjectName(userInfo.Username)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	users, err := store.GetUserUserName(c, project.ID)
	if err != nil {
		c.String(500, "Error getting user list. %s", err)
		return
	}
	c.JSON(200, users)
}

func GetUser(c *gin.Context) {
	user, err := store.GetUserLogin(c, c.Param("login"))
	if err != nil {
		c.String(404, "Cannot find user. %s", err)
		return
	}
	c.JSON(200, user)
}

func PatchUser(c *gin.Context) {
	in := &model.User{}
	err := c.Bind(in)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	user, err := store.GetUserLogin(c, c.Param("login"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	user.Active = in.Active

	err = store.UpdateUser(c, user)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusOK, user)
}

func PostUser(c *gin.Context) {
	in := &model.User{}
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	project, err := store.FromContext(c).GetProject(in.ProjectID)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	user := &model.User{
		Active:    true,
		Login:     in.Login,
		Token:     in.Token,
		Email:     in.Email,
		Avatar:    in.Avatar,
		Provider:  in.Provider,
		TokenName: in.TokenName,
		ProjectID: project.ID,
		Synced:    time.Now().Unix(),
		Hash: base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		),
	}
	//token := token.New(token.UserToken, user.Login)
	//tokenstr, err := token.Sign(user.Hash)
	//if err != nil {
	//	c.AbortWithError(http.StatusInternalServerError, err)
	//	return
	//}
	//user.Token = tokenstr
	if err = user.Validate(); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if err = store.CreateUser(c, user); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	sync := syncer{
		remote:  remote.FromContext(c),
		store:   store.FromContext(c),
		perms:   store.FromContext(c),
		limiter: Config.Services.Limiter,
	}
	if err := sync.Sync(user); err != nil {
		logrus.Debugf("sync error: %s: %s", user.Login, err)
	} else {
		logrus.Debugf("sync complete: %s", user.Login)
	}
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	user, err := store.GetUserLogin(c, c.Param("login"))
	if err != nil {
		c.String(404, "Cannot find user. %s", err)
		return
	}
	if err = store.DeleteUser(c, user); err != nil {
		c.String(500, "Error deleting user. %s", err)
		return
	}
	c.String(200, "")
}
