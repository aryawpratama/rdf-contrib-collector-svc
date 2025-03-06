package repository

import (
	"context"

	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *repository) CreateActionHistory(ctx context.Context, payload *model.ActionHistory) (*mongo.InsertOneResult, error) {
	res, err := r.mongo.Collection(r.col.ActionHistories).InsertOne(ctx, payload)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *repository) GetActionHistories(ctx context.Context, offset int64, limit int64, filter *model.ActionHistory) ([]model.ActionHistory, error) {
	var histories []model.ActionHistory
	cursor, err := r.mongo.Collection(r.col.ActionHistories).Find(ctx, filter, options.Find().SetSkip(offset).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var history model.ActionHistory
		if err := cursor.Decode(&history); err != nil {
			return nil, err
		}
		histories = append(histories, history)
	}
	return histories, nil
}

func (r *repository) GetActionHistory(ctx context.Context, filter *model.ActionHistory) (model.ActionHistory, error) {
	var history model.ActionHistory

	collection := r.mongo.Collection(r.col.ActionHistories)
	err := collection.FindOne(ctx, filter).Decode(&history)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.ActionHistory{}, nil
		}
		return model.ActionHistory{}, err
	}

	return history, nil
}
