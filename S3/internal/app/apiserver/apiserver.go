package apiserver

import (
	"S3_project/S3/internal/app/store/filestore"
	"database/sql"
	"net/http"

	_ "github.com/lib/pq"
)

func Start(config *Config) error {
	db, err := newDB(config.DatabaseURL)
	if err != nil {
		return err
	}

	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	fileStore := filestore.New(db, config.storePath)
	srv := NewServer(fileStore, config.apiGatewayUrl)

	return http.ListenAndServe(config.BindAddr, srv)
}

func newDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
