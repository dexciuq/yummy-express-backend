package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"
)

type Country struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Alpha2      string `json:"alpha2"`
	Alpha3      string `json:"alpha3"`
}

type CountryModel struct {
	DB *sql.DB
}

func ValidateCountry(v *validator.Validator, country *Country) {
	v.Check(country.Name != "", "name", "must be provided")
	v.Check(len(country.Name) <= 20, "name", "must not be more than 20 bytes long")
	v.Check(country.Description != "", "description", "must be provided")
	v.Check(len(country.Alpha2) == 2, "alpha2", "must be 2 bytes long")
	v.Check(len(country.Alpha3) == 3, "alpha3", "must be 3 bytes long")
}

func (c CountryModel) Insert(country *Country) error {
	query := `
	INSERT INTO countries (name, description, alpha2, alpha3)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	args := []any{
		country.Name,
		country.Description,
		country.Alpha2,
		country.Alpha3,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(&country.ID)

	if err != nil {
		return err
	}

	return nil
}

func (c CountryModel) GetAll() ([]*Country, error) {
	query := `SELECT count(*) OVER(), id, name, description, alpha2, alpha3 FROM countries`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	countries := []*Country{}

	for rows.Next() {
		var country Country
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&country.ID,
			&country.Name,
			&country.Description,
			&country.Alpha2,
			&country.Alpha3,
		)
		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}
		countries = append(countries, &country)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}
	return countries, nil
}

func (c CountryModel) Get(id int64) (*Country, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, name, description, alpha2, alpha3
		FROM countries
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var country Country
	err := c.DB.QueryRow(query, id).Scan(
		&country.ID,
		&country.Name,
		&country.Description,
		&country.Alpha2,
		&country.Alpha3,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &country, nil
}

func (c CountryModel) Update(country *Country) error {
	query := `UPDATE countries
	SET name = $1, description = $2, alpha2 = $3, alpha3 = $4
	WHERE id = $5
	RETURNING id`

	args := []any{
		country.Name,
		country.Description,
		country.Alpha2,
		country.Alpha3,
		country.ID,
	}

	return c.DB.QueryRow(query, args...).Scan(&country.ID)
}

func (c CountryModel) Delete(id int64) error {
	query := `
		DELETE FROM countries
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

func (c CountryModel) Init() error {
	var count int
	err := c.DB.QueryRow("SELECT COUNT(*) FROM countries").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		countries := []*Country{
			{
				Name:        "Kazakhstan",
				Description: "Country in Central Asia",
				Alpha2:      "KZ",
				Alpha3:      "KAZ",
			},
			{
				Name:        "Russia",
				Description: "Largest country in the world",
				Alpha2:      "RU",
				Alpha3:      "RUS",
			},
			{
				Name:        "United States",
				Description: "Powerful nation in North America",
				Alpha2:      "US",
				Alpha3:      "USA",
			},
			{
				Name:        "China",
				Description: "World's most populous country",
				Alpha2:      "CN",
				Alpha3:      "CHN",
			},
		}

		for _, country := range countries {
			err := c.Insert(country)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
