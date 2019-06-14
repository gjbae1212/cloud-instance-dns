package main

import (
	"flag"
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v2"
)

var (
	configPath = flag.String("c", "", "config yaml path")
)

func main() {
	flag.Parse()

	config := make(map[interface{}]interface{})
	bys, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Panic(err)
	}
	if err = yaml.Unmarshal(bys, &config); err != nil {
		log.Panic(err)
	}
}
