package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"
)

type Unit struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UnitModel struct {
	DB *sql.DB
}

func ValidateUnit(v *validator.Validator, unit *Unit) {
	v.Check(unit.Name != "", "name", "must be provided")
	v.Check(len(unit.Name) <= 20, "name", "must not be more than 20 bytes long")
	v.Check(unit.Description != "", "description", "must be provided")
}

func (u UnitModel) Insert(unit *Unit) error {
	query := `
	INSERT INTO units (name, description)
	VALUES ($1, $2)
	RETURNING id`

	args := []any{
		unit.Name,
		unit.Description,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(&unit.ID)

	if err != nil {
		return err
	}

	return nil
}

func (u UnitModel) GetAll() ([]*Unit, error) {
	query := `SELECT count(*) OVER(), id, name, description FROM units`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := u.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	units := []*Unit{}

	for rows.Next() {
		var unit Unit
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&unit.ID,
			&unit.Name,
			&unit.Description,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		units = append(units, &unit)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return units, nil
}

func (u UnitModel) Get(id int64) (*Unit, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, name, description
		FROM units
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var unit Unit
	err := u.DB.QueryRow(query, id).Scan(
		&unit.ID,
		&unit.Name,
		&unit.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &unit, nil
}

func (u UnitModel) Update(unit *Unit) error {
	query := `UPDATE units
	SET name = $1, description = $2
	WHERE id = $3
	RETURNING id`

	args := []any{
		unit.Name,
		unit.Description,
		unit.ID,
	}

	return u.DB.QueryRow(query, args...).Scan(&unit.ID)
}

func (u UnitModel) Delete(id int64) error {
	query := `
		DELETE FROM units
		WHERE id = $1`
	result, err := u.DB.Exec(query, id)
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

func (u UnitModel) Init() error {
	var count int
	err := u.DB.QueryRow("SELECT COUNT(*) FROM units").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		units := []*Unit{
			{Name: "kg", Description: "Unit of mass, one of the seven basic units of the International System of Units (SI)."},
			{Name: "pcs", Description: "Pieces, a unit of count."},
		}

		for _, unit := range units {
			err := u.Insert(unit)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
