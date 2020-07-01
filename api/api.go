package api

import (
	"context"
	"encoding/json"
	"errors"
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
	Client FTBClient
	Router *mux.Router
}

type FTBClient interface {
	GetData(ctx context.Context, url string) (ftb.Entity, error)
}

type Authenticator func(http.Handler) http.Handler

func Setup(ctx context.Context, r *mux.Router, auth Authenticator, client FTBClient) *API {
	api := &API{
		Client: client,
		Router: r,
	}

	r.PathPrefix("/v6/datasets").Handler(auth(api.Handler())).Methods(http.MethodGet)
	r.PathPrefix("/v6/datasets").HandlerFunc(api.preflightRequestHandler).Methods(http.MethodOptions)

	r.PathPrefix("/v6/codebook").Handler(auth(api.Handler())).Methods(http.MethodGet)
	r.PathPrefix("/v6/codebook").HandlerFunc(api.preflightRequestHandler).Methods(http.MethodOptions)

	r.PathPrefix("/v6/query").Handler(auth(api.Handler())).Methods(http.MethodGet)
	r.PathPrefix("/v6/query").HandlerFunc(api.preflightRequestHandler).Methods(http.MethodOptions)
	return api
}

func (api *API) preflightRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", http.MethodGet)
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		url := r.URL.String()

		entity, err := api.Client.GetData(ctx, url)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}

		WriteBody(ctx, w, entity, http.StatusOK)
	})
}

func getErrorResponse(ctx context.Context, err error) (SimpleEntity, int) {
	log.Event(ctx, "returning http error response", log.ERROR, log.Error(err))

	status := http.StatusInternalServerError
	msg := "internal server error"

	var ftbErr ftb.Error
	if errors.As(err, &ftbErr) {
		status = ftbErr.StatusCode
		msg = ftbErr.Message
	}

	return SimpleEntity{Message: msg}, status
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
