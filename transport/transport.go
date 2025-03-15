package transport

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ryakadev/rdf-contrib-collector/config"
	"github.com/ryakadev/rdf-contrib-collector/usecase"
	"go.uber.org/zap"
)

type transport struct {
	usecase usecase.UseCase
	log     *zap.Logger
	config  config.Config
}

func NewTransport(cfg config.Config, log *zap.Logger, usecase usecase.UseCase) {
	t := &transport{
		usecase: usecase,
		log:     log,
		config:  cfg,
	}
	r := mux.NewRouter()
	t.registerGitRoute(r)
	http.Handle("/", r)
}

func (t transport) registerGitRoute(r *mux.Router) {
	r.HandleFunc("/git/webhook", t.GitWebhookHandler).Methods(http.MethodPost)
	r.HandleFunc("/_status/healthz",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(http.StatusText(200)))
		},
	).Methods(http.MethodGet)
}
