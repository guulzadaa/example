package repository

import (
	"errors"
	"sync"

	"bookstore/internal/models"
)

type CartRepository interface {
	Create(customerID int) models.Cart
	GetAll() []models.Cart
	GetByID(id int) (models.Cart, []models.CartItem, error)
	Update(cart models.Cart) error
	Delete(id int) error

	AddItem(cartID int, bookID int, qty int) (models.CartItem, error)
	UpdateItem(cartID int, itemID int, qty int) error
	DeleteItem(cartID int, itemID int) error

	ClearCart(cartID int) error
}

type CartRepo struct {
	mu sync.RWMutex

	nextCartID int
	nextItemID int

	carts map[int]models.Cart
	items map[int][]models.CartItem
}

func NewCartRepo() *CartRepo {
	return &CartRepo{
		nextCartID: 1,
		nextItemID: 1,
		carts:      make(map[int]models.Cart),
		items:      make(map[int][]models.CartItem),
	}
}

func (r *CartRepo) Create(customerID int) models.Cart {
	r.mu.Lock()
	defer r.mu.Unlock()

	c := models.Cart{
		ID:         r.nextCartID,
		CustomerID: customerID,
	}
	r.nextCartID++
	r.carts[c.ID] = c
	r.items[c.ID] = []models.CartItem{}
	return c
}

func (r *CartRepo) GetAll() []models.Cart {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]models.Cart, 0, len(r.carts))
	for _, c := range r.carts {
		out = append(out, c)
	}
	return out
}

func (r *CartRepo) GetByID(id int) (models.Cart, []models.CartItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.carts[id]
	if !ok {
		return models.Cart{}, nil, errors.New("cart not found")
	}
	items := append([]models.CartItem(nil), r.items[id]...)
	return c, items, nil
}

func (r *CartRepo) Update(cart models.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.carts[cart.ID]; !ok {
		return errors.New("cart not found")
	}
	r.carts[cart.ID] = cart
	return nil
}

func (r *CartRepo) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.carts[id]; !ok {
		return errors.New("cart not found")
	}
	delete(r.carts, id)
	delete(r.items, id)
	return nil
}

func (r *CartRepo) AddItem(cartID int, bookID int, qty int) (models.CartItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.carts[cartID]; !ok {
		return models.CartItem{}, errors.New("cart not found")
	}
	if qty <= 0 {
		return models.CartItem{}, errors.New("qty must be positive")
	}

	items := r.items[cartID]
	for i := range items {
		if items[i].BookID == bookID {
			items[i].Qty += qty
			r.items[cartID] = items
			return items[i], nil
		}
	}

	it := models.CartItem{
		ID:     r.nextItemID,
		CartID: cartID,
		BookID: bookID,
		Qty:    qty,
	}
	r.nextItemID++
	r.items[cartID] = append(r.items[cartID], it)
	return it, nil
}

func (r *CartRepo) UpdateItem(cartID int, itemID int, qty int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if qty <= 0 {
		return errors.New("qty must be positive")
	}

	items := r.items[cartID]
	for i := range items {
		if items[i].ID == itemID {
			items[i].Qty = qty
			r.items[cartID] = items
			return nil
		}
	}
	return errors.New("item not found")
}

func (r *CartRepo) DeleteItem(cartID int, itemID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := r.items[cartID]
	out := make([]models.CartItem, 0, len(items))
	found := false
	for _, it := range items {
		if it.ID == itemID {
			found = true
			continue
		}
		out = append(out, it)
	}
	if !found {
		return errors.New("item not found")
	}
	r.items[cartID] = out
	return nil
}

func (r *CartRepo) ClearCart(cartID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.carts[cartID]; !ok {
		return errors.New("cart not found")
	}
	r.items[cartID] = []models.CartItem{}
	return nil
}
