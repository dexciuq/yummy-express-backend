package main

import (
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/data"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"net/http"
)

func (app *application) addProductHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string  `json:"name"`
		Price       int64   `json:"price"`
		Description string  `json:"description"`
		CategoryID  int64   `json:"category_id"`
		UPC         string  `json:"upc"`
		DiscountID  int64   `json:"discount_id"`
		Quantity    int64   `json:"quantity"`
		UnitID      int64   `json:"unit_id"`
		Image       string  `json:"image"`
		BrandID     int64   `json:"brand_id"`
		CountryID   int64   `json:"country_id"`
		Step        float64 `json:"step"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	product := &data.Product{
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
		CategoryID:  input.CategoryID,
		UPC:         input.UPC,
		DiscountID:  input.DiscountID,
		Quantity:    input.Quantity,
		UnitID:      input.UnitID,
		Image:       input.Image,
		BrandID:     input.BrandID,
		CountryID:   input.CountryID,
		Step:        input.Step,
	}

	v := validator.New()
	if data.ValidateProduct(v, product); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Products.Insert(product)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"product": product}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listProductsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CategoryID int
		BrandID    int
		CountryID  int
		data.Filters
	}
	// v := validator.New()
	qs := r.URL.Query()
	input.CategoryID = app.readInt(qs, "category", 0)
	input.BrandID = app.readInt(qs, "brand", 0)
	input.CountryID = app.readInt(qs, "country", 0)
	input.Filters.Page = app.readInt(qs, "page", 1)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "price",
		"-id", "-name", "-price"}
	// if data.ValidateFilters(v, input.Filters); !v.Valid() {
	// 	app.failedValidationResponse(w, r, v.Errors)
	// 	return
	// }
	// Call the GetAll() method to retrieve the movies, passing in the various filter
	// parameters.
	// Accept the metadata struct as a return value.
	products, metadata, err := app.models.Products.GetAll(input.CategoryID, input.BrandID, input.CountryID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Include the metadata in the response envelope.
	err = app.writeJSON(w, http.StatusOK, envelope{"products": products, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	product, err := app.models.Products.Get(id)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"product": product}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	product, err := app.models.Products.Get(id)
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
		Name        *string  `json:"name"`
		Price       *int64   `json:"price"`
		Description *string  `json:"description"`
		CategoryID  *int64   `json:"category_id"`
		UPC         *string  `json:"upc"`
		DiscountID  *int64   `json:"-"`
		Quantity    *int64   `json:"-"`
		UnitID      *int64   `json:"unit_id"`
		Image       *string  `json:"image"`
		BrandID     *int64   `json:"brand_id"`
		CountryID   *int64   `json:"country_id"`
		Step        *float64 `json:"step"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		product.Name = *input.Name
	}

	if input.Price != nil {
		product.Price = *input.Price
	}

	if input.Description != nil {
		product.Description = *input.Description
	}

	if input.CategoryID != nil {
		product.CategoryID = *input.CategoryID
	}

	if input.UPC != nil {
		product.UPC = *input.UPC
	}

	if input.DiscountID != nil {
		product.DiscountID = *input.DiscountID
	}

	if input.Quantity != nil {
		product.Quantity = *input.Quantity
	}

	if input.UnitID != nil {
		product.UnitID = *input.UnitID
	}

	if input.Image != nil {
		product.Image = *input.Image
	}

	if input.BrandID != nil {
		product.BrandID = *input.BrandID
	}

	if input.CountryID != nil {
		product.CountryID = *input.CountryID
	}

	if input.Step != nil {
		product.Step = *input.Step
	}

	v := validator.New()
	if data.ValidateProduct(v, product); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Products.Update(product)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"product": product}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Products.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "product successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
