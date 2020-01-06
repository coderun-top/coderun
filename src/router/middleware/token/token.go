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

package token

import (
	"context"
	"encoding/json"
	"time"
	"fmt"

	"github.com/coderun-top/coderun/src/remote"
	"github.com/coderun-top/coderun/src/router/middleware/session"
	"github.com/coderun-top/coderun/src/store"

	// "github.com/coderun-top/coderun/src/server"
	log "github.com/sirupsen/logrus"
	pb "github.com/coderun-top/coderun/src/grpc/user"

	"github.com/gin-gonic/gin"
)

func Refresh(c *gin.Context) {
	user := session.User(c)
	if user == nil {
		c.Next()
		return
	}

	// check if the remote includes the ability to
	// refresh the user token.
	remote_ := remote.FromContext(c)
	refresher, ok := remote_.(remote.Refresher)
	if !ok {
		c.Next()
		return
	}

	// check to see if the user token is expired or
	// will expire within the next 30 minutes (1800 seconds).
	// If not, there is nothing we really need to do here.
	if time.Now().UTC().Unix() < (user.Expiry - 1800) {
		c.Next()
		return
	}

	// attempts to refresh the access token. If the
	// token is refreshed, we must also persist to the
	// database.
	ok, _ = refresher.Refresh(user)
	if ok {
		err := store.UpdateUser(c, user)
		if err != nil {
			// we only log the error at this time. not sure
			// if we really want to fail the request, do we?
			log.Errorf("cannot refresh access token for %s. %s", user.Login, err)
		} else {
			log.Infof("refreshed access token for %s", user.Login)
		}
	}

	c.Next()
}

type Project struct {
	ProjectId         int64             `json:"project_id"`
	OwnerId           int               `json:"owner_id"`
	Name              string            `json:"name"`
	CreationTime      time.Time         `json:"creation_time"`
	UpdateTime        time.Time         `json:"update_time"`
	Deleted           int               `json:"deleted"`
	OwnerName         string            `json:"owner_name"`
	Togglable         bool              `json:"togglable"`
	CurrentUserRoleId int               `json:"current_user_role_id"`
	RepoCount         int               `json:"repo_count"`
	Metadata          map[string]string `json:"metadata"`
}

type ProjectMember struct {
	ProjectId int `json:"project_id"`
	UserId    int `json:"user_id"`
	RoleId    int `json:"role_id"`
}

type UserCurrent struct {
	UserId         int             `json:"user_id"`
	Username       string          `json:"username"`
	Email          string          `json:"email"`
	Realname       string          `json:"realname"`
	Company        string          `json:"company"`
	CompanyScale   int             `json:"company_scale"`
	Phone          string          `json:"phone"`
	Role_name      string          `json:"role_name"`
	RoleDd         int             `json:"role_id"`
	HasAdminRole   int             `json:"has_admin_role"`
	ProjectId      int64           `json:"project_id"`
	ProjectName    string          `json:"project_name"`
	ResetUuid      string          `json:"reset_uuid"`
	Avatar         string          `json:"avatar"`
	Projects       []Project       `json:"projects"`
	ProjectMembers []ProjectMember `json:"project_members"`
}

// type TOKENS struct {
// 	UserId int    `json:"user_id"`
// 	Token  string `json:"token"`
// }

// func ToToken(c *gin.Context) *UserCurrent {
// 	v, ok := c.Get("user_info")
// 	if !ok {
// 		return nil
// 	}
// 	u, ok := v.(*UserCurrent)
// 	if !ok {
// 		return nil
// 	}
// 	return u
// }

func GetUserInfo(c *gin.Context) *UserCurrent {
	v, ok := c.Get("user_info")
	if !ok {
		return nil
	}
	u, ok := v.(*UserCurrent)
	if !ok {
		return nil
	}
	return u
}

