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

type CartHandler struct {
	service *logic.CartCRUDService
}

func NewCartHandler(service *logic.CartCRUDService) *CartHandler {
	return &CartHandler{service: service}
}

func (h *CartHandler) Carts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	role := middleware.Role(r)

	switch r.Method {
	case http.MethodGet:
		if role == "admin" {
			writeJSON(w, http.StatusOK, h.service.ListCarts())
			return
		}

		all := h.service.ListCarts()
		out := make([]models.Cart, 0)
		for _, c := range all {
			if c.CustomerID == userID {
				out = append(out, c)
			}
		}
		writeJSON(w, http.StatusOK, out)

	case http.MethodPost:
		c := h.service.CreateCart(userID)
		writeJSON(w, http.StatusCreated, c)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *CartHandler) CartByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	role := middleware.Role(r)

	idStr := strings.TrimPrefix(r.URL.Path, "/carts/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	c, items, err := h.service.GetCart(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	if role != "admin" && c.CustomerID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"cart": c, "items": items})

	case http.MethodPut:
		var in models.Cart
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		in.ID = id

		if role != "admin" {
			in.CustomerID = userID
		}

		if err := h.service.UpdateCart(in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})

	case http.MethodDelete:
		if err := h.service.DeleteCart(id); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *CartHandler) CartItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	role := middleware.Role(r)

	path := strings.TrimPrefix(r.URL.Path, "/carts/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "items" {
		http.NotFound(w, r)
		return
	}

	cartID, err := strconv.Atoi(parts[0])
	if err != nil || cartID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cart id"})
		return
	}

	c, _, err := h.service.GetCart(cartID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if role != "admin" && c.CustomerID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		var in struct {
			BookID int `json:"bookId"`
			Qty    int `json:"qty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}

		item, err := h.service.AddItem(cartID, in.BookID, in.Qty)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, item)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *CartHandler) CartItemByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	role := middleware.Role(r)

	path := strings.TrimPrefix(r.URL.Path, "/carts/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "items" {
		http.NotFound(w, r)
		return
	}

	cartID, err := strconv.Atoi(parts[0])
	if err != nil || cartID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cart id"})
		return
	}

	itemID, err := strconv.Atoi(parts[2])
	if err != nil || itemID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid item id"})
		return
	}

	c, _, err := h.service.GetCart(cartID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if role != "admin" && c.CustomerID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var in struct {
			Qty int `json:"qty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if err := h.service.UpdateItem(cartID, itemID, in.Qty); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})

	case http.MethodDelete:
		if err := h.service.DeleteItem(cartID, itemID); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}
