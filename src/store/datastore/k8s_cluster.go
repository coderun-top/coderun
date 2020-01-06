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

func (db *datastore) K8sClusterFind(project, name string) (*model.K8sCluster, error) {
	// data := new(model.K8sCluster)
	// err := meddler.QueryRow(db, data, rebind(k8sClusterFind), project, name)
	// return data, err


	data := new(model.K8sCluster)
	err := db.Where("k8s_cluster_project_id = ? AND k8s_cluster_name = ?", project, name).First(&data).Error
	return data, err

}


func (db *datastore) K8sClusterList(project string) ([]*model.K8sCluster, error) {
	// data := []*model.K8sCluster{}
	// err := meddler.QueryAll(db, &data, rebind(k8sClusterFindProject), project)
	// return data, err

	data := []*model.K8sCluster{}
	err := db.Where("k8s_cluster_project_id = ?",project).Find(&data).Error
	return data, err
}

func (db *datastore) K8sClusterCreate(k8sCluster *model.K8sCluster) error {
	// return meddler.Insert(db, "k8s_cluster", k8sCluster)
	err := db.Create(k8sCluster).Error
	return err
}

func (db *datastore) K8sClusterUpdate(k8sCluster *model.K8sCluster) error {
	// return meddler.Update(db, "k8s_cluster", k8sCluster)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.K8sCluster{}).Where("k8s_cluster_id = ?",k8sCluster.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}

	return db.Save(k8sCluster).Error
}

func (db *datastore) K8sClusterDelete(project, name string) error {
	// _, err := db.Exec(rebind(k8sClusterDelete), project, name)
	// return err

	err := db.Where("k8s_cluster_project_id = ? AND k8s_cluster_name = ?", project, name).Delete(model.K8sCluster{}).Error
	return err
}

const k8sClusterFind = `
SELECT
 k8s_cluster_id
,k8s_cluster_project
,k8s_cluster_name
,k8s_cluster_provider
,k8s_cluster_address
,k8s_cluster_cert
,k8s_cluster_token
FROM k8s_cluster
WHERE k8s_cluster_project = ?
  AND k8s_cluster_name = ?
`

const k8sClusterFindProject = `
SELECT
 k8s_cluster_id
,k8s_cluster_project
,k8s_cluster_name
,k8s_cluster_provider
,k8s_cluster_address
,k8s_cluster_cert
,k8s_cluster_token
FROM k8s_cluster
WHERE k8s_cluster_project = ?
`

const k8sClusterDelete = `
DELETE FROM k8s_cluster WHERE k8s_cluster_project = ? AND k8s_cluster_name = ?
`
