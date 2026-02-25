package http

import (
	"database/sql"
	"net/http"
)

type HTTPConfig struct {
	Port  int
	DBURL string
}

func NewHTTPServer(cfg HTTPConfig, mux *http.ServeMux, db *sql.DB) *http.Server {
	return nil
}
