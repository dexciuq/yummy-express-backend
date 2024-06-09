package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dexciuq/yummy-express-backend/internal/validator"
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

type OrderDB struct {
	ID                int64     `json:"id"`
	UserID            int64     `json:"user_id"`
	FirstName         string    `json:"firstname"`
	LastName          string    `json:"lastname"`
	Email             string    `json:"email"`
	Total             int64     `json:"total"`
	Address           string    `json:"address"`
	StatusID          int64     `json:"status_id"`
	CreatedAt         time.Time `json:"created_at"`
	DeliveredAt       time.Time `json:"delivered_at"`
	StatusName        string    `json:"status_name"`
	StatusDescription string    `json:"status_description"`
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

func (o OrderModel) GetAll() ([]*OrderDB, error) {
	query := `
		SELECT 
			count(*) OVER(), 
			o.id, 
			o.user_id, 
			u.firstname,
			u.lastname,
			u.email,
			o.total, 
			o.address, 
			o.status_id, 
			o.created_at, 
			o.delivered_at,
			s.name AS status_name,
			s.description AS status_description
		FROM orders o
		INNER JOIN users u ON o.user_id = u.id
		INNER JOIN statuses s ON o.status_id = s.id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := o.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalRecords := 0
	orders := []*OrderDB{}

	for rows.Next() {
		var order OrderDB
		err := rows.Scan(
			&totalRecords,
			&order.ID,
			&order.UserID,
			&order.FirstName,
			&order.LastName,
			&order.Email,
			&order.Total,
			&order.Address,
			&order.StatusID,
			&order.CreatedAt,
			&order.DeliveredAt,
			&order.StatusName,
			&order.StatusDescription,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (o OrderModel) GetAllForUser(id int) ([]*OrderDB, error) {
	query := `
		SELECT 
			count(*) OVER(), 
			o.id, 
			o.user_id, 
			u.firstname,
			u.lastname,
			u.email,
			o.total, 
			o.address, 
			o.status_id, 
			o.created_at, 
			o.delivered_at,
			s.name AS status_name,
			s.description AS status_description
		FROM orders o
		INNER JOIN users u ON o.user_id = u.id
		INNER JOIN statuses s ON o.status_id = s.id
		WHERE o.user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := o.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalRecords := 0
	orders := []*OrderDB{}

	for rows.Next() {
		var order OrderDB
		err := rows.Scan(
			&totalRecords,
			&order.ID,
			&order.UserID,
			&order.FirstName,
			&order.LastName,
			&order.Email,
			&order.Total,
			&order.Address,
			&order.StatusID,
			&order.CreatedAt,
			&order.DeliveredAt,
			&order.StatusName,
			&order.StatusDescription,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
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

func (o OrderModel) GetDB(id int64) (*OrderDB, error) {
	query := `
		SELECT 
			o.id, 
			o.user_id, 
			u.firstname,
			u.lastname,
			u.email,
			o.total, 
			o.address, 
			o.status_id, 
			o.created_at, 
			o.delivered_at,
			s.name AS status_name,
			s.description AS status_description
		FROM orders o
		INNER JOIN statuses s ON o.status_id = s.id
		INNER JOIN users u ON o.user_id = u.id
		WHERE o.id = $1
	`

	var order OrderDB
	err := o.DB.QueryRow(query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.FirstName,
		&order.LastName,
		&order.Email,
		&order.Total,
		&order.Address,
		&order.StatusID,
		&order.CreatedAt,
		&order.DeliveredAt,
		&order.StatusName,
		&order.StatusDescription,
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
