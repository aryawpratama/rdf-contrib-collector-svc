package transport

import (
	"context"
	"net/http"

	"github.com/google/go-github/v69/github"
)

func (t transport) GitWebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	payload, err := github.ValidatePayload(r, []byte(t.config.App.Secret))
	if err != nil {
		t.log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
	}
	event, _ := github.ParseWebHook(github.WebHookType(r), payload)
	if err := t.usecase.HandleWebhook(ctx, event); err != nil {
		t.log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(200)))
}
