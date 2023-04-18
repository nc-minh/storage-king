package main

import (
	"log"

	"github.com/nc-minh/storage-king/api"
	"github.com/nc-minh/storage-king/utils"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	runGinServer(config)
}

func runGinServer(config utils.Config) {
	server, err := api.NewServer(config)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
