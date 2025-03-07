package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ActionHistory struct {
	ID               bson.ObjectID `json:"id,omitempty" bson:"_id"`
	CmdActionHistory `bson:",inline"`
}
type CmdActionHistory struct {
	Repo        GitRepo      `json:"repo,omitempty" bson:"repo"`
	Contributor Contributor  `json:"contributor,omitempty" bson:"contributor"`
	PullRequest *PullRequest `json:"pull_request,omitempty" bson:"pull_request"`
	Action      string       `json:"action,omitempty" bson:"action"`
	CreatedAt   time.Time    `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at,omitempty" bson:"updated_at"`
}
