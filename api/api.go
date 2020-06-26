package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-census-alpha-api-proxy/ftb"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

type SimpleEntity struct {
	Message string
}

//API provides a struct to wrap the api around
type API struct {
	Router *mux.Router
}

type FlexibleTableBuilder interface {
	GetData(ctx context.Context, url string) (ftb.Entity, error)
}

type Authenticator func(handlerFunc http.HandlerFunc) http.HandlerFunc

func Setup(ctx context.Context, r *mux.Router, authCheck Authenticator, ftb FlexibleTableBuilder) *API {
	api := &API{Router: r}

	r.PathPrefix("/v6/datasets").HandlerFunc(authCheck(api.Handle(ftb))).Methods("GET")
	r.PathPrefix("/v6/codebook").HandlerFunc(authCheck(api.Handle(ftb))).Methods("GET")
	return api
}

func (a *API) Handle(ftb FlexibleTableBuilder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		url := r.URL.String()

		entity, err := ftb.GetData(ctx, url)
		if err != nil {
			WriteBody(ctx, w, SimpleEntity{Message: err.Error()}, http.StatusInternalServerError)
			return
		}

		WriteBody(ctx, w, entity, http.StatusOK)
	}
}

func WriteBody(ctx context.Context, w http.ResponseWriter, entity interface{}, status int) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(entity)
	if err != nil {
		log.Event(ctx, "failed to write entity to response body", log.Error(err), log.ERROR)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (*API) Close(ctx context.Context) error {
	// Close any dependencies
	log.Event(ctx, "graceful shutdown of api complete", log.INFO)
	return nil
}
