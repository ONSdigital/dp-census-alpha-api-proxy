package ftb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

type Client struct {
	Host    string
	HttpCli dphttp.Clienter
}

type Entity interface{}

func (c *Client) GetData(ctx context.Context, url string) (Entity, error) {
	queryURL := c.Host + url
	logD := log.Data{"url": queryURL}
	log.Event(ctx, "making request to FTB API", log.INFO, logD)

	outReq, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HttpCli.Do(ctx, outReq)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("incorrect status code returned from ftb api expected 200 but was %d", resp.StatusCode)
	}

	logD["status"] = resp.StatusCode
	log.Event(ctx, "flexible table builder returned successful response", log.INFO, logD)

	var entity interface{}
	if err := json.NewDecoder(resp.Body).Decode(&entity); err != nil {
		return nil, err
	}

	return entity, nil
}
