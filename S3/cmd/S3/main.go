package main

import (
	"S3_project/S3/internal/app/apiserver"
	"flag"
	"github.com/BurntSushi/toml"
	"log"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "s3/configs/apiserver.toml", "config file path")
}

func main() {
	flag.Parse()

	config := apiserver.NewConfig()

	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("S3 server started successfully")
	if err := apiserver.Start(config); err != nil {
		log.Fatal(err)
	}
}
