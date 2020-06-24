package auth

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/log.go/log"
)

const (
	authHeader   = "Authorization"
)

var (
	unauthorizedJson, _ = json.Marshal(simpleMessage{"unauthorized"})
	unauthorizedStr     = string(unauthorizedJson)
)

type simpleMessage struct {
	Message string
}

func Handler(token string) func(http.HandlerFunc) http.HandlerFunc {

	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			callerToken := r.Header.Get(authHeader)

			if len(callerToken) == 0 {
				log.Event(ctx, "unauthorized no token provided", log.ERROR)
				http.Error(w, unauthorizedStr, http.StatusUnauthorized)
				return
			}

			if token != callerToken {
				log.Event(ctx, "unauthorized incorrect token provided", log.ERROR)
				http.Error(w, unauthorizedStr, http.StatusForbidden)
				return
			}

			log.Event(ctx, "valid credentials provided proceeding with request", log.INFO)
			h.ServeHTTP(w, r)
		}
	}
}
