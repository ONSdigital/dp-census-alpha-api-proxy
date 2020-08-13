package api

type GetDimensionsResponse struct {
	Dimensions []string `json:"dimensions,omitempty"`
}

type DimensionResponse struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Code  string `json:"code"`
}