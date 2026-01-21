package handlers

import (
	"net/http"
	"database/sql"
	"log"
	"errors"
	
	"github.com/google/uuid" 
	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/config"
	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/response"
	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/database"
)

func RegisterRoutes(mux *http.ServeMux, cfg *config.Config) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "order-service"})
	})

	mux.HandleFunc("POST /api/orders", func(w http.ResponseWriter, r *http.Request) {
		handlerCreateOrder(cfg, w, r)
	})

	mux.HandleFunc("GET /api/orders", func(w http.ResponseWriter, r *http.Request) {
		handlerGetOrders(cfg, w, r)
	})

	mux.HandleFunc("GET /api/orders/{orderID}", func(w http.ResponseWriter, r *http.Request) {
		handlerGetOrder(cfg, w, r)
	})

	mux.HandleFunc("DELETE /api/orders/{orderID}", func(w http.ResponseWriter, r *http.Request) {
		handlerCancelOrder(cfg, w, r)
	})

	mux.HandleFunc("PATCH /internal/orders/{orderID}/status", func(w http.ResponseWriter, r *http.Request) {
		handlerUpdateStatus(cfg, w, r)
	})
}

type OrderResponse struct {
    Order database.Order       `json:"order"`
    Items []database.OrderItem `json:"items"`
}

func handlerCreateOrder(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
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

	cart, exists, err := cfg.CartClient.GetCart(r.Context(), userID)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "error getting cart", err)
		return
	}

	if !exists {
		response.RespondWithError(w, http.StatusNotFound, "cart not found", nil)
		return
	}

	if len(cart.Items) == 0 {
		response.RespondWithError(w, http.StatusBadRequest, "cart is empty", nil)
		return
	}

	order, err := cfg.DB.CreateOrder(r.Context(), database.CreateOrderParams{
		UserID: userID,
		Status: "pending",
		TotalCents: int32(cart.TotalCents),
	})
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "could not create order", err)
		return
	}

	for _, item := range cart.Items {
		_, err := cfg.DB.CreateOrderItem(r.Context(), database.CreateOrderItemParams{
			OrderID : order.ID,
			ProductID : item.ProductID,
			Quantity : item.Quantity,
			PriceCents : item.PriceCents,
		})
		if err != nil {
    	response.RespondWithError(w, http.StatusInternalServerError, "could not create order item", err)
    	return
		}
	}

	err = cfg.CartClient.ClearCart(r.Context(), userID)
	if err != nil {
    log.Printf("warning: failed to clear cart: %v", err)
	}
	
	response.RespondWithJSON(w, http.StatusCreated, order)
}

func handlerGetOrders(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
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

	orders, err := cfg.DB.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "could not get orders", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, orders)
}

func handlerGetOrder(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
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

	orderIDStr := r.PathValue("orderID")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid order ID", err)
		return
	}

	order, err := cfg.DB.GetOrderByID(r.Context(), orderID)
  if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.RespondWithError(w, http.StatusNotFound, "order not found", nil)
			return
		}
		response.RespondWithError(w, http.StatusInternalServerError, "could not get order", err)
		return
	}

	if order.UserID != userID {
		response.RespondWithError(w, http.StatusForbidden, "order does not belong to you", nil)
    return
	}

	items, err := cfg.DB.GetOrderItems(r.Context(), orderID)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "could not get order items", err)
		return
  }

	response.RespondWithJSON(w, http.StatusOK, OrderResponse{
		Order: order,
		Items: items,
	})
}

func handlerCancelOrder(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.PathValue("orderID")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid order ID", err)
		return
	}
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid user ID", err)
		return
	}

	order, err := cfg.DB.GetOrderByID(r.Context(), orderID)
  if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.RespondWithError(w, http.StatusNotFound, "order not found", nil)
			return
		}
		response.RespondWithError(w, http.StatusInternalServerError, "could not get order", err)
		return
	}

	if order.UserID != userID {
		response.RespondWithError(w, http.StatusForbidden, "order does not belong to you", nil)
    return
	}

	if order.Status != "pending" {
		response.RespondWithError(w, http.StatusBadRequest, "can only cancel pending orders", nil)
    return
	}
	
	updatedOrder, err := cfg.DB.UpdateOrderStatus(r.Context(), database.UpdateOrderStatusParams{
		ID: order.ID,
		Status:  database.OrderStatusCancelled,
	})
	if err != nil {
    response.RespondWithError(w, http.StatusInternalServerError, "could not cancel order", err)
    return
	}

	response.RespondWithJSON(w, http.StatusOK, updatedOrder)
}

func handlerUpdateStatus(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	type updateStatusRequest struct {
    Status string `json:"status"`
	}

	orderIDStr := r.PathValue("orderID")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid order ID", err)
		return
	}
	
	decoder := json.NewDecoder(r.Body)
	params := updateStatusRequest{}
	err = decoder.Decode(&params)
	if err != nil {
    response.RespondWithError(w, http.StatusBadRequest, "invalid request body", err)
    return
	}

	var status database.OrderStatus
	switch params.Status {
	case "pending":
	    status = database.OrderStatusPending
	case "paid":
	    status = database.OrderStatusPaid
	case "shipped":
	    status = database.OrderStatusShipped
	case "delivered":
	    status = database.OrderStatusDelivered
	case "cancelled":
	    status = database.OrderStatusCancelled
	default:
	    response.RespondWithError(w, http.StatusBadRequest, "invalid status value", nil)
	    return
	}

	updatedOrder, err := cfg.DB.UpdateOrderStatus(r.Context(), database.UpdateOrderStatusParams{
		ID: orderID,
		Status: status,
	})
	if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
			response.RespondWithError(w, http.StatusNotFound, "order not found", nil)
			return
    }
    response.RespondWithError(w, http.StatusInternalServerError, "could not update order status", err)
    return
	}

	response.RespondWithJSON(w, http.StatusOK, updatedOrder)
}
