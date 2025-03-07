package repository

import (
	"context"
	"errors"

	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *repository) CreateGitRepo(ctx context.Context, payload *model.GitRepo) (*mongo.InsertOneResult, error) {
	res, err := r.mongo.Collection(r.col.GitRepos).InsertOne(ctx, payload)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *repository) GetGitRepo(ctx context.Context, filter *model.GitRepo) (model.GitRepo, error) {
	var contrib model.GitRepo

	collection := r.mongo.Collection(r.col.GitRepos)
	err := collection.FindOne(ctx, filter).Decode(&contrib)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.GitRepo{}, errors.New("GitRepo not found")
		}
		return model.GitRepo{}, err
	}

	return contrib, nil
}

func (r *repository) GetGitRepos(ctx context.Context, offset int64, limit int64, filter *model.GitRepo) ([]model.GitRepo, error) {
	var repos []model.GitRepo
	cursor, err := r.mongo.Collection(r.col.GitRepos).Find(ctx, filter, options.Find().SetSkip(offset).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var repo model.GitRepo
		if err := cursor.Decode(&repo); err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func (r *repository) UpdateGitRepo(ctx context.Context, payload *model.GitRepo, filter *model.GitRepo) (*mongo.UpdateResult, error) {
	collection := r.mongo.Collection(r.col.GitRepos)
	update := bson.M{"$set": payload}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return res, nil
}
