package repository

import (
	"context"
	"errors"
	"time"

	"bookstore/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookRepository interface {
	Create(book models.Book) (models.Book, error)
	GetByID(id int) (models.Book, error)
	GetAll() []models.Book
	Update(book models.Book) error
	Delete(id int) error
}

type BookRepo struct {
	col      *mongo.Collection
	counters *CounterRepo
}

func NewBookRepo(db *mongo.Database) *BookRepo {
	return &BookRepo{
		col:      db.Collection("books"),
		counters: NewCounterRepo(db),
	}
}

func (r *BookRepo) Create(book models.Book) (models.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, err := r.counters.Next("books")
	if err != nil {
		return models.Book{}, err
	}
	book.ID = id

	_, err = r.col.InsertOne(ctx, book)
	if err != nil {
		return models.Book{}, err
	}

	return book, nil
}

func (r *BookRepo) GetByID(id int) (models.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var b models.Book
	err := r.col.FindOne(ctx, bson.M{"id": id}).Decode(&b)
	if err == mongo.ErrNoDocuments {
		return models.Book{}, errors.New("book not found")
	}
	return b, err
}

func (r *BookRepo) GetAll() []models.Book {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return []models.Book{}
	}
	defer cur.Close(ctx)

	out := []models.Book{}
	for cur.Next(ctx) {
		var b models.Book
		if cur.Decode(&b) == nil {
			out = append(out, b)
		}
	}
	return out
}

func (r *BookRepo) Update(book models.Book) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.col.UpdateOne(ctx, bson.M{"id": book.ID}, bson.M{"$set": book})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("book not found")
	}
	return nil
}

func (r *BookRepo) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.col.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("book not found")
	}
	return nil
}
