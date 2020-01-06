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
	// "github.com/russross/meddler"
)

// func check_config_exist(db *datastore ,config_id int64,repo *model.Repo) (*model.Config,error) {
// 	check_exist_stmt := sql.Lookup(db.driver, "config_id-find")
// 	config_data := new(model.Config)
// 	err := meddler.QueryRow(db,config_data,check_exist_stmt,config_id,repo.ID)
// 	return config_data,err
// }

func (db *datastore) GetProjectEnvs(project string) ([]*model.ProjectEnv, error) {
	// data := []*model.ProjectEnv{}
    // err := meddler.QueryAll(db, &data, rebind(projectEnvFind), project)
	// return data, err
	data := []*model.ProjectEnv{}
	err := db.Where("env_project_id = ?", project).Find(&data).Error
	return data, err

}

func (db *datastore) UpdateProjectEnvs(project string, projectEnvs []*model.ProjectEnv) error {
    // // 开启事务
	// tx, err := db.Begin()
	// if err != nil{
	// 	return err
	// }

	// // 我们首先需要把数据库中该用户的所有环境变量删除，然后用新的赋值
    // _, err = tx.Exec(rebind(projectEnvDeleteAll), project)
	// if err != nil{
	// 	err1 := tx.Rollback()
	// 	if err1 != nil{
	// 		return err1
	// 	}
	// 	return err
	// }

	// // 循环取出key value 插入数据库
	// for _ , projectEnv := range projectEnvs {
    //     projectEnv.ID = 0
    //     projectEnv.Project = project
    //     err := meddler.Insert(tx, "project_env", projectEnv)
	// 	if err != nil {
	// 		err1 := tx.Rollback()
	// 		if err1 != nil {
	// 			return err1
	// 		}
	// 		return err
    //     }
	// }
    // // 提交事务
    // err = tx.Commit()
	// if err != nil{
	// 	err1 := tx.Rollback()
	// 	if err1 != nil{
	// 		return err1
	// 	}
	// 	return err
	// }
	// return nil

	// 开启事务
	tx := db.Begin()

	// 我们首先需要把数据库中该用户的所有环境变量删除，然后用新的赋值
	err := tx.Where("env_project_id = ?", project).Delete(model.ProjectEnv{}).Error
	if err != nil{
		tx.Rollback()
		return err
	}

	// 循环取出key value 插入数据库
	for _ , projectEnv := range projectEnvs {
		// stmt := sql.Lookup(db.driver, "pipeline_env-insert")
		// _, err = tx.Exec(stmt, config_id,env_key,env_value)

		projectEnv.ID = 0
		projectEnv.ProjectID = project
		err := tx.Create(projectEnv).Error
		if err != nil{
			tx.Rollback()
			return err
		}
	}
	// 提交事务
	tx.Commit()
	return nil
}

const projectEnvFind = `
SELECT
project_env_id,
env_project,
env_key,
env_value
FROM project_env
WHERE env_project = ?
`

const projectEnvDeleteAll = `
DELETE FROM project_env
WHERE env_project = ?
`
