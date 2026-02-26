package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	flagapp "github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/features/flag/application"
	flaghttp "github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/features/flag/http"
	flagrepo "github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/features/flag/infra/postgres"
)

type HTTPConfig struct {
	Port  int
}

func NewHTTPServer(cfg HTTPConfig, db *sql.DB) *http.Server {
	repo := flagrepo.NewRepository(db)
	flagService := flagapp.NewService(repo)

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
