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
	"fmt"
	"time"

	"github.com/coderun-top/coderun/src/model"
)

func (db *datastore) GetBuild(id string) (*model.Build, error) {
	var build = new(model.Build)
	err := db.Where("build_id = ?", id).First(&build).Error
	return build, err
}

func (db *datastore) GetBuildNumber(repo *model.Repo, num int) (*model.Build, error) {
	var build = new(model.Build)
	err := db.Where("build_repo_id = ? AND build_number = ?", repo.ID, num).First(&build).Error
	return build, err
}

// func (db *datastore) GetBuildNumber2(repo *model.Repo, num int) (*model.BuildRepo, error) {
// 	var build = new(model.BuildRepo)
// 	err := db.Table("builds").Select("builds.*,repos.repo_owner,repos.repo_name,repos.repo_full_name,repos.repo_avatar,repos.repo_branch,users.user_token_name").
// 		Joins("INNER JOIN repos ON repos.repo_id = builds.build_repo_id").
// 		Joins("INNER JOIN users ON users.user_id = repos.repo_user_id").
// 		Where("build_repo_id = ? AND build_number = ?", repo.ID, num).First(&build).Error
// 
// 	return build, err
// }

func (db *datastore) GetBuildRef(repo *model.Repo, ref string) (*model.Build, error) {
	var build = new(model.Build)
	err := db.Where("build_repo_id = ? AND build_ref = ?", repo.ID, ref).First(&build).Error
	return build, err
}

func (db *datastore) GetBuildCommit(repo *model.Repo, sha, branch string) (*model.Build, error) {
	var build = new(model.Build)
	err := db.Where("build_repo_id = ? AND build_commit  = ? AND build_branch  = ?", repo.ID, sha, branch).First(&build).Error
	return build, err
}

func (db *datastore) GetBuildLastByPipeline(pipelineID string) (*model.Build, error) {
	var build = new(model.Build)
	err := db.Where("build_config_id = ?", pipelineID).Order("build_number DESC").First(&build).Error
	return build, err
}

// func (db *datastore) GetBuildLast(repo *model.Repo, branch string) (*model.Build, error) {
// 	var build = new(model.Build)
// 	err := db.Where("build_repo_id = ? AND build_branch = ? AND build_event = ?", repo.ID, branch, "push").Order("build_number DESC").First(&build).Error
// 	return build, err
// }

func (db *datastore) GetBuildLast(repo *model.Repo) (*model.Build, error) {
	var build = new(model.Build)
	err := db.Where("build_repo_id = ?", repo.ID).Order("build_number DESC").First(&build).Error
	return build, err
}

func (db *datastore) GetBuildLastBefore(repo *model.Repo, branch string, buildID string) (*model.Build, error) {
	var build = new(model.Build)
	err := db.Where("build_repo_id = ? AND build_branch  = ? AND build_id < ?", repo.ID, branch, buildID).First(&build).Error
	return build, err
}

func (db *datastore) GetBuildList(repo *model.Repo, state string, page, pageSize int) ([]*model.BuildRepo, error) {
	var builds = []*model.BuildRepo{}
	query := db.Table("builds").Select("builds.*,repos.repo_owner,repos.repo_name,repos.repo_full_name,repos.repo_avatar,repos.repo_branch,users.user_token_name").
		Joins("INNER JOIN repos ON repos.repo_id = builds.build_repo_id").
		Joins("INNER JOIN users ON users.user_id = repos.repo_user_id").
	    Where("builds.build_repo_id = ?", repo.ID)
	if state != "" {
		query = query.Where("builds.build_status = ?", state)
	}
	err := query.Order("build_number DESC").Limit(pageSize).Offset(pageSize * (page - 1)).Find(&builds).Error
	return builds, err
}

func (db *datastore) GetBuildListProject(project string, state string, page, pageSize int) ([]*model.BuildRepo, error) {
	var builds = []*model.BuildRepo{}
	var err error

	query := db.Table("builds").Select("builds.*,repos.repo_owner,repos.repo_name,repos.repo_full_name,repos.repo_avatar,repos.repo_branch,users.user_token_name").
		Joins("INNER JOIN repos ON repos.repo_id = builds.build_repo_id").
		Joins("INNER JOIN users ON users.user_id = repos.repo_user_id").
		Where("users.user_project_id = ?", project)

	if state != "" {
		query = query.Where("builds.build_status = ?", state)
	}
	err = query.Order("builds.build_created DESC").Limit(pageSize).Offset(pageSize * (page - 1)).Scan(&builds).Error
	return builds, err
}

