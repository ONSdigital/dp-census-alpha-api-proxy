package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-census-alpha-api-proxy/cantabular"
	"github.com/ONSdigital/dp-census-alpha-api-proxy/config"
	hierarchy "github.com/ONSdigital/dp-hierarchy-api/models"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

func (api *API) GetHierarchy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataset := mux.Vars(r)["dataset"]
		h := mux.Vars(r)["name"]

		cb, err := api.Store.GetDatasetCodebook(ctx, dataset)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}

		dim := cb.GetDimension(h)
		if dim == nil {
			WriteBody(ctx, w, SimpleEntity{Message: "not found"}, http.StatusNotFound)
			return
		}

		hierarchyCodes := getHierarchyLevel(dataset, dim.Name, cb)
		WriteBody(ctx, w, hierarchyCodes, http.StatusOK)
	})
}

func getHierarchyLevel(dataset, rootDim string, cb *cantabular.Codebook) *hierarchy.Response {
	dim := cb.GetDimension(rootDim)

	elements := make([]*hierarchy.Element, 0)
	for i, code := range dim.Codes {
		index, found := dim.GetDescendantCodeIndices(code)

		el := &hierarchy.Element{
			Label:        dim.Labels[i],
			NoOfChildren: int64(index.Count),
			Links: map[string]hierarchy.Link{
				"code":     newLink(code, fmt.Sprintf("/v6/datasets/%s/hierarchies/%s/code/%s", dataset, dim.Name, code)),
				"self":     newLink(dim.Name, fmt.Sprintf("/v6/datasets/%s/hierarchies/%s", dataset, dim.Name)),
				"children": newLink(dim.MapFrom[0], fmt.Sprintf("/v6/datasets/%s/hierarchies/%s", dataset, dim.MapFrom[0])),
			},
			HasData: found,
		}

		elements = append(elements, el)
	}

	return &hierarchy.Response{
		ID:           dim.MapFrom[0],
		Label:        dim.Label,
		Children:     elements,
		NoOfChildren: int64(len(elements)),
		Links:        nil,
		HasData:      len(elements) > 0,
		Breadcrumbs:  nil,
	}
}

func (api *API) GetHierarchyForCode() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataset := mux.Vars(r)["dataset"]
		d := mux.Vars(r)["name"]
		dimensionCode := mux.Vars(r)["code"]

		codebook, err := api.Store.GetDatasetCodebook(ctx, dataset)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}

		rootDim := codebook.GetDimension(d)
		if rootDim == nil {
			WriteBody(ctx, w, SimpleEntity{Message: "not found"}, http.StatusNotFound)
			return
		}

		h := getHierarchyEntry(dataset, dimensionCode, rootDim, codebook)
		WriteBody(ctx, w, h, http.StatusOK)
	})
}

func getHierarchyEntry(dataset, dimensionCode string, rootDim *cantabular.Dimension, cb *cantabular.Codebook) *hierarchy.Response {
	if len(rootDim.MapFrom) == 0 {
		return &hierarchy.Response{
			ID:           dimensionCode,
			Label:        dimensionCode,
			Children:     nil,
			NoOfChildren: 0,
			Links:        nil,
			HasData:      false,
			Breadcrumbs:  nil,
		}
	}

	elements := make([]*hierarchy.Element, 0)

	childName := rootDim.MapFrom[0]
	childDim := cb.GetDimension(childName)

	index, found := rootDim.GetDescendantCodeIndices(dimensionCode)

	if found {
		for i := index.Start; i <= index.End; i++ {
			descendentCode := childDim.Codes[i]
			descendentIndex, descendentsExist := childDim.GetDescendantCodeIndices(descendentCode)

			el := &hierarchy.Element{
				Label:        descendentCode,
				NoOfChildren: 0,
				Links: map[string]hierarchy.Link{
					"code": newLink(descendentCode, fmt.Sprintf("/v6/datasets/%s/hierarchies/%s/code/%s", dataset, childName, descendentCode)),
				},
				HasData: descendentsExist,
			}

			if descendentsExist {
				el.Label = childDim.Labels[i]
				el.NoOfChildren = int64(descendentIndex.Count)
			}

			elements = append(elements, el)
		}
	}

	return &hierarchy.Response{
		ID:           dimensionCode,
		Label:        childDim.Label,
		Children:     elements,
		NoOfChildren: int64(len(elements)),
		Links:        nil,
		HasData:      len(elements) > 0,
		Breadcrumbs:  nil,
	}
}

func (api *API) BuildFullHierarchy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataset := mux.Vars(r)["dataset"]
		dimensionName := mux.Vars(r)["name"]

		depth, err := strconv.Atoi(r.URL.Query().Get("depth"))
		if err != nil {
			log.Event(ctx, "invalid depth provided applying default", log.WARN)
			depth = 2 // default to a sensible value
		}

		codebook, err := api.Store.GetDatasetCodebook(context.Background(), dataset)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}

		rootDimension := codebook.GetDimension(dimensionName)
		h := cantabular.BuildHierarchyFrom(rootDimension, codebook, depth)
		WriteBody(ctx, w, h, http.StatusOK)
	})
}

func newLink(id, path string) hierarchy.Link {
	cfg, _ := config.Get()
	return hierarchy.Link{
		ID:   id,
		HRef: fmt.Sprintf("http://%s%s%s", cfg.IPAddr, cfg.BindAddr, path),
	}
}
