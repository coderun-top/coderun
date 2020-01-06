package model

import (
	"github.com/coderun-top/coderun/server/utils"
)

// agent key
type AgentKey struct {
	ID        string `json:"agent_key_id"  gorm:"type:varchar(250);column:agent_key_id;primary_key"`
	Key       string `json:"agent_key"     gorm:"type:varchar(500);column:agent_key;unique_index:idx_agent_key"`
	ProjectID string `json:"agent_key_project_id" gorm:"type:varchar(250);column:agent_key_project_id;unique_index:idx_agent_key"`
	// Project      *Project `json:"project"`
}

func (t *AgentKey) BeforeCreate() (err error) {
	t.ID = utils.GeneratorId()
	return nil
}

func (AgentKey) TableName() string {
	return "agent_keys"
}

// agent
type Agent struct {
	ID        string `json:"agent_id"         gorm:"type:varchar(250);column:agent_id;primary_key"`
	ProjectID string `json:"agent_project_id" gorm:"type:varchar(250);column:agent_project_id;unique_index:idx_agent"`
	// user model: agent for coderun for user to deploy heself
	// dougo model: agent for dougo for all user
	Type     string `json:"agent_type"       gorm:"type:varchar(250);column:agent_type"`
	ClientID string `json:"agent_client_id"  gorm:"type:varchar(250);column:agent_client_id;unique_index:idx_agent"`
	Platform string `json:"agent_platform"   gorm:"type:varchar(250);column:agent_platform"`
	Name     string `json:"agent_name"       gorm:"type:varchar(250);column:agent_name"`
	Addr     string `json:"agent_addr"       gorm:"type:varchar(250);column:agent_addr"`
	Tags     string `json:"agent_tags"       gorm:"type:varchar(500);column:agent_tags"`
	MaxProcs int    `json:"agent_max_procs"  gorm:"type:int;column:agent_max_procs"`
	// 1 is started, 0 is invaild
	Status  int `json:"agent_status"     gorm:"type:integer;column:agent_status"`
	Created int `json:"-"                gorm:"type:integer;column:agent_created"`
	Updated int `json:"-"                gorm:"type:integer;column:agent_updated"`
}

func (t *Agent) BeforeCreate() (err error) {
	t.ID = utils.GeneratorId()
	return nil
}

func (Agent) TableName() string {
	return "agents"
}
