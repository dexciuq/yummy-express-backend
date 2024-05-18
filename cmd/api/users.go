package main

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dexciuq/yummy-express-backend/internal/data"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
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
			fmt.Println("Duplicate email users")
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			fmt.Println("Not duplicate email users")
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	uuidCode := strings.Replace(uuid.New().String(), "-", "", -1)
	err = app.models.ActivationLinks.Insert(user, uuidCode)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"name":  user.FirstName + " " + user.LastName,
			"email": user.Email,
			"uuid":  uuidCode,
		}
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	templName := "account_activation"
	uuidParam, err := app.readUUIDParam(r)
	if err != nil {
		app.renderTemplate(w, r, templName, map[string]interface{}{
			"title":   "Error",
			"message": "Invalid activation link.",
		})
		return
	}

	activationLink, err := app.models.ActivationLinks.Get(uuidParam)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.renderTemplate(w, r, templName, map[string]interface{}{
				"title":   "Error",
				"message": "Activation link not found.",
			})
		default:
			app.renderTemplate(w, r, templName, map[string]interface{}{
				"title":   "Error",
				"message": "Internal server error.",
			})
		}
		return
	}

	user, err := app.models.Users.GetById(activationLink.UserID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.renderTemplate(w, r, templName, map[string]interface{}{
				"title":   "Error",
				"message": "User not found.",
			})
		default:
			app.renderTemplate(w, r, templName, map[string]interface{}{
				"title":   "Error",
				"message": "Internal server error.",
			})
		}
		return
	}

	if uuidParam == activationLink.Link {
		user.Activated = true
		activationLink.Activated = true
	} else {
		app.renderTemplate(w, r, templName, map[string]interface{}{
			"title":   "Error",
			"message": "Invalid activation link.",
		})
	}

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.renderTemplate(w, r, templName, map[string]interface{}{
				"title":   "Error",
				"message": "Edit conflict. Please try again.",
			})
		default:
			app.renderTemplate(w, r, templName, map[string]interface{}{
				"title":   "Error",
				"message": "Internal server error.",
			})
		}
		return
	}

	err = app.models.ActivationLinks.Update(activationLink)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.renderTemplate(w, r, templName, map[string]interface{}{
				"title":   "Error",
				"message": "Edit conflict. Please try again.",
			})
		default:
			app.renderTemplate(w, r, templName, map[string]interface{}{
				"title":   "Error",
				"message": "Internal server error.",
			})
		}
		return
	}

	app.renderTemplate(w, r, templName, map[string]interface{}{
		"title":   "Account Activated",
		"message": "Your account has been activated successfully.",
	})
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

	if !user.Activated {
		app.notActivatedResponse(w, r)
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
	fmt.Println("Autorization tokens time")
	fmt.Println("Access token")
	accessTokenMap, _ := data.DecodeAccessToken(token.AccessToken)
	exp := int64(accessTokenMap["exp"].(float64))
	expUnix := time.Unix(exp, 0)
	fmt.Println(exp, reflect.TypeOf(exp))
	fmt.Println(expUnix, reflect.TypeOf(expUnix))
	fmt.Println(time.Now().After(expUnix), time.Now())
	if time.Now().After(expUnix) {
		fmt.Println("token expired")
		//				app.errorResponse(w, r, http.StatusUnauthorized, "access token was expired")
	}
	fmt.Println("Refresh token")
	refreshTokenMap, _ := data.DecodeRefreshToken(token.RefreshToken)
	exp = int64(refreshTokenMap["exp"].(float64))
	expUnix = time.Unix(exp, 0)
	fmt.Println(exp, reflect.TypeOf(exp))
	fmt.Println(expUnix, reflect.TypeOf(expUnix))
	fmt.Println(time.Now().After(expUnix), time.Now())
	if time.Now().After(expUnix) {
		fmt.Println("token expired")
		//				app.errorResponse(w, r, http.StatusUnauthorized, "access token was expired")
	}
	fmt.Println("-------------------------------")

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

	refreshToken := strings.TrimPrefix(authorizationHeader, "Bearer ")
	fmt.Println("Refreshing access token, refresh token:", refreshToken)
	accessToken, err := app.models.Tokens.RefreshAccessToken(refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrTokenExpired):
			fmt.Println("Refresh/maybe token expired")
			app.errorResponse(w, r, http.StatusUnauthorized, "Refresh/maybe token expired")
		default:
			fmt.Println("Not refresh/maybe token expired")
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if accessToken == "" {
		app.serverErrorResponse(w, r, data.ErrTokenExpired)
	}

	if err = app.writeJSON(w, http.StatusOK, envelope{"accessToken": accessToken}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserInformationByToken(w http.ResponseWriter, r *http.Request) {
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
	user, err := app.models.Users.GetById(int64(userId))

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

func (app *application) logoutUserHandler(w http.ResponseWriter, r *http.Request) {
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

	err = app.models.Tokens.RemoveToken(token.RefreshToken)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	logoutCookie := http.Cookie{
		Name:   "refreshToken",
		MaxAge: -1,
	}
	http.SetCookie(w, &logoutCookie)
	app.writeJSON(w, http.StatusOK, envelope{"message": "you was successfully logouted"}, nil)

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

func (app *application) requestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the email
	v := validator.New()
	v.Check(validator.Matches(input.Email, validator.EmailRX), "email", "must be a valid email address")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if user == nil {
		app.notFoundResponse(w, r)
		return
	}

	code, err := data.GenerateResetCode()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	expiration := time.Now().Add(1 * time.Hour)
	err = app.models.Users.InsertPasswordResetCode(user.ID, code, expiration)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send email with the reset code
	app.background(func() {
		data := map[string]any{
			"name":      user.FirstName + " " + user.LastName,
			"email":     user.Email,
			"resetCode": code,
		}
		err = app.mailer.Send(user.Email, "reset_password_code.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Password reset code has been sent to your email"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) verifyResetCodeHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code string `json:"code"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	resetCode, err := app.models.Users.ValidatePasswordResetCode(input.Code)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if resetCode == nil || time.Now().After(resetCode.ExpiresAt) {
		app.invalidCredentialsResponse(w, r)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Code is valid"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) setNewPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	resetCode, err := app.models.Users.ValidatePasswordResetCode(input.Code)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if resetCode == nil || time.Now().After(resetCode.ExpiresAt) {
		app.invalidCredentialsResponse(w, r)
		return
	}

	user, err := app.models.Users.GetById(resetCode.User_ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = user.Password.Set(input.NewPassword)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.Update(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.DeletePasswordResetCode(input.Code)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Password has been reset successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	resetCode, err := app.models.Users.ValidatePasswordResetCode(input.Code)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if resetCode == nil || time.Now().After(resetCode.ExpiresAt) {
		app.errorResponse(w, r, http.StatusBadRequest, "Invalid or expired code")
		return
	}
	user, err := app.models.Users.GetById(resetCode.User_ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if input.NewPassword != "" {
		user.Password.Set(input.NewPassword)
	}

	//err = app.models.Users.UpdateUserPassword(resetCode.User_ID, input.NewPassword)
	//if err != nil {
	//	http.Error(w, "Internal server error", http.StatusInternalServerError)
	//	return
	//}
	err = app.models.Users.Update(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.DeletePasswordResetCode(input.Code)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Password has been reset successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
