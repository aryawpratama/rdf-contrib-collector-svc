package usecase

import (
	"context"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (u usecase) PullRequestReviewCommentEvent(ctx context.Context, evt *github.PullRequestReviewCommentEvent, point *model.PointActionData) error {
	// PR COMMENT CREATED EVENT
	u.log.Info("Pull Request Review Comment Event Called")
	var prrc = model.Webhook{
		Avatar:       evt.Comment.User.GetAvatarURL(),
		RepoName:     evt.PullRequest.Base.Repo.GetFullName(),
		PrUrl:        evt.PullRequest.GetHTMLURL(),
		SrcPrUrl:     evt.PullRequest.Head.Repo.GetHTMLURL(),
		Action:       evt.GetAction(),
		HRef:         evt.PullRequest.Head.GetRef(),
		BRef:         evt.PullRequest.Base.GetRef(),
		ContribUname: evt.Comment.User.GetLogin(),
		ContribUrl:   evt.Comment.User.GetHTMLURL(),
		Comment:      evt.Comment.GetBody(),
		CommentUrl:   evt.Comment.GetHTMLURL(),
		DiffHunk:     evt.Comment.GetDiffHunk(),
	}

	// Repository
	repoData := model.CmdGitRepo{
		FullName:  prrc.RepoName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	r, err := u.repo.GetGitRepo(ctx, &bson.M{
		"full_name": prrc.RepoName,
	})
	if err != nil {
		if err.Error() == "GitRepo not found" {
			res, err := u.repo.CreateGitRepo(ctx, &repoData)
			if err != nil {
				u.log.Error(err.Error())
				return err
			}
			ID, _ := res.InsertedID.(bson.ObjectID)
			r = model.GitRepo{
				ID:         ID,
				CmdGitRepo: repoData,
			}
			u.log.Info("GitRepo Created")
		} else {
			u.log.Error(err.Error())
			return err
		}
	}

	// Contributor
	contribData := model.CmdContributor{
		Username:   prrc.ContribUname,
		Avatar:     prrc.Avatar,
		ProfileURL: prrc.ContribUrl,
		IsLead:     false,
		IsCTO:      false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	c, err := u.repo.GetContributor(ctx, &bson.M{"username": prrc.ContribUname})
	if err != nil {
		if err.Error() == "Contributor not found" {
			res, err := u.repo.CreateContributor(ctx, &contribData)
			if err != nil {
				u.log.Error(err.Error())
				return err
			}
			ID, _ := res.InsertedID.(bson.ObjectID)
			c = model.Contributor{CmdContributor: contribData, ID: ID}
			u.log.Info("Contributor Created")
		} else {
			u.log.Error(err.Error())
			return err
		}
	}

	// Pull Request
	prData := model.CmdPullRequest{
		Contributor:       c,
		Repo:              r,
		PullRequestURL:    prrc.PrUrl,
		SrcPullRequestURL: prrc.SrcPrUrl,
		SrcBranch:         prrc.HRef,
		DstBranch:         prrc.BRef,
		Action:            prrc.Action,
		IsMerged:          prrc.IsMerged,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	pr, err := u.repo.GetPullRequest(ctx, &bson.M{"pull_request_url": prrc.PrUrl})
	if err != nil {
		if err.Error() == "Pull request not found" && prrc.Action == "opened" {
			res, err := u.repo.CreatePullRequest(ctx, &prData)
			if err != nil {
				u.log.Error(err.Error())
				return err
			}
			ID, _ := res.InsertedID.(bson.ObjectID)
			pr = model.PullRequest{CmdPullRequest: prData, ID: ID}
			u.log.Info("PR Created")
		} else {
			u.log.Error(err.Error())
			return err
		}
	}
	ahModel := model.CmdActionHistory{
		Repo:           r,
		Contributor:    c,
		PullRequest:    &pr,
		Event:          "pull_request_review_comment",
		Action:         prrc.Action,
		CommentContent: prrc.Comment,
		CommentUrl:     prrc.CommentUrl,
		DiffHunk:       prrc.DiffHunk,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = u.repo.CreateActionHistory(ctx, &ahModel)
	if err != nil {
		u.log.Error(err.Error())
		return err
	}
	lPointData := model.CmdPoint{
		Contributor: c,
		Point:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	p, err := u.repo.GetPoint(ctx, &bson.M{"contributor._id": c.ID})
	if err != nil {
		if err.Error() == "Point not found" {
			res, err := u.repo.CreatePoint(ctx, &lPointData)
			if err != nil {
				u.log.Error(err.Error())
				return err
			}
			ID, _ := res.InsertedID.(bson.ObjectID)
			p = model.Point{CmdPoint: lPointData, ID: ID}
		} else {
			u.log.Error(err.Error())
			return err

		}
	}
	var contribPoint int64
	if c.IsLead {
		contribPoint = int64(point.CommentLead)
	} else {
		contribPoint = int64(point.CommentContrib)
	}
	_, err = u.repo.CreatePointHistory(ctx, &model.CmdPointHistory{
		ActionHistory: ahModel,
		Point:         int64(contribPoint),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})
	if err != nil {
		u.log.Error(err.Error())
		return err
	}
	_, err = u.repo.UpdatePoint(ctx, &model.CmdPoint{
		Contributor: c,
		Point:       p.Point + int64(contribPoint),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   time.Now(),
	}, &bson.M{"_id": p.ID})
	if err != nil {
		u.log.Error(err.Error())
		return err
	}
	u.log.Info("Success Insert Point")
	return nil
} // PR COMMENT CREATED EVENT END
