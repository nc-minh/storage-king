package main

import (
	"database/sql"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nc-minh/storage-king/api"
	db "github.com/nc-minh/storage-king/db/sqlc"
	gg "github.com/nc-minh/storage-king/google"
	"github.com/nc-minh/storage-king/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Env == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if err != nil {
		log.Fatal().Err(err).Msg("cannot marshal config to JSON")
	}
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURIs[0],
		Scopes: []string{
			drive.DriveMetadataReadonlyScope,
			drive.DriveFileScope,
			"email",
		},
		Endpoint: google.Endpoint,
	}

	if err != nil {
		log.Fatal().Msg("cannot create oauth2 config")
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Msg("cannot connect to db")
	}

	store := db.NewStore(conn)

	google := gg.NewDrive(oauth2Config, store)

	runGinServer(config, oauth2Config, store, google)
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)

	if err != nil {
		log.Fatal().Msg("cannot create migration")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("cannot run migration")
	}

	log.Info().Msg("migration is done")
}

func runGinServer(config utils.Config, oauth2Config *oauth2.Config, store db.Store, google gg.GoogleAuthService) {
	server, err := api.NewServer(config, oauth2Config, store, google)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	server.HttpLogger()

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
