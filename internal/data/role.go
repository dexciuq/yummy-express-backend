package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleModel struct {
	DB *sql.DB
}

func ValidateRole(v *validator.Validator, role *Role) {
	v.Check(role.Name != "", "name", "must be provided")
	v.Check(len(role.Name) <= 20, "name", "must not be more than 20 bytes long")
	v.Check(role.Description != "", "description", "must be provided")
}

func (r RoleModel) Insert(role *Role) error {
	query := `
	INSERT INTO roles (name, description)
	VALUES ($1, $2)
	RETURNING id`

	args := []any{
		role.Name,
		role.Description,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&role.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r RoleModel) GetAll() ([]*Role, error) {
	query := `SELECT count(*) OVER(), id, name, description FROM roles`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	roles := []*Role{}

	for rows.Next() {
		var role Role
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&role.ID,
			&role.Name,
			&role.Description,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		roles = append(roles, &role)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return roles, nil
}

func (r RoleModel) Get(id int64) (*Role, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, name, description
		FROM roles
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var role Role
	err := r.DB.QueryRow(query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &role, nil
}

func (r RoleModel) Update(role *Role) error {
	query := `UPDATE roles
	SET name = $1, description = $2
	WHERE id = $3
	RETURNING id`

	args := []any{
		role.Name,
		role.Description,
		role.ID,
	}

	return r.DB.QueryRow(query, args...).Scan(&role.ID)
}

func (r RoleModel) Delete(id int64) error {
	query := `
		DELETE FROM roles
		WHERE id = $1`
	result, err := r.DB.Exec(query, id)
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
