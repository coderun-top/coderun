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

package client

import (
	"encoding/json"
	// "github.com/sirupsen/logrus"
)

const (
	droneServiceUrl = "/projects/:id/services/drone-ci"
	hookUrl = "/repos/:id/hooks"
	DelhookUrl = "/repos/:id/hooks/:host_id"
)

type Hooks struct {
	ID    int `json:"id"`
	Url     string `json:"url"`
}


func (c *Client) GetHook(id string, t int) ([]*Hooks, error) {
	var hooks []*Hooks
	url, opaque := c.ResourceUrl(
		hookUrl,
		QMap{":id": id},
		nil,
	)



	contents, err := c.Do("GET", url, opaque, nil, t)

	if err == nil {
		err = json.Unmarshal(contents, &hooks)
	}
	return hooks, err
}

func (c *Client) AddHook(id string, params QMap, t int) error {
	// url, opaque := c.ResourceUrl(
	// 	hookUrl,
	// 	QMap{":id": id},
	// 	params,
	// )
	
	url := c.BaseUrl + c.ApiPath + "/repos/" + id + "/hooks"
	
	data,_ := json.Marshal(params)
	_, err := c.Do("POST", url, "", data, t)
	return err
}

func (c *Client) DeleteHook(id, host_id string, t int) error {
	url, opaque := c.ResourceUrl(
		DelhookUrl,
		QMap{":id": id,
		     ":host_id": host_id},
		nil,
	)
	_, err := c.Do("DELETE", url, opaque, nil, t)
	return err
}

//func (c *Client) AddDroneService(id string, params QMap, t int) error {
//	url, opaque := c.ResourceUrl(
//		droneServiceUrl,
//		QMap{":id": id},
//		params,
//	)
//
//	_, err := c.Do("PUT", url, opaque, nil, t)
//	return err
//}
//
//func (c *Client) DeleteDroneService(id string, t int) error {
//	url, opaque := c.ResourceUrl(
//		droneServiceUrl,
//		QMap{":id": id},
//		nil,
//	)
//
//	_, err := c.Do("DELETE", url, opaque, nil, t)
//	return err
//}
