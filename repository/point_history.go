package repository

import (
	"context"
	"errors"

	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CreatePointHistory implements Repository.
func (r *repository) CreatePointHistory(ctx context.Context, payload *model.CmdPointHistory) (*mongo.InsertOneResult, error) {
	res, err := r.mongo.Collection(r.col.PointHistories).InsertOne(ctx, payload)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetPointHistories implements Repository.
func (r *repository) GetPointHistories(ctx context.Context, offset int64, limit int64, filter *bson.M) ([]model.PointHistory, error) {
	var pHistories []model.PointHistory
	cursor, err := r.mongo.Collection(r.col.PointHistories).Find(ctx, filter, options.Find().SetSkip(offset).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var history model.PointHistory
		if err := cursor.Decode(&history); err != nil {
			return nil, err
		}
		pHistories = append(pHistories, history)
	}
	return pHistories, nil
}

// GetPointHistory implements Repository.
func (r *repository) GetPointHistory(ctx context.Context, filter *bson.M) (model.PointHistory, error) {
	var ph model.PointHistory

	collection := r.mongo.Collection(r.col.PointHistories)
	err := collection.FindOne(ctx, filter).Decode(&ph)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.PointHistory{}, errors.New("Point History not found")
		}
		return model.PointHistory{}, err
	}

	return ph, nil
}
