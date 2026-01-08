package product

import (
	"net/http"

	"github.com/herodragmon/scalable-ecommerce/internal/config"
)

func RegisterRoutes(mux *http.ServeMux, cfg *config.APIConfig) {
	mux.HandleFunc("GET /api/products", func(w http.ResponseWriter, r *http.Request) {
		handlerProductsGet(cfg, w, r)
	})

	mux.HandleFunc("GET /api/products/{productID}", func(w http.ResponseWriter, r *http.Request) {
		handlerProductsGetByID(cfg, w, r)
	})
}
