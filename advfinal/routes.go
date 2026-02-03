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

	// Repositories
	bookRepo := repository.NewBookRepo(mongoDB)
	userRepo := repository.NewUserRepo(mongoDB)
	cartRepo := repository.NewCartRepo()
	wishlistRepo := repository.NewWishlistRepo(mongoDB)
	orderRepo := repository.NewOrderRepo(mongoDB)

	// Background workers
	logic.StartOrderWorkerPool(2, cartRepo, wishlistRepo)

	// Services
	bookService := logic.NewBookService(bookRepo)
	cartCRUDService := logic.NewCartCRUDService(cartRepo, bookRepo)

	orderSvc := logic.NewOrderService(orderRepo, bookRepo, cartRepo)
	orderCRUD := logic.NewOrderCRUDService(orderRepo)

	wishlistService := logic.NewWishlistService(wishlistRepo, bookRepo, orderRepo)
	authService := logic.NewAuthService(userRepo, secret)

	// Handlers
	bookHandler := handlers.NewBookHandler(bookService)
	cartHandler := handlers.NewCartHandler(cartCRUDService)

	orderHandler := handlers.NewOrderHandler(orderSvc)
	orderCRUDHandler := handlers.NewOrderCRUDHandler(orderCRUD)

	wishlistHandler := handlers.NewWishlistHandler(wishlistService)
	authHandler := handlers.NewAuthHandler(authService)

	frontend, err := handlers.NewFrontendHandler(bookService, authService)
	if err != nil {
		log.Fatal(err)
	}

	// ---------------- FRONTEND ----------------
	mux.HandleFunc("GET /", frontend.Catalog)

	mux.HandleFunc("GET /login", frontend.Login)
	mux.HandleFunc("POST /login", frontend.LoginPost)

	mux.HandleFunc("GET /register", frontend.Register)
	mux.HandleFunc("POST /register", frontend.RegisterPost)

	mux.HandleFunc("GET /admin/books", frontend.AdminBooks)

	// ---------------- HEALTH ----------------
	mux.HandleFunc("GET /health", handlers.Health)

	// ---------------- AUTH API ----------------
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// ---------------- BOOKS API ----------------
	mux.HandleFunc("GET /books", bookHandler.Books)
	mux.HandleFunc("POST /books", middleware.AdminOnly(secret, bookHandler.Books))

	// /books/{id}
	mux.HandleFunc("GET /books/", bookHandler.BookByID)
	mux.HandleFunc("PUT /books/", middleware.AdminOnly(secret, bookHandler.BookByID))
	mux.HandleFunc("DELETE /books/", middleware.AdminOnly(secret, bookHandler.BookByID))

	// ---------------- CARTS API ----------------
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

	// ---------------- ORDERS API ----------------
	mux.HandleFunc("POST /orders", middleware.AuthOnly(secret, orderHandler.Orders))
	mux.HandleFunc("GET /orders", middleware.AuthOnly(secret, orderCRUDHandler.Orders))

	ordersByID := middleware.AuthOnly(secret, orderCRUDHandler.OrderByID)
	mux.HandleFunc("GET /orders/", ordersByID)
	mux.HandleFunc("POST /orders/", ordersByID)
	mux.HandleFunc("PUT /orders/", ordersByID)
	mux.HandleFunc("DELETE /orders/", ordersByID)

	// ---------------- WISHLISTS API ----------------
	mux.HandleFunc("GET /wishlists", middleware.AuthOnly(secret, wishlistHandler.Wishlists))
	mux.HandleFunc("POST /wishlists", middleware.AuthOnly(secret, wishlistHandler.Wishlists))

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

	mux.HandleFunc("GET /wishlists/", wishlistsPrefixHandler)
	mux.HandleFunc("POST /wishlists/", wishlistsPrefixHandler)
	mux.HandleFunc("PUT /wishlists/", wishlistsPrefixHandler)
	mux.HandleFunc("DELETE /wishlists/", wishlistsPrefixHandler)

	// ---------------- ADMIN PING ----------------
	mux.HandleFunc("GET /admin/ping", middleware.AdminOnly(secret, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"admin ok"}`))
	}))

	// Optional old demo handler
	mux.HandleFunc("POST /cart/add", handlers.AddToCartHandler)
}
