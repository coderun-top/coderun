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

import "errors"

var (
	errRegistryAddressInvalid  = errors.New("Invalid Registry Address")
	errRegistryUsernameInvalid = errors.New("Invalid Registry Username")
	errRegistryPasswordInvalid = errors.New("Invalid Registry Password")
)

// RegistryService defines a service for managing registries.
type RegistryService interface {
	RegistryFind(string, string) (*Registry, error)
	RegistryList(string) ([]*Registry, error)
	RegistryCreate(*Registry) error
	RegistryUpdate(*Registry) error
	RegistryDelete(string, string) error
}

// RegistryStore persists registry information to storage.
type RegistryStore interface {
	RegistryFind(string, string) (*Registry, error)
	RegistryList(string) ([]*Registry, error)
	RegistryCreate(*Registry) error
	RegistryUpdate(*Registry) error
	RegistryDelete(string, string) error
}

// Registry represents a docker registry with credentials.
// swagger:model registry
type Registry struct {
	ID          int64  `json:"id"           meddler:"registry_id,pk"                  gorm:"AUTO_INCREMENT;primary_key;column:registry_id"`
	ProjectID  string  `json:"-"         gorm:"type:varchar(250);column:registry_project_id;unique_index:idx_registry"`
	Type        string `json:"type"         meddler:"registry_type"                   gorm:"type:varchar(250);column:registry_type"`
	Name        string `json:"name"         meddler:"registry_name"                   gorm:"type:varchar(250);column:registry_name;unique_index:idx_registry"`
	Address     string `json:"address"      meddler:"registry_addr"                   gorm:"type:varchar(250);column:registry_addr"`
	Username    string `json:"username"     meddler:"registry_username"               gorm:"type:varchar(2000);column:registry_username"`
	Password    string `json:"password"     meddler:"registry_password"               gorm:"type:varchar(8000);column:registry_password"`
	Email       string `json:"email"        meddler:"registry_email"                  gorm:"type:varchar(500);column:registry_email"`
	Token       string `json:"token"        meddler:"registry_token"                  gorm:"type:varchar(2000);column:registry_token"`
	Prefix      string `json:"prefix"       meddler:"registry_prefix"                 gorm:"type:varchar(250);column:registry_prefix"`
	Project      *Project `json:"project"`
}

type RegistryRepo struct {
	ID          int64  `json:"id"           meddler:"registry_id,pk"`
	RepoID int64  `json:"repo_id"    meddler:"registry_repo_repo_id"`
	Branche string `json:"branche" meddler:"registry_repo_branche"`
	RegistryID int64  `json:"registry_id"    meddler:"registry_repo_registry_id"`
}

// Validate validates the registry information.
func (r *Registry) Validate() error {
	switch {
	case len(r.Address) == 0:
		return errRegistryAddressInvalid
	case len(r.Username) == 0:
		return errRegistryUsernameInvalid
	case len(r.Password) == 0:
		return errRegistryPasswordInvalid
	default:
		return nil
	}
}

// Copy makes a copy of the registry without the password.
func (r *Registry) Copy() *Registry {
	return &Registry{
		ID:           r.ID,
		Type:         r.Type,
		Name:         r.Name,
		Address:      r.Address,
		Username:     r.Username,
		Email:        r.Email,
		Token:        r.Token,
		Prefix:       r.Prefix,
	}
}

// 定义生成表的名称
func (Registry) TableName() string {
	return "registry"
}
