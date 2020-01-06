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

import (
	"errors"
	"regexp"
)

// validate a username (e.g. from github)
var reUsername = regexp.MustCompile("^[a-zA-Z0-9-_.]+$")

var errUserLoginInvalid = errors.New("Invalid User Login")

// User represents a registered user.
//
// swagger:model user
type User struct {
	// the id for this user.
	//
	// required: true
	ID int64 `json:"id" meddler:"user_id,pk"                           gorm:"AUTO_INCREMENT;primary_key;column:user_id"`

	// Login is the username for this user.
	//
	// required: true
	Login string `json:"login"  meddler:"user_login"                   gorm:"type:varchar(250);column:user_login"`


	// Token is the oauth2 token.
	Token string `json:"-"  meddler:"user_token"                   gorm:"type:varchar(500);column:user_token"`

	// Secret is the oauth2 token secret.
	Secret string `json:"-" meddler:"user_secret"                      gorm:"type:varchar(500);column:user_secret"`

	// Expiry is the token and secret expiration timestamp.
	Expiry int64 `json:"-" meddler:"user_expiry"                       gorm:"type:integer;column:user_expiry"`

	// Email is the email address for this user.
	//
	// required: true
	Email string `json:"email" meddler:"user_email"                    gorm:"type:varchar(500);column:user_email"`

	// the avatar url for this user.
	Avatar string `json:"avatar_url" meddler:"user_avatar"             gorm:"type:varchar(500);column:user_avatar"`

	// Activate indicates the user is active in the system.
	Active bool `json:"active" meddler:"user_active"                   gorm:"column:user_active"`

	// Synced is the timestamp when the user was synced with the remote system.
	Synced int64 `json:"synced" meddler:"user_synced"                  gorm:"type:integer;column:user_synced"`

	// Admin indicates the user is a system administrator.
	//
	// NOTE: This is sourced from the DRONE_ADMINS environment variable and is no
	// longer persisted in the database.
	Admin bool `json:"admin,omitempty" meddler:"-"                     gorm:"-"`

	// Hash is a unique token used to sign tokens.
	Hash string `json:"-" meddler:"user_hash"                         gorm:"type:varchar(500);column:user_hash"`

	// DEPRECATED Admin indicates the user is a system administrator.
	XAdmin bool `json:"-" meddler:"user_admin"                        gorm:"column:user_admin"`

	Provider int64 `json:"provider" meddler:"user_provider"           gorm:"type:integer;column:user_provider"`
	Oauth bool `json:"oauth" meddler:"oauth"                          gorm:"column:oauth"`
	Host string `json:"host" meddler:"user_host"                      gorm:"type:varchar(500);column:user_host"`
	TokenName string `json:"token_name" meddler:"user_token_name"     gorm:"type:varchar(500);column:user_token_name"`
	ProjectID string `json:"-"                                        gorm:"type:varchar(250);column:user_project_id"`

	Project  *Project `json:"project"`

}

// Validate validates the required fields and formats.
func (u *User) Validate() error {
	switch {
	case len(u.Login) == 0:
		return errUserLoginInvalid
	case len(u.Login) > 250:
		return errUserLoginInvalid
	case !reUsername.MatchString(u.Login):
		return errUserLoginInvalid
	default:
		return nil
	}
}
