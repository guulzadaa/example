package logic

import (
	"errors"
	"log"

	"bookstore/internal/models"
	"bookstore/internal/repository"
)

type WishlistService struct {
	wRepo     repository.WishlistRepository
	bookRepo  repository.BookRepository
	orderRepo repository.OrderRepository
}

func NewWishlistService(
	wRepo repository.WishlistRepository,
	bookRepo repository.BookRepository,
	orderRepo repository.OrderRepository,
) *WishlistService {
	return &WishlistService{
		wRepo:     wRepo,
		bookRepo:  bookRepo,
		orderRepo: orderRepo,
	}
}

func (s *WishlistService) CreateWishlist(customerID int) models.Wishlist {
	if customerID <= 0 {
		customerID = 1
	}
	return s.wRepo.Create(customerID)
}

func (s *WishlistService) ListWishlists() []models.Wishlist {
	return s.wRepo.GetAll()
}

func (s *WishlistService) GetWishlist(id int) (models.Wishlist, []models.WishlistItem, error) {
	return s.wRepo.GetByID(id)
}

func (s *WishlistService) AddItem(wishlistID, bookID, qty int) (models.WishlistItem, error) {
	if wishlistID <= 0 {
		return models.WishlistItem{}, errors.New("wishlistId must be positive")
	}
	if bookID <= 0 {
		return models.WishlistItem{}, errors.New("bookId must be positive")
	}
	if qty <= 0 {
		return models.WishlistItem{}, errors.New("qty must be > 0")
	}
	if _, err := s.bookRepo.GetByID(bookID); err != nil {
		return models.WishlistItem{}, errors.New("book not found")
	}
	return s.wRepo.AddItem(wishlistID, bookID, qty)
}

func (s *WishlistService) GiftFromWishlist(wishlistID int, buyerID int) (models.Order, []models.OrderItem, int, error) {
	if wishlistID <= 0 {
		return models.Order{}, nil, 0, errors.New("wishlistId must be positive")
	}
	if buyerID <= 0 {
		return models.Order{}, nil, 0, errors.New("buyerCustomerId must be positive")
	}

	w, items, err := s.wRepo.GetByID(wishlistID)
	if err != nil {
		return models.Order{}, nil, 0, err
	}
	if len(items) == 0 {
		return models.Order{}, nil, 0, errors.New("wishlist is empty")
	}

	orderItems := make([]models.OrderItem, 0, len(items))
	var total float64

	for _, wi := range items {
		if wi.BookID <= 0 {
			return models.Order{}, nil, 0, errors.New("invalid bookId in wishlist")
		}
		if wi.Qty <= 0 {
			return models.Order{}, nil, 0, errors.New("invalid qty in wishlist")
		}

		book, err := s.bookRepo.GetByID(wi.BookID)
		if err != nil {
			return models.Order{}, nil, 0, errors.New("book not found")
		}
		if book.Price < 0 {
			return models.Order{}, nil, 0, errors.New("book price cannot be negative")
		}

		orderItems = append(orderItems, models.OrderItem{
			BookID: wi.BookID,
			Qty:    wi.Qty,
			Price:  book.Price,
		})

		total += book.Price * float64(wi.Qty)
	}

	order := models.Order{
		CustomerID: buyerID,
		CartID:     wishlistID,
		Total:      total,
	}

	createdOrder, createdItems, err := s.orderRepo.Create(order, orderItems)
	if err != nil {
		return models.Order{}, nil, 0, err
	}

	log.Printf("[GIFT] enqueue clear wishlist job: wishlistId=%d orderId=%d\n", wishlistID, createdOrder.ID)

	OrderJobQueue <- OrderJob{
		Type:       JobClearWishlist,
		OrderID:    createdOrder.ID,
		WishlistID: wishlistID,
	}

	return createdOrder, createdItems, w.CustomerID, nil
}
