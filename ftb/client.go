package ftb

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	dphttp "github.com/ONSdigital/dp-net/http"
)

type Client struct {
	URL     string
	HttpCli dphttp.Clienter
}

func (c *Client) GetDatasets(ctx context.Context) (*Datasets, error) {
	outReq, err := http.NewRequest("GET", c.URL+"/datasets", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HttpCli.Do(ctx, outReq)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("incorrect status code returned from ftb api")
	}

	var datasets []*Dataset
	if err := json.NewDecoder(resp.Body).Decode(&datasets); err != nil {
		return nil, err
	}

	return &Datasets{Items: datasets}, nil
}
