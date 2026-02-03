package models

import "time"

type Book struct {
	ID          int
	Title       string
	Author      string
	Genre       string
	Price       float64
	Description string
}

type User struct {
	ID       int    `json:"id" bson:"id"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
	Role     string `json:"role" bson:"role"`
	Address  string `json:"address,omitempty" bson:"address,omitempty"`
}

type Cart struct {
	ID         int
	CustomerID int
	CreatedAt  time.Time
}

type CartItem struct {
	ID     int
	CartID int
	BookID int
	Qty    int
}

type Order struct {
	ID         int     `json:"id" bson:"id"`
	CustomerID int     `json:"customerId" bson:"customerId"`
	CartID     int     `json:"cartId" bson:"cartId"`
	Total      float64 `json:"total" bson:"total"`
}

type OrderItem struct {
	ID      int     `json:"id" bson:"id"`
	OrderID int     `json:"orderId" bson:"orderId"`
	BookID  int     `json:"bookId" bson:"bookId"`
	Qty     int     `json:"qty" bson:"qty"`
	Price   float64 `json:"price" bson:"price"`
}

type Payment struct {
	ID      int
	OrderID int
	Total   float64
	Status  string
}

type Wishlist struct {
	ID         int `json:"id" bson:"id"`
	CustomerID int `json:"customerId" bson:"customerId"`
}

type WishlistItem struct {
	ID         int `json:"id" bson:"id"`
	WishlistID int `json:"wishlistId" bson:"wishlistId"`
	BookID     int `json:"bookId" bson:"bookId"`
	Qty        int `json:"qty" bson:"qty"`
}
