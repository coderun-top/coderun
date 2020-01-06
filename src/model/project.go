package model

import (
	"github.com/coderun-top/coderun/src/utils"
)

type Project struct {
	ID        string `json:"id"          gorm:"type:varchar(200);column:project_id;primary_key"`
	Name      string `json:"name"        gorm:"type:varchar(250);column:project_name;unique"`
	OwnerName string `json:"owern_name"  gorm:"type:varchar(250);column:project_owner_name"`
}

func (p *Project) BeforeCreate() (err error) {
	p.ID = utils.GeneratorId()
	return nil
}

func (Project) TableName() string {
	return "projects"
}
