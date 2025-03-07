package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ActionHistory struct {
	ID            bson.ObjectID `json:"id,omitempty" bson:"_id"`
	RepoID        bson.ObjectID `json:"repo_id,omitempty" bson:"repo_id"`
	Repo          GitRepo       `json:"repo,omitempty" bson:"repo"`
	ContribID     bson.ObjectID `json:"contrib_id,omitempty" bson:"contrib_id"`
	Contributor   Contributor   `json:"contributor,omitempty" bson:"contributor"`
	PullRequestID bson.ObjectID `json:"pull_request_id,omitempty" bson:"pull_request_id"`
	PullRequest   PullRequest   `json:"pull_request,omitempty" bson:"pull_request"`
	Action        string        `json:"action,omitempty" bson:"action"`
	CreatedAt     time.Time     `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at,omitempty" bson:"updated_at"`
}
