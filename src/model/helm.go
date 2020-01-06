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
	errHelmAddressInvalid  = errors.New("Invalid Helm Address")
	errHelmUsernameInvalid = errors.New("Invalid Helm Username")
	errHelmPasswordInvalid = errors.New("Invalid Helm Password")
)

// HelmService defines a service for managing registries.
type HelmService interface {
	HelmFind(string, string) (*Helm, error)
	HelmList(string) ([]*Helm, error)
	HelmCreate(*Helm) error
	HelmUpdate(*Helm) error
	HelmDelete(string, string) error
}

// HelmStore persists registry information to storage.
type HelmStore interface {
	HelmFind(string, string) (*Helm, error)
	HelmList(string) ([]*Helm, error)
	HelmCreate(*Helm) error
	HelmUpdate(*Helm) error
	HelmDelete(string, string) error
}

// Helm represents a docker registry with credentials.
// swagger:model registry
type Helm struct {
	ID          int64  `json:"id"           meddler:"helm_id,pk"                  gorm:"AUTO_INCREMENT;primary_key;column:helm_id"`
	ProjectID  string  `json:"-"         gorm:"type:varchar(250);column:helm_project_id;unique_index:idx_helm"`
	Type        string `json:"type"         meddler:"helm_type"                   gorm:"type:varchar(250);column:helm_type"`
	Name        string `json:"name"         meddler:"helm_name"                   gorm:"type:varchar(250);column:helm_name;unique_index:idx_helm"`
	Address     string `json:"address"      meddler:"helm_addr"                   gorm:"type:varchar(250);column:helm_addr"`
	Username    string `json:"username"     meddler:"helm_username"               gorm:"type:varchar(2000);column:helm_username"`
	Password    string `json:"password"     meddler:"helm_password"               gorm:"type:varchar(8000);column:helm_password"`
	Prefix      string `json:"prefix"       meddler:"helm_prefix"                 gorm:"type:varchar(250);column:helm_prefix"`
	Project      *Project `json:"project"`
}

// Validate validates the registry information.
func (r *Helm) Validate() error {
	switch {
	case len(r.Address) == 0:
		return errHelmAddressInvalid
	case len(r.Username) == 0:
		return errHelmUsernameInvalid
	case len(r.Password) == 0:
		return errHelmPasswordInvalid
	default:
		return nil
	}
}

// Copy makes a copy of the registry without the password.
func (r *Helm) Copy() *Helm {
	return &Helm{
		ID:           r.ID,
		//ProjectName:  r.ProjectName,
		Type:         r.Type,
		Name:         r.Name,
		Address:      r.Address,
		Username:     r.Username,
		Prefix:       r.Prefix,
	}
}

// 定义生成表的名称
func (Helm) TableName() string {
	return "helm"
}
