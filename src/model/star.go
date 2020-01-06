package model

// swagger:model build
type Star struct {
	ID        int64   `json:"star_id"            meddler:"star_id,pk"     gorm:"AUTO_INCREMENT;primary_key;column:star_id"`
	UserName string `json:"username"            meddler:"username"            gorm:"type:varchar(500);column:username"`
	RepoID int64 `json:"repo_id"            meddler:"repo_id"            gorm:"type:integer;column:repo_id"`
}



// 定义生成表的名称
func (Star) TableName() string {
	return "star"
  }
