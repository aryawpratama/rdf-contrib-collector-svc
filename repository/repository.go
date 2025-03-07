package repository

import (
	"context"

	"github.com/ryakadev/rdf-contrib-collector/internal/database"
	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository interface {
	CreateActionHistory(ctx context.Context, payload *model.CmdActionHistory) (*mongo.InsertOneResult, error)
	GetActionHistory(ctx context.Context, filter *bson.M) (model.ActionHistory, error)
	GetActionHistories(ctx context.Context, offset int64, limit int64, filter *bson.M) ([]model.ActionHistory, error)

	CreateContributor(ctx context.Context, payload *model.CmdContributor) (*mongo.InsertOneResult, error)
	UpdateContributor(ctx context.Context, payload *model.CmdContributor, filter *bson.M) (*mongo.UpdateResult, error)
	GetContributor(ctx context.Context, filter *bson.M) (model.Contributor, error)
	GetContributors(ctx context.Context, offset int64, limit int64, filter *bson.M) ([]model.Contributor, error)

	CreatePoint(ctx context.Context, payload *model.CmdPoint) (*mongo.InsertOneResult, error)
	UpdatePoint(ctx context.Context, payload *model.CmdPoint, filter *bson.M) (*mongo.UpdateResult, error)
	GetPoint(ctx context.Context, filter *bson.M) (model.Point, error)
	GetPoints(ctx context.Context, offset int64, limit int64, filter *bson.M) ([]model.Point, error)

	CreatePointHistory(ctx context.Context, payload *model.CmdPointHistory) (*mongo.InsertOneResult, error)
	GetPointHistory(ctx context.Context, filter *bson.M) (model.PointHistory, error)
	GetPointHistories(ctx context.Context, offset int64, limit int64, filter *bson.M) ([]model.PointHistory, error)

	CreatePullRequest(ctx context.Context, payload *model.CmdPullRequest) (*mongo.InsertOneResult, error)
	UpdatePullRequest(ctx context.Context, payload *model.CmdPullRequest, filter *bson.M) (*mongo.UpdateResult, error)
	GetPullRequest(ctx context.Context, filter *bson.M) (model.PullRequest, error)
	GetPullRequests(ctx context.Context, offset int64, limit int64, filter *bson.M) ([]model.PullRequest, error)

	CreateGitRepo(ctx context.Context, payload *model.CmdGitRepo) (*mongo.InsertOneResult, error)
	UpdateGitRepo(ctx context.Context, payload *model.CmdGitRepo, filter *bson.M) (*mongo.UpdateResult, error)
	GetGitRepo(ctx context.Context, filter *bson.M) (model.GitRepo, error)
	GetGitRepos(ctx context.Context, offset int64, limit int64, filter *bson.M) ([]model.GitRepo, error)
}

type repository struct {
	mongo *mongo.Database
	col   database.MongoDBCollections
}

func New(mongo *mongo.Client) Repository {
	col := database.MongoDBCollections{
		ActionHistories: "action_histories",
		Contributors:    "contributors",
		GitRepos:        "git_repos",
		Points:          "points",
		PointHistories:  "point_histories",
		PullRequests:    "pull_requests",
	}
	m := mongo.Database("rdf")
	return &repository{
		mongo: m,
		col:   col,
	}
}
