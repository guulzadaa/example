package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bookstore/internal/logic"
	"bookstore/internal/middleware"
	"bookstore/internal/models"
)

type OrderCRUDHandler struct {
	crud *logic.OrderCRUDService
}

func NewOrderCRUDHandler(crud *logic.OrderCRUDService) *OrderCRUDHandler {
	return &OrderCRUDHandler{crud: crud}
}

func (h *OrderCRUDHandler) Orders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	role := middleware.Role(r)

	switch r.Method {
	case http.MethodGet:
		all := h.crud.ListOrders()

		if role == "admin" {
			writeJSON(w, http.StatusOK, all)
			return
		}

		out := make([]models.Order, 0)
		for _, o := range all {
			if o.CustomerID == userID {
				out = append(out, o)
			}
		}
		writeJSON(w, http.StatusOK, out)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *OrderCRUDHandler) OrderByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	role := middleware.Role(r)

	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	o, items, err := h.crud.GetOrder(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	if role != "admin" && o.CustomerID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"order": o, "items": items})

	case http.MethodPut:
		if role != "admin" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin only"})
			return
		}

		var in models.Order
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		in.ID = id

		if err := h.crud.UpdateOrder(in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})

	case http.MethodDelete:
		if role != "admin" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin only"})
			return
		}

		if err := h.crud.DeleteOrder(id); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}
