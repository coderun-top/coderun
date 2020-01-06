// package model
// 
// type PipelineEnv struct {
// 	ID int64 `json:"id" meddler:"pipeline_env_id,pk"`
// 	UserId string `json:"user_id"  meddler:"user_id"`
// 	EnvKey string `json:"env_key" meddler:"env_key"`
// 	EnvValue string `json:"env_value"  meddler:"env_value"`
// }
// 
// type PipelineEnvSchema struct {
// 	EnvKey string `json:"env_key" meddler:"env_key"`
// 	EnvValue string `json:"env_value"  meddler:"env_value"`
// }

package model

// Config represents a pipeline configuration.
type ProjectEnv struct {
	ID       int64  `json:"id"    meddler:"project_env_id,pk"           gorm:"AUTO_INCREMENT;primary_key;column:project_env_id"`
	ProjectID  string  `json:"-"         gorm:"type:varchar(250);column:env_project_id"`
	Key      string `json:"key" meddler:"env_key"                       gorm:"type:varchar(250);column:env_key"`
	Value    string `json:"value" meddler:"env_value"                   gorm:"type:varchar(3000);column:env_value"`
	Project      *Project `json:"project"`
}



// 定义生成表的名称
func (ProjectEnv) TableName() string {
	return "project_env"
}
