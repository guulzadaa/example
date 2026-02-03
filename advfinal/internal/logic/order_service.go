package logic

import (
	"errors"

	"bookstore/internal/models"
	"bookstore/internal/repository"
)

type OrderService struct {
	repo     repository.OrderRepository
	bookRepo repository.BookRepository
	cartRepo repository.CartRepository
}

func NewOrderService(repo repository.OrderRepository, bookRepo repository.BookRepository, cartRepo repository.CartRepository) *OrderService {
	return &OrderService{repo: repo, bookRepo: bookRepo, cartRepo: cartRepo}
}

func (s *OrderService) CreateOrderFromCart(customerID int, cartID int) (models.Order, []models.OrderItem, error) {
	if customerID <= 0 {
		return models.Order{}, nil, errors.New("customerId must be positive")
	}
	if cartID <= 0 {
		return models.Order{}, nil, errors.New("cartId must be positive")
	}

	_, cartItems, err := s.cartRepo.GetByID(cartID)
	if err != nil {
		return models.Order{}, nil, err
	}
	if len(cartItems) == 0 {
		return models.Order{}, nil, errors.New("cart is empty")
	}

	items := make([]models.OrderItem, 0, len(cartItems))
	var total float64

	for _, ci := range cartItems {
		if ci.BookID <= 0 {
			return models.Order{}, nil, errors.New("invalid bookId in cart")
		}
		if ci.Qty <= 0 {
			return models.Order{}, nil, errors.New("invalid qty in cart")
		}

		b, err := s.bookRepo.GetByID(ci.BookID)
		if err != nil {
			return models.Order{}, nil, errors.New("book not found")
		}

		items = append(items, models.OrderItem{
			BookID: ci.BookID,
			Qty:    ci.Qty,
			Price:  b.Price,
		})

		total += b.Price * float64(ci.Qty)
	}

	order := models.Order{
		CustomerID: customerID,
		CartID:     cartID,
		Total:      total,
	}

	createdOrder, createdItems, err := s.repo.Create(order, items)
	if err != nil {
		return models.Order{}, nil, err
	}

	select {
	case OrderJobQueue <- OrderJob{Type: JobAuditOrderCreated, OrderID: createdOrder.ID, CartID: cartID}:
	default:
	}
	select {
	case OrderJobQueue <- OrderJob{Type: JobClearCart, OrderID: createdOrder.ID, CartID: cartID}:
	default:
	}

	return createdOrder, createdItems, nil
}
