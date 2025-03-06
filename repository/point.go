package repository

import (
	"context"

	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CreatePoint implements Repository.
func (r *repository) CreatePoint(ctx context.Context, payload *model.Point) (*mongo.InsertOneResult, error) {
	res, err := r.mongo.Collection(r.col.Points).InsertOne(ctx, payload)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// UpdatePoint implements Repository.
func (r *repository) UpdatePoint(ctx context.Context, payload *model.Point, filter *model.Point) (*mongo.UpdateResult, error) {
	collection := r.mongo.Collection(r.col.Points)
	update := bson.M{"$set": payload}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetPoint implements Repository.
func (r *repository) GetPoint(ctx context.Context, filter *model.Point) (model.Point, error) {
	var p model.Point

	collection := r.mongo.Collection(r.col.Points)
	err := collection.FindOne(ctx, filter).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Point{}, nil
		}
		return model.Point{}, err
	}

	return p, nil
}

// GetPoints implements Repository.
func (r *repository) GetPoints(ctx context.Context, offset int64, limit int64, filter *model.Point) ([]model.Point, error) {
	var pHistories []model.Point
	cursor, err := r.mongo.Collection(r.col.Points).Find(ctx, filter, options.Find().SetSkip(offset).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var history model.Point
		if err := cursor.Decode(&history); err != nil {
			return nil, err
		}
		pHistories = append(pHistories, history)
	}
	return pHistories, nil
}
