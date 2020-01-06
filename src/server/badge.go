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
	// "fmt"

	log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"github.com/coderun-top/coderun/src/model"
	// "github.com/coderun-top/coderun/src/shared/httputil"
	"github.com/coderun-top/coderun/src/store"
	"github.com/coderun-top/coderun/src/router/middleware/session"
)

var (
	badgeSuccess = `<svg xmlns="http://www.w3.org/2000/svg" width="111" height="20"><linearGradient id="a" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><rect rx="3" width="91" height="20" fill="#555"/><rect rx="3" x="57" width="54" height="20" fill="#4c1"/><rect rx="3" width="111" height="20" fill="url(#a)"/><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="27.5" y="15" fill="#010101" fill-opacity=".3">CodeRun</text><text x="27.5" y="14">CodeRun</text><text x="83" y="15" fill="#010101" fill-opacity=".3">success</text><text x="83" y="14">success</text></g></svg>`
	badgeFailure = `<svg xmlns="http://www.w3.org/2000/svg" width="111" height="20"><linearGradient id="a" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><rect rx="3" width="91" height="20" fill="#555"/><rect rx="3" x="57" width="54" height="20" fill="#4c1"/><rect rx="3" width="111" height="20" fill="url(#a)"/><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="27.5" y="15" fill="#010101" fill-opacity=".3">CodeRun</text><text x="27.5" y="14">CodeRun</text><text x="83" y="15" fill="#010101" fill-opacity=".3">failure</text><text x="83" y="14">failure</text></g></svg>`
	badgeStarted = `<svg xmlns="http://www.w3.org/2000/svg" width="111" height="20"><linearGradient id="a" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><rect rx="3" width="91" height="20" fill="#555"/><rect rx="3" x="57" width="54" height="20" fill="#4c1"/><rect rx="3" width="111" height="20" fill="url(#a)"/><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="27.5" y="15" fill="#010101" fill-opacity=".3">CodeRun</text><text x="27.5" y="14">CodeRun</text><text x="83" y="15" fill="#010101" fill-opacity=".3">started</text><text x="83" y="14">started</text></g></svg>`
	badgeError = `<svg xmlns="http://www.w3.org/2000/svg" width="111" height="20"><linearGradient id="a" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><rect rx="3" width="91" height="20" fill="#555"/><rect rx="3" x="57" width="54" height="20" fill="#4c1"/><rect rx="3" width="111" height="20" fill="url(#a)"/><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="27.5" y="15" fill="#010101" fill-opacity=".3">CodeRun</text><text x="27.5" y="14">CodeRun</text><text x="83" y="15" fill="#010101" fill-opacity=".3">error</text><text x="83" y="14">error</text></g></svg>`
	badgeNone = `<svg xmlns="http://www.w3.org/2000/svg" width="111" height="20"><linearGradient id="a" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><rect rx="3" width="91" height="20" fill="#555"/><rect rx="3" x="57" width="54" height="20" fill="#4c1"/><rect rx="3" width="111" height="20" fill="url(#a)"/><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="27.5" y="15" fill="#010101" fill-opacity=".3">CodeRun</text><text x="27.5" y="14">CodeRun</text><text x="83" y="15" fill="#010101" fill-opacity=".3">none</text><text x="83" y="14">none</text></g></svg>`
)

func GetBadge(c *gin.Context) {
	//config_id := c.Param("config_id")
	// repo, err := store.GetRepoOwnerName(c,
	// 	// c.Param("owner"),
	// 	// c.Param("name"),
	// 	full_name,
	// )

	// config_id, err := strconv.ParseInt(config_id_string, 10, 64)

	// repo, err := store.GetRepoConfigId(c,config_id)
	// if err != nil {
	// 	c.AbortWithStatus(404)
	// 	return
	// }
	repo := session.Repo(c)

	// an SVG response is always served, even when error, so
	// we can go ahead and set the content type appropriately.
	c.Writer.Header().Set("Content-Type", "image/svg+xml")

	// if no commit was found then display
	// the 'none' badge, instead of throwing
	// an error response
	branch := c.Query("branch")
	if len(branch) == 0 {
		branch = repo.Branch
	}

	log.Debugf("repo %s", repo)
	// build, err := store.GetBuildLast(c, repo, branch)
	build, err := store.GetBuildLast(c, repo)
	//build, err := store.GetBuildLastByPipeline(c, config_id)
	if err != nil {
		log.Warning(err)
		c.String(200, badgeNone)
		return
	}

	switch build.Status {
	case model.StatusSuccess:
		c.String(200, badgeSuccess)
	case model.StatusFailure:
		c.String(200, badgeFailure)
	case model.StatusError, model.StatusKilled:
		c.String(200, badgeError)
	case model.StatusPending, model.StatusRunning:
		c.String(200, badgeStarted)
	default:
		c.String(200, badgeNone)
	}
}

func GetCC(c *gin.Context) {
	//full_name := c.Query("full_name")
	//if full_name == ""{
	//	c.AbortWithStatus(404)
	//	return
	//}
	//repo, err := store.GetRepoOwnerName(c,
	//	full_name,
	//)
	//if err != nil {
	//	c.AbortWithStatus(404)
	//	return
	//}

	// repo := session.Repo(c)

	// builds, err := store.GetBuildList(c, repo, "", 1, 10)
	// if err != nil || len(builds) == 0 {
	// 	c.AbortWithStatus(404)
	// 	return
	// }

	// url := fmt.Sprintf("%s/%s/%d", httputil.GetURL(c.Request), repo.FullName, builds[0].Number)
	// cc := model.NewCC(repo, builds[0], url)
	// c.XML(200, cc)
}
