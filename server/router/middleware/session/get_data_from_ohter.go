package session

import (


	"errors"
	"io/ioutil"
	"net/http"
)


func get_data(url string, token string) ([]byte,error){
	client := &http.Client{}
	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)
	
	//增加header选项
	reqest.Header.Add("Authorization", token)
	
	if err != nil {
		panic(err)
	}
	//处理返回结果
	response, _ := client.Do(reqest)
	if response != nil{
		defer response.Body.Close()
	}
	// 检查状态码
	status := response.StatusCode
	if status != 200 {
		return nil,errors.New("http请求失败")
	}
	
	bodyByte, err := ioutil.ReadAll(response.Body)
	
	if err != nil {
		return nil,err
	}
	
	return bodyByte,nil
}
