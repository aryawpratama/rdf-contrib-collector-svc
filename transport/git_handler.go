package transport

import (
	"context"
	"net/http"

	"github.com/google/go-github/v69/github"
	"github.com/ryakadev/rdf-contrib-collector/model"
)

func (t transport) GitWebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	point := model.PointActionData{
		CreatePR:       40,
		ForkRepo:       5,
		ResolveComment: 50,
		MergeContrib:   100,
		MergeLead:      20,
		CommentLead:    2,
		CommentContrib: 2,
	}

	// Validate Payload
	payload, err := github.ValidatePayload(r, []byte(t.config.AppSecret))
	if err != nil {
		t.log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
	}

	// Parse Webhook
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		t.log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
	}

	// Group Event
	if evt, ok := event.(*github.PullRequestEvent); ok {
		err = t.usecase.PullRequestEvent(ctx, evt, &point)
	}
	if evt, ok := event.(*github.ForkEvent); ok {
		err = t.usecase.ForkEvent(ctx, evt, &point)
	}
	if evt, ok := event.(*github.PullRequestReviewCommentEvent); ok {
		err = t.usecase.PullRequestReviewCommentEvent(ctx, evt, &point)
	}

	if evt, ok := event.(*github.PullRequestReviewThreadEvent); ok {
		err = t.usecase.PullRequestReviewThreadEvent(ctx, evt, &point)
	}
	if evt, ok := event.(*github.PullRequestReviewEvent); ok {
		err = t.usecase.PullRequestReviewApproved(ctx, evt, &point)
	}

	if err != nil {
		t.log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(200)))
}
