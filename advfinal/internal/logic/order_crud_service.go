package logic

import (
	"errors"

	"bookstore/internal/models"
	"bookstore/internal/repository"
)

type OrderCRUDService struct {
	repo repository.OrderRepository
}

func NewOrderCRUDService(repo repository.OrderRepository) *OrderCRUDService {
	return &OrderCRUDService{repo: repo}
}

func (s *OrderCRUDService) ListOrders() []models.Order {
	return s.repo.GetAll()
}

func (s *OrderCRUDService) GetOrder(id int) (models.Order, []models.OrderItem, error) {
	return s.repo.GetByID(id)
}

func (s *OrderCRUDService) UpdateOrder(o models.Order) error {
	if o.ID <= 0 {
		return errors.New("order id must be positive")
	}
	if o.Total < 0 {
		return errors.New("total cannot be negative")
	}
	return s.repo.Update(o)
}

func (s *OrderCRUDService) DeleteOrder(id int) error {
	return s.repo.Delete(id)
}
