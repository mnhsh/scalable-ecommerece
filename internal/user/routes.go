package user

import (
	"net/http"

	"github.com/herodragmon/scalable-ecommerce/internal/config"
	"github.com/herodragmon/scalable-ecommerce/internal/middleware"
)

func RegisterRoutes(mux *http.ServeMux, cfg *config.APIConfig) {
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		handlerLogin(cfg, w, r)
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		handlerUsersCreate(cfg, w, r)
	})

	mux.Handle(
		"GET /api/me",
		middleware.Auth(cfg)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerMe(cfg, w, r)
			}),
		),
	)

	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		handlerRefresh(cfg, w, r)
	})

	mux.HandleFunc("POST /api/revoke", func(w http.ResponseWriter, r *http.Request) {
		handlerRevoke(cfg, w, r)
	})
}
