package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	//categories
	router.HandlerFunc(http.MethodPost, "/v1/categories", app.addCategoryHandler)
	router.HandlerFunc(http.MethodGet, "/v1/categories", app.listCategoriesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/categories/:id", app.showCategoryHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/categories/:id", app.deleteCategoryHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/categories/:id", app.updateCategoryHandler)

	//units
	router.HandlerFunc(http.MethodPost, "/v1/units", app.addUnitHandler)
	router.HandlerFunc(http.MethodGet, "/v1/units", app.listUnitsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/units/:id", app.showUnitHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/units/:id", app.deleteUnitHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/units/:id", app.updateUnitHandler)

	//brands
	router.HandlerFunc(http.MethodPost, "/v1/brands", app.addBrandHandler)
	router.HandlerFunc(http.MethodGet, "/v1/brands", app.listBrandsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/brands/:id", app.showBrandHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/brands/:id", app.deleteBrandHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/brands/:id", app.updateBrandHandler)

	//countries
	router.HandlerFunc(http.MethodPost, "/v1/countries", app.addCountryHandler)
	router.HandlerFunc(http.MethodGet, "/v1/countries", app.listCountriesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/countries/:id", app.showCountryHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/countries/:id", app.deleteCountryHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/countries/:id", app.updateCountryHandler)

	//discounts
	router.HandlerFunc(http.MethodPost, "/v1/discounts", app.addDiscountHandler)
	router.HandlerFunc(http.MethodGet, "/v1/discounts", app.listDiscountsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/discounts/:id", app.showDiscountHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/discounts/:id", app.deleteDiscountHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/discounts/:id", app.updateDiscountHandler)

	return router
}
