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

package model

// ConfigStore persists pipeline configuration to storage.
type ConfigStore interface {
	ConfigLoad(string) (*Config, error)
	ConfigFind(*Repo, string) (*Config, error)
	ConfigFindApproved(*Config) (bool, error)
	ConfigCreate(*Config) error
	ConfigRepoFind(*Repo) ([]*Config, error)
}

// Config represents a pipeline configuration.
type Config struct {
	ID          string `json:"id"    meddler:"config_id,pk"             gorm:"type:varchar(250);primary_key;column:config_id"`
	RepoID      int64  `json:"repo_id"    meddler:"config_repo_id"       gorm:"type:integer;column:config_repo_id;unique_index:idx_config"`
	Data        string `json:"data" meddler:"config_data"                gorm:"type:mediumblob;column:config_data"`
	Hash        string `json:"hash" meddler:"config_hash"                gorm:"type:varchar(250);column:config_hash;unique_index:idx_config"`
	FilePath    string `json:"file_path" meddler:"config_file_path"   gorm:"type:varchar(250);column:config_file_path"`
	File        int64  `json:"file" meddler:"config_file"                 gorm:"type:integer;column:config_file;default:1"`
	Branche     string `json:"branche" meddler:"config_branche"         gorm:"type:varchar(250);column:config_branche;unique_index:idx_config"`
	AgentPublic bool   `json:"agent_public" meddler:"agent_public"    gorm:"column:agent_public"`
	AgentFilter string `json:"agent_filter" meddler:"agent_filter"  gorm:"type:varchar(500);column:agent_filter"`

	ConfigType          string `json:"config_type"         gorm:"type:varchar(50);column:config_type"`
	StepBuild           string `json:"step_build"            gorm:"type:mediumblob;column:step_build"`
	StepUnitTest        string `json:"step_unit_test"        gorm:"type:mediumblob;column:step_unit_test"`
	StepIntegrationTest string `json:"step_integration_test" gorm:"type:mediumblob;column:step_integration_test"`
	StepDeploy          string `json:"step_deploy"           gorm:"type:mediumblob;column:step_deploy"`
}

// 定义生成表的名称
func (Config) TableName() string {
	return "config"
}
