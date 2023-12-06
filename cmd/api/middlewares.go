package main

import (
	"fmt"
	"github.com/dexciuq/yummy-express-backend/internal/data"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func (app *application) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var flag bool
		flag = false
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			flag = true
		}

		accessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")

		fmt.Println("accessToken |", accessToken, "|")

		if accessToken == "" {
			flag = true
		}

		accessTokenMap, err := data.DecodeAccessToken(accessToken)
		fmt.Println(accessTokenMap)
		fmt.Println(len(accessTokenMap))

		if err != nil {
			flag = true
		}

		if len(accessTokenMap) == 0 {
			flag = true
			//			app.errorResponse(w, r, http.StatusUnauthorized, "access token was expired")
		}

		if flag == false {
			exp := int64(accessTokenMap["exp"].(float64))
			expUnix := time.Unix(exp, 0)
			fmt.Println(exp, reflect.TypeOf(exp))
			fmt.Println(expUnix, reflect.TypeOf(expUnix))
			fmt.Println(time.Now().After(expUnix), time.Now())
			if time.Now().After(expUnix) {
				flag = true
				fmt.Println("token expired")
				//				app.errorResponse(w, r, http.StatusUnauthorized, "access token was expired")
			}
			userId := accessTokenMap["user_id"].(float64)
			_, err = app.models.Users.GetById(int64(userId))

			if err != nil {
				flag = true
			}
		}

		if flag == false {
			next.ServeHTTP(w, r)
		} else {
			app.UserUnauthorizedResponse(w, r)
		}
	})
}
