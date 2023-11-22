package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"
)

type Discount struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DiscountPercent int       `json:"discount_percent"`
	CreatedAt       time.Time `json:"created_at"`
	StartedAt       time.Time `json:"started_at"`
	EndedAt         time.Time `json:"ended_at"`
}

type DiscountModel struct {
	DB *sql.DB
}

func ValidateDiscount(v *validator.Validator, discount *Discount) {
	v.Check(discount.Name != "", "name", "must be provided")
	v.Check(len(discount.Name) <= 20, "name", "must not be more than 20 bytes long")
	v.Check(discount.Description != "", "description", "must be provided")
	v.Check(discount.StartedAt.Before(discount.EndedAt), "ended_at", "must be later than started_at")
}

func (d DiscountModel) Insert(discount *Discount) error {
	query := `
	INSERT INTO discounts (name, description, discount_percent, started_at, ended_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at`

	args := []any{
		discount.Name,
		discount.Description,
		discount.DiscountPercent,
		discount.StartedAt,
		discount.EndedAt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := d.DB.QueryRowContext(ctx, query, args...).Scan(&discount.ID, &discount.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (d DiscountModel) GetAll() ([]*Discount, error) {
	// Update the SQL query to include the window function which counts the total
	// (filtered) records.
	query := `SELECT id, name, description, discount_percent, created_at, started_at, ended_at
		FROM discounts`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	discounts := []*Discount{}

	for rows.Next() {
		var discount Discount
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&discount.ID,
			&discount.Name,
			&discount.Description,
			&discount.DiscountPercent,
			&discount.CreatedAt,
			&discount.StartedAt,
			&discount.EndedAt,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		discounts = append(discounts, &discount)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return discounts, nil
}

func (d DiscountModel) Get(id int64) (*Discount, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, name, description, discount_percent, created_at, started_at, ended_at
		FROM discounts
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var discount Discount
	err := d.DB.QueryRow(query, id).Scan(
		&discount.ID,
		&discount.Name,
		&discount.Description,
		&discount.DiscountPercent,
		&discount.CreatedAt,
		&discount.StartedAt,
		&discount.EndedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &discount, nil
}

func (d DiscountModel) Update(discount *Discount) error {
	query := `UPDATE discounts
	SET name = $1, description = $2, discount_percent = $3, created_at = $4, started_at = $5, ended_at = $6
	WHERE id = $7
	RETURNING id`

	args := []any{
		discount.Name,
		discount.Description,
		discount.DiscountPercent,
		discount.CreatedAt,
		discount.StartedAt,
		discount.EndedAt,
		discount.ID,
	}

	return d.DB.QueryRow(query, args...).Scan(&discount.ID)
}

func (d DiscountModel) Delete(id int64) error {
	query := `
		DELETE FROM discounts
		WHERE id = $1`
	result, err := d.DB.Exec(query, id)
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

func (d DiscountModel) Init() error {
	var count int
	err := d.DB.QueryRow("SELECT COUNT(*) FROM discounts").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		discounts := []*Discount{
			{
				Name:            "0% discount",
				Description:     "0",
				DiscountPercent: 0,
				StartedAt:       time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndedAt:         time.Date(2050, time.December, 31, 23, 59, 59, 999999999, time.UTC),
			},
			{
				Name:            "15% discount",
				Description:     "Special discount for the holiday season",
				DiscountPercent: 15,
				StartedAt:       time.Date(2023, time.November, 1, 0, 0, 0, 0, time.UTC),
				EndedAt:         time.Date(2023, time.December, 31, 23, 59, 59, 999999999, time.UTC),
			},
			{
				Name:            "10% discount",
				Description:     "Back-to-school promotion",
				DiscountPercent: 10,
				StartedAt:       time.Date(2023, time.November, 1, 0, 0, 0, 0, time.UTC),
				EndedAt:         time.Date(2023, time.December, 31, 23, 59, 59, 999999999, time.UTC),
			},
			{
				Name:            "5% discount",
				Description:     "End-of-year clearance sale",
				DiscountPercent: 5,
				StartedAt:       time.Date(2023, time.November, 1, 0, 0, 0, 0, time.UTC),
				EndedAt:         time.Date(2023, time.December, 31, 23, 59, 59, 999999999, time.UTC),
			},
		}

		for _, discount := range discounts {
			err := d.Insert(discount)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
