package usecase

import (
	"context"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (u usecase) PullRequestEvent(ctx context.Context, evt *github.PullRequestEvent, point *model.PointActionData) error {
	u.log.Info("PR Event Called")
	var contribPoint int64 = 0
	var leadPoint int64 = 0
	w := model.Webhook{
		RepoName:     evt.Repo.GetFullName(),
		PrUrl:        evt.PullRequest.GetHTMLURL(),
		SrcPrUrl:     evt.PullRequest.Head.Repo.GetHTMLURL(),
		ContribUname: evt.PullRequest.User.GetLogin(),
		HRef:         evt.PullRequest.Head.GetRef(),
		BRef:         evt.PullRequest.Base.GetRef(),
		IsMerged:     evt.PullRequest.GetMerged(),
		MergedBy:     evt.PullRequest.MergedBy.GetLogin(),
		Action:       evt.GetAction(),
	}

	// Extract event data to variable

	// If pull request action is not opened and closed, pass
	if w.Action != "opened" && w.Action != "closed" {
		u.log.Info("PR Action is not opened or closed")
		return nil
	}

	if w.Action == "opened" {
		contribPoint = point.CreatePR
	}

	// Repository
	repoData := model.CmdGitRepo{
		FullName:  w.RepoName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	r, err := u.repo.GetGitRepo(ctx, &bson.M{
		"full_name": w.RepoName,
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
			u.log.Info("Created GitRepo")
		} else {
			u.log.Error(err.Error())
			return err
		}
	}

	// Contributor
	contribData := model.CmdContributor{
		Username:   w.ContribUname,
		Avatar:     w.Avatar,
		ProfileURL: w.ContribUrl,
		IsLead:     false,
		IsCTO:      false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	c, err := u.repo.GetContributor(ctx, &bson.M{"username": w.ContribUname})
	if err != nil {
		if err.Error() == "Contributor not found" {
			res, err := u.repo.CreateContributor(ctx, &contribData)
			if err != nil {
				u.log.Error(err.Error())
				return err
			}
			ID, _ := res.InsertedID.(bson.ObjectID)
			c = model.Contributor{CmdContributor: contribData, ID: ID}
			u.log.Info("Created Contributor")
		} else {
			u.log.Error(err.Error())
			return err
		}
	}

	var l model.Contributor
	if w.IsMerged {
		l, err = u.repo.GetContributor(ctx, &bson.M{
			"username": w.MergedBy,
		})
		if !l.IsLead && !l.IsCTO {
			u.log.Info("Merge Action is not from lead or CTO! Passing point addition")
			return nil
		}
		if c.IsLead && !l.IsCTO {
			u.log.Info("Merge Action Lead PR is not from CTO! Passing point addition")
			return nil
		}
		if err != nil {
			u.log.Error(err.Error())
		}
	}

	// Pull Request
	prData := model.CmdPullRequest{
		Contributor:       c,
		Repo:              r,
		PullRequestURL:    w.PrUrl,
		SrcPullRequestURL: w.SrcPrUrl,
		SrcBranch:         w.HRef,
		DstBranch:         w.BRef,
		Action:            w.Action,
		MergedBy:          &l,
		IsMerged:          w.IsMerged,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	pr, err := u.repo.GetPullRequest(ctx, &bson.M{"pull_request_url": w.PrUrl})
	if err != nil {
		if err.Error() == "Pull request not found" && w.Action == "opened" {
			res, err := u.repo.CreatePullRequest(ctx, &prData)
			if err != nil {
				u.log.Error(err.Error())
				return err
			}
			ID, _ := res.InsertedID.(bson.ObjectID)
			pr = model.PullRequest{CmdPullRequest: prData, ID: ID}
			u.log.Info("Created Pull Request")
		} else {
			u.log.Error(err.Error())
			return err
		}
	}

	if pr.Action != w.Action {
		u.repo.UpdatePullRequest(ctx, &model.CmdPullRequest{Action: w.Action, UpdatedAt: time.Now()}, &bson.M{"_id": pr.ID})
		if w.Action == "closed" && w.IsMerged {
			u.repo.UpdatePullRequest(ctx, &prData, &bson.M{"_id": pr.ID})
			contribPoint = point.MergeContrib
			leadPoint = point.MergeLead
			u.log.Info("Update pull request to closed and merged")
		}
	}

	// INSERT POINT
	ahModel := model.CmdActionHistory{
		Repo:        r,
		Contributor: c,
		PullRequest: &pr,
		Event:       "pull_request",
		Action:      w.Action,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_, err = u.repo.CreateActionHistory(ctx, &ahModel)
	if err != nil {
		u.log.Error(err.Error())
		return err
	}

	// Insert Contributor Point
	if contribPoint > 0 {
		pointData := model.CmdPoint{
			Contributor: c,
			Point:       0,
		}
		p, err := u.repo.GetPoint(ctx, &bson.M{"contributor._id": c.ID})
		if err != nil {
			if err.Error() == "Point not found" {
				pointData.CreatedAt = time.Now()
				pointData.UpdatedAt = time.Now()
				res, err := u.repo.CreatePoint(ctx, &pointData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				ID, _ := res.InsertedID.(bson.ObjectID)
				p = model.Point{CmdPoint: pointData, ID: ID}
			} else {
				u.log.Error(err.Error())
				return err
			}
		}
		_, err = u.repo.UpdatePoint(ctx, &model.CmdPoint{
			Contributor: c,
			Point:       int64(p.Point + int64(contribPoint)),
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   time.Now(),
		}, &bson.M{"_id": p.ID})
		if err != nil {
			u.log.Error(err.Error())
			return err
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
	}

	// Insert Lead Point
	if leadPoint > 0 {
		lPointData := model.CmdPoint{
			Contributor: c,
			Point:       0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		p, err := u.repo.GetPoint(ctx, &bson.M{"contributor._id": c})
		if err != nil {
			if err.Error() == "Point not found" {
				_, err := u.repo.CreatePoint(ctx, &lPointData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				p = model.Point{CmdPoint: lPointData}
			} else {
				u.log.Error(err.Error())
				return err

			}
		}
		_, err = u.repo.CreatePointHistory(ctx, &model.CmdPointHistory{
			ActionHistory: ahModel,
			Point:         int64(leadPoint),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		if err != nil {
			u.log.Error(err.Error())
			return err
		}
		_, err = u.repo.UpdatePoint(ctx, &model.CmdPoint{
			Contributor: l,
			Point:       p.Point + int64(leadPoint),
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   time.Now(),
		}, &bson.M{"_id": l.ID})
		if err != nil {
			u.log.Error(err.Error())
			return err
		}
		u.log.Info("Update Point Success!")
	}
	return nil
}
