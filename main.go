package main

import (
	"log"
	"os"

	"github.com/nc-minh/storage-king/api"
	"github.com/nc-minh/storage-king/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	oauth2Config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope, drive.DriveFileScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	runGinServer(config, oauth2Config)
}

func runGinServer(config utils.Config, oauth2Config *oauth2.Config) {
	server, err := api.NewServer(config, oauth2Config)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
