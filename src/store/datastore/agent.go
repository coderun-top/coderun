package datastore

import (
	"github.com/coderun-top/coderun/src/model"
)

// AgentKey
func (db *datastore) GetAgentKey(projectID string) (*model.AgentKey, error) {
	data := new(model.AgentKey)
	err := db.Where("agent_key_project_id = ?", projectID).First(&data).Error
	return data, err
}

func (db *datastore) GetAgentKeyByKey(key string) (*model.AgentKey, error) {
	data := new(model.AgentKey)
	err := db.Where("agent_key = ?", key).First(&data).Error
	return data, err
}

func (db *datastore) CreateAgentKey(key *model.AgentKey) error {
	err := db.Create(key).Error
	return err
}

// Agent
func (db *datastore) GetAgents(projectID string) ([]*model.Agent, error) {
	data := []*model.Agent{}
	err := db.Find(&data, "agent_project_id = ?", projectID).Error
	return data, err
}

func (db *datastore) GetAgent(agentID string) (*model.Agent, error) {
	data := new(model.Agent)
	err := db.Where("agent_id = ?", agentID).Find(&data).Error
	return data, err
}

func (db *datastore) GetAgentClient(projectID string, clientID string) (*model.Agent, error) {
	data := new(model.Agent)
	err := db.Where("agent_project_id = ? and agent_client_id = ?", projectID, clientID).First(&data).Error
	return data, err
}

func (db *datastore) GetAgentCount(projectID string) (int, error) {
	var count int
	err := db.Model(&model.Agent{}).Where("agent_project_id = ? AND agent_status = 1", projectID).Count(&count).Error
	return count, err
}

func (db *datastore) CreateAgent(agent *model.Agent) error {
	err := db.Create(agent).Error
	return err
}

func (db *datastore) UpdateAgent(agent *model.Agent) error {
	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.Agent{}).Where("agent_id = ?", agent.ID).Count(&count).Error
	if err != nil || count == 0 {
		return nil
	}

	return db.Save(agent).Error
}
