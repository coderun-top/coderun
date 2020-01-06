
package main

import (
    "fmt"
)

func main() {
	haha := "123123"
	tmp_content := `
kind: `+haha+`
`
    fmt.Println(tmp_content)
}
