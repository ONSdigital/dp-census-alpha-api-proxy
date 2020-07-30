package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-census-alpha-api-proxy/cantabular"
	hierarchy "github.com/ONSdigital/dp-hierarchy-api/models"
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
				"code": {
					ID:   code,
					HRef: fmt.Sprintf("http://localhost:10100/v6/datasets/%s/hierarchies/%s/%s", dataset, dim.Name, code),
				},
				"self": {
					ID:   dim.Name,
					HRef: fmt.Sprintf("http://localhost:10100/v6/datasets/%s/hierarchies/%s", dataset, dim.Name),
				},
				"children": {
					ID:   dim.MapFrom[0],
					HRef: fmt.Sprintf("http://localhost:10100/v6/datasets/%s/hierarchies/%s", dataset, dim.MapFrom[0]),
				},
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
					"code": {
						ID:   descendentCode,
						HRef: fmt.Sprintf("http://localhost:10100/v6/datasets/%s/hierarchies/%s/%s", dataset, childName, descendentCode),
					},
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
		dimensionName := mux.Vars(r)["dimension"]

		codebook, err := api.Store.GetDatasetCodebook(context.Background(), dataset)
		if err != nil {
			errEntity, status := getErrorResponse(ctx, err)
			WriteBody(ctx, w, errEntity, status)
			return
		}

		rootDimension := codebook.GetDimension(dimensionName)
		h := cantabular.BuildHierarchyFrom(rootDimension, codebook)
		WriteBody(ctx, w, h, http.StatusOK)
	})
}
