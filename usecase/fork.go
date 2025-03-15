package usecase

import (
	"context"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/ryakadev/rdf-contrib-collector/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (u usecase) ForkEvent(ctx context.Context, evt *github.ForkEvent, point *model.PointActionData) error {
	u.log.Info("Fork Event Called")

	f := model.Webhook{
		ContribUname: evt.Forkee.Owner.GetLogin(),
		ContribUrl:   evt.Forkee.Owner.GetHTMLURL(),
		Avatar:       evt.Forkee.Owner.GetAvatarURL(),
		RepoName:     evt.Repo.GetFullName(),
		Action:       "fork",
	}

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

	ah, err := u.repo.GetActionHistory(ctx, &bson.M{
		"repo._id":        r.ID,
		"contributor._id": c.ID,
		"action":          f.Action,
	})

	if err != nil {
		if err.Error() != "Action History not found" {
			u.log.Error("Fork Err: " + err.Error())
			return err
		}
	}
	// Check fork duplicate
	if !ah.ID.IsZero() {
		u.log.Info("Already Forked! Passing point addition")
		return nil
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
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   time.Now(),
	}, &bson.M{"_id": p.ID})
	if err != nil {
		u.log.Error(err.Error())
	}

	iahData := model.CmdActionHistory{
		Repo:        r,
		Contributor: c,
		PullRequest: nil,
		Event:       f.Action,
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
	return nil
}
