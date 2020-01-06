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
	// "gitlab.com/douwa/helm/dougo/src/dougo/store/datastore/sql"
	// "github.com/russross/meddler"
)

func (db *datastore) HelmFind(projectName, name string) (*model.Helm, error) {
	// stmt := sql.Lookup(db.driver, "helm-find-name-project")
	// data := new(model.Helm)
	// err := meddler.QueryRow(db, data, stmt, projectName, name)
	// return data, err

	data := new(model.Helm)
	err := db.Where("helm_project_id = ? AND helm_name = ?",projectName, name).First(&data).Error
	return data, err

}

func (db *datastore) HelmList(projectName string) ([]*model.Helm, error) {
	// stmt := sql.Lookup(db.driver, "helm-find-project")
	// data := []*model.Helm{}
	// err := meddler.QueryAll(db, &data, stmt, projectName)
	// return data, err

	data := []*model.Helm{}
	err := db.Where("helm_project_id = ?", projectName).Find(&data).Error
	return data, err
}

func (db *datastore) HelmCreate(helm *model.Helm) error {
	// return meddler.Insert(db, "helm", helm)
	err := db.Create(helm).Error
	return err
}

func (db *datastore) HelmUpdate(helm *model.Helm) error {
	// return meddler.Update(db, "helm", helm)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.Helm{}).Where("helm_id = ?",helm.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}

	return db.Save(helm).Error
}

func (db *datastore) HelmDeleteProject(projectName string) error {
	// stmt := sql.Lookup(db.driver, "helm-delete-project")
	// _, err := db.Exec(stmt, projectName)
	// return err

	err := db.Where("helm_project_id = ?",projectName).Delete(model.Helm{}).Error
	return err
}

func (db *datastore) HelmDelete(projectName, name string) error {
	// stmt := sql.Lookup(db.driver, "helm-delete")
	// _, err := db.Exec(stmt, projectName, name)
	// return err

	err := db.Where("helm_project_id = ? AND helm_name = ?",projectName, name).Delete(model.Helm{}).Error
	return err
}
