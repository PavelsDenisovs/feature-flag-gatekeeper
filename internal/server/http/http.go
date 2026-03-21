package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	flagapp "github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/application"
	flaghttp "github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/http"
	flagpq "github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/infra/postgres"
)

type HTTPConfig struct {
	Port int
}

func New(cfg HTTPConfig, db *sql.DB) *http.Server {
	repo := flagpq.New(db)
	flagService := flagapp.New(repo)

	mux := http.NewServeMux()

	flaghttp.RegisterEndpoints(mux, flagService)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}
