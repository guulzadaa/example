package logic

import (
	"errors"

	"bookstore/internal/models"
	"bookstore/internal/repository"
)

type CartCRUDService struct {
	repo     repository.CartRepository
	bookRepo repository.BookRepository
}

func NewCartCRUDService(repo repository.CartRepository, bookRepo repository.BookRepository) *CartCRUDService {
	return &CartCRUDService{repo: repo, bookRepo: bookRepo}
}

func (s *CartCRUDService) CreateCart(customerID int) models.Cart {
	if customerID <= 0 {
		customerID = 1
	}
	return s.repo.Create(customerID)
}

func (s *CartCRUDService) ListCarts() []models.Cart {
	return s.repo.GetAll()
}

func (s *CartCRUDService) GetCart(id int) (models.Cart, []models.CartItem, error) {
	return s.repo.GetByID(id)
}

func (s *CartCRUDService) UpdateCart(c models.Cart) error {
	if c.ID <= 0 {
		return errors.New("cart id must be positive")
	}
	return s.repo.Update(c)
}

func (s *CartCRUDService) DeleteCart(id int) error {
	return s.repo.Delete(id)
}

func (s *CartCRUDService) AddItem(cartID int, bookID int, qty int) (models.CartItem, error) {
	if _, err := s.bookRepo.GetByID(bookID); err != nil {
		return models.CartItem{}, errors.New("book not found")
	}
	return s.repo.AddItem(cartID, bookID, qty)
}

func (s *CartCRUDService) UpdateItem(cartID int, itemID int, qty int) error {
	return s.repo.UpdateItem(cartID, itemID, qty)
}

func (s *CartCRUDService) DeleteItem(cartID int, itemID int) error {
	return s.repo.DeleteItem(cartID, itemID)
}
