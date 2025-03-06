package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Contributor struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	Username   string             `json:"username,omitempty" bson:"username"`
	Avatar     string             `json:"avatar,omitempty" bson:"avatar"`
	ProfileURL string             `json:"profile_url,omitempty" bson:"profile_url"`
	CreatedAt  time.Time          `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at,omitempty" bson:"updated_at"`
}
