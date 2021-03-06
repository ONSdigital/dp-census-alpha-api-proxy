package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-census-alpha-api-proxy/cantabular"
	filterModel "github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-code-list-api/models"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

type SimpleEntity struct {
	Message string
}

//API provides a struct to wrap the api around
type API struct {
	Store  DataStore
	Router *mux.Router
}

type DataStore interface {
	GetData(ctx context.Context, url string) (cantabular.Entity, error)
	GetDatasetCodebook(ctx context.Context, dataset string) (*cantabular.Codebook, error)
}

type Authenticator func(http.Handler) http.Handler

func Setup(ctx context.Context, r *mux.Router, auth Authenticator, client DataStore) *API {
	api := &API{
		Store:  client,
		Router: r,
	}

	r.Handle("/v6/datasets/{dataset}/filter/dimensions/{name}/options", auth(api.GetFilterDimensions())).Methods(http.MethodGet)

	r.Handle("/v6/datasets/{dataset}/dimensions", auth(api.GetDatasetDimensions())).Methods(http.MethodGet)
	r.Handle("/v6/datasets/{dataset}/dimensions/{name}", auth(api.GetDatasetDimension())).Methods(http.MethodGet)
	r.Handle("/v6/datasets/{dataset}/dimensions/{name}/codes", auth(api.GetDatasetDimensionCodes())).Methods(http.MethodGet)
	r.Handle("/v6/datasets/{dataset}/dimensions/{name}/index/{index}", auth(api.GetDatasetDimensionByIndex())).Methods(http.MethodGet)

	r.Handle("/v6/datasets/{dataset}/hierarchies/{name}", auth(api.GetHierarchy())).Methods(http.MethodGet)
	r.Handle("/v6/datasets/{dataset}/hierarchies/{name}/full", auth(api.BuildFullHierarchy())).Methods(http.MethodGet)
	r.Handle("/v6/datasets/{dataset}/hierarchies/{name}/code/{code}", auth(api.GetHierarchyForCode())).Methods(http.MethodGet)

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
	w.Header().Set("Access-Control-Allow-Headers", "Authorization")
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		url := r.URL.String()

		entity, err := api.Store.GetData(ctx, url)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}

		WriteBody(ctx, w, entity, http.StatusOK)
	})
}

func (api *API) GetDatasetDimensions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataset := mux.Vars(r)["dataset"]

		codebook, err := api.Store.GetDatasetCodebook(ctx, dataset)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}

		dims := make([]string, 0)
		for _, cb := range codebook.CodeBook {
			dims = append(dims, cb.Label+": "+cb.Name)
		}

		WriteBody(ctx, w, GetDimensionsResponse{Dimensions: dims}, http.StatusOK)
	})
}

func (api *API) GetDatasetDimension() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataset := mux.Vars(r)["dataset"]
		dimension := mux.Vars(r)["name"]

		codebook, err := api.Store.GetDatasetCodebook(ctx, dataset)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}

		var result *cantabular.Dimension
		for _, cb := range codebook.CodeBook {
			if cb.Name == dimension {
				result = &cb
				break
			}
		}

		if result == nil {
			entity := SimpleEntity{Message: "dimension not found"}
			WriteBody(ctx, w, entity, http.StatusNotFound)
			return
		}

		WriteBody(ctx, w, result, http.StatusOK)
	})
}

func (api *API) GetDatasetDimensionCodes() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataset := mux.Vars(r)["dataset"]
		dimension := mux.Vars(r)["name"]

		codebook, err := api.Store.GetDatasetCodebook(ctx, dataset)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}

		dim := codebook.GetDimension(dimension)
		if dim == nil {
			WriteBody(ctx, w, SimpleEntity{Message: "not found"}, http.StatusNotFound)
			return
		}

		codelist := mapToCMDCodeList(dim)
		WriteBody(ctx, w, codelist, http.StatusOK)
	})
}

func mapToCMDCodeList(dimension *cantabular.Dimension) *models.CodeResults {
	codes := make([]models.Code, 0)
	for i, c := range dimension.Codes {
		codes = append(codes, models.Code{
			ID:    c,
			Label: dimension.Labels[i],
			Links: nil,
		})
	}

	length := len(codes)
	return &models.CodeResults{
		Items:      codes,
		Count:      length,
		Offset:     0,
		Limit:      length,
		TotalCount: length,
	}
}

func (api *API) GetFilterDimensions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataset := mux.Vars(r)["dataset"]
		dimensionName := mux.Vars(r)["name"]

		codebook, err := api.Store.GetDatasetCodebook(ctx, dataset)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}


		dim := codebook.GetDimension(dimensionName)

		var options []*filterModel.PublicDimensionOption
		for _, code := range dim.Codes {
			options = append(options, &filterModel.PublicDimensionOption{
				Links:  nil,
				Option: code,
			})
		}

		WriteBody(ctx, w, options, http.StatusOK)
	})
}

func (api *API) GetDatasetDimensionByIndex() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataset := mux.Vars(r)["dataset"]
		dimension := mux.Vars(r)["name"]

		index, err := strconv.Atoi(mux.Vars(r)["index"])
		if err != nil {
			WriteBody(ctx, w, SimpleEntity{"invalid dimension index"}, http.StatusBadRequest)
			return
		}

		codebook, err := api.Store.GetDatasetCodebook(ctx, dataset)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}

		dim := codebook.GetDimension(dimension)
		if dim == nil {
			WriteBody(ctx, w, SimpleEntity{Message: "not found"}, http.StatusNotFound)
			return
		}

		if index < 0 || index > len(dim.Codes) {
			WriteBody(ctx, w, SimpleEntity{"invalid dimension index"}, http.StatusBadRequest)
			return
		}

		code := dim.Codes[index]
		label := code
		if dim.Labels != nil && len(dim.Labels) > index {
			label = code
		}

		d := DimensionResponse{
			Index: index,
			Name:  code,
			Code:  label,
		}

		WriteBody(ctx, w, d, http.StatusOK)
	})
}


func getErrorResponse(ctx context.Context, err error) (SimpleEntity, int) {
	log.Event(ctx, "returning http error response", log.ERROR, log.Error(err))

	status := http.StatusInternalServerError
	msg := "internal server error"

	var ftbErr cantabular.Error
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
