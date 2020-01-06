package main

import (
	//"errors"
	"io/ioutil"
	"encoding/json"
	//"os"
	"fmt"
	"net/http"
)

type Helm struct {
	ID          int64  `json:"id"           meddler:"helm_id,pk"                  gorm:"AUTO_INCREMENT;primary_key;column:helm_id"`
	ProjectName string `json:"projectname"  meddler:"helm_project_name"           gorm:"type:varchar(250);column:helm_project_name"`
	Name        string `json:"name"         meddler:"helm_name"                   gorm:"type:varchar(250);column:helm_name"`
	Address     string `json:"address"      meddler:"helm_addr"                   gorm:"type:varchar(250);column:helm_addr"`
	Username    string `json:"username"     meddler:"helm_username"               gorm:"type:varchar(2000);column:helm_username"`
	Password    string `json:"password"     meddler:"helm_password"               gorm:"type:varchar(8000);column:helm_password"`
	Prefix      string `json:"prefix"       meddler:"helm_prefix"                 gorm:"type:varchar(250);column:helm_prefix"`
}


func main() {
	client := &http.Client{}
	//提交请求
	username := "watasihakamidesu"
	name := "coderun"
	url := "https://dev.crun.top/dougo/api/user/"+username+"/inside/helm/"+name
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
	data := new(Helm)
	err = json.Unmarshal(bodyByte,data)
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println("data",data)
	//return data,nil
}
