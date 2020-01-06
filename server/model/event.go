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

// EventType defines the possible types of build events.
type EventType string

const (
	// Retry  EventType = "retry"
	Enqueued  EventType = "enqueued"
	Started   EventType = "started"
	Finished  EventType = "finished"
	Cancelled EventType = "cancelled"
)

// Event represents a build event.
type Event struct {
	Type  EventType `json:"type"`
	Repo  Repo      `json:"repo"`
	Build Build     `json:"build"`
	Proc  Proc      `json:"proc"`
}