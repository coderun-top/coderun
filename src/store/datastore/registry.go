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
	// "github.com/coderun-top/coderun/src/store/datastore/sql"
	// "github.com/russross/meddler"
)

func (db *datastore) RegistryFind(projectName, name string) (*model.Registry, error) {
	// stmt := sql.Lookup(db.driver, "registry-find-name-project")
	// data := new(model.Registry)
	// err := meddler.QueryRow(db, data, stmt, projectName, name)
	// return data, err

	data := new(model.Registry)
	err := db.Where("registry_project_id = ? AND registry_name = ?",projectName, name).First(&data).Error
	return data, err

}

func (db *datastore) RegistryList(projectName string) ([]*model.Registry, error) {
	// stmt := sql.Lookup(db.driver, "registry-find-project")
	// data := []*model.Registry{}
	// err := meddler.QueryAll(db, &data, stmt, projectName)
	// return data, err

	data := []*model.Registry{}
	err := db.Where("registry_project_id = ?", projectName).Find(&data).Error
	return data, err
}

func (db *datastore) RegistryCreate(registry *model.Registry) error {
	// return meddler.Insert(db, "registry", registry)
	err := db.Create(registry).Error
	return err
}

func (db *datastore) RegistryUpdate(registry *model.Registry) error {
	// return meddler.Update(db, "registry", registry)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.Registry{}).Where("registry_id = ?",registry.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}

	return db.Save(registry).Error
}

func (db *datastore) RegistryDeleteProject(projectName string) error {
	// stmt := sql.Lookup(db.driver, "registry-delete-project")
	// _, err := db.Exec(stmt, projectName)
	// return err

	err := db.Where("registry_project_id = ?",projectName).Delete(model.Registry{}).Error
	return err
}

func (db *datastore) RegistryDelete(projectName, name string) error {
	// stmt := sql.Lookup(db.driver, "registry-delete")
	// _, err := db.Exec(stmt, projectName, name)
	// return err

	err := db.Where("registry_project_id = ? AND registry_name = ?",projectName, name).Delete(model.Registry{}).Error
	return err
}
