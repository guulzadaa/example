package handlers

import (
	"encoding/json"
	"net/http"

	"bookstore/internal/logic"
	"bookstore/internal/middleware"
)

type OrderHandler struct {
	svc *logic.OrderService
}

func NewOrderHandler(svc *logic.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) Orders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		var in struct {
			CartID int `json:"cartId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}

		o, items, err := h.svc.CreateOrderFromCart(userID, in.CartID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{"order": o, "items": items})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}
