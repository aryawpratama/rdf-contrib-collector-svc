package usecase

import (
	"context"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (u usecase) HandleWebhook(ctx context.Context, event interface{}) error {
	// Point Data
	var point model.PointActionData = model.PointActionData{
		CreatePR:       40,
		ForkRepo:       5,
		ResolveComment: 50,
		MergeContrib:   100,
		MergeLead:      20,
		CommentLead:    20,
	}

	// PULL REQUEST EVENT
	if evt, ok := event.(*github.PullRequestEvent); ok {
		u.log.Info("PR Event Called")
		var contribPoint int = 0
		var leadPoint int = 0
		var w = model.Webhook{}

		// Extract event data to variable
		w.RepoName = evt.Repo.GetFullName()
		w.PrUrl = evt.PullRequest.GetHTMLURL()
		w.ContribUname = evt.PullRequest.User.GetLogin()
		w.HRef = evt.PullRequest.Head.GetRef()
		w.BRef = evt.PullRequest.Base.GetRef()
		w.Action = evt.GetAction()

		// Repository
		repoData := model.GitRepo{
			FullName:  w.RepoName,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		r, err := u.repo.GetGitRepo(ctx, &model.GitRepo{
			FullName: w.RepoName,
		})
		if err != nil {
			if err.Error() == "GitRepo not found" {
				res, err := u.repo.CreateGitRepo(ctx, &repoData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				repoData.ID, ok = res.InsertedID.(bson.ObjectID)
				r = repoData
			}
			u.log.Error(err.Error())
			return err
		}

		// Contributor
		contribData := model.Contributor{
			Username:   w.ContribUname,
			Avatar:     w.Avatar,
			ProfileURL: w.ContribUrl,
			IsLead:     false,
			IsCTO:      false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		c, err := u.repo.GetContributor(ctx, &model.Contributor{Username: w.ContribUname})
		if err != nil {
			if err.Error() == "Contributor not found" {
				res, err := u.repo.CreateContributor(ctx, &contribData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				contribData.ID, ok = res.InsertedID.(bson.ObjectID)
				c = contribData
			}
			u.log.Error(err.Error())
			return err
		}

		// Pull Request
		prData := model.PullRequest{
			ContributorID:  c.ID,
			RepoID:         r.ID,
			PullRequestURL: w.PrUrl,
			SrcBranch:      w.HRef,
			DstBranch:      w.BRef,
			Action:         w.Action,
			IsMerged:       w.IsMerged,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		pr, err := u.repo.GetPullRequest(ctx, &model.PullRequest{PullRequestURL: w.PrUrl})
		if err != nil {
			if err.Error() == "Pull request not found" && w.Action == "opened" {
				contribPoint = point.CreatePR
				res, err := u.repo.CreatePullRequest(ctx, &prData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				prData.ID, ok = res.InsertedID.(bson.ObjectID)
				pr = prData
			}
			u.log.Error(err.Error())
			return err
		}
		var l model.Contributor
		if w.IsMerged {
			l, err = u.repo.GetContributor(ctx, &model.Contributor{
				Username: w.MergedBy,
			})
			if err != nil {
				u.log.Error(err.Error())
			}
		}
		prData.MergedByID = l.ID
		if pr.Action != w.Action {
			u.repo.UpdatePullRequest(ctx, &model.PullRequest{Action: w.Action, UpdatedAt: time.Now()}, &model.PullRequest{ID: pr.ID})
			if w.Action == "closed" && w.IsMerged {
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
			Action:        w.Action,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		if err != nil {
			u.log.Error(err.Error())
			return err
		}

		// Insert Contributor Point
		if contribPoint > 0 {
			pointData := model.Point{
				ContribID: c.ID,
				Point:     0,
			}
			p, err := u.repo.GetPoint(ctx, &model.Point{ContribID: c.ID})
			if err != nil {
				if err.Error() == "Point not found" {
					pointData.CreatedAt = time.Now()
					pointData.UpdatedAt = time.Now()
					res, err := u.repo.CreatePoint(ctx, &pointData)
					if err != nil {
						u.log.Error(err.Error())
						return err
					}
					p.ID, ok = res.InsertedID.(bson.ObjectID)
					p = pointData
				}
				u.log.Error(err.Error())
				return err
			}
			u.repo.UpdatePoint(ctx, &model.Point{
				Point: int64(p.Point + int64(contribPoint)),
			}, &model.Point{ID: p.ID})
			u.repo.CreatePointHistory(ctx, &model.PointHistory{
				ActionHistoryId: ah.InsertedID.(bson.ObjectID),
				Point:           int64(contribPoint),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			})
		}

		// Insert Lead Point
		if leadPoint > 0 {
			lPointData := model.Point{
				ContribID: c.ID,
				Point:     0,
			}
			p, err := u.repo.GetPoint(ctx, &model.Point{ContribID: c.ID})
			if err != nil {
				if err.Error() == "Point not found" {
					lPointData.CreatedAt = time.Now()
					lPointData.UpdatedAt = time.Now()
					res, err := u.repo.CreatePoint(ctx, &lPointData)
					if err != nil {
						u.log.Error(err.Error())
						return err
					}
					p.ID, ok = res.InsertedID.(bson.ObjectID)
					p = lPointData
				}
				u.log.Error(err.Error())
				return err
			}
			u.repo.CreatePointHistory(ctx, &model.PointHistory{
				ActionHistoryId: ah.InsertedID.(bson.ObjectID),
				Point:           int64(leadPoint),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			})
			u.repo.UpdatePoint(ctx, &model.Point{
				Point: p.Point + int64(leadPoint),
			}, &model.Point{ID: l.ID})
		}
	} // PULL REQUEST EVENT END

	// FORK EVENT
	if evt, ok := event.(*github.ForkEvent); ok {
		u.log.Info("Fork Event Called")

		var f = model.Webhook{}

		f.ContribUname = evt.Forkee.Owner.GetLogin()
		f.ContribUrl = evt.Forkee.Owner.GetURL()
		f.Avatar = evt.Forkee.Owner.GetAvatarURL()
		f.RepoName = evt.Repo.GetFullName()
		f.Action = "fork"

		repoData := model.GitRepo{
			FullName:  f.RepoName,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		r, err := u.repo.GetGitRepo(ctx, &model.GitRepo{FullName: f.RepoName})
		if err != nil {
			if err.Error() == "GitRepo not found" {
				res, err := u.repo.CreateGitRepo(ctx, &repoData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				if repoData.ID, ok = res.InsertedID.(bson.ObjectID); ok {
					r = repoData
					u.log.Info("Repo Data inserted")
				}
			} else {
				u.log.Error(err.Error())
				return err
			}
		}

		contribData := model.Contributor{
			Username:   f.ContribUname,
			Avatar:     f.Avatar,
			ProfileURL: f.ContribUrl,
			IsLead:     false,
			IsCTO:      false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		c, err := u.repo.GetContributor(ctx, &model.Contributor{Username: f.ContribUname})
		if err != nil {
			if err.Error() == "Contributor not found" {
				res, err := u.repo.CreateContributor(ctx, &contribData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				contribData.ID, ok = res.InsertedID.(bson.ObjectID)
				c = contribData
			}
			u.log.Error(err.Error())
			return err
		}

		_, err = u.repo.GetActionHistory(ctx, &model.ActionHistory{
			RepoID:    r.ID,
			ContribID: c.ID,
			Action:    f.Action,
		})

		if err != nil {
			if err.Error() == "Action History not found" {
				u.log.Info("New Forkee")
				pointData := model.Point{
					ContribID: c.ID,
					Point:     0,
				}
				iah, err := u.repo.CreateActionHistory(ctx, &model.ActionHistory{
					RepoID:    r.ID,
					ContribID: c.ID,
					Action:    f.Action,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				})
				p, err := u.repo.GetPoint(ctx, &model.Point{ContribID: c.ID})
				if err != nil {
					if err.Error() == "Point not found" {
						pointData.CreatedAt = time.Now()
						pointData.UpdatedAt = time.Now()
						res, err := u.repo.CreatePoint(ctx, &pointData)
						if err != nil {
							u.log.Error(err.Error())
							return err
						}
						p.ID, ok = res.InsertedID.(bson.ObjectID)
						p = pointData
					}
					u.log.Error(err.Error())
					return err
				}
				u.repo.UpdatePoint(ctx, &model.Point{
					Point: int64(p.Point + int64(point.ForkRepo)),
				}, &model.Point{ID: p.ID})
				u.repo.CreatePointHistory(ctx, &model.PointHistory{
					ActionHistoryId: iah.InsertedID.(bson.ObjectID),
					Point:           int64(point.ForkRepo),
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				})

			}
			u.log.Error("Fork Err: " + err.Error())
			return err
		}

	} // FORK EVENT END

	// PR COMMENT EVENT
	// if evt, ok := event.(*github.IssueCommentEvent); ok {

	// } // PR COMMENT EVENT END
	return nil
}
