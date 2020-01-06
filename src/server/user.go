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
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"

	"github.com/coderun-top/coderun/src/model"
	//"github.com/coderun-top/coderun/src/remote"
	"github.com/coderun-top/coderun/src/router/middleware/session"
	//"github.com/coderun-top/coderun/src/shared/token"
	"encoding/base64"
	pb "github.com/coderun-top/coderun/src/grpc/user"
	tt "github.com/coderun-top/coderun/src/router/middleware/token"
	"github.com/coderun-top/coderun/src/store"
)

type XToken struct {
	Provider int64  `json:"provider"`
	Token    string `json:"token"`
	Name     string `json:"name"`
}

type XToken2 struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

func GetSelf(c *gin.Context) {
	v, ok := c.Get("user_info")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	c.JSON(200, v)
}

func GetFeed(c *gin.Context) {
	user := session.User(c)
	latest, _ := strconv.ParseBool(c.Query("latest"))

	if time.Unix(user.Synced, 0).Add(time.Hour * 72).Before(time.Now()) {
		logrus.Debugf("sync begin: %s", user.Login)

		user.Synced = time.Now().Unix()
		store.FromContext(c).UpdateUser(user)
		remote_, err := SetupRemote(user.Provider, user.Host)
		if err != nil {
			logrus.Fatal("get remote error. %s", err)
			c.AbortWithError(400, err)
			return
		}

		sync := syncer{
			remote:  remote_,
			store:   store.FromContext(c),
			perms:   store.FromContext(c),
			limiter: Config.Services.Limiter,
		}
		if err := sync.Sync(user); err != nil {
			logrus.Debugf("sync error: %s: %s", user.Login, err)
		} else {
			logrus.Debugf("sync complete: %s", user.Login)
		}
	}

	if latest {
		feed, err := store.FromContext(c).RepoListLatest(user)
		if err != nil {
			c.String(500, "Error fetching feed. %s", err)
		} else {
			c.JSON(200, feed)
		}
		return
	}

	feed, err := store.FromContext(c).UserFeed(user)
	if err != nil {
		c.String(500, "Error fetching user feed. %s", err)
		return
	}
	c.JSON(200, feed)
}

func GetRepos(c *gin.Context) {
	//v, ok := c.Get("headimage")
	//if !ok {
	//	c.String(500, "Error fetching feed.")
	//	return
	//}
	projectname := c.Param("projectname")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pagesize, _ := strconv.Atoi(c.DefaultQuery("pagesize", "10"))
	search := c.DefaultQuery("search", "")
	repos, err := store.FromContext(c).RepoListProjectPage(projectname, search, page, pagesize)
	if err != nil {
		c.String(http.StatusInternalServerError, "get users error: %s", err)
		return
	}

	count, err := store.FromContext(c).RepoListProjectCount(projectname, search)
	if err != nil {
		c.String(http.StatusInternalServerError, "get users error: %s", err)
		return
	}
	data := &model.ReposData{
		//Data:     model.ListCopy(repos, v.(string)),
		Data:     repos,
		Page:     page,
		PageSize: pagesize,
		Count:    count,
	}

	c.JSON(http.StatusOK, data)
}

func GetRepoName(c *gin.Context) {
	user_id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logrus.Debugf("string to int64 err. %s", err)
		c.AbortWithError(400, err)
		return
	}

	user, err := store.GetUser(c, user_id)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	remote_, err := SetupRemote(user.Provider, user.Host)
	if err != nil {
		logrus.Debugf("get remote error. %s", err)
		c.AbortWithError(400, err)
		return
	}

	if user != nil {
		repos, err := remote_.Repos(user)
		if err != nil {
			c.String(404, "Cannot find repos. %s", err)
			return
		}

		c.JSON(http.StatusOK, repos)
		return
	}
	c.String(400, "user does not exist")
	return

}

// 获取Douser中的Token，虽然是列表但是只有一个
func GetOAuthToken(token string) (*XToken, error) {
	r, err := Config.Services.Grpc.GetAllTokens(context.Background(), &pb.UserTokenRequest{Token: token, Name: "self", Username: ""})
	if err != nil {
		logrus.Errorf("GetAllTokens error: %v", err)
		return nil, err
	}

	var data []XToken
	err = json.Unmarshal([]byte(r.Message), &data)
	if err != nil {
		return nil, err
	}
	if len(data) <= 0 {
		return nil, errors.New("数据为空")
	}
	return &data[0], nil
}

