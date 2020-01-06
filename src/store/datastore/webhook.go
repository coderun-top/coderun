package datastore

import (
	"github.com/coderun-top/coderun/src/model"
	// "github.com/coderun-top/coderun/src/store/datastore/sql"
	// "github.com/russross/meddler"
)

func (db *datastore) GetWebHook(repo *model.Repo) (*model.WebHook, error) {
	// stmt := sql.Lookup(db.driver, "webhook-find-repo")
	// data := new(model.WebHook)
	// err := meddler.QueryRow(db, data, stmt, repo.ID)
	// return data, err


	data := new(model.WebHook)
	err := db.Where("repo_id = ?",repo.ID).First(&data).Error
	return data, err
}

func (db *datastore) CreateWebHook(webhook *model.WebHook) error {
	// return meddler.Insert(db, "webhook", webhook)
	err := db.Create(webhook).Error
	return err
}

func (db *datastore) UpdateWebHook(webhook *model.WebHook) error {
	// return meddler.Update(db, "webhook", webhook)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.WebHook{}).Where("id = ?",webhook.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}
	return db.Save(webhook).Error
}

func (db *datastore) GetWebHookProject(project string) (*model.WebHookProject, error) {
	// stmt := sql.Lookup(db.driver, "webhook-project-find-repo")
	// data := new(model.WebHookProject)
	// err := meddler.QueryRow(db, data, stmt, project)
	// return data, err

	data := new(model.WebHookProject)
	err := db.Where("project_id = ?", project).First(&data).Error
	return data, err
}

func (db *datastore) CreateWebHookProject(webhook *model.WebHookProject) error {
	// return meddler.Insert(db, "webhook_project", webhook)
	err := db.Create(webhook).Error
	return err
}

func (db *datastore) UpdateWebHookProject(webhook *model.WebHookProject) error {
	// return meddler.Update(db, "webhook_project", webhook)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.WebHookProject{}).Where("id = ?",webhook.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}
	return db.Save(webhook).Error
}

//func (db *datastore) DeleteUser(user *model.User) error {
//	stmt := sql.Lookup(db.driver, "user-delete")
//	_, err := db.Exec(stmt, user.ID)
//	return err
//}
