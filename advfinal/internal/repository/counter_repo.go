package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CounterRepo struct {
	col *mongo.Collection
}

func NewCounterRepo(db *mongo.Database) *CounterRepo {
	return &CounterRepo{col: db.Collection("counters")}
}

func (r *CounterRepo) Next(name string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	res := r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": name},
		bson.M{"$inc": bson.M{"seq": 1}},
		opts,
	)

	var out struct {
		Seq int `bson:"seq"`
	}
	err := res.Decode(&out)
	return out.Seq, err
}
