package admin

import (
	"net/http"

	"github.com/herodragmon/scalable-ecommerce/internal/config"
)

func RegisterRoutes(mux *http.ServeMux, cfg *config.APIConfig) {
	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		handlerReset(cfg, w, r)
	})
}
