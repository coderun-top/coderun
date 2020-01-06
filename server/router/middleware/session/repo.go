package session

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/coderun-top/coderun/server/model"
	"github.com/coderun-top/coderun/server/remote"
	"github.com/coderun-top/coderun/server/store"

	log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func Repo(c *gin.Context) *model.Repo {
	v, ok := c.Get("repo")
	if !ok {
		return nil
	}
	r, ok := v.(*model.Repo)
	if !ok {
		return nil
	}
	return r
}

func Repos(c *gin.Context) []*model.RepoLite {
	v, ok := c.Get("repos")
	if !ok {
		return nil
	}
	r, ok := v.([]*model.RepoLite)
	if !ok {
		return nil
	}
	return r
}

func SetRepo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			//owner = c.Param("owner")
			//user  = User(c)
			id_str = c.Param("id")
		)
		//log.Debugf("Cannot find repository %s",
		//	name,
		//)
		id, err := strconv.ParseInt(id_str, 10, 64)

		repo, err := store.GetRepo(c, id)
		if err == nil {
			c.Set("repo", repo)
			c.Next()
			return
		}

		// debugging
		log.Debugf("Cannot find repository %s. %s",
			id_str,
			err.Error(),
		)

		//if user != nil {
		//	c.AbortWithStatus(http.StatusNotFound)
		//} else {
		//	c.AbortWithStatus(http.StatusUnauthorized)
		//}
	}
}

func Perm(c *gin.Context) *model.Perm {
	v, ok := c.Get("perm")
	if !ok {
		return nil
	}
	u, ok := v.(*model.Perm)
	if !ok {
		return nil
	}
	return u
}

func SetPerm() gin.HandlerFunc {

	return func(c *gin.Context) {
		user := User(c)
		repo := Repo(c)
		perm := new(model.Perm)

		switch {
		case user != nil:
			var err error
			perm, err = store.FromContext(c).PermFind(user, repo)
			if err != nil {
				log.Errorf("Error fetching permission for %s %s. %s",
					user.Login, repo.FullName, err)
			}
			if time.Unix(perm.Synced, 0).Add(time.Hour).Before(time.Now()) {
				perm, err = remote.FromContext(c).Perm(user, repo.Owner, repo.Name)
				if err == nil {
					log.Debugf("Synced user permission for %s %s", user.Login, repo.FullName)
					perm.Repo = repo.FullName
					perm.UserID = user.ID
					perm.Synced = time.Now().Unix()
					store.FromContext(c).PermUpsert(perm)
				}
			}
		}

		if perm == nil {
			perm = new(model.Perm)
		}

		if user != nil && user.Admin {
			perm.Pull = true
			perm.Push = true
			perm.Admin = true
		}

		switch {
		case repo.Visibility == model.VisibilityPublic:
			perm.Pull = true
		case repo.Visibility == model.VisibilityInternal && user != nil:
			perm.Pull = true
		}

		if user != nil {
			log.Debugf("%s granted %+v permission to %s",
				user.Login, perm, repo.FullName)

		} else {
			log.Debugf("Guest granted %+v to %s", perm, repo.FullName)
		}

		c.Set("perm", perm)
		c.Next()
	}
}

func MustPull(c *gin.Context) {
	user := User(c)
	perm := Perm(c)

	if perm.Pull {
		c.Next()
		return
	}

	// debugging
	if user != nil {
		c.AbortWithStatus(http.StatusNotFound)
		log.Debugf("User %s denied read access to %s",
			user.Login, c.Request.URL.Path)
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Debugf("Guest denied read access to %s %s",
			c.Request.Method,
			c.Request.URL.Path,
		)
	}
}

func MustPush(c *gin.Context) {
	c.Next()
	return
	user := User(c)
	perm := Perm(c)

	// if the user has push access, immediately proceed
	// the middleware execution chain.
	if perm.Push {
		c.Next()
		return
	}

	// debugging
	if user != nil {
		c.AbortWithStatus(http.StatusNotFound)
		log.Debugf("User %s denied write access to %s",
			user.Login, c.Request.URL.Path)

	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Debugf("Guest denied write access to %s %s",
			c.Request.Method,
			c.Request.URL.Path,
		)
	}
}

//func SetRepoAndUser() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		var (
//			id_str  = c.Param("id")
//		)
//		id, err := strconv.ParseInt(id_str, 10, 64)
//
//		repo, err := store.GetRepo(c, id)
//		if err == nil {
//			user, err := store.GetUser(c, repo.UserID)
//			if err == nil {
//				c.Set("user", user)
//				c.Set("repo", repo)
//				c.Next()
//				return
//			}
//		}
//
//		log.Debugf("cannot find repository %s. %s",
//			id_str,
//			err.Error(),
//		)
//
//	}
//}

func SetRepoAndUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userName, ok := c.Get("username")
		if !ok {
			c.String(500, "Login UserName get error")
			c.Abort()
			return
		}
		// log.Debugf("login user: %s", v)

		var (
			tokenName  = c.Query("token_name")
			repoName = c.Query("repo_name")
		)

		if tokenName == "" {
			c.String(500, "Gitname not null")
			c.Abort()
			return
		}
		if repoName == "" {
			c.String(500, "Repo not null")
			c.Abort()
			return
		}

		projectName := c.Param("projectname")
		project, err := store.FromContext(c).GetProjectName(projectName)
		if err != nil {
			c.String(404, "Cannot find project. %s", err)
			return
		}

		oauth2 := false
		if userName == tokenName {
			oauth2 = true
		}

		user, err := store.GetUserNameToken(c, project.ID, tokenName, oauth2)
		if err != nil {
			log.Errorf("Cannot find user %s", err.Error())
			c.String(500, fmt.Sprintf("Cannot find user: %s", err.Error()))
			c.Abort()
			return
		}
		//new_reponame := strings.Replace(reponame, "%2F", "/", -1)
		repo, err := store.FromContext(c).GetRepoFullName(user, repoName)
		if err != nil {
			log.Errorf("Cannot find repository %s", err.Error())
			c.String(500, fmt.Sprintf("Cannot find repository: %s", err.Error()))
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Set("repo", repo)
		c.Next()
		return
	}
}
