package datastore

import (
	"github.com/coderun-top/coderun/src/model"
	// "github.com/coderun-top/coderun/src/store/datastore/sql"
	// "github.com/russross/meddler"
)

func (db *datastore) GetEmail(repo *model.Repo) (*model.Email, error) {
	// stmt := sql.Lookup(db.driver, "email-find-repo")
	// data := new(model.Email)
	// err := meddler.QueryRow(db, data, stmt, repo.ID)
	// return data, err


	data := new(model.Email)
	err := db.Where("repo_id = ?",repo.ID).First(&data).Error
	return data, err
}

func (db *datastore) CreateEmail(email *model.Email) error {
	// return meddler.Insert(db, "email", email)
	err := db.Create(email).Error
	return err
}

func (db *datastore) UpdateEmail(email *model.Email) error {
	// return meddler.Update(db, "email", email)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.Email{}).Where("id = ?",email.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}
	return db.Save(email).Error
}

func (db *datastore) GetEmailProject(project string) (*model.EmailProject, error) {
	// stmt := sql.Lookup(db.driver, "email-project-find-repo")
	// data := new(model.EmailProject)
	// err := meddler.QueryRow(db, data, stmt, project)
	// return data, err

	data := new(model.EmailProject)
	err := db.Where("project_id = ?", project).First(&data).Error
	return data, err
}

func (db *datastore) CreateEmailProject(email *model.EmailProject) error {
	// return meddler.Insert(db, "email_project", email)
	err := db.Create(email).Error
	return err
}

func (db *datastore) UpdateEmailProject(email *model.EmailProject) error {
	// return meddler.Update(db, "email_project", email)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.EmailProject{}).Where("id = ?",email.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}
	return db.Save(email).Error
}
