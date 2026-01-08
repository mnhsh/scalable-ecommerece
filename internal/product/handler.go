package product

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/herodragmon/scalable-ecommerce/internal/config"
	"github.com/herodragmon/scalable-ecommerce/internal/response"
)

func handlerProductsGet(cfg *config.APIConfig, w http.ResponseWriter, r *http.Request) {
	products, err := cfg.DB.GetProducts(r.Context())
	if err != nil {
		response.RespondWithError(
			w,
			http.StatusInternalServerError,
			"Couldn't get products",
			err,
		)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, products)
}

func handlerProductsGetByID(cfg *config.APIConfig, w http.ResponseWriter, r *http.Request) {
	productIDStr := r.PathValue("productID")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	product, err := cfg.DB.GetProductByID(r.Context(), productID)
	if err != nil {
		response.RespondWithError(w, http.StatusNotFound, "Couldn't get product", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, product)
}
