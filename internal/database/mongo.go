package database

import (
	"context"
	"fmt"
	"strconv"

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
	creds := options.Credential{
		AuthSource: "admin",
		Username:   config.MongoUsername,
		Password:   config.MongoPassword,
	}
	log.Info("Connecting to MongoDB...")
	p, err := strconv.Atoi(config.MongoPort)
	if err != nil {
		log.Fatal("Port is not int")
		panic(err)
	}
	client, err := mongo.Connect(
		options.Client().ApplyURI(
			fmt.Sprintf(
				"mongodb://%s:%d",
				config.MongoHost,
				p,
			),
		).SetAuth(creds),
	)

	if err != nil {
		log.Fatal("Failed connect to MongoDB", zap.Error(err))
		panic(err)
	}

	var result bson.M

	if err := client.Database("ryakadevforum").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		log.Fatal("Failed ping to MongoDB", zap.Error(err))
	}

	log.Info("Successfully connected to MongoDB")

	return client
}
