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
	"github.com/coderun-top/coderun/src/utils"
)

func (db *datastore) GetStarRepo(username string, page, pagesize int) ([]*model.Repo, error) {

	data := []*model.Repo{}
	err := db.Joins("INNER JOIN projects on projects.project_id = repos.repo_project_id").Joins("INNER JOIN star ON star.repo_id = repos.repo_id").
		Where("star.username = ?", username).Limit(pagesize).Offset((page - 1) * pagesize).Preload("User").Preload("Project").Find(&data).Error
	return data, err
}

func (db *datastore) GetStarRepoCount(username string) (count int, err error) {
	err = db.Table("repos").Joins("INNER JOIN star ON star.repo_id = repos.repo_id").
		Where("star.username = ?", username).Count(&count).Error
	return
}

func (db *datastore) GetRepo(id int64) (*model.Repo, error) {
	var repo = new(model.Repo)
	err := db.Where("repo_id = ?", id).Preload("Project").Preload("User").First(&repo).Error
	return repo, err
}

func (db *datastore) GetRepoFullName(user *model.User, repoFullName string) (*model.Repo, error) {
	// stmt := sql.Lookup(db.driver, "repo-find-fullname")
	// var repo = new(model.Repo)
	// var err = meddler.QueryRow(db, repo, stmt, user.ID, repoFullName)
	// return repo, err
	var repo = new(model.Repo)
	err := db.Where(" repo_user_id = ? AND repo_full_name = ?", user.ID, repoFullName).Preload("User").First(&repo).Error
	return repo, err

}

func (db *datastore) GetRepoName(name string) (*model.Repo, error) {
	// var repo = new(model.Repo)
	// var err = meddler.QueryRow(db, repo, rebind(repoNameQuery), name)
	// return repo, err

	var repo = new(model.Repo)
	err := db.Where("repo_full_name = ?", name).Preload("User").First(&repo).Error
	return repo, err
}

func (db *datastore) GetRepoCount() (count int, err error) {
	// err = db.QueryRow(
	// 	sql.Lookup(db.driver, "count-repos"),
	// ).Scan(&count)
	err = db.Model(&model.Repo{}).Where("repo_active = ?", true).Count(&count).Error
	return
}

func (db *datastore) CreateRepo(repo *model.Repo) error {
	// return meddler.Insert(db, repoTable, repo)
	err := db.Create(repo).Error
	return err
}

func (db *datastore) UpdateRepo(repo *model.Repo) error {
	// return meddler.Update(db, repoTable, repo)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.Repo{}).Where("repo_id = ?", repo.ID).Count(&count).Error
	if err != nil || count == 0 {
		return nil
	}

	return db.Save(repo).Error
}

func (db *datastore) DeleteRepo(repo *model.Repo) error {
	// stmt := sql.Lookup(db.driver, "repo-delete")
	// logrus.Debugf("id:%s",repo.ID)
	// _, err := db.Exec(stmt, repo.ID)
	// return err
	// 必须确保主键有值，否则会将数据全部删除（大坑啊）
	// if repo.ID == 0{
	// 	return errors.New("要删除的对象不存在，请传递ID")
	// }
	// err := db.Delete(repo).Error
	// return err
	err := db.Where("repo_id = ?", repo.ID).Delete(model.Repo{}).Error
	return err
}

func (db *datastore) RepoList(user *model.User) ([]*model.Repo, error) {
	// stmt := sql.Lookup(db.driver, "repo-find-user")
	// data := []*model.Repo{}
	// err := meddler.QueryAll(db, &data, stmt, user.ID)
	// return data, err

	data := []*model.Repo{}
	err := db.Where("repo_user_id = ?", user.ID).Preload("User").Find(&data).Error
	return data, err
}

func (db *datastore) RepoListProject(projectName string) ([]*model.Repo, error) {
	// stmt := sql.Lookup(db.driver, "repo-find-project")
	// data := []*model.Repo{}
	// err := meddler.QueryAll(db, &data, stmt, projectName)
	// return data, err

	data := []*model.Repo{}
	err := db.Joins("INNER JOIN projects on projects.project_id = repos.repo_project_id").Where("projects.project_name = ?", projectName).Preload("User").Find(&data).Error
	return data, err
}

