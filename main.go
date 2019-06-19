package main

import (
	"flag"
	"path/filepath"

	"log"

	"github.com/gjbae1212/cloud-instance-dns/server"
)

var (
	configPath = flag.String("c", "", "config yaml path")
)

func main() {
	flag.Parse()

	yamlPath, err := filepath.Abs(*configPath)
	if err != nil {
		log.Panic(err)
	}

	s, err := server.NewServer(yamlPath)
	if err != nil {
		log.Panic(err)
	}
	s.Start()
}