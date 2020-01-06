package main

import (
	//"errors"
	"io/ioutil"
	"encoding/json"
	//"os"
	"fmt"
	"net/http"
)

type K8sCluster struct {
	ID       int64  `json:"id"         meddler:"k8s_cluster_id,pk"     gorm:"AUTO_INCREMENT;primary_key;column:k8s_cluster_id"`
	Project  string `json:"project"    meddler:"k8s_cluster_project"   gorm:"type:varchar(250);column:k8s_cluster_project"`
	Name     string `json:"name"       meddler:"k8s_cluster_name"      gorm:"type:varchar(250);column:k8s_cluster_name"`
	Provider string `json:"provider"   meddler:"k8s_cluster_provider"   gorm:"type:varchar(250);column:k8s_cluster_provider"`
	Address  string `json:"address"    meddler:"k8s_cluster_address"    gorm:"type:varchar(250);column:k8s_cluster_address"`
	Cert     string `json:"cert"       meddler:"k8s_cluster_cert"       gorm:"type:varchar(250);column:k8s_cluster_cert"`
	Token    string `json:"token"      meddler:"k8s_cluster_token"      gorm:"type:varchar(250);column:k8s_cluster_token"`
}


func main() {
	client := &http.Client{}
	//提交请求
	username := "watasihakamidesu"
	name := "testkube"
	url := "https://dev.crun.top/dougo/api/user/"+username+"/k8s/cluster/"+name
	reqest, err := http.NewRequest("GET", url, nil)

	//增加header选项
	reqest.Header.Add("Authorization", "bfm0ls225cqg00f1upkg")

	if err != nil {
		fmt.Println(err)
	}
	//处理返回结果
	response, _ := client.Do(reqest)
	if response != nil{
		defer response.Body.Close()
	}
	// 检查状态码
	status := response.StatusCode
	if status != 200 {
		fmt.Println("http请求失败")
	}

	bodyByte, err := ioutil.ReadAll(response.Body)

	if err != nil {
		//return nil,err
		fmt.Println(err)
	}
	//fmt.Println("body",string(bodyByte))
	data := new(K8sCluster)
	err = json.Unmarshal(bodyByte,data)
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println("data",data)
	//return data,nil
}
