package repository

import (
	"context"
	"errors"

	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CreateContributor implements Repository.
func (r *repository) CreateContributor(ctx context.Context, payload *model.CmdContributor) (*mongo.InsertOneResult, error) {
	res, err := r.mongo.Collection(r.col.Contributors).InsertOne(ctx, payload)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetContributor implements Repository.
func (r *repository) GetContributor(ctx context.Context, filter *bson.M) (model.Contributor, error) {
	var contrib model.Contributor

	collection := r.mongo.Collection(r.col.Contributors)
	err := collection.FindOne(ctx, filter).Decode(&contrib)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Contributor{}, errors.New("Contributor not found")
		}
		return model.Contributor{}, err
	}

	return contrib, nil
}

// GetContributors implements Repository.
func (r *repository) GetContributors(ctx context.Context, offset int64, limit int64, filter *bson.M) ([]model.Contributor, error) {
	var contributors []model.Contributor
	cursor, err := r.mongo.Collection(r.col.Contributors).Find(ctx, filter, options.Find().SetSkip(offset).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var contrib model.Contributor
		if err := cursor.Decode(&contrib); err != nil {
			return nil, err
		}
		contributors = append(contributors, contrib)
	}
	return contributors, nil
}

// UpdateContributor implements Repository.
func (r *repository) UpdateContributor(ctx context.Context, payload *model.CmdContributor, filter *bson.M) (*mongo.UpdateResult, error) {
	collection := r.mongo.Collection(r.col.Contributors)
	update := bson.M{"$set": payload}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return res, nil
}
