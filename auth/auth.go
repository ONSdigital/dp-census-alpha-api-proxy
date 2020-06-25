package auth

import (
	"net/http"

	"github.com/ONSdigital/dp-census-alpha-api-proxy/api"
	"github.com/ONSdigital/log.go/log"
)

const (
	authHeader = "Authorization"
)

func Handler(token string) func(http.HandlerFunc) http.HandlerFunc {

	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			callerToken := r.Header.Get(authHeader)

			if len(callerToken) == 0 {
				log.Event(ctx, "unauthorized no token provided", log.INFO)
				api.WriteBody(ctx, w, api.SimpleEntity{Message: "unauthorized no token provided"}, http.StatusUnauthorized)
				return
			}

			if token != callerToken {
				log.Event(ctx, "unauthorized incorrect token provided", log.INFO)
				api.WriteBody(ctx, w, api.SimpleEntity{Message: "unauthorized incorrect token provided"}, http.StatusUnauthorized)
				return
			}

			log.Event(ctx, "valid credentials provided proceeding with request", log.INFO)
			h.ServeHTTP(w, r)
		}
	}
}
