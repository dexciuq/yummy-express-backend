package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"
)

type Brand struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BrandModel struct {
	DB *sql.DB
}

func ValidateBrand(v *validator.Validator, brand *Brand) {
	v.Check(brand.Name != "", "name", "must be provided")
	v.Check(len(brand.Name) <= 20, "name", "must not be more than 20 bytes long")
	v.Check(brand.Description != "", "description", "must be provided")
}

func (b BrandModel) Insert(brand *Brand) error {
	query := `
	INSERT INTO brands (name, description)
	VALUES ($1, $2)
	RETURNING id`

	args := []any{
		brand.Name,
		brand.Description,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, args...).Scan(&brand.ID)

	if err != nil {
		return err
	}

	return nil
}

func (b BrandModel) GetAll() ([]*Brand, error) {
	query := `SELECT count(*) OVER(), id, name, description FROM brands`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := b.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	brands := []*Brand{}

	for rows.Next() {
		var brand Brand
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&brand.ID,
			&brand.Name,
			&brand.Description,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		brands = append(brands, &brand)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return brands, nil
}

func (b BrandModel) Get(id int64) (*Brand, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, name, description
		FROM brands
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var brand Brand
	err := b.DB.QueryRow(query, id).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &brand, nil
}

func (b BrandModel) Update(brand *Brand) error {
	query := `UPDATE brands
	SET name = $1, description = $2
	WHERE id = $3
	RETURNING id`

	args := []any{
		brand.Name,
		brand.Description,
		brand.ID,
	}

	return b.DB.QueryRow(query, args...).Scan(&brand.ID)
}

func (b BrandModel) Delete(id int64) error {
	query := `
		DELETE FROM brands
		WHERE id = $1`
	result, err := b.DB.Exec(query, id)
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

func (b BrandModel) Init() error {
	var count int
	err := b.DB.QueryRow("SELECT COUNT(*) FROM brands").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		brands := []*Brand{
			{
				Name:        "ASI Mart",
				Description: "The closest market",
			},
			{
				Name:        "Coca-Cola",
				Description: "Beverage company known for its soft drinks.",
			},
			{
				Name:        "Nestlé",
				Description: "Multinational food and beverage company.",
			},
			{
				Name:        "PepsiCo",
				Description: "Global food and beverage company.",
			},
			{
				Name:        "Kellogg's",
				Description: "Producer of cereal and convenience foods.",
			},
			{
				Name:        "Unilever",
				Description: "British-Dutch multinational consumer goods company.",
			},
		}

		for _, brand := range brands {
			err := b.Insert(brand)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
