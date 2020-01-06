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
	// "errors"
)

func (db *datastore) GetUser(id int64) (*model.User, error) {
	// var usr = new(model.User)
	// var err = meddler.Load(db, "users", usr, id)
	// return usr, err

	var usr = new(model.User)
	err := db.Where("user_id = ?",id).First(&usr).Error
	return usr,err
}

func (db *datastore) GetUserLogin(login string) (*model.User, error) {
	// stmt := sql.Lookup(db.driver, "user-find-login")
	// data := new(model.User)
	// err := meddler.QueryRow(db, data, stmt, login)
	// return data, err

	data := new(model.User)
	err := db.Where("user_login = ?",login).First(&data).Error
	return data,err
}

func (db *datastore) GetUserName(login string) (*model.User, error) {
	// stmt := sql.Lookup(db.driver, "user-find-name")
	// data := new(model.User)
	// err := meddler.QueryRow(db, data, stmt, login)
	// return data, err

	data := new(model.User)
	err := db.Where("user_token_name = ?",login).First(&data).Error
	return data,err
}

// func (db *datastore) GetUserUserName2(name string,login string) (*model.User, error) {
// 	data := new(model.User)
// 	err := db.Where("user_token_name = ? and oauth = 1", name).Order("user_id desc").Preload("Project").Limit(1).First(&data).Error
// 	return data,err
// }

func (db *datastore) GetUserUserName(projectName string) ([]*model.User, error) {
	data := []*model.User{}
	// err := db.Where("oauth=0 and user_project_id = ?", projectName).Order("oauth desc, user_token_name").Preload("Project").Find(&data).Error
	err := db.Where("user_project_id = ?", projectName).Order("oauth desc, user_token_name").Preload("Project").Find(&data).Error
	return data, err
}

func (db *datastore) GetUserUserNameOauth(projectName string, oauth bool) ([]*model.User, error) {
	// stmt := sql.Lookup(db.driver, "user-find-user-name-oauth")
	// data := []*model.User{}
	// err := meddler.QueryAll(db, &data, stmt, projectName, oauth)
	// return data, err

	data := []*model.User{}
	// err := db.Where("user_project_id = ? AND oauth = ?",projectName,oauth).Order("user_token_name").Preload("Project").Find(&data).Error
	err := db.Where("user_project_id = ?",projectName).Order("user_token_name").Preload("Project").Find(&data).Error
	return data, err
}

func (db *datastore) GetUserNameToken(projectName, token string, oauth bool) (*model.User, error) {
	// data := new(model.User)
	// stmt := sql.Lookup(db.driver, "user-find-user-name-token")
	// err := meddler.QueryRow(db, data, stmt, projectName, token)
	// return data, err

	data := new(model.User)
	var err error
	// if oauth == true {
	// 	err = db.Where("user_token_name = ? AND oauth = ?", token, oauth).First(&data).Error
	// } else {
	// 	err = db.Where("user_project_id = ? AND user_token_name = ?",projectName,token).First(&data).Error
	// }
	err = db.Where("user_project_id = ? AND user_token_name = ?",projectName,token).First(&data).Error
	return data, err
}

func (db *datastore) GetUserNameTokenOauth(token string, oauth bool) (*model.User, error) {
	data := new(model.User)
	// err := db.Where("user_token_name = ? AND oauth = ?", token, oauth).First(&data).Error
	err := db.Where("user_token_name = ?", token).First(&data).Error
	return data, err
}

func (db *datastore) GetUserNameTokenOauth2(projectName, token string, oauth bool) (*model.User, error) {
	data := new(model.User)
	// err := db.Where("user_project_id and user_token_name = ? AND oauth = ?", projectName, token, oauth).First(&data).Error
	err := db.Where("user_project_id = ? and user_token_name = ?", projectName, token).First(&data).Error
	return data, err
}

func (db *datastore) GetUserByRepoId(repoId int64) (*model.User, error) {
	// stmt := sql.Lookup(db.driver, "user-find-repo-id")
	// data := new(model.User)
	// err := meddler.QueryRow(db, data, stmt, repoId)
	// return data, err

	data := new(model.User)
	err := db.Joins("INNER JOIN repos ON repos.repo_user_id = users.user_id").Where("repos.repo_id = ?", repoId).First(&data).Error
	return data, err
}

func (db *datastore) GetUserList() ([]*model.User, error) {
	// stmt := sql.Lookup(db.driver, "user-find")
	// data := []*model.User{}
	// err := meddler.QueryAll(db, &data, stmt)
	// return data, err
	data := []*model.User{}
	err := db.Order("user_login").Order("user_login").Find(&data).Error
	return data,err
}

func (db *datastore) GetUserCount() (count int, err error) {
	// err = db.QueryRow(
	// 	sql.Lookup(db.driver, "count-users"),
	// ).Scan(&count)
	// return
	err = db.Model(&model.User{}).Count(&count).Error
	return
}

func (db *datastore) CreateUser(user *model.User) error {
	// return meddler.Insert(db, "users", user)
	err := db.Create(user).Error
	return err
}

func (db *datastore) UpdateUser(user *model.User) error {
	// return meddler.Update(db, "users", user)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.User{}).Where("user_id = ?",user.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}

	return db.Save(user).Error
}

func (db *datastore) DeleteUser(user *model.User) error {
	// stmt := sql.Lookup(db.driver, "user-delete")
	// _, err := db.Exec(stmt, user.ID)
	// return err

	// 必须确保主键有值，否则会将数据全部删除（大坑啊）
	// if user.ID == 0{
	// 	return errors.New("要删除的对象不存在，请传递ID")
	// }
	// err := db.Delete(user).Error
	// return err
	err := db.Where("user_id = ?",user.ID).Delete(model.User{}).Error
	return err
}

func (db *datastore) UserFeed(user *model.User) ([]*model.Feed, error) {
	// stmt := sql.Lookup(db.driver, "feed")
	// data := []*model.Feed{}
	// err := meddler.QueryAll(db, &data, stmt, user.ID)
	// return data, err

	data := []*model.Feed{}

	err := db.Table("repos").Select("repos.repo_owner,repos.repo_name,repos.repo_full_name,builds.build_number,builds.build_event,builds.build_status,builds.build_created,builds.build_started,builds.build_finished,builds.build_commit,builds.build_branch,builds.build_ref,builds.build_refspec,builds.build_remote,builds.build_title,builds.build_message,builds.build_author,builds.build_email,builds.build_avatar").
			Joins("INNER JOIN perms  ON perms.perm_repo_id   = repos.repo_id").
			Joins("INNER JOIN builds ON builds.build_repo_id = repos.repo_id").
			Where("perms.perm_user_id = ?", user.ID).Order("builds.build_id desc").Scan(&data).Error
	
	return data,err
}
