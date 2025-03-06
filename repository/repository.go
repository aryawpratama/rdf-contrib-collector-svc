package repository

import (
	"context"

	"github.com/ryakadev/rdf-contrib-collector/internal/database"
	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository interface {
	CreateActionHistory(ctx context.Context, payload *model.ActionHistory) (*mongo.InsertOneResult, error)
	GetActionHistory(ctx context.Context, filter *model.ActionHistory) (model.ActionHistory, error)
	GetActionHistories(ctx context.Context, offset int64, limit int64, filter *model.ActionHistory) ([]model.ActionHistory, error)

	CreateContributor(ctx context.Context, payload *model.Contributor) (*mongo.InsertOneResult, error)
	UpdateContributor(ctx context.Context, payload *model.Contributor) (*mongo.UpdateResult, error)
	GetContributor(ctx context.Context, filter *model.Contributor) (model.Contributor, error)
	GetContributors(ctx context.Context, offset int64, limit int64, filter *model.Contributor) ([]model.Contributor, error)

	CreatePoint(ctx context.Context, payload *model.Point) (*mongo.InsertOneResult, error)
	UpdatePoint(ctx context.Context, payload *model.Point) (*mongo.UpdateResult, error)
	GetPoint(ctx context.Context, filter *model.Point) (model.Point, error)
	GetPoints(ctx context.Context, offset int64, limit int64, filter *model.Point) ([]model.Point, error)

	CreatePointHistory(ctx context.Context, payload *model.PointHistory) (*mongo.InsertOneResult, error)
	GetPointHistory(ctx context.Context, filter *model.PointHistory) (model.PointHistory, error)
	GetPointHistories(ctx context.Context, offset int64, limit int64, filter *model.PointHistory) ([]model.PointHistory, error)

	CreatePullRequest(ctx context.Context, payload *model.PullRequest) (*mongo.InsertOneResult, error)
	UpdatePullRequest(ctx context.Context, payload *model.PullRequest) (*mongo.UpdateResult, error)
	GetPullRequest(ctx context.Context, filter *model.PullRequest) (model.PullRequest, error)
	GetPullRequests(ctx context.Context, offset int64, limit int64, filter *model.PullRequest) ([]model.PullRequest, error)

	CreateGitRepo(ctx context.Context, payload *model.GitRepo) (*mongo.InsertOneResult, error)
	UpdateGitRepo(ctx context.Context, payload *model.GitRepo) (*mongo.UpdateResult, error)
	GetGitRepo(ctx context.Context, filter *model.GitRepo) (model.GitRepo, error)
	GetGitRepos(ctx context.Context, offset int64, limit int64, filter *model.GitRepo) ([]model.GitRepo, error)
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
