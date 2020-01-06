UPDATE users r 
LEFT JOIN  project p 
ON r.project_name=p.project_name
set r.project_id=p.project_id;
