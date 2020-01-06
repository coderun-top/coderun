package model

// Team represents a team or organization in the remote version control system.
//
// swagger:model user
type Email struct {
	ID int64 `json:"id" meddler:"id,pk"                                 gorm:"AUTO_INCREMENT;primary_key;column:id"`
	RepoID    int64   `json:"repo_id"       meddler:"repo_id"           gorm:"type:integer;column:repo_id;unique"`
	Email string `json:"email" meddler:"email"                                gorm:"type:varchar(500);column:email"`
	State bool `json:"state" meddler:"state"                            gorm:"column:state"`
}

type EmailProject struct {
	ID int64 `json:"id" meddler:"id,pk"                                 gorm:"AUTO_INCREMENT;primary_key;column:id"`
	ProjectID      string  `json:"-"                        meddler:"project_id"             gorm:"type:varchar(250);column:project_id;unique"`
	Email string `json:"email" meddler:"email"                                gorm:"type:varchar(500);column:email"`
	State bool `json:"state" meddler:"state"                            gorm:"column:state"`
}

// 定义生成表的名字
func (Email) TableName() string {
	return "email"
}

func (EmailProject) TableName() string {
	return "email_project"
}
