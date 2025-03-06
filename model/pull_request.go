package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PullRequest struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	ContributorID  primitive.ObjectID `json:"contributor_id,omitempty" bson:"contributor_id"`
	PullRequestURL string             `json:"pull_request_url,omitempty" bson:"pull_request_url"`
	SrcBranch      string             `json:"src_branch,omitempty" bson:"src_branch"`
	DstBranch      string             `json:"dst_branch,omitempty" bson:"dst_branch"`
	Status         string             `json:"status,omitempty" bson:"status"`
	CreatedAt      time.Time          `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at,omitempty" bson:"updated_at"`
}
