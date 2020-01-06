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

func (db *datastore) K8sDeployFind(id int64) (*model.K8sDeploy, error) {
	data := new(model.K8sDeploy)
	err := db.Where("k8s_deploy_id = ?",id).First(&data).Error
	return data, err

}

func (db *datastore) K8sDeployFindProject(project string, name string) (*model.K8sDeploy, error) {
	data := new(model.K8sDeploy)
	err := db.Where("k8s_deploy_project_id = ? AND k8s_deploy_name = ?",project, name).First(&data).Error
	return data, err

}

func (db *datastore) K8sDeployList(project string) ([]*model.K8sDeploy, error) {
	data := []*model.K8sDeploy{}
	err := db.Where("k8s_deploy_project_id = ?",project).Find(&data).Error
	return data, err
}


func (db *datastore) K8sDeployCreate(k8sDeploy *model.K8sDeploy) error {
	err := db.Create(k8sDeploy).Error
	return err
}

func (db *datastore) K8sDeployUpdate(k8sDeploy *model.K8sDeploy) error {
	var count int
	err := db.Model(&model.K8sDeploy{}).Where("k8s_deploy_id = ?",k8sDeploy.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}

	return db.Save(k8sDeploy).Error
}

func (db *datastore) K8sDeployDelete(id int64) error {
	err := db.Where("k8s_deploy_id = ?", id).Delete(model.K8sDeploy{}).Error
	return err
}
