package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PointHistory struct {
	ID              bson.ObjectID `json:"id,omitempty" bson:"_id"`
	CmdPointHistory `bson:",inline"`
}
type CmdPointHistory struct {
	ActionHistory CmdActionHistory `json:"action_history,omitempty" bson:"action_history"`
	Event         string           `json:"event,omitempty" bson:"event"`
	Point         int64            `json:"point,omitempty" bson:"point"`
	CreatedAt     time.Time        `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at,omitempty" bson:"updated_at"`
}
