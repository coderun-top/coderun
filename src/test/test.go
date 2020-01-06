package main

import (
	"fmt"
	"github.com/drone/expr"
)

func main() {
	var st *expr.Selector
	var err error

	st, err = expr.ParseString("repo GLOB 'watasihakamidesu/test_dougo'")
	if err != nil {
		fmt.Println("error:%s", err)
	}
	fmt.Printf("hello:%v", st)

	vv := map[string]string{
		"repo": "watasihakamidesu/test_dougo1",
		"project":"watasihakamidesu-123",
	}
	if st != nil {
		match, _ := st.Eval(expr.NewRow(vv))
		fmt.Printf("match:%v", match)
	}

	//	logrus.Debugf("task.Labels: %v", task.Labels)
	//	map[platform:linux/amd64 project:watasihakamidesu repo:watasihakamidesu/test_dougo]
	//	if st != nil {
	//		match, _ := st.Eval(expr.NewRow(task.Labels))
	//		return match
	//	}

	//	for k, v := range filter.Labels {
	//		if task.Labels[k] != v {
	//			return false
	//		}
	//	}
	//	return true
}
