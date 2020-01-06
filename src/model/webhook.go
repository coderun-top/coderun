package model

// Team represents a team or organization in the remote version control system.
//
// swagger:model user
type WebHook struct {
	// Login is the username for this team.
	ID int64 `json:"id" meddler:"id,pk"                                 gorm:"AUTO_INCREMENT;primary_key;column:id"`

	RepoID    int64   `json:"repo_id"       meddler:"repo_id"           gorm:"type:integer;column:repo_id;unique"`

	// the avatar url for this team.
	Url string `json:"url" meddler:"url"                                gorm:"type:varchar(500);column:url"`
	State bool `json:"state" meddler:"state"                            gorm:"column:state"`
}

type WebHookData struct {
	Build   *Build `json:"build"`
	Repo  *Repo `json:"repo"`
}

type WebHookProject struct {
	// Login is the username for this team.
	ID int64 `json:"id" meddler:"id,pk"                                 gorm:"AUTO_INCREMENT;primary_key;column:id"`

	//Project string `json:"project"       meddler:"project"              gorm:"type:varchar(250);column:project;unique"`

	ProjectID      string  `json:"-"                        meddler:"project_id"             gorm:"type:varchar(250);column:project_id;unique"`

	// the avatar url for this team. 
	Url string `json:"url" meddler:"url"                                gorm:"type:varchar(500);column:url"`
	State bool `json:"state" meddler:"state"                            gorm:"column:state"`
}

// 定义生成表的名字
func (WebHook) TableName() string {
	return "webhook"
}

func (WebHookProject) TableName() string {
	return "webhook_project"
}
