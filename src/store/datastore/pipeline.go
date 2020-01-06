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

// func check_config_exist(db *datastore ,config_id int64,repo *model.Repo) (*model.Config,error) {
// 	check_exist_stmt := sql.Lookup(db.driver, "config_id-find")
// 	config_data := new(model.Config)
// 	err := meddler.QueryRow(db,config_data,check_exist_stmt,config_id,repo.ID)
// 	return config_data,err
// }

// var db *gorm.DB


func (db *datastore) GetPipelineEnvs(configID string) ([]*model.PipelineEnv, error) {
	// config_data,err := check_config_exist(db,config_id,repo)
	// if err != nil{
	// 	return nil,err
	// }
	// if config_data == nil {
	// 	return nil,errors.New("config_id不正确")
	// }

	// stmt := sql.Lookup(db.driver, "pipeline-env-find")
	// data := []*model.PipelineEnv{}
    // err := meddler.QueryAll(db, &data, stmt, configID)
	// return data, err


	data := []*model.PipelineEnv{}

	db.Where("env_config_id = ?", configID).Find(&data)
	return data,nil
}

func (db *datastore) UpdatePipelineEnvs(configID string, pipelineEnvs []*model.PipelineEnv) error {

    // // 开启事务
	// tx, err := db.Begin()
	// if err != nil{
	// 	return err
	// }

	// // 我们首先需要把数据库中该用户的所有环境变量删除，然后用新的赋值
	// stmt := sql.Lookup(db.driver, "pipeline-env-delete-config")
    // _, err = tx.Exec(stmt, configID)
	// if err != nil{
	// 	err1 := tx.Rollback()
	// 	if err1 != nil{
	// 		return err1
	// 	}
	// 	return err
	// }

	// // 循环取出key value 插入数据库
	// for _ , pipelineEnv := range pipelineEnvs {
	// 	// stmt := sql.Lookup(db.driver, "pipeline_env-insert")
	// 	// _, err = tx.Exec(stmt, config_id,env_key,env_value)

    //     pipelineEnv.ID = 0
    //     pipelineEnv.ConfigID = configID
    //     err := meddler.Insert(tx, "pipeline_env", pipelineEnv)
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
	err := tx.Where("env_config_id = ?", configID).Delete(model.PipelineEnv{}).Error
	if err != nil{
		tx.Rollback()
		return err
	}

	// 循环取出key value 插入数据库
	for _ , pipelineEnv := range pipelineEnvs {
		// stmt := sql.Lookup(db.driver, "pipeline_env-insert")
		// _, err = tx.Exec(stmt, config_id,env_key,env_value)

        pipelineEnv.ID = 0
        pipelineEnv.ConfigID = configID
		err := tx.Create(pipelineEnv).Error
		if err != nil{
			tx.Rollback()
			return err
		}	
	}

    // 提交事务
	tx.Commit()
	return nil
}

// func (db *datastore) GetAllPilelineEnv(repo *model.Repo) ([]*model.PipelineEnvSchema, error) {
// 	
// 	stmt := sql.Lookup(db.driver, "all_pipeline_env-find")
// 	data := []*model.PipelineEnvSchema{}
// 
// 	err := meddler.QueryAll(db, &data, stmt, repo.ID)
// 	return data, err
// }
