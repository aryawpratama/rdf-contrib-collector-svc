package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GitRepo struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	FullName  string             `json:"full_name,omitempty" bson:"full_name"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
