package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"
)

type Status struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type StatusModel struct {
	DB *sql.DB
}

func ValidateStatus(v *validator.Validator, status *Status) {
	v.Check(status.Name != "", "name", "must be provided")
	v.Check(len(status.Name) <= 20, "name", "must not be more than 20 bytes long")
	v.Check(status.Description != "", "description", "must be provided")
}

func (s StatusModel) Insert(status *Status) error {
	query := `
	INSERT INTO statuses (name, description)
	VALUES ($1, $2)
	RETURNING id`

	args := []any{
		status.Name,
		status.Description,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(ctx, query, args...).Scan(&status.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s StatusModel) GetAll() ([]*Status, error) {
	query := `SELECT count(*) OVER(), id, name, description FROM statuses`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	statuses := []*Status{}

	for rows.Next() {
		var status Status
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&status.ID,
			&status.Name,
			&status.Description,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		statuses = append(statuses, &status)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return statuses, nil
}

func (s StatusModel) Get(id int64) (*Status, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, name, description
		FROM statuses
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var status Status
	err := s.DB.QueryRow(query, id).Scan(
		&status.ID,
		&status.Name,
		&status.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &status, nil
}

func (s StatusModel) Update(status *Status) error {
	query := `UPDATE statuses
	SET name = $1, description = $2
	WHERE id = $3
	RETURNING id`

	args := []any{
		status.Name,
		status.Description,
		status.ID,
	}

	return s.DB.QueryRow(query, args...).Scan(&status.ID)
}

func (s StatusModel) Delete(id int64) error {
	query := `
		DELETE FROM statuses
		WHERE id = $1`
	result, err := s.DB.Exec(query, id)
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
