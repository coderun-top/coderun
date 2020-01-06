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

func (db *datastore) SecretFind(repo *model.Repo, name string) (*model.Secret, error) {
	// stmt := sql.Lookup(db.driver, "secret-find-repo-name")
	// data := new(model.Secret)
	// err := meddler.QueryRow(db, data, stmt, repo.ID, name)
	// return data, err
	data := new(model.Secret)
	err := db.Where("secret_repo_id = ? AND secret_name = ?", repo.ID, name).First(&data).Error
	return data, err
}

func (db *datastore) SecretList(repo *model.Repo) ([]*model.Secret, error) {
	// stmt := sql.Lookup(db.driver, "secret-find-repo")
	// data := []*model.Secret{}
	// err := meddler.QueryAll(db, &data, stmt, repo.ID)
	// return data, err
	data := []*model.Secret{}
	err := db.Where("secret_repo_id = ?", repo.ID).Find(&data).Error
	return data, err
}

func (db *datastore) SecretCreate(secret *model.Secret) error {
	// return meddler.Insert(db, "secrets", secret)
	err := db.Create(secret).Error
	return err
}

func (db *datastore) SecretUpdate(secret *model.Secret) error {
	// return meddler.Update(db, "secrets", secret)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.Secret{}).Where("secret_id = ?",secret.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}
	return db.Save(secret).Error
}

func (db *datastore) SecretDelete(secret *model.Secret) error {
	// stmt := sql.Lookup(db.driver, "secret-delete")
	// _, err := db.Exec(stmt, secret.ID)
	// return err

	err := db.Where("secret_id = ?",secret.ID).Delete(model.Secret{}).Error
	return err
}
