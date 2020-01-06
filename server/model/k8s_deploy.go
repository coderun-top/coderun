package model

// Config represents a pipeline configuration.
type K8sDeploy struct {
	ID       int64  `json:"id"         gorm:"AUTO_INCREMENT;primary_key;column:k8s_deploy_id"`
	ClusterName string `json:"cluster_name"   gorm:"type:varchar(250);column:k8s_deploy_cluster"`
	ProjectID  string  `json:"-"         gorm:"type:varchar(250);column:k8s_deploy_project_id"`
	Namespace string `json:"namespace"      gorm:"type:varchar(250);column:k8s_deploy_namespace"`
	Name string `json:"name"       gorm:"type:varchar(250);column:k8s_deploy_name;unique_index:idx_k8s_deploy"`
	Image string `json:"image"  gorm:"type:varchar(250);column:k8s_deploy_image"`
	ExposePort int `json:"expose_port" gorm:"type:integer;column:k8s_deploy_export_port"`
	Replicas int `json:"replicas"           gorm:"type:integer;column:k8s_deploy_replicas"`
	InternalPort int `json:"internal_port"           gorm:"type:integer;column:k8s_deploy_internal_prot"`
	DeployContent string `json:"deploy_content" gorm:"type:varchar(5000);column:k8s_deploy_deploy_content"`
	ServiceContent string `json:"service_content" gorm:"type:varchar(5000);column:k8s_deploy_service_content"`
	Type int `json:"type" gorm:"type:integer;column:k8s_deploy_type"`
	Project      *Project `json:"project"`
}
// 定义生成表的名称
func (K8sDeploy) TableName() string {
	return "k8s_deploy"
}
