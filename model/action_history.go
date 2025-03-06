package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActionHistory struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	RepoID        primitive.ObjectID `json:"repo_id,omitempty" bson:"repo_id"`
	ContribID     primitive.ObjectID `json:"contrib_id,omitempty" bson:"contrib_id"`
	PullRequestID primitive.ObjectID `json:"pull_request_id,omitempty" bson:"pull_request_id"`
	Action        string             `json:"action,omitempty" bson:"action"`
	CreatedAt     time.Time          `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at,omitempty" bson:"updated_at"`
}
