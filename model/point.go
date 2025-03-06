package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Point struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	ContribID primitive.ObjectID `json:"contrib_id,omitempty" bson:"contrib_id"`
	Point     int64              `json:"point,omitempty" bson:"point"`
	CreatedAt time.Time          `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at,omitempty" bson:"updated_at"`
}
