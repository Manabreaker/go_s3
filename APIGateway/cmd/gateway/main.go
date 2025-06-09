package main

import (
	"S3_project/APIGateway/internal/app/apiserver"
	"fmt"
	"log"
)

const (
	configPath = "APIGateway/configs/apiserver.yml"
)

func main() {
	config, err := apiserver.NewConfig(configPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Ошибка при обработке файла `%s`: %e", configPath, err))
	}
	GatewayServer := apiserver.NewServer()
	log.Println("APIGateway server started successfully")
	if err := GatewayServer.Start(config); err != nil {
		log.Fatal(fmt.Sprintf("Ошибка при запуске сервера: %e", err))
	}
}
