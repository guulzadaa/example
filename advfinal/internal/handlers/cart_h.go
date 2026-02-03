package handlers

import (
	"bookstore/internal/logic"
	"bookstore/internal/models"
	"net/http"
	"strconv"
)

func AddToCartHandler(w http.ResponseWriter, r *http.Request) {

	bookID, _ := strconv.Atoi(r.URL.Query().Get("book_id"))
	cartID, _ := strconv.Atoi(r.URL.Query().Get("cart_id"))

	item := models.CartItem{
		BookID: bookID,
		CartID: cartID,
		Qty:    1,
	}

	service := logic.CartService{}
	service.AddItemToCart(item)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Success: Item added and stock reservation started in background."))
}
