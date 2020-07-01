package middleware

import (
	"net/http"

	"github.com/ONSdigital/dp-census-alpha-api-proxy/api"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"
)

const (
	authHeader = "Authorization"
)

func RequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := common.WithRequestId(r.Context(), dphttp.NewRequestID(16))
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Auth(token string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		})
	}
}
