package main

import (
	"github.com/dexciuq/yummy-express-backend/internal/data"
	"net/http"
	"strings"
)

func (app *application) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			app.UserUnauthorizedResponse(w, r)
		}

		accessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")

		if accessToken == "" {
			app.UserUnauthorizedResponse(w, r)
		}

		accessTokenMap, err := data.DecodeAccessToken(accessToken)

		if err != nil {
			app.UserUnauthorizedResponse(w, r)
		}

		userId := accessTokenMap["user_id"].(float64)
		_, err = app.models.Users.GetById(int64(userId))

		if err != nil {
			app.UserUnauthorizedResponse(w, r)
		}

		next.ServeHTTP(w, r)
	})
}
