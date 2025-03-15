package database

import (
	"context"

	"github.com/ryakadev/rdf-contrib-collector/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type MongoDBCollections struct {
	ActionHistories string
	Contributors    string
	GitRepos        string
	Points          string
	PointHistories  string
	PullRequests    string
}

func NewConnection(ctx context.Context, config config.Config, log *zap.Logger) *mongo.Client {

	log.Info("Connecting to MongoDB...")
	client, err := mongo.Connect(
		options.Client().ApplyURI(
			config.MongoStringConn,
		),
	)
	if err != nil {
		log.Fatal("Failed connect to MongoDB", zap.Error(err))
		panic(err)
	}

	var result bson.M

	if err := client.Database("rdf").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		log.Fatal("Failed ping to MongoDB", zap.Error(err))
	}

	log.Info("Successfully connected to MongoDB")

	return client
}
