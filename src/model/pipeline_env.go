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
type PipelineEnv struct {
	ID       int64  `json:"id"    meddler:"pipeline_env_id,pk"     gorm:"AUTO_INCREMENT;primary_key;column:pipeline_env_id"`
	ConfigID string  `json:"conf_id"    meddler:"env_config_id"     gorm:"type:varchar(250);column:env_config_id;not null;unique_index:idx_pipeline"`
	Key      string `json:"key" meddler:"env_key"                  gorm:"type:varchar(250);column:env_key;unique_index:idx_pipeline"`
	Value    string `json:"value" meddler:"env_value"              gorm:"type:varchar(3000);column:env_value"`
}

// 定义生成表的名称
func (PipelineEnv) TableName() string {
	return "pipeline_env"
}
