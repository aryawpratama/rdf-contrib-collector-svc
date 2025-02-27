package database

import (
	"context"
	"fmt"

	"github.com/ryakadev/rdf-contrib-collector/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

func NewConnection(ctx context.Context, config config.Config, log *zap.Logger) *mongo.Client {
	client, err := mongo.Connect(
		options.Client().ApplyURI(
			fmt.Sprintf(
				"mongodb://%s:%s@%s:%d",
				config.Mongo.Username,
				config.Mongo.Password,
				config.Mongo.Host,
				config.Mongo.Port,
			),
		),
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
