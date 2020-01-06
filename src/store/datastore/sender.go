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

func (db *datastore) SenderFind(repo *model.Repo, login string) (*model.Sender, error) {
	// stmt := sql.Lookup(db.driver, "sender-find-repo-login")
	// data := new(model.Sender)
	// err := meddler.QueryRow(db, data, stmt, repo.ID, login)
	// return data, err

	data := new(model.Sender)
	err := db.Where("sender_repo_id = ? AND sender_login = ?",repo.ID, login).First(&data).Error
	return data, err
}

func (db *datastore) SenderList(repo *model.Repo) ([]*model.Sender, error) {
	// stmt := sql.Lookup(db.driver, "sender-find-repo")
	// data := []*model.Sender{}
	// err := meddler.QueryAll(db, &data, stmt, repo.ID)
	// return data, err
	data := []*model.Sender{}
	err := db.Where("sender_repo_id = ?", repo.ID).Find(&data).Error
	return data, err
}

func (db *datastore) SenderCreate(sender *model.Sender) error {
	// return meddler.Insert(db, "senders", sender)
	err := db.Create(sender).Error
	return err
}

func (db *datastore) SenderUpdate(sender *model.Sender) error {
	// return meddler.Update(db, "senders", sender)
	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.Sender{}).Where("sender_id = ?",sender.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}
	return db.Save(sender).Error
}

func (db *datastore) SenderDelete(sender *model.Sender) error {
	// stmt := sql.Lookup(db.driver, "sender-delete")
	// _, err := db.Exec(stmt, sender.ID)
	// return err

	err := db.Where("sender_id = ?",sender.ID).Delete(model.Sender{}).Error
	return err
}
