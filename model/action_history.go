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
	Repo                    GitRepo      `json:"repo,omitempty" bson:"repo"`
	Contributor             Contributor  `json:"contributor,omitempty" bson:"contributor"`
	PullRequest             *PullRequest `json:"pull_request,omitempty" bson:"pull_request"`
	Action                  string       `json:"action,omitempty" bson:"action"`
	Event                   string       `json:"event,omitempty" bson:"event"`
	ApproveComment          string       `json:"approve_comment,omitempty" bson:"approve_comment"`
	ApproveUrl              string       `json:"approve_url,omitempty" bson:"approve_url"`
	CommentUrl              string       `json:"comment_url,omitempty" bson:"comment_url"`
	CommentContent          string       `json:"comment_content,omitempty" bson:"comment_content"`
	DiffHunk                string       `json:"diff_hunk,omitempty" bson:"diff_hunk"`
	IsReportedAbuse         bool         `json:"is_reported_abuse" bson:"is_reported_abuse"`
	IsReportedAbuseApproved bool         `json:"is_reported_abuse_approved" bson:"is_reported_abuse_approved"`
	CreatedAt               time.Time    `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt               time.Time    `json:"updated_at,omitempty" bson:"updated_at"`
}
