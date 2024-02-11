package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"
)

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	Total     int64   `json:"total"`
}

type OrderItemModel struct {
	DB *sql.DB
}

func ValidateOrderItem(v *validator.Validator, order *OrderItem) {
	//v.Check(order.Name != "", "name", "must be provided")
	//v.Check(len(order.Name) <= 20, "name", "must not be more than 20 bytes long")
	//v.Check(order.Price >= 0, "price", "can not be negative")
	//v.Check(order.Description != "", "description", "must be provided")
	//v.Check(order.Quantity >= 0, "quantity", "can not be negative")
}

func (o OrderItemModel) Insert(item *OrderItem) error {
	query := `
	INSERT INTO order_items (order_id, product_id, quantity, total)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	args := []any{
		item.OrderID,
		item.ProductID,
		item.Quantity,
		item.Total,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := o.DB.QueryRowContext(ctx, query, args...).Scan(&item.ID)

	if err != nil {
		return err
	}

	return nil
}

func (o OrderItemModel) GetAll() ([]*OrderItem, error) {
	// Update the SQL query to include the window function which counts the total
	// (filtered) records.
	query := `SELECT count(*) OVER(), id, order_id, product_id, quantity, total FROM order_items`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := o.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	items := []*OrderItem{}

	for rows.Next() {
		var item OrderItem
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Total,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return items, nil
}

func (o OrderItemModel) GetAllByOrder(order_id int64) ([]*OrderItem, error) {
	// Update the SQL query to include the window function which counts the total
	// (filtered) records.
	query := `SELECT count(*) OVER(), id, order_id, product_id, quantity, total FROM order_items where order_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := o.DB.QueryContext(ctx, query, order_id)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	items := []*OrderItem{}

	for rows.Next() {
		var item OrderItem
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Total,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return items, nil
}

func (o OrderItemModel) Get(id int64) (*OrderItem, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, order_id, product_id, quantity, total 
		FROM order_items
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var item OrderItem
	err := o.DB.QueryRow(query, id).Scan(
		&item.ID,
		&item.OrderID,
		&item.ProductID,
		&item.Quantity,
		&item.Total,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &item, nil
}

func (o OrderItemModel) Update(item *OrderItem) error {
	query := `UPDATE order_items
	SET order_id = $1, product_id = $2, quantity = $3, total = $4
	WHERE id = $5
	RETURNING id`

	args := []any{
		item.OrderID,
		item.ProductID,
		item.Quantity,
		item.Total,
		item.ID,
	}

	return o.DB.QueryRow(query, args...).Scan(&item.ID)
}

func (o OrderItemModel) Delete(id int64) error {
	query := `
		DELETE FROM order_items
		WHERE id = $1`
	result, err := o.DB.Exec(query, id)
	if err != nil {
		return nil
	}

	// Checking how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Check if the row was in the database before the query
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
