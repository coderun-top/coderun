package main

import (
    "fmt"
    "io/ioutil"
    "log"

    "gopkg.in/yaml.v2"
)

type conf struct {
    Enabled bool   `yaml:"enabled"`
    Path    string `yaml:"path"`
}

func (c *conf) getConf() *conf {

    yamlFile, err := ioutil.ReadFile("conf.yaml")
    if err != nil {
        log.Printf("yamlFile.Get err   #%v ", err)
    }
    err = yaml.Unmarshal(yamlFile, c)
    if err != nil {
        log.Fatalf("Unmarshal: %v", err)
    }

    return c
}

func main() {
    var c conf
    c.getConf()

    fmt.Println(c.Path)
}
