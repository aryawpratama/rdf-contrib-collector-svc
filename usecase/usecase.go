package usecase

import (
	"context"

	"github.com/google/go-github/v69/github"
	"github.com/ryakadev/rdf-contrib-collector/model"
	"github.com/ryakadev/rdf-contrib-collector/repository"
	"go.uber.org/zap"
)

type UseCase interface {
	PullRequestEvent(ctx context.Context, evt *github.PullRequestEvent, point *model.PointActionData) error
	ForkEvent(ctx context.Context, evt *github.ForkEvent, point *model.PointActionData) error
	PullRequestReviewCommentEvent(ctx context.Context, evt *github.PullRequestReviewCommentEvent, point *model.PointActionData) error
	PullRequestReviewThreadEvent(ctx context.Context, evt *github.PullRequestReviewThreadEvent, point *model.PointActionData) error
	PullRequestReviewApproved(ctx context.Context, evt *github.PullRequestReviewEvent, point *model.PointActionData) error
}

type usecase struct {
	repo repository.Repository
	log  *zap.Logger
}

func New(repo repository.Repository, log *zap.Logger) UseCase {
	return &usecase{
		repo: repo,
		log:  log,
	}
}
