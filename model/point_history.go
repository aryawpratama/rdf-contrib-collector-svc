package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PointHistory struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	ActionHistoryId primitive.ObjectID `json:"action_history_id,omitempty" bson:"action_history_id"`
	ActionHistory   ActionHistory      `json:"action_history,omitempty"`
	Point           int64              `json:"point,omitempty" bson:"point"`
	CreatedAt       time.Time          `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at,omitempty" bson:"updated_at"`
}
