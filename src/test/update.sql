UPDATE users r 
LEFT JOIN  project p 
ON r.project_name=p.project_name
set r.project_id=p.project_id;

UPDATE project_env r 
LEFT JOIN  project p 
ON r.env_project=p.project_name
set r.env_project_id=p.project_id;

UPDATE agent_keys r 
LEFT JOIN  project p 
ON r.project=p.project_name
set r.project_id=p.project_id;

UPDATE k8s_cluster r 
LEFT JOIN  project p 
ON r.k8s_cluster_project=p.project_name
set r.k8s_cluster_project_id=p.project_id;

UPDATE k8s_deploy r 
LEFT JOIN  project p 
ON r.k8s_deploy_project=p.project_name
set r.k8s_deploy_project_id=p.project_id;

UPDATE helm r 
LEFT JOIN  project p 
ON r.helm_project_name=p.project_name
set r.helm_project_id=p.project_id;

UPDATE registry r 
LEFT JOIN  project p 
ON r.registry_project_name=p.project_name
set r.registry_project_id=p.project_id;

UPDATE repos r 
left join users u
ON u.user_id=r.repo_user_id
LEFT JOIN  project p 
ON p.project_id=u.project_id
set r.repo_project_id=p.project_id;

