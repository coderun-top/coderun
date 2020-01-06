package session

import (
	// "os"
	"context"
	"github.com/coderun-top/coderun/server/model"
	"strconv"
	// "github.com/coderun-top/coderun/server/shared/token"
	"github.com/gin-gonic/gin"
	"github.com/coderun-top/coderun/server/store"

	"encoding/json"
	"github.com/sirupsen/logrus"
	pb "github.com/coderun-top/coderun/server/grpc/user"
)

func User(c *gin.Context) *model.User {
	v, ok := c.Get("user")
	if !ok {
		return nil
	}
	u, ok := v.(*model.User)
	if !ok {
		return nil
	}
	return u
}

// func Token(c *gin.Context) *token.Token {
// 	v, ok := c.Get("token")
// 	if !ok {
// 		return nil
// 	}
// 	u, ok := v.(*token.Token)
// 	if !ok {
// 		return nil
// 	}
// 	return u
// }

func MustAdmin(grpc_client pb.GreeterClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := GetRole(c, grpc_client)
		if err != nil {
			c.AbortWithError(404, err)
		}

		if role == "admin" {
			c.Next()
		} else {
			c.String(403, "User not authorized for admin")
			c.Abort()
		}
	}
}

func MustRepoAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := User(c)
		//perm := Perm(c)
		switch {
		case user == nil:
			c.String(403, "User not authorized")
			c.Abort()
		//case perm.Admin == false:
		//	c.String(403, "User not authorized")
		//	c.Abort()
		default:
			c.Next()
		}
	}
}

func MustUser(grpc_client pb.GreeterClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := GetRole(c, grpc_client)
		if err != nil {
			c.AbortWithError(404, err)
		}

		if role == "admin" || role == "developer" {
			c.Next()
		} else {
			c.String(403, "User not authorized for admin or developer")
			c.Abort()
		}
	}
}

func SetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// oauth, _ := strconv.ParseBool(c.Query("oauth"))
		// user, err := store.GetUserNameToken(c, c.Param("projectname"), c.Param("tokenname"), oauth)
		user_id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			logrus.Debugf("string to int64 err. %s", err)
			c.AbortWithError(400, err)
			return
		}
		user, err := store.GetUser(c, user_id)
		if err == nil {
			c.Set("user", user)
		}
		c.Next()
	}
}

type UserName struct {
	Username string `json:"username"`
}

func GetRole(c *gin.Context, grpc_client pb.GreeterClient) (string, error) {
	user := User(c)
	if user == nil {
		return "", nil
	}

	// 修改成使用参数获取ProjectName而不是user下
	// 拿project_name跟token去douser去验证，确定是否有权限操作
	// project, err := store.FromContext(c).GetProject(user.ProjectID)
	// if err != nil {
	// 	logrus.Errorf("Cannot find project. %s", err)
	// 	return "", err
	// }
	projectName := c.Param("projectname")

	token := c.Request.Header.Get("Authorization")
	r, err := grpc_client.GetRole(context.Background(), &pb.ProjectRequest{Token: token, Projectname: projectName})
	if err != nil {
		logrus.Errorf("GetRole error: %v", err)
		return "", err
	}

	var data []int
	err = json.Unmarshal([]byte(r.Message), &data)
	if err != nil {
		return "", err
	}

	for _, role := range data {
		if role == 1 {
			return "admin", nil
		} else if role == 2 {
			return "developer", nil
		}
		break
	}
	return "", nil
}

// func ProjectJurisdiction(user_role int) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		user := User(c)
// 		if user == nil{
// 			// 如果用户不存在则退出
// 			c.String(401, "User not authorized")
// 			c.Abort()
// 		}else{
// 			// 拿project_name跟token去douser去验证，确定是否有权限操作
// 			project_name := user.Name
// 			log.Debug(project_name)
// 			token := c.Request.Header.Get("Authorization")
//             url := os.Getenv("DOUSER_ROLE_URL") + "?name=" + project_name
// 			data_byte,err := get_data(url,token)
// 			// 如果发生错误则报错
// 			if err != nil{
// 				log.Error(err)
// 				c.String(500, "内部错误")
// 				c.Abort()
// 			}
// 			var data []int
// 			err = json.Unmarshal(data_byte,&data)
// 			if err != nil{
// 				log.Error(err)
// 				c.String(500, "内部错误")
// 				c.Abort()
// 			}
// 			if len(data) == 0{
// 				log.Error(err)
// 				c.String(401, "无任何权限")
// 				c.Abort()
// 			}
// 			role := data[0]
// 			if user_role == 1{
// 				if role != 1{
// 					c.String(401, "User not authorized")
// 					c.Abort()
// 				}
// 			}else if user_role == 2{
// 				if role != 1 && role != 2{
// 					c.String(401, "User not authorized")
// 					c.Abort()
// 				}
// 			}
// 		}
// 		c.Next()
// 	}
// }
