
package main

import (
    "fmt"
)

func main() {

	m := make(map[interface{}]interface{})
	m["branches"] = "13123"
	m2 := make(map[interface{}]interface{})
	m2["clsd"] = "asda"
	m2["clsd2"] = "asda2"
	m["steps"] = map[interface{}]interface{}{"bind_id": m2,}
	//m["steps"]["aaa
	mm, ok := m["steps"].(map[interface{}]interface{})
	if ok {
	    //mm = make(map[string]string)
	    mm["b2"] =  m2
		fmt.Println(mm)
	}
	fmt.Println(m)

	//m["steps"]["bind_id"]="qqq"
	//	Vargs: map[string]interface{}{"bind_id": k8s_deploy_id,
	//				      "cluster": k8s_deploy.ClusterName,
	//			              "tags": "latest",
	//				      "template": path,
	//			      },
	//qq := m["steps"]
	//qq["a1"] = "1111"
	//map[branches:[master] steps:map[docker2:map[image:crun/kube cluster:testkube namespace:default tags:latest template:deployment.yaml]]]
	fmt.Println(m)
}