func (db *datastore) RepoListProjectPage(projectName string, search string, page, pagesize int) ([]*model.Repo, error) {
	data := []*model.Repo{}
	err := db.Joins("INNER JOIN projects on projects.project_id = repos.repo_project_id").
		Where("projects.project_name = ? AND repos.repo_full_name like ?", projectName, "%"+search+"%").
		Limit(pagesize).Offset((page - 1) * pagesize).Preload("User").Preload("Project").Find(&data).Error
	return data, err
}

func (db *datastore) RepoListProjectCount(projectName string, search string) (count int, err error) {
	// err = db.QueryRow(
	// 	sql.Lookup(db.driver, "count-repo-find-project"),
	// projectName, "%"+search+"%").Scan(&count)
	// return
	err = db.Table("repos").
		Joins("INNER JOIN projects on projects.project_id = repos.repo_project_id").
		Where("projects.project_name = ? AND repos.repo_full_name like ?", projectName, "%"+search+"%").
		Count(&count).Error

	return
}

func (db *datastore) RepoListLatest(user *model.User) ([]*model.Feed, error) {
	// stmt := sql.Lookup(db.driver, "feed-latest-build")
	// data := []*model.Feed{}
	// err := meddler.QueryAll(db, &data, stmt, user.ID)
	// return data, err

	// stmt := sql.Lookup(db.driver, "feed-latest-build")
	// data := []*model.Feed{}
	// err := db.Raw(stmt, user.ID).Scan(&data).Error
	// return data, err

	data := []*model.Feed{}
	err := db.Table("repos").Select("repos.repo_owner,repos.repo_name,repos.repo_full_name,builds.*").
		Joins("LEFT OUTER JOIN builds ON build_id = (SELECT build_id FROM builds WHERE builds.build_repo_id = repos.repo_id ORDER BY build_id DESC LIMIT 1)").
		Joins("INNER JOIN perms ON perms.perm_repo_id = repos.repo_id").
		Where("perms.perm_user_id = ? AND repos.repo_active = true", 1).Scan(&data).Error
	return data, err
}

func (db *datastore) RepoBatch(repos []*model.Repo) error {
	// stmt := sql.Lookup(db.driver, "repo-insert-ignore")
	// for _, repo := range repos {
	// 	_, err := db.Exec(stmt,
	// 		repo.UserID,
	// 		repo.Owner,
	// 		repo.Name,
	// 		repo.FullName,
	// 		repo.Avatar,
	// 		repo.Link,
	// 		repo.Clone,
	// 		repo.Branch,
	// 		repo.Timeout,
	// 		repo.IsPrivate,
	// 		repo.IsTrusted,
	// 		repo.IsActive,
	// 		repo.AllowPull,
	// 		repo.AllowPush,
	// 		repo.AllowDeploy,
	// 		repo.AllowTag,
	// 		repo.Hash,
	// 		repo.Kind,
	// 		repo.Config,
	// 		repo.IsGated,
	// 		repo.Visibility,
	// 		repo.Counter,
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// return nil

	// 开启事务
	// Omit 插入的时候忽略repo_id
	tx := db.Begin()
	for _, repo := range repos {
		err := tx.Omit("repo_id").Create(&repo).Error

		if err != nil {
			tx.Rollback()
			return nil
		}
	}
	tx.Commit()
	return nil
}

func (db *datastore) GetRepoConfigId(config_id int64) (*model.Repo, error) {
	var config = new(model.Config)
	var repo = new(model.Repo)

	_ = db.Where("config_id = ?", config_id).First(&config).Error
	err := db.Where("repo_id = ?", config.RepoID).Preload("User").First(&repo).Error

	return repo, err

}

func (db *datastore) CreateRepoConfigs(repo *model.Repo, config *model.Config) error {
	tx := db.Begin()
	if err := tx.Create(repo).Error; err != nil {
		tx.Rollback()
		return err
	}

	config.ID = utils.GeneratorId()
	config.RepoID = repo.ID
	if err := tx.Create(config).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

const repoTable = "repos"

const repoNameQuery = `
SELECT *
FROM repos
WHERE repo_full_name = ?
LIMIT 1;
`

const repoDeleteStmt = `
DELETE FROM repos
WHERE repo_id = ?
`
