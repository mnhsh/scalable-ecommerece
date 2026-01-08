package admin

import (
	"net/http"

	"github.com/herodragmon/scalable-ecommerce/internal/config"
	"github.com/herodragmon/scalable-ecommerce/internal/response"
)

// Used only for local dev and manual testing.
func handlerReset(cfg *config.APIConfig, w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}

	err := cfg.DB.Reset(r.Context())
	if err != nil {
		response.RespondWithError(
			w,
			http.StatusInternalServerError,
			"Couldn't reset database",
			err,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Database reset to initial state"))
}