func (db *datastore) GetBuildListProjectCountByMonth(project string, state string) (count int, err error) {
	query := db.Table("builds").
		Joins("INNER JOIN repos ON repos.repo_id = builds.build_repo_id").
		Joins("INNER JOIN users ON users.user_id = repos.repo_user_id").
		Where("users.user_project_id = ? and build_created >= unix_timestamp(date_add(curdate(),interval -day(curdate())+1 day))", project)

	if state != "" {
		query = query.Where("builds.build_status = ?", state)
	}
	err = query.Count(&count).Error
	return
}

func (db *datastore) GetBuildCount2(repo *model.Repo, state string) (count int, err error) {
	query := db.Table("builds").Where("build_repo_id = ?", repo.ID)
	if state != "" {
		query = query.Where("builds.build_status = ?", state)
	}
	err = query.Count(&count).Error
	return
}

func (db *datastore) GetBuildQueue() ([]*model.Feed, error) {
	feed := []*model.Feed{}
	err := db.Table("builds").Select("repos.repo_owner,repos.repo_name,repos.repo_full_name,builds.*").
		Joins("INNER JOIN repos ON builds.build_repo_id = repos.repo_id").
		Where("builds.build_status in (?)", []string{"pending", "running"}).Scan(&feed).Error
	return feed, err
}

func (db *datastore) CreateBuild(build *model.Build, procs ...*model.Proc) error {
	id, err := db.incrementRepoRetry(build.RepoID)
	if err != nil {
		return err
	}
	build.Number = id
	build.Created = time.Now().UTC().Unix()
	build.Enqueued = build.Created
	err = db.Create(build).Error
	if err != nil {
		return err
	}
	for _, proc := range procs {
		proc.BuildID = build.ID
		err = db.Create(proc).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *datastore) incrementRepoRetry(id int64) (int, error) {
	repo, err := db.GetRepo(id)
	if err != nil {
		return 0, fmt.Errorf("database: cannot fetch repository: %s", err)
	}
	for i := 0; i < 10; i++ {
		seq, err := db.incrementRepo(repo.ID, repo.Counter+i, repo.Counter+i+1)
		if err != nil {
			return 0, err
		}
		if seq == 0 {
			continue
		}
		return seq, nil
	}
	return 0, fmt.Errorf("cannot increment next build number")
}

func (db *datastore) incrementRepo(id int64, old, new int) (int, error) {
	query := db.Model(&model.Repo{}).Where("repo_counter = ? AND repo_id = ?", old, id).
		Update(map[string]interface{}{"repo_counter": new})
	err := query.Error
	affect_count := query.RowsAffected

	if err != nil {
		return 0, fmt.Errorf("database: update repository counter: %s", err)
	}
	if affect_count == 0 {
		return 0, nil
	}
	return new, nil
}

func (db *datastore) UpdateBuild(build *model.Build) error {
	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.Build{}).Where("build_id = ?", build.ID).Count(&count).Error
	if err != nil || count == 0 {
		return nil
	}
	return db.Save(build).Error
}

func (db *datastore) GetBuildCount() (count int, err error) {
	err = db.Table("builds").Count(&count).Error
	return
}

func (db *datastore) GetBuildTimeProject(projectID string) (int64, error) {
	type ProjectBuildTime struct {
		Total int64
	}

	var result = new(ProjectBuildTime)
	err := db.Table("builds").Select("SUM(builds.build_finished-builds.build_started) as total").
		Joins("INNER JOIN repos ON repos.repo_id = builds.build_repo_id").
		Where("repos.repo_project_id = ? and builds.build_started>0 and builds.build_finished>0 and build_created >= unix_timestamp(date_add(curdate(),interval -day(curdate())+1 day))", projectID).Scan(&result).Error

	return result.Total, err
}
