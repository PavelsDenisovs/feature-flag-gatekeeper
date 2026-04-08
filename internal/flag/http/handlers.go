package http

import (
	"net/http"

	flagapp "github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/application"
)

func RegisterEndpoints(
	mux *http.ServeMux,
	service flagapp.FlagService,
) {
}
