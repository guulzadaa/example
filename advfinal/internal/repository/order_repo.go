package repository

import (
	"context"
	"errors"
	"time"

	"bookstore/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository interface {
	Create(order models.Order, items []models.OrderItem) (models.Order, []models.OrderItem, error)
	GetByID(id int) (models.Order, []models.OrderItem, error)
	GetAll() []models.Order
	Update(order models.Order) error
	Delete(id int) error
}

type OrderRepo struct {
	ordersCol *mongo.Collection
	itemsCol  *mongo.Collection
	counters  *CounterRepo
}

func NewOrderRepo(db *mongo.Database) *OrderRepo {
	return &OrderRepo{
		ordersCol: db.Collection("orders"),
		itemsCol:  db.Collection("order_items"),
		counters:  NewCounterRepo(db),
	}
}

func (r *OrderRepo) Create(order models.Order, items []models.OrderItem) (models.Order, []models.OrderItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if order.CustomerID <= 0 {
		return models.Order{}, nil, errors.New("customerId must be positive")
	}
	if order.CartID <= 0 {
		return models.Order{}, nil, errors.New("cartId must be positive")
	}
	if len(items) == 0 {
		return models.Order{}, nil, errors.New("order items required")
	}
	if order.Total < 0 {
		return models.Order{}, nil, errors.New("total cannot be negative")
	}

	orderID, err := r.counters.Next("orders")
	if err != nil {
		return models.Order{}, nil, err
	}
	order.ID = orderID

	_, err = r.ordersCol.InsertOne(ctx, order)
	if err != nil {
		return models.Order{}, nil, err
	}

	outItems := make([]models.OrderItem, 0, len(items))
	docs := make([]any, 0, len(items))

	for _, it := range items {
		if it.BookID <= 0 {
			_, _ = r.ordersCol.DeleteOne(ctx, bson.M{"id": order.ID})
			return models.Order{}, nil, errors.New("bookId must be positive")
		}
		if it.Qty <= 0 {
			_, _ = r.ordersCol.DeleteOne(ctx, bson.M{"id": order.ID})
			return models.Order{}, nil, errors.New("qty must be positive")
		}
		if it.Price < 0 {
			_, _ = r.ordersCol.DeleteOne(ctx, bson.M{"id": order.ID})
			return models.Order{}, nil, errors.New("price cannot be negative")
		}

		itemID, err := r.counters.Next("order_items")
		if err != nil {
			_, _ = r.ordersCol.DeleteOne(ctx, bson.M{"id": order.ID})
			return models.Order{}, nil, err
		}

		it.ID = itemID
		it.OrderID = order.ID

		docs = append(docs, it)
		outItems = append(outItems, it)
	}

	if len(docs) > 0 {
		_, err = r.itemsCol.InsertMany(ctx, docs)
		if err != nil {
			_, _ = r.ordersCol.DeleteOne(ctx, bson.M{"id": order.ID})
			_, _ = r.itemsCol.DeleteMany(ctx, bson.M{"orderId": order.ID})
			return models.Order{}, nil, err
		}
	}

	return order, outItems, nil
}

func (r *OrderRepo) GetByID(id int) (models.Order, []models.OrderItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var o models.Order
	err := r.ordersCol.FindOne(ctx, bson.M{"id": id}).Decode(&o)
	if err == mongo.ErrNoDocuments {
		return models.Order{}, nil, errors.New("order not found")
	}
	if err != nil {
		return models.Order{}, nil, err
	}

	cur, err := r.itemsCol.Find(ctx, bson.M{"orderId": id})
	if err != nil {
		return models.Order{}, nil, err
	}
	defer cur.Close(ctx)

	items := []models.OrderItem{}
	for cur.Next(ctx) {
		var it models.OrderItem
		if cur.Decode(&it) == nil {
			items = append(items, it)
		}
	}

	return o, items, nil
}

func (r *OrderRepo) GetAll() []models.Order {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := r.ordersCol.Find(ctx, bson.M{})
	if err != nil {
		return []models.Order{}
	}
	defer cur.Close(ctx)

	out := []models.Order{}
	for cur.Next(ctx) {
		var o models.Order
		if cur.Decode(&o) == nil {
			out = append(out, o)
		}
	}

	return out
}

func (r *OrderRepo) Update(order models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if order.ID <= 0 {
		return errors.New("order id must be positive")
	}
	if order.CustomerID <= 0 {
		return errors.New("customerId must be positive")
	}
	if order.CartID <= 0 {
		return errors.New("cartId must be positive")
	}
	if order.Total < 0 {
		return errors.New("total cannot be negative")
	}

	res, err := r.ordersCol.UpdateOne(ctx, bson.M{"id": order.ID}, bson.M{"$set": bson.M{
		"customerId": order.CustomerID,
		"cartId":     order.CartID,
		"total":      order.Total,
	}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *OrderRepo) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := r.ordersCol.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("order not found")
	}

	_, _ = r.itemsCol.DeleteMany(ctx, bson.M{"orderId": id})
	return nil
}
