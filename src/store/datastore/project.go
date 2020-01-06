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

package datastore

import (
	"github.com/coderun-top/coderun/src/model"
)

func (db *datastore) GetProject(id string) (*model.Project, error) {
	var usr = new(model.Project)
	err := db.Where("project_id = ?", id).First(&usr).Error
	return usr,err
}


func (db *datastore) GetProjectName(name string) (*model.Project, error) {
	data := new(model.Project)
	err := db.Where("project_name = ?", name).First(&data).Error
	return data,err
}


func (db *datastore) CreateProject(project *model.Project) error {
	err := db.Create(project).Error
	return err
}

func (db *datastore) DeleteProject(project *model.Project) error {
	err := db.Where("project_id = ?",project.ID).Delete(model.Project{}).Error
	return err
}
