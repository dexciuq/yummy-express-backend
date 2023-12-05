package main

import (
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/data"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"net/http"
	"strings"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName   string `json:"firstname"`
		LastName    string `json:"lastname"`
		PhoneNumber string `json:"phone_number"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		Role_ID     int64  `json:"role_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		PhoneNumber: input.PhoneNumber,
		Email:       input.Email,
		Role_ID:     input.Role_ID,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) authenticateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := data.GenerateTokens(user.ID, user.Role_ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	if err = app.models.Tokens.SaveToken(token); err != nil {
		app.serverErrorResponse(w, r, err)
	}

	refreshTokenCookie := http.Cookie{
		Name:     "refreshToken",
		Value:    token.RefreshToken,
		HttpOnly: true,
		MaxAge:   30 * 24 * 60 * 60,
	}

	http.SetCookie(w, &refreshTokenCookie)
	if err = app.writeJSON(w, http.StatusOK, envelope{"refreshToken": token.RefreshToken, "accessToken": token.AccessToken}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) refreshHandler(w http.ResponseWriter, r *http.Request) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		app.UserUnauthorizedResponse(w, r)
	}

	accessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")
	accessTokenMap, err := data.DecodeAccessToken(accessToken)

	if err != nil {
		app.UserUnauthorizedResponse(w, r)
	}

	userId := accessTokenMap["user_id"].(float64)
	refreshToken, err := app.models.Tokens.FindTokenByUserId(int64(userId))
	if err != nil {
		app.UserUnauthorizedResponse(w, r)
	}
	_, err = data.DecodeRefreshToken(refreshToken.RefreshToken)
	if err != nil {
		app.UserUnauthorizedResponse(w, r)
	}
	userForToken, err := app.models.Users.GetById(refreshToken.UserID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	token, err := data.GenerateTokens(userForToken.ID, userForToken.Role_ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	if err = app.models.Tokens.SaveToken(token); err != nil {
		app.serverErrorResponse(w, r, err)
	}

	refreshTokenCookie := http.Cookie{
		Name:     "refreshToken",
		Value:    token.RefreshToken,
		HttpOnly: true,
		MaxAge:   30 * 24 * 60 * 60,
	}

	http.SetCookie(w, &refreshTokenCookie)
	if err = app.writeJSON(w, http.StatusOK, envelope{"refreshToken": token.RefreshToken, "accessToken": token.AccessToken}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) logoutUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refreshToken"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	authorizationHeader := r.Header.Get("Authorization")
	accessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")

	accessTokenMap, err := data.DecodeAccessToken(accessToken)
	if err != nil {
		app.UserUnauthorizedResponse(w, r)
	}

	userId := accessTokenMap["user_id"].(float64)

	token, err := app.models.Tokens.FindTokenByUserId(int64(userId))
	if err != nil {
		app.UserUnauthorizedResponse(w, r)
	}

	if token.RefreshToken != input.RefreshToken {
		app.UserUnauthorizedResponse(w, r)
	}

	err = app.models.Tokens.RemoveToken(token.RefreshToken)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	logoutCookie := http.Cookie{
		Name:   "refreshToken",
		MaxAge: -1,
	}
	http.SetCookie(w, &logoutCookie)
	app.writeJSON(w, http.StatusOK, envelope{"refreshToken": token.RefreshToken}, nil)

}

func (app *application) showUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user, err := app.models.Users.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user, err := app.models.Users.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		FirstName   *string `json:"firstname"`
		LastName    *string `json:"lastname"`
		PhoneNumber *string `json:"phone_number"`
		Email       *string `json:"email"`
		Password    *string `json:"password"`
		Role_ID     *int64  `json:"role_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}
	if input.PhoneNumber != nil {
		user.PhoneNumber = *input.PhoneNumber
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Password != nil {
		user.Password.Set(*input.Password)
	}
	if input.Role_ID != nil {
		user.Role_ID = *input.Role_ID
	}

	err = app.models.Users.Update(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Users.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "user successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
