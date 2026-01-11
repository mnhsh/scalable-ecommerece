package handlers

import (
	"database/sql"
	"net/http"
	"errors"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/herodragmon/scalable-ecommerce/services/cart-service/internal/config"
	"github.com/herodragmon/scalable-ecommerce/services/cart-service/internal/response"
	"github.com/herodragmon/scalable-ecommerce/services/cart-service/internal/database"
)

func RegisterRoutes(mux *http.ServeMux, cfg *config.Config) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "cart-service"})
	})

	mux.HandleFunc("GET /api/cart", func(w http.ResponseWriter, r *http.Request) {
		handlerCartGet(cfg, w, r)
	})

	mux.HandleFunc("POST /api/cart/items", func(w http.ResponseWriter, r *http.Request) {
		handlerCartAddItem(cfg, w, r)
	})
}


func handlerCartGet(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	type CartItemResponse struct {
    ID         uuid.UUID `json:"id"`
    ProductID  uuid.UUID `json:"product_id"`
    Quantity   int32     `json:"quantity"`
    PriceCents int32     `json:"price_cents"`
	}

	type CartResponse struct {
    ID         uuid.UUID          `json:"id,omitempty"`
    Items      []CartItemResponse `json:"items"`
    TotalCents int64              `json:"total_cents"`
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		response.RespondWithError(w, http.StatusUnauthorized, "missing user ID", nil)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid user ID", err)
		return
	}
	
	cart, err := cfg.DB.GetCartByUserID(r.Context(), userID)
	if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        response.RespondWithJSON(w, http.StatusOK, CartResponse{
            Items:      []CartItemResponse{},
            TotalCents: 0,
        })
        return
    }
    response.RespondWithError(w, http.StatusInternalServerError, "couldn't get cart", err)
    return
	}

	items, err := cfg.DB.GetCartItems(r.Context(), cart.ID)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "couldn't get cart items", err)
  	return
	}

	var totalCents int64
    itemResponses := make([]CartItemResponse, len(items))
    for i, item := range items {
			totalCents += int64(item.PriceCents) * int64(item.Quantity)
      itemResponses[i] = CartItemResponse{
				ID:         item.ID,
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
				PriceCents: item.PriceCents,
			}
    }
    
  response.RespondWithJSON(w, http.StatusOK, CartResponse{
		ID:         cart.ID,
		Items:      itemResponses,
		TotalCents: totalCents,
	})
}

func handlerCartAddItem(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	type addItemRequest struct {
    ProductID uuid.UUID `json:"product_id"`
    Quantity  int32     `json:"quantity"`
	}
	type CartItemResponse struct {
    ID         uuid.UUID `json:"id"`
    ProductID  uuid.UUID `json:"product_id"`
    Quantity   int32     `json:"quantity"`
    PriceCents int32     `json:"price_cents"`
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		response.RespondWithError(w, http.StatusUnauthorized, "missing user ID", nil)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid user ID", err)
		return
	}

	var item addItemRequest
	err = json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "couldn't decode json", err)
		return
	}

	if item.ProductID == uuid.Nil || item.Quantity <= 0 {
		response.RespondWithError(w, http.StatusBadRequest, "invalid product_id or quantity", err)
		return
	}

	product, exists, err:= cfg.ProductClient.GetProduct(r.Context(), item.ProductID)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "error checking product", err)
		return
	}
	if !exists {
    response.RespondWithError(w, http.StatusNotFound, "product not found", nil)
    return
	}

	cart, err := cfg.DB.GetCartByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			cart, err = cfg.DB.CreateCart(r.Context(), userID)
			if err != nil {
				response.RespondWithError(w, http.StatusInternalServerError, "couldn't create cart", err)
				return
      }
    } else {
			response.RespondWithError(w, http.StatusInternalServerError, "couldn't get cart", err)
			return
    }
	}

	existingItem, err := cfg.DB.GetCartItemByProductID(r.Context(), database.GetCartItemByProductIDParams{
    CartID:    cart.ID,
    ProductID: item.ProductID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
    response.RespondWithError(w, http.StatusInternalServerError, "couldn't check cart", err)
    return
	}
	var cartItem database.CartItem
	if err == nil {
		cartItem, err = cfg.DB.UpdateCartItemQuantity(r.Context(), database.UpdateCartItemQuantityParams{
			ID:       existingItem.ID,
      Quantity: existingItem.Quantity + item.Quantity,
    })
	} else {
			cartItem, err = cfg.DB.AddCartItem(r.Context(), database.AddCartItemParams{
        CartID:     cart.ID,
        ProductID:  item.ProductID,
        Quantity:   item.Quantity,
        PriceCents: product.PriceCents,
    	})
		}
	if err != nil {
    response.RespondWithError(w, http.StatusInternalServerError, "couldn't add item to cart", err)
    return
	}
	response.RespondWithJSON(w, http.StatusCreated, CartItemResponse{
    ID:         cartItem.ID,
    ProductID:  cartItem.ProductID,
    Quantity:   cartItem.Quantity,
    PriceCents: cartItem.PriceCents,
	})
}
