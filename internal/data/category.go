package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"
)

type Category struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CategoryModel struct {
	DB *sql.DB
}

func ValidateCategory(v *validator.Validator, category *Category) {
	v.Check(category.Name != "", "name", "must be provided")
	v.Check(len(category.Name) <= 20, "name", "must not be more than 20 bytes long")
	v.Check(category.Description != "", "description", "must be provided")
}

func (c CategoryModel) Insert(category *Category) error {
	query := `
	INSERT INTO categories (name, description)
	VALUES ($1, $2)
	RETURNING id`

	args := []any{
		category.Name,
		category.Description,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(&category.ID)

	if err != nil {
		return err
	}

	return nil
}

func (c CategoryModel) GetAll() ([]*Category, error) {
	query := `SELECT count(*) OVER(), id, name, description FROM categories`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	categories := []*Category{}

	for rows.Next() {
		var category Category
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&category.ID,
			&category.Name,
			&category.Description,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return categories, nil
}

func (c CategoryModel) Get(id int64) (*Category, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, name, description
		FROM categories
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var category Category
	err := c.DB.QueryRow(query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &category, nil
}

func (c CategoryModel) Update(category *Category) error {
	query := `UPDATE categories
	SET name = $1, description = $2
	WHERE id = $3
	RETURNING id`

	args := []any{
		category.Name,
		category.Description,
		category.ID,
	}

	return c.DB.QueryRow(query, args...).Scan(&category.ID)
}

func (c CategoryModel) Delete(id int64) error {
	query := `
		DELETE FROM categories
		WHERE id = $1`
	result, err := c.DB.Exec(query, id)
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