//Config.Services.
// func VerifyToken(grpc_client pb.GreeterClient) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		token := c.Request.Header.Get("Authorization")
// 		r, err := grpc_client.TokenVerify2(context.Background(), &pb.HelloRequest{Token: token})
// 		if err != nil {
// 			log.Errorf("grpc token verify 2 %s", err)
// 			c.String(401, "用户没有登陆")
// 			c.Abort()
// 			return
// 		}
// 		data := new(UserCurrent)
// 		err = json.Unmarshal([]byte(r.Message), data)
// 		if err != nil {
// 			log.Errorf("json error: %v", err)
// 			c.String(401, "用户没有登陆")
// 			c.Abort()
// 			return
// 		}
// 		c.Set("user_info", data)
// 		c.Set("username", data.Username)
// 		c.Set("headimage", data.Avatar)
// 	}
// }
// 
// func VerifyToken2(grpc_client pb.GreeterClient) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		token := c.Request.Header.Get("Authorization")
// 
// 		r2, err := grpc_client.GetUserToken(context.Background(), &pb.UserTokenRequest{Token: token, Name: "", Username: c.Param("username")})
// 		if err != nil {
// 			log.Errorf("GetUserToken error: %v", err)
// 			c.String(401, "用户没有登陆")
// 			c.Abort()
// 			return
// 		}
// 
// 		data := []*TOKENS{}
// 		err = json.Unmarshal([]byte(r2.Message), &data)
// 		if err != nil {
// 			log.Errorf("json error: %v", err)
// 			c.String(401, "用户没有登陆")
// 			c.Abort()
// 			return
// 		}
// 		if len(data) > 0 {
// 			c.Set("username", c.Param("username"))
// 			c.Set("token", data[0].Token)
// 		}
// 
// 	}
// }
// 
// func VerifyToken3(grpc_client pb.GreeterClient) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		token := c.Request.Header.Get("Authorization")
// 		r, err := grpc_client.TokenVerify2(context.Background(), &pb.HelloRequest{Token: token})
// 		if err != nil {
// 			log.Errorf("grpc token verify 2 %s", err)
// 			c.String(401, "用户没有登陆")
// 			c.Abort()
// 			return
// 		}
// 		data := new(UserCurrent)
// 		err = json.Unmarshal([]byte(r.Message), data)
// 		if err != nil {
// 			log.Errorf("json error: %v", err)
// 			c.String(401, "用户没有登陆")
// 			c.Abort()
// 			return
// 		}
// 
// 		r2, err := grpc_client.GetUserToken(context.Background(), &pb.UserTokenRequest{Token: token, Name: "", Username: data.Username})
// 
// 		data2 := []*TOKENS{}
// 		err = json.Unmarshal([]byte(r2.Message), &data2)
// 		if err != nil {
// 			log.Errorf("json error: %v", err)
// 			c.String(401, "用户没有登陆")
// 			c.Abort()
// 			return
// 		}
// 		if len(data2) > 0 {
// 			c.Set("token", data2[0].Token)
// 		}
// 
// 	}
// }
// 
// func VerifyToken4(grpc_client pb.GreeterClient) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		token := c.Query("Authorization")
// 		r, err := grpc_client.TokenVerify2(context.Background(), &pb.HelloRequest{Token: token})
// 		if err != nil {
// 			log.Errorf("grpc token verify 2 %s", err)
// 			c.String(401, "用户没有登陆")
// 			c.Abort()
// 			return
// 		}
// 		data := new(UserCurrent)
// 		err = json.Unmarshal([]byte(r.Message), data)
// 		if err != nil {
// 			log.Errorf("json error: %v", err)
// 			c.String(401, "用户没有登陆")
// 			c.Abort()
// 			return
// 		}
// 		c.Set("user_info", data)
// 		c.Set("username", data.Username)
// 		c.Set("headimage", data.Avatar)
// 	}
// }

func TokenVerify(grpc pb.GreeterClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")

		userInfo, err := DouserTokenVerify(grpc, token)
		if err != nil {
			log.Errorf("grpc token verify error: %s", err.Error())
			c.String(401, "Token验证错误")
			return
		}

		c.Set("user_info", userInfo)
		c.Set("username", userInfo.Username)
		c.Set("headimage", userInfo.Avatar)

		log.Debugf("Current Login User: %s", userInfo.Username)
	}
}

func TokenVerifyByQuery(grpc pb.GreeterClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("Authorization")

		userInfo, err := DouserTokenVerify(grpc, token)
		if err != nil {
			log.Errorf("grpc token verify error: %s", err.Error())
			c.String(401, "Token验证错误")
			return
		}

		c.Set("user_info", userInfo)
		c.Set("username", userInfo.Username)
		c.Set("headimage", userInfo.Avatar)

		log.Debugf("Current Login User: %s", userInfo.Username)
	}
}

func DouserTokenVerify(grpc pb.GreeterClient, token string) (*UserCurrent, error) {
	r, err := grpc.TokenVerify2(context.Background(), &pb.HelloRequest{Token: token})
	if err != nil {
		log.Errorf("grpc token verify error: %s", err.Error())
		return nil, fmt.Errorf("grpc token verify error: %s", err.Error())
	}

	userInfo := new(UserCurrent)
	err = json.Unmarshal([]byte(r.Message), userInfo)
	if err != nil {
		log.Errorf("user info json unmarshal error: %s", err.Error())
		return nil, fmt.Errorf("user info json unmarshal error: %s", err.Error())
	}
	return userInfo, nil
}

type UserAccessToken struct {
	UserId int    `json:"user_id"`
	Token  string `json:"token"`
}

func GetUserAccessToken(grpc_client pb.GreeterClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")

		userName, ok := c.Get("username")
		if !ok {
			c.String(500, "用户名获取失败")
			return
		}

		rpcRes, err := grpc_client.GetUserToken(context.Background(), &pb.UserTokenRequest{Token: token, Name: "", Username: userName.(string)})
		if err != nil {
			log.Errorf("User Access Token fetch error: %s", err.Error())
			c.String(401, "用户访问Token获取失败")
			return
		}

		uaTokens := []*UserAccessToken{}
		err = json.Unmarshal([]byte(rpcRes.Message), &uaTokens)
		if err != nil {
			log.Errorf("User Access token unmarshal error: %s", err.Error())
			c.String(401, "用户访问Token解析失败")
			return
		}

		if len(uaTokens) > 0 {
			c.Set("token", uaTokens[0].Token)
		}
	}
}
