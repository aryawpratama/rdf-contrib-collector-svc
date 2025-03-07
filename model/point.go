package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Point struct {
	ID          bson.ObjectID `json:"id,omitempty" bson:"_id"`
	ContribID   bson.ObjectID `json:"contrib_id,omitempty" bson:"contrib_id"`
	Contributor Contributor   `json:"contributor,omitempty"`
	Point       int64         `json:"point,omitempty" bson:"point"`
	CreatedAt   time.Time     `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at,omitempty" bson:"updated_at"`
}

type PointActionData struct {
	CreatePR       int
	ForkRepo       int
	ResolveComment int
	MergeContrib   int
	MergeLead      int
	CommentLead    int
}
