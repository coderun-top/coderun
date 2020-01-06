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

package internal

import (
	// "encoding/json"
	// log "github.com/sirupsen/logrus"
	"github.com/coderun-top/coderun/src/model"
)


type Email2 struct{
	Email string `json:"email"`
}
type UserEmail struct{
	Values []Email2 `json:"values"`
}

func (c *Client) GetCurrentUser() (*model.User, error) {
	
	user_current,err := c.FindCurrent()
	if err != nil {
		return nil, err
	}
	email,err := c.ListEmail()
	if err != nil {
		return nil, err
	}
	
	return &model.User{
		Login : user_current.Login,
		Email : email.Values[0].Email,
		Avatar: user_current.Links.Avatar.Href,
		Token : c.AccessToken,
	}, nil
}
