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
	"net/http"
	"errors"
	"io/ioutil"
	"os"

	"github.com/coderun-top/coderun/src/model"
	// "gitlab.com/douwa/helm/dougo/src/dougo/router/middleware/session"

	"github.com/gin-gonic/gin"
	"github.com/coderun-top/coderun/src/store"
)

// GetHelm gets the name helm from the database and writes
// to the response in json format.
func GetHelm(c *gin.Context) {
	var (
		name        = c.Param("name")
	)
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}
	helm, err := store.FromContext(c).HelmFind(project.ID, name)
	if err != nil {
		c.String(404, "Error getting helm %q. %s", name, err)
		return
	}
	c.JSON(200, helm.Copy())
}

func GetHelmInside(c *gin.Context) {
	var (
		name        = c.Param("name")
	)
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}
	helm, err := store.FromContext(c).HelmFind(project.ID, name)
	if err != nil {
		c.String(404, "Error getting helm %q. %s", name, err)
		return
	}
	c.JSON(200, helm)
}

// PostHelm persists the helm to the database.
func PostHelm(c *gin.Context) {
	var (
		projectName = c.Param("projectname")
	)

	project, err := store.FromContext(c).GetProjectName(projectName)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	in := new(model.Helm)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}
	helm := &model.Helm{
		ProjectID:   project.ID,
		Name:          in.Name,
		Address:       in.Address,
		Username:      in.Username,
		Password:      in.Password,
		Prefix:        in.Prefix,
		Type: in.Type,
	}
	if err := helm.Validate(); err != nil {
		c.String(400, "Error inserting helm. %s", err)
		return
	}
	if err := store.FromContext(c).HelmCreate(helm); err != nil {
		c.String(500, "Error inserting helm %s. %s", helm.Name, err)
		return
	}

	reg, err := store.FromContext(c).HelmFind(project.ID, helm.Name)
	if err != nil {
		c.String(500, "Error find helm %s. %s", helm.Name, err)
		return
	}

	c.JSON(200, reg)
}

// PatchHelm updates the helm in the database.
func PatchHelm(c *gin.Context) {
	var (
		projectName = c.Param("projectname")
		name        = c.Param("name")
	)

	in := new(model.Helm)
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}

	project, err := store.FromContext(c).GetProjectName(projectName)
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	helm, err := store.FromContext(c).HelmFind(project.ID, name)
	if err != nil {
		c.String(404, "Error getting helm %q. %s", name, err)
		return
	}
	if in.Address != "" {
		helm.Address = in.Address
	}
	if in.Username != "" {
		helm.Username = in.Username
	}
	if in.Password != "" {
		helm.Password = in.Password
	}
	if in.Prefix != "" {
		helm.Prefix = in.Prefix
	}

	if err := helm.Validate(); err != nil {
		c.String(400, "Error updating helm. %s", err)
		return
	}
	if err := store.FromContext(c).HelmUpdate(helm); err != nil {
		c.String(500, "Error updating helm %q. %s", in.Address, err)
		return
	}
	c.JSON(200, in.Copy())
}

// GetHelmList gets the helm list from the database and writes
// to the response in json format.
func GetHelmList(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	list, err := store.FromContext(c).HelmList(project.ID)
	if err != nil {
		c.String(500, "Error getting helm list. %s", err)
		return
	}
	// copy the helm detail to remove the sensitive
	// password and token fields.
	for i, helm := range list {
		list[i] = helm.Copy()
	}
	c.JSON(200, list)
}

// DeleteHelm deletes the named helm from the database.
func DeleteHelm(c *gin.Context) {
	var (
		name        = c.Param("name")
	)
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}
	if err := store.FromContext(c).HelmDelete(project.ID, name); err != nil {
		c.String(500, "Error deleting helm %q. %s", name, err)
		return
	}
	c.String(204, "")
}

func GetList(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	token, ok := c.Get("token")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	data, err := GetData(os.Getenv("DOUCHARTS_API")+username.(string)+"/"+c.Param("prefix")+"/charts", token.(string))
	if err != nil {
		c.String(500, "Error deleting helm %q. %s", username.(string), err)
		return
	}
	c.String(200, data)
}

func GetHelmName(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	token, ok := c.Get("token")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	data, err := GetData(os.Getenv("DOUCHARTS_API")+username.(string)+"/"+c.Param("prefix")+"/charts/"+c.Param("name"), token.(string))
	if err != nil {
		c.String(500, "Error deleting helm %q. %s", username.(string), err)
		return
	}
	c.String(200, data)
}

func GetHelmNameVersion(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	token, ok := c.Get("token")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	data, err := GetData(os.Getenv("DOUCHARTS_API")+username.(string)+"/"+c.Param("prefix")+"/charts/"+c.Param("name")+"/"+c.Param("version"), token.(string))
	if err != nil {
		c.String(500, "Error deleting helm %q. %s", username.(string), err)
		return
	}
	c.String(200, data)
}

func DeleteHelmNameVersion(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	token, ok := c.Get("token")
	if !ok {
		c.String(500, "Error fetching feed.")
		return
	}
	data, err := DelData(os.Getenv("DOUCHARTS_API")+username.(string)+"/"+c.Param("prefix")+"/charts/"+c.Param("name")+"/"+c.Param("version"), token.(string))
	if err != nil {
		c.String(500, "Error deleting helm %q. %s", username.(string), err)
		return
	}
	c.String(200, data)
}

func GetData(url string,token string) (string,error){
	client := &http.Client{}
	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)

	//增加header选项
	reqest.Header.Add("Authorization", token)

	if err != nil {
		panic(err)
	}
	//处理返回结果
	response, _ := client.Do(reqest)
	if response != nil{
		defer response.Body.Close()
	}
	// 检查状态码
	status := response.StatusCode
	if status != 200 {
		return "",errors.New("http请求失败")
	}

	bodyByte, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "",err
	}
	return string(bodyByte), nil
}

func DelData(url string,token string) (string,error){
	client := &http.Client{}
	//提交请求
	reqest, err := http.NewRequest("DELETE", url, nil)

	//增加header选项
	reqest.Header.Add("Authorization", token)

	if err != nil {
		panic(err)
	}
	//处理返回结果
	response, _ := client.Do(reqest)
	if response != nil{
		defer response.Body.Close()
	}
	// 检查状态码
	status := response.StatusCode
	if status != 200 {
		return "",errors.New("http请求失败")
	}
	return "ok", nil
}
