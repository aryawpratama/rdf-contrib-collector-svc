package repository

import (
	"context"

	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CreatePullRequest implements Repository.
func (r *repository) CreatePullRequest(ctx context.Context, payload *model.PullRequest) (*mongo.InsertOneResult, error) {
	res, err := r.mongo.Collection(r.col.PullRequests).InsertOne(ctx, payload)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetPullRequest implements Repository.
func (r *repository) GetPullRequest(ctx context.Context, filter *model.PullRequest) (model.PullRequest, error) {
	var p model.PullRequest

	collection := r.mongo.Collection(r.col.PullRequests)
	err := collection.FindOne(ctx, filter).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.PullRequest{}, nil
		}
		return model.PullRequest{}, err
	}

	return p, nil
}

// GetPullRequests implements Repository.
func (r *repository) GetPullRequests(ctx context.Context, offset int64, limit int64, filter *model.PullRequest) ([]model.PullRequest, error) {
	var pr []model.PullRequest
	cursor, err := r.mongo.Collection(r.col.Points).Find(ctx, filter, options.Find().SetSkip(offset).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var pRequest model.PullRequest
		if err := cursor.Decode(&pRequest); err != nil {
			return nil, err
		}
		pr = append(pr, pRequest)
	}
	return pr, nil
}

// UpdatePullRequest implements Repository.
func (r *repository) UpdatePullRequest(ctx context.Context, payload *model.PullRequest, filter *model.PullRequest) (*mongo.UpdateResult, error) {
	collection := r.mongo.Collection(r.col.PullRequests)
	update := bson.M{"$set": payload}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return res, nil
}
