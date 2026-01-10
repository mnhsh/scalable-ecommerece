package product

import (
	"net/http"

	"github.com/herodragmon/scalable-ecommerce/internal/config"
	"github.com/herodragmon/scalable-ecommerce/internal/middleware"
)

func RegisterRoutes(mux *http.ServeMux, cfg *config.APIConfig) {
	// -------- Public routes --------
	mux.HandleFunc("GET /api/products", func(w http.ResponseWriter, r *http.Request) {
		handlerProductsGet(cfg, w, r)
	})

	mux.HandleFunc("GET /api/products/{productID}", func(w http.ResponseWriter, r *http.Request) {
		handlerProductsGetByID(cfg, w, r)
	})

	// -------- Admin routes --------
	mux.Handle(
		"POST /admin/products",
		middleware.Auth(cfg)(
			middleware.RequireAdmin(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					handlerProductsCreate(cfg, w, r)
				}),
			),
		),
	)
}

