package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/dexciuq/yummy-express-backend/internal/data"
)

type envelope map[string]any

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) readUUIDParam(r *http.Request) (string, error) {
	params := httprouter.ParamsFromContext(r.Context())
	uuid := params.ByName("uuid")
	if len(uuid) == 0 {
		return "", errors.New("invalid uuid parameter")
	}
	return uuid, nil
}

func (app *application) readParamByNurik(r *http.Request, name string) (string, error) {
	params := httprouter.ParamsFromContext(r.Context())
	param := params.ByName(name)
	return param, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	err := json.NewDecoder(r.Body).Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		if errors.As(err, &syntaxError) {
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		} else if errors.As(err, &unmarshalTypeError) {
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", unmarshalTypeError.Offset)

		} else if errors.As(err, &invalidUnmarshalError) {
			panic(err)

		} else if errors.Is(err, io.ErrUnexpectedEOF) {
			return errors.New("body contains badly-formed JSON")

		} else if errors.Is(err, io.EOF) {
			return errors.New("body must not be empty")

		} else {
			return err
		}
	}

	return nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}
	return s
}

func (app *application) readIntArray(qs url.Values, key string, defaultValue []int) []int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	ids := strings.Split(s, ",")
	var result []int
	for _, id := range ids {
		elem, err := strconv.Atoi(id)
		if err == nil {
			result = append(result, elem)
		}
	}
	return result
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return i
}

func (app *application) background(function func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		function()
	}()
}

func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	tmpl, err := template.ParseFiles(fmt.Sprintf("templates/%s.tmpl", name))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserIDFromHeader(w http.ResponseWriter, r *http.Request) float64 {
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
	return userId
}
