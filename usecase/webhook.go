package usecase

import (
	"context"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (u usecase) HandleWebhook(ctx context.Context, event interface{}) error {
	var point model.PointActionData = model.PointActionData{
		CreatePR:       10,
		ForkRepo:       5,
		ResolveComment: 50,
		MergeContrib:   100,
		MergeLead:      20,
		CommentLead:    10,
	}

	// PULL REQUEST EVENT
	if evt, ok := event.(*github.PullRequestEvent); ok {
		var contribPoint int = 0
		var leadPoint int = 0
		var (
			avatar       string
			repoName     string
			prUrl        string
			action       string
			hRef         string
			bRef         string
			contribUname string
			contribUrl   string
			isMerged     bool
			mergedBy     string
		)

		// Extract event data to variable
		repoName = evt.Repo.GetFullName()
		prUrl = evt.PullRequest.GetHTMLURL()
		contribUname = evt.PullRequest.User.GetLogin()
		hRef = evt.PullRequest.Head.GetRef()
		bRef = evt.PullRequest.Base.GetRef()
		action = evt.GetAction()

		// Repository
		repoData := model.GitRepo{
			FullName:  repoName,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		r, err := u.repo.GetGitRepo(ctx, &model.GitRepo{
			FullName: repoName,
		})
		if err != nil {
			if err.Error() == "GitRepo not found" {
				res, err := u.repo.CreateGitRepo(ctx, &repoData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				repoData.ID, ok = res.InsertedID.(primitive.ObjectID)
				r = repoData
			}
			u.log.Error(err.Error())
			return err
		}

		// Contributor
		contribData := model.Contributor{
			Username:   contribUname,
			Avatar:     avatar,
			ProfileURL: contribUrl,
			IsLead:     false,
			IsCTO:      false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		c, err := u.repo.GetContributor(ctx, &model.Contributor{Username: contribUname})
		if err != nil {
			if err.Error() == "Contributor not found" {
				res, err := u.repo.CreateContributor(ctx, &contribData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				contribData.ID, ok = res.InsertedID.(primitive.ObjectID)
				c = contribData
			}
			u.log.Error(err.Error())
			return err
		}

		// Pull Request
		prData := model.PullRequest{
			ContributorID:  c.ID,
			RepoID:         r.ID,
			PullRequestURL: prUrl,
			SrcBranch:      hRef,
			DstBranch:      bRef,
			Action:         action,
			IsMerged:       isMerged,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		pr, err := u.repo.GetPullRequest(ctx, &model.PullRequest{PullRequestURL: prUrl})
		if err != nil {
			if err.Error() == "Pull request not found" && action == "opened" {
				contribPoint = point.CreatePR
				res, err := u.repo.CreatePullRequest(ctx, &prData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				prData.ID, ok = res.InsertedID.(primitive.ObjectID)
				pr = prData
			}
			u.log.Error(err.Error())
			return err
		}
		var l model.Contributor
		if isMerged {
			l, err = u.repo.GetContributor(ctx, &model.Contributor{
				Username: mergedBy,
			})
			if err != nil {
				u.log.Error(err.Error())
			}
		}
		prData.MergedByID = l.ID
		if pr.Action != action {
			u.repo.UpdatePullRequest(ctx, &model.PullRequest{Action: action, UpdatedAt: time.Now()}, &model.PullRequest{ID: pr.ID})
			if action == "closed" && isMerged {
				u.repo.UpdatePullRequest(ctx, &prData, &model.PullRequest{ID: pr.ID})
				contribPoint = point.MergeContrib
				leadPoint = point.MergeLead
			}
		}

		// INSERT POINT
		ah, err := u.repo.CreateActionHistory(ctx, &model.ActionHistory{
			RepoID:        r.ID,
			ContribID:     c.ID,
			PullRequestID: pr.ID,
			Action:        action,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		if err != nil {
			u.log.Error(err.Error())
			return err
		}

		if contribPoint > 0 {
			u.repo.CreatePointHistory(ctx, &model.PointHistory{
				ActionHistoryId: ah.InsertedID.(primitive.ObjectID),
				Point:           int64(contribPoint),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			})
			u.repo.UpdatePoint(ctx, &model.Point{
				Point: int64(contribPoint),
			}, &model.Point{ContribID: c.ID})
		}
		if leadPoint > 0 {
			u.repo.CreatePointHistory(ctx, &model.PointHistory{
				ActionHistoryId: ah.InsertedID.(primitive.ObjectID),
				Point:           int64(leadPoint),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			})
			u.repo.UpdatePoint(ctx, &model.Point{
				Point: int64(leadPoint),
			}, &model.Point{ContribID: l.ID})
		}
	} // PULL REQUEST EVENT END
	return nil
}
