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

	"github.com/coderun-top/coderun/src/model"
	// "github.com/coderun-top/coderun/src/router/middleware/session"

	"github.com/coderun-top/coderun/src/store"

	"github.com/gin-gonic/gin"
)

// GetRegistry gets the name registry from the database and writes
// to the response in json format.
func GetRegistry(c *gin.Context) {
	var (
		name        = c.Param("name")
	)
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	registry, err := Config.Services.Registries.RegistryFind(project.ID, name)
	if err != nil {
		c.String(404, "Error getting registry %q. %s", name, err)
		return
	}
	c.JSON(200, registry.Copy())
}
// GetRegistry gets the name registry from the database and writes
// to the response in json format.
func GetRegistryInside(c *gin.Context) {
	var (
		name        = c.Param("name")
	)
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	registry, err := Config.Services.Registries.RegistryFind(project.ID, name)
	if err != nil {
		c.String(404, "Error getting registry %q. %s", name, err)
		return
	}
	c.JSON(200, registry)
}

// PostRegistry persists the registry to the database.
func PostRegistry(c *gin.Context) {

	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	in := new(model.Registry)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}
	registry := &model.Registry{
		ProjectID:   project.ID,
		Name:          in.Name,
		Type:          in.Type,
		Address:       in.Address,
		Username:      in.Username,
		Password:      in.Password,
		Token:         in.Token,
		Prefix:        in.Prefix,
	}
	if err := registry.Validate(); err != nil {
		c.String(400, "Error inserting registry. %s", err)
		return
	}
	if err := Config.Services.Registries.RegistryCreate(registry); err != nil {
		c.String(500, "Error inserting registry %s. %s", registry.Name, err)
		return
	}

	reg, err := Config.Services.Registries.RegistryFind(project.ID, registry.Name)
	if err != nil {
		c.String(500, "Error find registry %s. %s", registry.Name, err)
		return
	}

	c.JSON(200, reg)
}

// PatchRegistry updates the registry in the database.
func PatchRegistry(c *gin.Context) {
	var (
		name        = c.Param("name")
	)
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	in := new(model.Registry)
	err = c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}

	registry, err := Config.Services.Registries.RegistryFind(project.ID, name)
	if err != nil {
		c.String(404, "Error getting registry %q. %s", name, err)
		return
	}
	if in.Address != "" {
		registry.Address = in.Address
	}
	if in.Username != "" {
		registry.Username = in.Username
	}
	if in.Password != "" {
		registry.Password = in.Password
	}
	if in.Token != "" {
		registry.Token = in.Token
	}
	if in.Prefix != "" {
		registry.Prefix = in.Prefix
	}

	if err := registry.Validate(); err != nil {
		c.String(400, "Error updating registry. %s", err)
		return
	}
	if err := Config.Services.Registries.RegistryUpdate(registry); err != nil {
		c.String(500, "Error updating registry %q. %s", in.Address, err)
		return
	}
	c.JSON(200, in.Copy())
}

// GetRegistryList gets the registry list from the database and writes
// to the response in json format.
func GetRegistryList(c *gin.Context) {
	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	list, err := Config.Services.Registries.RegistryList(project.ID)
	if err != nil {
		c.String(500, "Error getting registry list. %s", err)
		return
	}
	// copy the registry detail to remove the sensitive
	// password and token fields.
	for i, registry := range list {
		list[i] = registry.Copy()
	}
	c.JSON(200, list)
}

// DeleteRegistry deletes the named registry from the database.
func DeleteRegistry(c *gin.Context) {
	var (
		name  =   c.Param("name")
	)

	project, err := store.FromContext(c).GetProjectName(c.Param("projectname"))
	if err != nil {
		c.String(404, "Cannot find project. %s", err)
		return
	}

	if err = Config.Services.Registries.RegistryDelete(project.ID, name); err != nil {
		c.String(500, "Error deleting registry %q. %s", name, err)
		return
	}
	c.String(204, "")
}
