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

func (db *datastore) TaskList() ([]*model.Task, error) {
	// stmt := sql.Lookup(db.driver, "task-list")
	// data := []*model.Task{}
	// err := meddler.QueryAll(db, &data, stmt)
	// return data, err

	data := []*model.Task{}
	err := db.Find(&data).Error
	return data, err
}

func (db *datastore) TaskInsert(task *model.Task) error {
	// return meddler.Insert(db, "tasks", task)
	err := db.Create(task).Error
	return err
}

func (db *datastore) TaskDelete(id string) error {
	// stmt := sql.Lookup(db.driver, "task-delete")
	// _, err := db.Exec(stmt, id)
	// return err

	err := db.Where("task_id = ?",id).Delete(model.Task{}).Error
	return err
}
