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
type K8sCluster struct {
	ID         int64  `json:"id"         meddler:"k8s_cluster_id,pk"     gorm:"AUTO_INCREMENT;primary_key;column:k8s_cluster_id"`
	ProjectID  string  `json:"-"         gorm:"type:varchar(250);column:k8s_cluster_project_id;unique_index:idx_k8s_cluster"`
	Name       string `json:"name"       meddler:"k8s_cluster_name"      gorm:"type:varchar(250);column:k8s_cluster_name;unique_index:idx_k8s_cluster"`
	Provider   string `json:"provider"   meddler:"k8s_cluster_provider"   gorm:"type:varchar(250);column:k8s_cluster_provider"`
	Address    string `json:"address"    meddler:"k8s_cluster_address"    gorm:"type:varchar(250);column:k8s_cluster_address"`
	Cert       string `json:"cert"       meddler:"k8s_cluster_cert"       gorm:"type:varchar(250);column:k8s_cluster_cert"`
	Token      string `json:"token"      meddler:"k8s_cluster_token"      gorm:"type:varchar(2000);column:k8s_cluster_token"`
	KubeConfig string `json:"kubeconfig" gorm:"type:mediumblob;column:k8s_cluster_kubeconfig"`
	Project      *Project `json:"project"`
}

// Copy makes a copy of the registry without the password.
func (r *K8sCluster) Copy() *K8sCluster {
	return &K8sCluster{
		ID:         r.ID,
		Name:       r.Name,
		Provider:   r.Provider,
		Address:    r.Address,
		Cert:       r.Cert,
		KubeConfig: r.KubeConfig,
	}
}

// 定义生成表的名称
func (K8sCluster) TableName() string {
	return "k8s_cluster"
}
