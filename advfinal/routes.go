package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"bookstore/internal/handlers"
	"bookstore/internal/logic"
	"bookstore/internal/middleware"
	"bookstore/internal/repository"

	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRoutes(mux *http.ServeMux, mongoDB *mongo.Database) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	// ---------------- Repositories ----------------
	bookRepo := repository.NewBookRepo(mongoDB)
	userRepo := repository.NewUserRepo(mongoDB)
	cartRepo := repository.NewCartRepo() // in-memory
	wishlistRepo := repository.NewWishlistRepo(mongoDB)
	orderRepo := repository.NewOrderRepo(mongoDB)

	// ---------------- Workers ----------------
	logic.StartOrderWorkerPool(2, cartRepo, wishlistRepo)

	// ---------------- Services ----------------
	bookService := logic.NewBookService(bookRepo)
	authService := logic.NewAuthService(userRepo, secret)
	cartCRUDService := logic.NewCartCRUDService(cartRepo, bookRepo)
	orderSvc := logic.NewOrderService(orderRepo, bookRepo, cartRepo)
	orderCRUD := logic.NewOrderCRUDService(orderRepo)
	wishlistService := logic.NewWishlistService(wishlistRepo, bookRepo, orderRepo)

	// ---------------- API Handlers ----------------
	bookHandler := handlers.NewBookHandler(bookService)
	cartHandler := handlers.NewCartHandler(cartCRUDService)
	orderHandler := handlers.NewOrderHandler(orderSvc)
	orderCRUDHandler := handlers.NewOrderCRUDHandler(orderCRUD)
	wishlistHandler := handlers.NewWishlistHandler(wishlistService)
	authHandler := handlers.NewAuthHandler(authService)

	// ---------------- Frontend ----------------
	frontend, err := handlers.NewFrontendHandler(
		bookService,
		authService,
		cartCRUDService,
		orderSvc,
		orderCRUD,
		wishlistService,
		secret,
	)
	if err != nil {
		log.Fatal(err)
	}

	// ================= FRONTEND PAGES =================
	mux.HandleFunc("GET /", frontend.Home)
	mux.HandleFunc("GET /catalog", frontend.Catalog)
	mux.HandleFunc("GET /about", frontend.About)

	mux.HandleFunc("GET /login", frontend.Login)
	mux.HandleFunc("POST /login", frontend.LoginPost)

	mux.HandleFunc("GET /register", frontend.Register)
	mux.HandleFunc("POST /register", frontend.RegisterPost)

	mux.HandleFunc("POST /logout", frontend.Logout)

	// Cart
	mux.HandleFunc("GET /cart", frontend.CartPage)
	mux.HandleFunc("POST /cart/add/{bookId}", frontend.CartAdd)
	mux.HandleFunc("POST /cart/item/{itemId}/update", frontend.CartUpdateQty)
	mux.HandleFunc("POST /cart/item/{itemId}/delete", frontend.CartDeleteItem)

	// Orders
	mux.HandleFunc("GET /orders", frontend.OrdersPage)
	mux.HandleFunc("GET /orders/{id}", frontend.OrderDetailsPage)
	mux.HandleFunc("POST /orders/create", frontend.CreateOrderFromCart)

	// Wishlists
	mux.HandleFunc("GET /wishlists", frontend.WishlistsPage)
	mux.HandleFunc("POST /wishlists/add/{bookId}", frontend.WishlistAdd)
	mux.HandleFunc("POST /wishlists/gift/{wishlistId}", frontend.WishlistGift)

	// ================= HEALTH =================
	mux.HandleFunc("GET /health", handlers.Health)

	// ================= AUTH API =================
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// ================= BOOKS API =================
	mux.HandleFunc("GET /books", bookHandler.Books)
	mux.HandleFunc("POST /books", middleware.AdminOnly(secret, bookHandler.Books))

	mux.HandleFunc("GET /books/{id}", bookHandler.BookByID)
	mux.HandleFunc("PUT /books/{id}", middleware.AdminOnly(secret, bookHandler.BookByID))
	mux.HandleFunc("DELETE /books/{id}", middleware.AdminOnly(secret, bookHandler.BookByID))

	// ================= CARTS API =================
	mux.HandleFunc("GET /carts", middleware.AuthOnly(secret, cartHandler.Carts))
	mux.HandleFunc("POST /carts", middleware.AuthOnly(secret, cartHandler.Carts))

	cartsPrefixHandler := middleware.AuthOnly(secret, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/items/") {
			cartHandler.CartItemByID(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/items") {
			cartHandler.CartItems(w, r)
			return
		}
		cartHandler.CartByID(w, r)
	})

	mux.HandleFunc("GET /carts/", cartsPrefixHandler)
	mux.HandleFunc("POST /carts/", cartsPrefixHandler)
	mux.HandleFunc("PUT /carts/", cartsPrefixHandler)
	mux.HandleFunc("DELETE /carts/", cartsPrefixHandler)

	// ================= ORDERS API =================
	mux.HandleFunc("POST /orders_api", middleware.AuthOnly(secret, orderHandler.Orders))
	mux.HandleFunc("GET /orders_api", middleware.AuthOnly(secret, orderCRUDHandler.Orders))

	ordersByID := middleware.AuthOnly(secret, orderCRUDHandler.OrderByID)
	mux.HandleFunc("GET /orders_api/", ordersByID)
	mux.HandleFunc("POST /orders_api/", ordersByID)
	mux.HandleFunc("PUT /orders_api/", ordersByID)
	mux.HandleFunc("DELETE /orders_api/", ordersByID)

	// ================= WISHLISTS API =================
	mux.HandleFunc("GET /wishlists_api", middleware.AuthOnly(secret, wishlistHandler.Wishlists))
	mux.HandleFunc("POST /wishlists_api", middleware.AuthOnly(secret, wishlistHandler.Wishlists))

	wishlistsPrefixHandler := middleware.AuthOnly(secret, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/items") {
			wishlistHandler.WishlistItems(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/gift") {
			wishlistHandler.Gift(w, r)
			return
		}
		wishlistHandler.WishlistByID(w, r)
	})

	mux.HandleFunc("GET /wishlists_api/", wishlistsPrefixHandler)
	mux.HandleFunc("POST /wishlists_api/", wishlistsPrefixHandler)
	mux.HandleFunc("PUT /wishlists_api/", wishlistsPrefixHandler)
	mux.HandleFunc("DELETE /wishlists_api/", wishlistsPrefixHandler)
}
