package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-census-alpha-api-proxy/ftb"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

type simpleMessage struct {
	Message string
}

//API provides a struct to wrap the api around
type API struct {
	Router *mux.Router
}

type FlexibleTableBuilder interface {
	GetDatasets(ctx context.Context) (*ftb.Datasets, error)
}

type Authenticator func(handlerFunc http.HandlerFunc) http.HandlerFunc

func Setup(ctx context.Context, r *mux.Router, authCheck Authenticator, ftb FlexibleTableBuilder) *API {
	api := &API{Router: r}

	r.PathPrefix("/v6/datasets").HandlerFunc(authCheck(api.GetDatasets(ftb))).Methods("GET")
	return api
}

func (a *API) GetDatasets(ftb FlexibleTableBuilder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		datasets, err := ftb.GetDatasets(ctx)
		if err != nil {
			errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(datasets); err != nil {
			errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Event(ctx, "get datasets request successful", log.INFO)
	}
}

func errorResponse(w http.ResponseWriter, msg string, status int) {
	b, err := json.Marshal(simpleMessage{Message: msg})
	if err != nil {
		msg = "internal server error"
		status = http.StatusInternalServerError
	}

	log.Event(nil, msg, log.ERROR, log.Data{"status": status})
	http.Error(w, string(b), status)
}

func (*API) Close(ctx context.Context) error {
	// Close any dependencies
	log.Event(ctx, "graceful shutdown of api complete", log.INFO)
	return nil
}
