package main

import (
	"errors"
	"fmt"
	"github.com/dexciuq/yummy-express-backend/internal/data"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"math"
	"net/http"
	"strings"
	"time"
)

func (app *application) addOrderHandler(w http.ResponseWriter, r *http.Request) {
	authorizationHeader := r.Header.Get("Authorization")
	accessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")
	accessTokenMap, _ := data.DecodeAccessToken(accessToken)

	userId := accessTokenMap["user_id"].(float64)

	type product struct {
		ID     int64   `json:"id"`
		Price  int64   `json:"price"`
		Amount float64 `json:"amount"`
	}
	var input struct {
		Total    int64     `json:"total"`
		Address  string    `json:"address"`
		Products []product `json:"products"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	order := &data.Order{
		UserID:   int64(userId),
		Total:    input.Total,
		Address:  input.Address,
		StatusID: int64(1),
	}

	v := validator.New()
	if data.ValidateOrder(v, order); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Orders.Insert(order)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	for _, product := range input.Products {
		item := &data.OrderItem{
			OrderID:   order.ID,
			ProductID: product.ID,
			Quantity:  product.Amount,
			Total:     int64(math.Floor(float64(product.Price) * product.Amount)),
		}
		if data.ValidateOrderItem(v, item); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}

		err = app.models.OrderItems.Insert(item)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"order": order}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders, err := app.models.Orders.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"orders": orders}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	authorizationHeader := r.Header.Get("Authorization")
	accessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")
	accessTokenMap, err := data.DecodeAccessToken(accessToken)
	userId := accessTokenMap["user_id"].(float64)

	orders, err := app.models.Orders.GetAllForUser(int(userId))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"orders": orders}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showOrderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}
	fmt.Println("order_id: ", id)

	order, err := app.models.Orders.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	items, err := app.models.OrderItems.GetAllByOrder(order.ID)

	type ProductItem struct {
		ID          int64   `json:"id"`
		ProductID   int64   `json:"product_id"`
		Name        string  `json:"name"`
		Price       int64   `json:"price"`
		Description string  `json:"description"`
		UPC         string  `json:"upc"`
		Quantity    int64   `json:"quantity"`
		Step        float64 `json:"step"`
		Amount      float64 `json:"amount"`
		Subtotal    int64   `json:"subtotal"`
		Image       string  `json:"image"`
		Unit        string  `json:"unit"`
		Category    string  `json:"category"`
		Brand       string  `json:"brand"`
		Country     string  `json:"country"`
	}

	var productItems []ProductItem
	for _, item := range items {
		product, _ := app.models.Products.Get(item.ProductID)
		unit, _ := app.models.Units.Get(product.UnitID)
		category, _ := app.models.Category.Get(product.CategoryID)
		brand, _ := app.models.Brands.Get(product.BrandID)
		country, _ := app.models.Country.Get(product.CountryID)

		productItem := ProductItem{
			ID:          item.ID,
			ProductID:   product.ID,
			Name:        product.Name,
			Price:       product.Price,
			Description: product.Description,
			Category:    category.Name,
			UPC:         product.UPC,
			Quantity:    product.Quantity,
			Unit:        unit.Name,
			Image:       product.Image,
			Brand:       brand.Name,
			Country:     country.Name,
			Step:        product.Step,
			Amount:      item.Quantity,
			Subtotal:    item.Total,
		}
		fmt.Println("order_id", order.ID, "subtotal", productItem.Subtotal, "item.total", item.Total)
		productItems = append(productItems, productItem)
	}
	// Encode the struct to JSON and send it as the HTTP response.
	// using envelope
	fmt.Println("order:", order)
	fmt.Println("order_items:", productItems)
	err = app.writeJSON(w, http.StatusOK, envelope{"order": order, "order_items": productItems}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateOrderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	order, err := app.models.Orders.Get(id)
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
		UserID      *int64     `json:"user_id"`
		Total       *int64     `json:"total"`
		Address     *string    `json:"address"`
		StatusID    *int64     `json:"status_id"`
		DeliveredAt *time.Time `json:"delivered_at"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.UserID != nil {
		order.UserID = *input.UserID
	}

	if input.Total != nil {
		order.Total = *input.Total
	}

	if input.Address != nil {
		order.Address = *input.Address
	}

	if input.StatusID != nil {
		order.StatusID = *input.StatusID
	}

	if input.DeliveredAt != nil {
		order.DeliveredAt = *input.DeliveredAt
	}

	v := validator.New()
	if data.ValidateOrder(v, order); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Orders.Update(order)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"order": order}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Orders.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "order successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
