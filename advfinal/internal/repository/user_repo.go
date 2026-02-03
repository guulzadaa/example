package repository

import (
	"context"
	"errors"
	"time"

	"bookstore/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	Create(user models.User) error
	GetByEmail(email string) (models.User, error)
	GetByID(id int) (models.User, error)
	Update(user models.User) error
}

type UserRepo struct {
	col      *mongo.Collection
	counters *CounterRepo
}

func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{
		col:      db.Collection("users"),
		counters: NewCounterRepo(db),
	}
}

func (r *UserRepo) Create(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if user.Email == "" {
		return errors.New("email required")
	}
	if user.Password == "" {
		return errors.New("password required")
	}
	if user.Role == "" {
		user.Role = "user"
	}

	exists, err := r.col.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("email already exists")
	}

	id, err := r.counters.Next("users")
	if err != nil {
		return err
	}
	user.ID = id

	_, err = r.col.InsertOne(ctx, user)
	return err
}

func (r *UserRepo) GetByEmail(email string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var u models.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return models.User{}, errors.New("user not found")
	}
	return u, err
}

func (r *UserRepo) GetByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var u models.User
	err := r.col.FindOne(ctx, bson.M{"id": id}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return models.User{}, errors.New("user not found")
	}
	return u, err
}
func (r *UserRepo) Update(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if user.ID <= 0 {
		return errors.New("invalid user id")
	}

	res, err := r.col.UpdateOne(
		ctx,
		bson.M{"id": user.ID},
		bson.M{"$set": user},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}
