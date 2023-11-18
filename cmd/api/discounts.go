package main

import (
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/data"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"net/http"
	"time"
)

func (app *application) addDiscountHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name            string    `json:"name"`
		Description     string    `json:"description"`
		DiscountPercent int       `json:"discount_percent"`
		StartedAt       time.Time `json:"started_at"`
		EndedAt         time.Time `json:"ended_at"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	discount := &data.Discount{
		Name:            input.Name,
		Description:     input.Description,
		DiscountPercent: input.DiscountPercent,
		StartedAt:       input.StartedAt,
		EndedAt:         input.EndedAt,
	}

	v := validator.New()
	if data.ValidateDiscount(v, discount); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Discount.Insert(discount)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"discount": discount}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listDiscountsHandler(w http.ResponseWriter, r *http.Request) {
	// v := validator.New()
	discounts, err := app.models.Discount.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"discounts": discounts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showDiscountHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	discount, err := app.models.Discount.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Encode the struct to JSON and send it as the HTTP response.
	// using envelope
	err = app.writeJSON(w, http.StatusOK, envelope{"discount": discount}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateDiscountHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	discount, err := app.models.Discount.Get(id)
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
		Name            *string    `json:"name"`
		Description     *string    `json:"description"`
		DiscountPercent *int       `json:"discount_percent"`
		StartedAt       *time.Time `json:"started_at"`
		EndedAt         *time.Time `json:"ended_at"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		discount.Name = *input.Name
	}

	if input.Description != nil {
		discount.Description = *input.Description
	}

	if input.DiscountPercent != nil {
		discount.DiscountPercent = *input.DiscountPercent
	}

	if input.StartedAt != nil {
		discount.StartedAt = *input.StartedAt
	}

	if input.EndedAt != nil {
		discount.EndedAt = *input.EndedAt
	}

	v := validator.New()
	if data.ValidateDiscount(v, discount); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Discount.Update(discount)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"discount": discount}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteDiscountHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Discount.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "discount successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
