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

func (db *datastore) StarFind(star_id int64) (*model.Star, error) {
	data := new(model.Star)
	err := db.Where("star_id = ?", star_id).First(&data).Error
	return data, err
}

func (db *datastore) StarFind2(repo_id int64, name string) (*model.Star, error) {
	data := new(model.Star)
	err := db.Where("username = ? AND repo_id = ?", name, repo_id).First(&data).Error
	return data, err
}

func (db *datastore) StarList(name string) ([]*model.Star, error) {
	data := []*model.Star{}
	err := db.Where("username = ?", name).Find(&data).Error
	return data, err
}

func (db *datastore) StarCreate(star *model.Star) error {
	err := db.Create(star).Error
	return err
}


func (db *datastore) StarDelete(star *model.Star) error {
	err := db.Where("star_id = ?", star.ID).Delete(model.Star{}).Error
	return err
}

