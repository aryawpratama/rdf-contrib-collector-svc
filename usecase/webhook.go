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
		CommentContrib: 2,
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

		// If pull request action is not opened and closed, pass
		if w.Action != "opened" && w.Action != "closed" {
			u.log.Info("PR Action is not opened or closed")
			return nil
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
				u.log.Info("Created Pull Request")
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
		if !l.IsLead {
			u.log.Info("Merge Action is not from lead! Passing point addition")
			return nil
		}
		prData.MergedBy = l
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
			_, err = u.repo.UpdatePoint(ctx, &model.CmdPoint{
				Point: int64(p.Point + int64(contribPoint)),
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
				Point: p.Point + int64(leadPoint),
			}, &bson.M{"_id": l.ID})
			if err != nil {
				u.log.Error(err.Error())
				return err
			}
			u.log.Info("Update Point Success!")
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
				u.log.Info("GitRepo created")
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
				u.log.Info("Contributor created")
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

		// Check fork duplicate
		if err != nil {
			if err.Error() != "Action History not found" {
				u.log.Error("Fork Err: " + err.Error())
				return err
			}
		}

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
				u.log.Info("Point Created")
			} else {
				u.log.Error(err.Error())
				return err
			}
		}
		_, err = u.repo.UpdatePoint(ctx, &model.CmdPoint{
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
			Event:       "fork",
			Action:      f.Action,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		_, err = u.repo.CreateActionHistory(ctx, &iahData)
		if err != nil {
			u.log.Error(err.Error())
			return err
		}

		_, err = u.repo.CreatePointHistory(ctx, &model.CmdPointHistory{
			ActionHistory: iahData,
			Point:         int64(point.ForkRepo),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		if err != nil {
			u.log.Error(err.Error())
			return err
		}

		u.log.Info("Success Insert Point")

	} // FORK EVENT END

	// PR COMMENT CREATED EVENT
	if evt, ok := event.(*github.PullRequestReviewCommentEvent); ok {
		u.log.Info("Pull Request Review Comment Event Called")
		var prrc = model.Webhook{
			Avatar:       evt.Comment.User.GetAvatarURL(),
			RepoName:     evt.PullRequest.Base.Repo.GetFullName(),
			PrUrl:        evt.PullRequest.GetHTMLURL(),
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
			Contributor:    c,
			Repo:           r,
			PullRequestURL: prrc.PrUrl,
			SrcBranch:      prrc.HRef,
			DstBranch:      prrc.BRef,
			Action:         prrc.Action,
			IsMerged:       prrc.IsMerged,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
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
			Point: p.Point + int64(contribPoint),
		}, &bson.M{"_id": p.ID})
		if err != nil {
			u.log.Error(err.Error())
			return err
		}
		u.log.Info("Success Insert Point")
	} // PR COMMENT CREATED EVENT END

	// PR COMMENT RESOLVED
	if evt, ok := event.(*github.PullRequestReviewThreadEvent); ok {
		u.log.Info("Pull Request Review Thread Event Called")
		var prrt = model.Webhook{
			Avatar:       evt.PullRequest.User.GetAvatarURL(),
			RepoName:     evt.PullRequest.Base.Repo.GetFullName(),
			PrUrl:        evt.PullRequest.GetHTMLURL(),
			Action:       evt.GetAction(),
			HRef:         evt.PullRequest.Head.GetRef(),
			BRef:         evt.PullRequest.Base.GetRef(),
			ContribUname: evt.PullRequest.User.GetLogin(),
			ContribUrl:   evt.PullRequest.User.GetHTMLURL(),
		}

		// If review is not in resolved action, pass
		if prrt.Action != "resolved" {
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
			} else {
				u.log.Error(err.Error())
				return err
			}
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
			} else {
				u.log.Error(err.Error())
				return err
			}
		}

		// Pull Request
		prData := model.CmdPullRequest{
			Contributor:    c,
			Repo:           r,
			PullRequestURL: prrt.PrUrl,
			SrcBranch:      prrt.HRef,
			DstBranch:      prrt.BRef,
			Action:         prrt.Action,
			IsMerged:       prrt.IsMerged,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
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
			} else {
				u.log.Error(err.Error())
				return err
			}
		}
		ahModel := model.CmdActionHistory{
			Repo:        r,
			Contributor: c,
			PullRequest: &pr,
			Event:       "pull_request_review_thread",
			Action:      prrt.Action,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
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
			Point: p.Point + int64(point.ResolveComment),
		}, &bson.M{"_id": p.ID})
		if err != nil {
			u.log.Error(err.Error())
			return err
		}
		u.log.Info("Point Inserted Successfully")
	} // PR COMMENT RESOLVED END

	// PR REVIEW APPROVED
	if evt, ok := event.(*github.PullRequestReviewEvent); ok {
		u.log.Info("Pull Request Review Event Called")
		var prrt = model.Webhook{
			Avatar:         evt.PullRequest.User.GetAvatarURL(),
			RepoName:       evt.PullRequest.Base.Repo.GetFullName(),
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
			Contributor:    c,
			Repo:           r,
			PullRequestURL: prrt.PrUrl,
			SrcBranch:      prrt.HRef,
			DstBranch:      prrt.BRef,
			Action:         prrt.Action,
			IsMerged:       prrt.IsMerged,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
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
			Point: p.Point + int64(point.ResolveComment),
		}, &bson.M{"_id": p.ID})
		if err != nil {
			u.log.Error(err.Error())
			return err
		}
	} // PR COMMENT RESOLVED END
	return nil
}
