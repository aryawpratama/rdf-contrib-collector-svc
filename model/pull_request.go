package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PullRequest struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	RepoID         primitive.ObjectID `json:"repo_id,omitempty" bson:"repo_id"`
	Repo           GitRepo            `json:"repo,omitempty" bson:"repo"`
	ContributorID  primitive.ObjectID `json:"contributor_id,omitempty" bson:"contributor_id"`
	Contributor    Contributor        `json:"contributor,omitempty" bson:"contributor"`
	MergedByID     primitive.ObjectID `json:"merged_by_id,omitempty" bson:"merged_by_id"`
	MergedBy       Contributor        `json:"merged_by,omitempty" bson:"merged_by"`
	PullRequestURL string             `json:"pull_request_url,omitempty" bson:"pull_request_url"`
	SrcBranch      string             `json:"src_branch,omitempty" bson:"src_branch"`
	DstBranch      string             `json:"dst_branch,omitempty" bson:"dst_branch"`
	Action         string             `json:"action,omitempty" bson:"action"`
	IsMerged       bool               `json:"is_merged,omitempty" bson:"is_merged"`
	CreatedAt      time.Time          `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at,omitempty" bson:"updated_at"`
}
