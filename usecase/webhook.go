package usecase

import (
	"context"
	"fmt"
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
			} else {
				u.log.Error(err.Error())
				return err
			}
		}

		// Pull Request
		prData := model.CmdPullRequest{
			Contributor:    c,
			Repo:           r,
			PullRequestURL: w.PrUrl,
			SrcBranch:      w.HRef,
			DstBranch:      w.BRef,
			Action:         w.Action,
			IsMerged:       w.IsMerged,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		pr, err := u.repo.GetPullRequest(ctx, &bson.M{"pull_request_url": w.PrUrl})
		if err != nil {
			if err.Error() == "Pull request not found" && w.Action == "opened" {
				contribPoint = point.CreatePR
				res, err := u.repo.CreatePullRequest(ctx, &prData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				ID, _ := res.InsertedID.(bson.ObjectID)
				pr = model.PullRequest{CmdPullRequest: prData, ID: ID}
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
			if err != nil {
				u.log.Error(err.Error())
			}
		}
		prData.MergedBy = l
		if pr.Action != w.Action {
			u.repo.UpdatePullRequest(ctx, &model.CmdPullRequest{Action: w.Action, UpdatedAt: time.Now()}, &bson.M{"_id": pr.ID})
			if w.Action == "closed" && w.IsMerged {
				u.repo.UpdatePullRequest(ctx, &prData, &bson.M{"_id": pr.ID})
				contribPoint = point.MergeContrib
				leadPoint = point.MergeLead
			}
		}

		// INSERT POINT
		ahModel := model.CmdActionHistory{
			Repo:        r,
			Contributor: c,
			PullRequest: &pr,
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
			p, err := u.repo.GetPoint(ctx, &bson.M{"contributor._id": c})
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
			u.repo.UpdatePoint(ctx, &model.CmdPoint{
				Point: int64(p.Point + int64(contribPoint)),
			}, &bson.M{"_id": p.ID})
			u.repo.CreatePointHistory(ctx, &model.CmdPointHistory{
				ActionHistory: ahModel,
				Point:         int64(contribPoint),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			})
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
			u.repo.CreatePointHistory(ctx, &model.CmdPointHistory{
				ActionHistory: ahModel,
				Point:         int64(leadPoint),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			})
			u.repo.UpdatePoint(ctx, &model.CmdPoint{
				Point: p.Point + int64(leadPoint),
			}, &bson.M{"_id": l.ID})
		}
	} // PULL REQUEST EVENT END

	// FORK EVENT
	if evt, ok := event.(*github.ForkEvent); ok {
		u.log.Info("Fork Event Called")

		var f = model.Webhook{}

		f.ContribUname = evt.Forkee.Owner.GetLogin()
		f.ContribUrl = evt.Forkee.Owner.GetHTMLURL()
		f.Avatar = evt.Forkee.Owner.GetAvatarURL()
		f.RepoName = evt.Repo.GetFullName()
		f.Action = "fork"

		repoData := model.CmdGitRepo{
			FullName:  f.RepoName,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		r, err := u.repo.GetGitRepo(ctx, &bson.M{"full_name": f.RepoName})
		if err != nil {
			if err.Error() == "GitRepo not found" {
				res, err := u.repo.CreateGitRepo(ctx, &repoData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				ID, _ := res.InsertedID.(bson.ObjectID)
				r = model.GitRepo{
					CmdGitRepo: repoData,
					ID:         ID,
				}
				u.log.Info("Repo Data inserted")
			} else {
				u.log.Error(err.Error())
				return err
			}
		}

		contribData := model.CmdContributor{
			Username:   f.ContribUname,
			Avatar:     f.Avatar,
			ProfileURL: f.ContribUrl,
			IsLead:     false,
			IsCTO:      false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		c, err := u.repo.GetContributor(ctx, &bson.M{"username": f.ContribUname})
		if err != nil {
			if err.Error() == "Contributor not found" {
				res, err := u.repo.CreateContributor(ctx, &contribData)
				if err != nil {
					u.log.Error(err.Error())
					return err
				}
				ID, _ := res.InsertedID.(bson.ObjectID)
				c = model.Contributor{
					CmdContributor: contribData,
					ID:             ID,
				}
				u.log.Info("Contributor Data Inserted")
			} else {
				u.log.Error(err.Error())
				return err
			}
		}

		_, err = u.repo.GetActionHistory(ctx, &bson.M{
			"repo._id":        r.ID,
			"contributor._id": c.ID,
			"action":          f.Action,
		})

		if err != nil {
			if err.Error() == "Action History not found" {
				u.log.Info("New Forkee")

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
						p = model.Point{
							CmdPoint: pointData,
							ID:       ID,
						}
					} else {
						u.log.Error(err.Error())
						return err
					}
				}
				up, err := u.repo.UpdatePoint(ctx, &model.CmdPoint{
					Contributor: c,
					Point:       (int64(p.Point) + int64(point.ForkRepo)),
					UpdatedAt:   time.Now(),
				}, &bson.M{"_id": p.ID})
				if err != nil {
					u.log.Error(err.Error())
				}
				iahData := model.CmdActionHistory{
					Repo:        r,
					Contributor: c,
					PullRequest: nil,
					Action:      f.Action,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				u.log.Info(fmt.Sprintf("%d, %s", up.MatchedCount, up.UpsertedID))
				_, err = u.repo.CreateActionHistory(ctx, &iahData)
				u.repo.CreatePointHistory(ctx, &model.CmdPointHistory{
					ActionHistory: iahData,
					Point:         int64(point.ForkRepo),
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				})

			} else {
				u.log.Error("Fork Err: " + err.Error())
				return err
			}
		}

	} // FORK EVENT END

	// PR COMMENT EVENT
	// if evt, ok := event.(*github.IssueCommentEvent); ok {

	// } // PR COMMENT EVENT END
	return nil
}