func GetUserToken(token, username string) (*XToken2, error) {
	r, err := Config.Services.Grpc.PostUserToken(context.Background(), &pb.UserTokenRequest{Token: token, Name: "coderun", Username: username})

	if err != nil {
		logrus.Errorf("PostUserToken error: %v", err)
		// 如果失败发送get请求获取，因为重名直接使用，不需要生成
		// return nil, errors.New("http请求失败")
		r2, err := Config.Services.Grpc.GetUserToken(context.Background(), &pb.UserTokenRequest{Token: token, Name: "", Username: username})
		if err != nil {
			logrus.Errorf("GetUserToken error: %v", err)
			return nil, err
		}

		data := []*XToken2{}
		err = json.Unmarshal([]byte(r2.Message), &data)
		if err != nil {
			return nil, err
		}
		return data[0], nil
	}

	data := new(XToken2)
	err = json.Unmarshal([]byte(r.Message), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func AddUser(c *gin.Context, projectName string, provider int64, tokenName string, token string, t int, host string) (*model.User, error) {
	remote_, err := SetupRemote(provider, host)
	if err != nil {
		logrus.Errorf("create remote error. %s", err)
		// c.AbortWithError(400, err)
		return nil, err
	}

	// tmpuser, err := remote_.AddUser(token, t)
	tmpuser, err := remote_.User(token, t)
	logrus.Debug("tmpuser is ", tmpuser)

	if err != nil {
		logrus.Errorf("get remote user. %s", err)
		// c.AbortWithError(400, err)
		return nil, err
	}
	if tmpuser == nil {
		return nil, fmt.Errorf("get remote user is nil")
	}

	// 根据t类型判断oauth
	oauth := false
	if t == 2 {
		oauth = true
	}

	project, err := store.FromContext(c).GetProjectName(projectName)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return nil, err
	}
	project_id := project.ID

	// 总是添加project_id
	// if oauth == true {
	// 	project_id = ""
	// }

	user := &model.User{
		Active:    true,
		Login:     tmpuser.Login,
		Token:     tmpuser.Token,
		Email:     tmpuser.Email,
		Avatar:    tmpuser.Avatar,
		Provider:  provider,
		TokenName: tokenName,
		ProjectID: project_id,
		Synced:    time.Now().Unix(),
		Hash: base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		),
		Oauth: oauth,
		Host:  host,
	}
	logrus.Debug("login is ", user.Login)

	if err = user.Validate(); err != nil {
		// c.String(http.StatusBadRequest, err.Error())
		logrus.Error("err1 is ", err)
		return nil, err
	}
	if err = store.CreateUser(c, user); err != nil {
		// c.String(http.StatusInternalServerError, err.Error())
		logrus.Error("err2 is ", err)
		return nil, err
	}

	return user, nil
}

func GetToken(c *gin.Context) {
	var (
		// include_oauth, _ = strconv.ParseBool(c.Query("include_oauth"))
		projectName      = c.Param("projectname")
	)
	// userInfo := tt.GetUserInfo(c)
	// token := c.Request.Header.Get("Authorization")

	project, err := store.FromContext(c).GetProjectName(projectName)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	//users, err := store.GetUserUserName(c,  userInfo.Username)

	users, err := store.GetUserUserName(c, project.ID)
	if err != nil {
		c.String(404, "Cannot find users. %s", err)
		return
	}

	// if !include_oauth {
	// 	users, err = store.GetUserUserNameOauth(c, project.ID, false)
	// 	c.JSON(http.StatusOK, users)
	// 	return
	// }

	// 不获取当前用户的token，仅使用项目下的Token
	// user, err := store.GetUserUserName2(c, userInfo.Username, project.ID)
	// if err != nil {
	// 	c.String(404, "Cannot find user. %s", err)
	// 	return
	// }
	// data := []*model.User{}
	// data = append(data, user)
	// for _, o := range users {
	// 	data = append(data, o)
	// }
	// c.JSON(http.StatusOK, data)

	c.JSON(http.StatusOK, users)
	return
}

func PostToken(c *gin.Context) {
	var (
		provider_str = c.Query("provider")
		token        = c.Query("token")
		login        = c.Query("login")
		projectName  = c.Param("projectname")
		tokenName    = c.Param("tokenname")
		host         = c.Query("host")
	)
	userInfo := tt.GetUserInfo(c)
	if userInfo.Username == tokenName {
		c.String(400, "Token Name does not using Username")
		return
	}

	var provider int64 = 0
	if provider_str == "github" {
		provider = 1
	} else if provider_str == "gitlab" {
		provider = 2
	} else if provider_str == "coding" {
		provider = 3
	} else if provider_str == "bitbucket" {
		provider = 4
		username_password := login + ":" + token
		token = base64.StdEncoding.EncodeToString([]byte(username_password))
	} else if provider_str == "gitee" {
		provider = 5
	}

	project, err := store.FromContext(c).GetProjectName(projectName)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	_, err = store.GetUserNameTokenOauth2(c, project.ID, tokenName, false)
	// 如果用户不存在的话
	if err != nil {
		user, err := AddUser(c, projectName, provider, tokenName, token, 1, host)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, user)
		return
		// 用户已经存在
	} else {
		c.String(400, "user already exists")
		return
	}
}

func PatchToken(c *gin.Context) {
	var (
		token = c.Query("token")
	)

	user_id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logrus.Debugf("string to int64 err. %s", err)
		c.AbortWithError(400, err)
		return
	}
	user, err := store.GetUser(c, user_id)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user.Token = token

	err = store.UpdateUser(c, user)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusOK, user)

}

func DeleteToken(c *gin.Context) {
	user_id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logrus.Debugf("string to int64 err. %s", err)
		c.AbortWithError(400, err)
		return
	}
	user, err := store.GetUser(c, user_id)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	repos, err := store.FromContext(c).RepoList(user)
	if len(repos) < 1 {
		if err = store.DeleteUser(c, user); err != nil {
			c.String(500, "Error deleting user. %s", err)
			return
		}
		c.String(200, "")
		return
	} else {
		c.String(400, "repos does not exist")
	}
}
