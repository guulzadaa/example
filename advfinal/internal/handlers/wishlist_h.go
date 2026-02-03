package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bookstore/internal/logic"
)

type WishlistHandler struct {
	service *logic.WishlistService
}

func NewWishlistHandler(service *logic.WishlistService) *WishlistHandler {
	return &WishlistHandler{service: service}
}

func (h *WishlistHandler) Wishlists(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.service.ListWishlists())
	case http.MethodPost:
		var in struct {
			CustomerID int `json:"customerId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		wl := h.service.CreateWishlist(in.CustomerID)
		writeJSON(w, http.StatusCreated, wl)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *WishlistHandler) WishlistByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/wishlists/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	wl, items, err := h.service.GetWishlist(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"wishlist": wl,
		"items":    items,
	})
}

func (h *WishlistHandler) WishlistItems(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/wishlists/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "items" {
		http.NotFound(w, r)
		return
	}

	wishlistID, err := strconv.Atoi(parts[0])
	if err != nil || wishlistID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid wishlist id"})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var in struct {
		BookID int `json:"bookId"`
		Qty    int `json:"qty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	item, err := h.service.AddItem(wishlistID, in.BookID, in.Qty)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *WishlistHandler) Gift(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/wishlists/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "gift" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	wishlistID, err := strconv.Atoi(parts[0])
	if err != nil || wishlistID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid wishlist id"})
		return
	}

	buyerID, _ := strconv.Atoi(r.URL.Query().Get("buyerCustomerId"))

	order, items, giftForCustomerID, err := h.service.GiftFromWishlist(wishlistID, buyerID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"order":             order,
		"items":             items,
		"giftForCustomerId": giftForCustomerID,
	})
}
