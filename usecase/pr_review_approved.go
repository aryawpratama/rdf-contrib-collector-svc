package usecase

import (
	"context"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (u usecase) PullRequestReviewApproved(ctx context.Context, evt *github.PullRequestReviewEvent, point *model.PointActionData) error {
	u.log.Info("Pull Request Review Event Called")
	var prrt = model.Webhook{
		Avatar:         evt.PullRequest.User.GetAvatarURL(),
		RepoName:       evt.PullRequest.Base.Repo.GetFullName(),
		SrcPrUrl:       evt.PullRequest.Head.Repo.GetHTMLURL(),
		PrUrl:          evt.PullRequest.GetHTMLURL(),
		Action:         evt.Review.GetState(),
		HRef:           evt.PullRequest.Head.GetRef(),
		BRef:           evt.PullRequest.Base.GetRef(),
		ContribUname:   evt.PullRequest.User.GetLogin(),
		ContribUrl:     evt.PullRequest.User.GetHTMLURL(),
		ApproveComment: evt.Review.GetBody(),
		ApproveUrl:     evt.Review.GetHTMLURL(),
	}

	// If PR Review state is not approved, pass
	if prrt.Action != "approved" {
		u.log.Info("PR Review submit is not in approved state")
		return nil
	}

	// Contributor
	contribData := model.CmdContributor{
		Username:   prrt.ContribUname,
		Avatar:     prrt.Avatar,
		ProfileURL: prrt.ContribUrl,
		IsLead:     false,
		IsCTO:      false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	c, err := u.repo.GetContributor(ctx, &bson.M{"username": prrt.ContribUname})
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

	// If contributor is not lead, pass
	if !c.IsLead {
		u.log.Info("PR Approval not from lead! passing point addition")
		return nil
	}

	// Repository
	repoData := model.CmdGitRepo{
		FullName:  prrt.RepoName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	r, err := u.repo.GetGitRepo(ctx, &bson.M{
		"full_name": prrt.RepoName,
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

	// Pull Request
	prData := model.CmdPullRequest{
		Contributor:       c,
		Repo:              r,
		PullRequestURL:    prrt.PrUrl,
		SrcPullRequestURL: prrt.SrcPrUrl,
		SrcBranch:         prrt.HRef,
		DstBranch:         prrt.BRef,
		Action:            prrt.Action,
		IsMerged:          prrt.IsMerged,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	pr, err := u.repo.GetPullRequest(ctx, &bson.M{"pull_request_url": prrt.PrUrl})
	if err != nil {
		if err.Error() == "Pull request not found" && prrt.Action == "resolved" {
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
		Event:          "pull_request_review",
		Action:         prrt.Action,
		ApproveComment: prrt.ApproveComment,
		ApproveUrl:     prrt.ApproveUrl,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = u.repo.CreateActionHistory(ctx, &ahModel)
	if err != nil {
		u.log.Error(err.Error())
		return err
	}
	pointData := model.CmdPoint{
		Contributor: c,
		Point:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	p, err := u.repo.GetPoint(ctx, &bson.M{"contributor._id": c.ID})
	if err != nil {
		if err.Error() == "Point not found" {
			res, err := u.repo.CreatePoint(ctx, &pointData)
			if err != nil {
				u.log.Error(err.Error())
				return err
			}
			ID, _ := res.InsertedID.(bson.ObjectID)
			p = model.Point{CmdPoint: pointData, ID: ID}
			u.log.Info("Point Created")
		} else {
			u.log.Error(err.Error())
			return err

		}
	}
	_, err = u.repo.CreatePointHistory(ctx, &model.CmdPointHistory{
		ActionHistory: ahModel,
		Point:         int64(point.ResolveComment),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})
	if err != nil {
		u.log.Error(err.Error())
		return err
	}

	_, err = u.repo.UpdatePoint(ctx, &model.CmdPoint{
		Contributor: c,
		Point:       p.Point + int64(point.ResolveComment),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   time.Now(),
	}, &bson.M{"_id": p.ID})
	if err != nil {
		u.log.Error(err.Error())
		return err
	}
	return nil
}
