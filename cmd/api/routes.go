package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	//products
	router.HandlerFunc(http.MethodPost, "/v1/products", app.addProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products", app.listProductsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products/:id", app.showProductHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/products/:id", app.deleteProductHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/products/:id", app.updateProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/upc/:upc", app.findProductByUPCHandler)

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

	//roles
	router.HandlerFunc(http.MethodPost, "/v1/roles", app.addRoleHandler)
	router.HandlerFunc(http.MethodGet, "/v1/roles", app.listRolesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/roles/:id", app.showRoleHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/roles/:id", app.deleteRoleHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/roles/:id", app.updateRoleHandler)

	router.HandlerFunc(http.MethodPost, "/v1/auth/register", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/auth/authenticate", app.authenticateUserHandler)
	router.HandlerFunc(http.MethodGet, "/v1/auth/logout", app.authMiddleware(app.logoutUserHandler))
	router.HandlerFunc(http.MethodGet, "/v1/auth/refresh", app.refreshHandler)
	router.HandlerFunc(http.MethodPost, "/v1/request-password-reset", app.requestPasswordResetHandler)
	router.HandlerFunc(http.MethodPost, "/v1/verify-reset-code", app.verifyResetCodeHandler)
	router.HandlerFunc(http.MethodPost, "/v1/reset-password", app.resetPasswordHandler)

	router.HandlerFunc(http.MethodGet, "/v1/users/:id", app.showUserHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/users/:id", app.authMiddleware(app.updateUserHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/users/:id", app.authMiddleware(app.deleteUserHandler))
	router.HandlerFunc(http.MethodGet, "/v1/profile/me", app.authMiddleware(app.getUserInformationByToken))
	router.HandlerFunc(http.MethodGet, "/v1/auth/activate/:uuid", app.activateUserHandler)

	//orders
	router.HandlerFunc(http.MethodPost, "/v1/orders", app.addOrderHandler)
	router.HandlerFunc(http.MethodGet, "/v1/orders", app.listOrdersHandler)
	router.HandlerFunc(http.MethodGet, "/v1/profile/orders", app.authMiddleware(app.listUserOrdersHandler))
	router.HandlerFunc(http.MethodGet, "/v1/orders/:id", app.showOrderHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/orders/:id", app.deleteOrderHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/orders/:id", app.updateOrderHandler)

	// Enable CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE"},
		AllowedHeaders: []string{"Content-Type"},
	})

	return c.Handler(router)
}
