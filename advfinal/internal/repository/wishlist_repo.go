package repository

import (
	"context"
	"errors"
	"time"

	"bookstore/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type WishlistRepository interface {
	Create(customerID int) models.Wishlist
	GetAll() []models.Wishlist
	GetByID(id int) (models.Wishlist, []models.WishlistItem, error)
	Delete(id int) error

	AddItem(wishlistID int, bookID int, qty int) (models.WishlistItem, error)
	DeleteItem(wishlistID int, itemID int) error
}

type WishlistRepo struct {
	wishlistsCol *mongo.Collection
	itemsCol     *mongo.Collection
	counters     *CounterRepo
}

func NewWishlistRepo(db *mongo.Database) *WishlistRepo {
	return &WishlistRepo{
		wishlistsCol: db.Collection("wishlists"),
		itemsCol:     db.Collection("wishlist_items"),
		counters:     NewCounterRepo(db),
	}
}

func (r *WishlistRepo) Create(customerID int) models.Wishlist {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if customerID <= 0 {
		customerID = 1
	}

	id, err := r.counters.Next("wishlists")
	if err != nil {
		return models.Wishlist{}
	}

	w := models.Wishlist{
		ID:         id,
		CustomerID: customerID,
	}

	_, err = r.wishlistsCol.InsertOne(ctx, w)
	if err != nil {
		return models.Wishlist{}
	}

	return w
}

func (r *WishlistRepo) GetAll() []models.Wishlist {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cur, err := r.wishlistsCol.Find(ctx, bson.M{})
	if err != nil {
		return []models.Wishlist{}
	}
	defer cur.Close(ctx)

	out := []models.Wishlist{}
	for cur.Next(ctx) {
		var w models.Wishlist
		if cur.Decode(&w) == nil {
			out = append(out, w)
		}
	}
	return out
}

func (r *WishlistRepo) GetByID(id int) (models.Wishlist, []models.WishlistItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var w models.Wishlist
	err := r.wishlistsCol.FindOne(ctx, bson.M{"id": id}).Decode(&w)
	if err == mongo.ErrNoDocuments {
		return models.Wishlist{}, nil, errors.New("wishlist not found")
	}
	if err != nil {
		return models.Wishlist{}, nil, err
	}

	cur, err := r.itemsCol.Find(ctx, bson.M{"wishlistId": id})
	if err != nil {
		return models.Wishlist{}, nil, err
	}
	defer cur.Close(ctx)

	items := []models.WishlistItem{}
	for cur.Next(ctx) {
		var it models.WishlistItem
		if cur.Decode(&it) == nil {
			items = append(items, it)
		}
	}

	return w, items, nil
}

func (r *WishlistRepo) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	res, err := r.wishlistsCol.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("wishlist not found")
	}

	_, _ = r.itemsCol.DeleteMany(ctx, bson.M{"wishlistId": id})
	return nil
}

func (r *WishlistRepo) AddItem(wishlistID int, bookID int, qty int) (models.WishlistItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if qty <= 0 {
		return models.WishlistItem{}, errors.New("qty must be > 0")
	}

	err := r.wishlistsCol.FindOne(ctx, bson.M{"id": wishlistID}).Err()
	if err == mongo.ErrNoDocuments {
		return models.WishlistItem{}, errors.New("wishlist not found")
	}
	if err != nil {
		return models.WishlistItem{}, err
	}

	res := r.itemsCol.FindOneAndUpdate(
		ctx,
		bson.M{"wishlistId": wishlistID, "bookId": bookID},
		bson.M{"$inc": bson.M{"qty": qty}},
	)
	if res.Err() == nil {
		var updated models.WishlistItem
		_ = res.Decode(&updated)
		return updated, nil
	}
	if res.Err() != mongo.ErrNoDocuments {
		return models.WishlistItem{}, res.Err()
	}

	itemID, err := r.counters.Next("wishlist_items")
	if err != nil {
		return models.WishlistItem{}, err
	}

	it := models.WishlistItem{
		ID:         itemID,
		WishlistID: wishlistID,
		BookID:     bookID,
		Qty:        qty,
	}

	_, err = r.itemsCol.InsertOne(ctx, it)
	if err != nil {
		return models.WishlistItem{}, err
	}

	return it, nil
}

func (r *WishlistRepo) DeleteItem(wishlistID int, itemID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	res, err := r.itemsCol.DeleteOne(ctx, bson.M{"wishlistId": wishlistID, "id": itemID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("item not found")
	}
	return nil
}
