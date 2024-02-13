package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"
)

type Order struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Total       int64     `json:"total"`
	Address     string    `json:"address"`
	StatusID    int64     `json:"status_id"`
	CreatedAt   time.Time `json:"created_at"`
	DeliveredAt time.Time `json:"delivered_at"`
}

type OrderModel struct {
	DB *sql.DB
}

func ValidateOrder(v *validator.Validator, order *Order) {
	//v.Check(order.Name != "", "name", "must be provided")
	//v.Check(len(order.Name) <= 20, "name", "must not be more than 20 bytes long")
	//v.Check(order.Price >= 0, "price", "can not be negative")
	//v.Check(order.Description != "", "description", "must be provided")
	//v.Check(order.Amount >= 0, "quantity", "can not be negative")
}

func (o OrderModel) Insert(order *Order) error {
	query := `
	INSERT INTO orders (user_id, total, address, status_id, delivered_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at`

	args := []any{
		order.UserID,
		order.Total,
		order.Address,
		order.StatusID,
		order.DeliveredAt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := o.DB.QueryRowContext(ctx, query, args...).Scan(&order.ID,
		&order.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (o OrderModel) GetAll() ([]*Order, error) {
	// Update the SQL query to include the window function which counts the total
	// (filtered) records.
	query := `SELECT count(*) OVER(), id, user_id, total, address, status_id, created_at, delivered_at FROM orders`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := o.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	orders := []*Order{}

	for rows.Next() {
		var order Order
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&order.ID,
			&order.UserID,
			&order.Total,
			&order.Address,
			&order.StatusID,
			&order.CreatedAt,
			&order.DeliveredAt,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return orders, nil
}

func (o OrderModel) GetAllForUser(id int) ([]*Order, error) {
	// Update the SQL query to include the window function which counts the total
	// (filtered) records.
	query := `SELECT count(*) OVER(), id, user_id, total, address, status_id, created_at, delivered_at FROM orders where user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := o.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	orders := []*Order{}

	for rows.Next() {
		var order Order
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&order.ID,
			&order.UserID,
			&order.Total,
			&order.Address,
			&order.StatusID,
			&order.CreatedAt,
			&order.DeliveredAt,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return orders, nil
}

func (o OrderModel) Get(id int64) (*Order, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, user_id, total, address, status_id, created_at, delivered_at 
		FROM orders
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var order Order
	err := o.DB.QueryRow(query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Total,
		&order.Address,
		&order.StatusID,
		&order.CreatedAt,
		&order.DeliveredAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &order, nil
}

func (o OrderModel) Update(order *Order) error {
	query := `UPDATE orders
	SET user_id = $1, total = $2, address = $3, status_id = $4, delivered_at = $5
	WHERE id = $6
	RETURNING id`

	args := []any{
		order.UserID,
		order.Total,
		order.Address,
		order.StatusID,
		order.DeliveredAt,
		order.ID,
	}

	return o.DB.QueryRow(query, args...).Scan(&order.ID)
}

func (o OrderModel) Delete(id int64) error {
	query := `
		DELETE FROM orders
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
