package main

import (
	"net/http"
	"os"

	"github.com/ONSdigital/dp-census-alpha-api-proxy/api"
	"github.com/ONSdigital/dp-census-alpha-api-proxy/auth"
	"github.com/ONSdigital/dp-census-alpha-api-proxy/config"
	"github.com/ONSdigital/dp-census-alpha-api-proxy/ftb"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

const serviceName = "dp-census-alpha-api-proxy"

func main() {
	log.Namespace = serviceName

	if err := run(); err != nil {
		log.Event(nil, "fatal runtime error", log.Error(err), log.FATAL)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Get()
	if err != nil {
		return err
	}

	log.Event(nil, "application configuration", log.INFO, log.Data{"values": cfg})

	ftbCli := &ftb.Client{
		URL:     cfg.FlexibleTableBuilderURL,
		HttpCli: dphttp.NewClient(),
	}

	r := mux.NewRouter()

	api.Setup(nil, r, auth.Handler(cfg.GetAuthToken()), ftbCli)

	log.Event(nil, "starting ftb proxy api", log.INFO, log.Data{"port": cfg.BindAddr})
	return http.ListenAndServe(cfg.BindAddr, r)
}
