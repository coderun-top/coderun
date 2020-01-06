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
	// gosql "database/sql"

	"github.com/coderun-top/coderun/src/model"
	// "github.com/coderun-top/coderun/src/store/datastore/sql"
	// "github.com/russross/meddler"
	"github.com/coderun-top/coderun/src/utils"
)

func (db *datastore) ConfigLoad(id string) (*model.Config, error) {
	// // stmt := sql.Lookup(db.driver, "config-find-id")
	// // conf := new(model.Config)
	// // err := meddler.QueryRow(db, conf, stmt, id)
	// // return conf, err
	// var conf = new(model.Config)
	// var err = meddler.Load(db, "config", conf, id)
	// return conf, err

	var conf = new(model.Config)
	err := db.Where("config_id = ?",id).First(&conf).Error
	return conf, err
}

func (db *datastore) ConfigFind(repo *model.Repo, hash string) (*model.Config, error) {
	// stmt := sql.Lookup(db.driver, "config-find-repo-hash")
	// conf := new(model.Config)
	// err := meddler.QueryRow(db, conf, stmt, repo.ID, hash)
	// return conf, err

	conf := new(model.Config)
	err := db.Where("config_repo_id = ? AND config_hash = ?",repo.ID, hash).First(&conf).Error
	return conf, err
}

func (db *datastore) ConfigFindApproved(config *model.Config) (bool, error) {
	// var dest int64
	// stmt := sql.Lookup(db.driver, "config-find-approved")
	// err := db.DB.QueryRow(stmt, config.RepoID, config.ID).Scan(&dest)
	// if err == gosql.ErrNoRows {
	// 	return false, nil
	// } else if err != nil {
	// 	return false, err
	// }
	// return true, nil


	var builds = []*model.Build{}
	err := db.Table("builds").Select("build_id").
		Where("build_repo_id = ? AND build_config_id = ?",config.RepoID, config.ID).
		Not("build_status", []string{"blocked", "pending"}).Find(&builds).Error
	
	if err != nil{
		return false,err
	} else if 0 == len(builds) {
		return false,nil
	}
	return true,nil
	// if err == gosql.ErrNoRows {
	// 	return false, nil
	// } else if err != nil {
	// 	return false, err
	// }
	// return true, nil
}

func (db *datastore) ConfigCreate(config *model.Config) error {
	// return meddler.Insert(db, "config", config)
	config.ID = utils.GeneratorId()
	err := db.Create(config).Error
	return err
}

func (db *datastore) ConfigUpdate(config *model.Config) error {
	// return meddler.Update(db, "config", config)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.Config{}).Where("config_id = ?",config.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}

	return db.Save(config).Error
}

func (db *datastore) ConfigRepoFind(repo *model.Repo) ([]*model.Config, error) {
	data := []*model.Config{}
	err := db.Where("config_repo_id = ?",repo.ID).Find(&data).Error
	return data, err

}

func (db *datastore) ConfigDeleteRepo(repo *model.Repo) error {
	// stmt := sql.Lookup(db.driver, "config-delete-repo")
	// _, err := db.Exec(stmt, repo.ID)
	// return err

	// 必须确保主键有值，否则会将数据全部删除（大坑啊）
	// if repo.ID == 0{
	// 	return errors.New("要删除的对象不存在，请传递ID")
	// }
	// err := db.Delete(repo).Error
	err := db.Where("repo_id = ?",repo.ID).Delete(model.Repo{}).Error
	return err
}
