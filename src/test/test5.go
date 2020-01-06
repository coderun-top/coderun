package main
import (
    "fmt"
    "strings"
)
func main() {
    lines :=  "asd\nasd\n"   
    fmt.Println(strings.Replace(lines, "\n", "\\n", -1 ))
}
