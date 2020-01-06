
package main

import (
	//"errors"
	"io/ioutil"
	"encoding/json"
	//"os"
	"fmt"
	"net/http"
)

type Registry struct {
	ID          int64  `json:"id"           meddler:"registry_id,pk"                  gorm:"AUTO_INCREMENT;primary_key;column:registry_id"`
	ProjectName string `json:"projectname"  meddler:"registry_project_name"           gorm:"type:varchar(250);column:registry_project_name"`
	Type        string `json:"type"         meddler:"registry_type"                   gorm:"type:varchar(250);column:registry_type"`
	Name        string `json:"name"         meddler:"registry_name"                   gorm:"type:varchar(250);column:registry_name"`
	Address     string `json:"address"      meddler:"registry_addr"                   gorm:"type:varchar(250);column:registry_addr"`
	Username    string `json:"username"     meddler:"registry_username"               gorm:"type:varchar(2000);column:registry_username"`
	Password    string `json:"password"     meddler:"registry_password"               gorm:"type:varchar(8000);column:registry_password"`
	Email    string `json:"email"           meddler:"registry_email"                  gorm:"type:varchar(500);column:registry_email"`
	Token    string `json:"token"           meddler:"registry_token"                  gorm:"type:varchar(2000);column:registry_token"`
	Prefix string `json:"prefix"           meddler:"registry_prefix"                  gorm:"type:varchar(500);column:registry_prefix"`
}


func main() {
	client := &http.Client{}
	//提交请求
	username := "watasihakamidesu"
	name := "coderun"
	url := "https://dev.crun.top/dougo/api/user/"+username+"/inside/registry/"+name
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
	data := new(Registry)
	err = json.Unmarshal(bodyByte,data)
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println("data",data)
	//return data,nil
}
