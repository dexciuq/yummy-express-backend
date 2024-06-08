package main

import (
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/data"
	"net/http"
)

func (app *application) updateOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	orderItem, err := app.models.OrderItems.Get(id)
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
		Quantity *float64 `json:"quantity"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Quantity == nil {
		app.badRequestResponse(w, r, errors.New("quantity is required"))
		return
	}

	price := int64(float64(orderItem.Total) / orderItem.Quantity)
	difference := int64((orderItem.Quantity - *input.Quantity) * float64(price))

	if orderItem.Quantity == 0 {
		err = app.models.OrderItems.Delete(orderItem.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	} else {
		orderItem.Total = int64(float64(price) * (*input.Quantity))
		err = app.models.OrderItems.Update(orderItem)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	order, err := app.models.Orders.Get(orderItem.OrderID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	order.Total -= difference

	err = app.models.Orders.Update(order)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"orderItem": orderItem}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
