package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"
)

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Price       int64     `json:"price"`
	Description string    `json:"description"`
	CategoryID  int64     `json:"category_id"`
	UPC         string    `json:"upc"`
	DiscountID  string    `json:"discount_id"`
	Quantity    int64     `json:"quantity"`
	UnitID      int64     `json:"unit_id"`
	Image       string    `json:"image"`
	BrandID     int64     `json:"brand_id"`
	CountryID   int64     `json:"country_id"`
	Step        float64   `json:"step"`
	CreatedAt   time.Time `json:"created_at"`
	Version     int       `json:"-"`
}

type ProductModel struct {
	DB *sql.DB
}

func ValidateProduct(v *validator.Validator, product *Product) {
	v.Check(product.Name != "", "name", "must be provided")
	v.Check(len(product.Name) <= 20, "name", "must not be more than 20 bytes long")
	v.Check(product.Price >= 0, "price", "can not be negative")
	v.Check(product.Description != "", "description", "must be provided")
	v.Check(product.Quantity >= 0, "quantity", "can not be negative")
}

func (p ProductModel) Insert(product *Product) error {
	query := `
	INSERT INTO products (name, price, description, category_id, upc, discount_id, quantity, unit_id, image, brand_id, country_id, step)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING id, created_at`

	args := []any{
		product.Name,
		product.Price,
		product.Description,
		product.CategoryID,
		product.UPC,
		product.DiscountID,
		product.Quantity,
		product.UnitID,
		product.Image,
		product.BrandID,
		product.CountryID,
		product.Step,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, args...).Scan(&product.ID,
		&product.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (p ProductModel) GetAll(category int, brand int, country int, filters Filters) ([]*Product, Metadata, error) {
	// Update the SQL query to include the window function which counts the total
	// (filtered) records.
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, name, price, description, category_id, upc, discount_id, quantity, unit_id, image, brand_id, country_id, step
		FROM products
		WHERE (category_id = $1 OR $1 = 0)
		AND (brand_id = $2 OR $2 = 0)
		AND (country_id = $3 OR $3 = 0)
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []any{category, brand, country, filters.limit(), filters.offset()}

	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	products := []*Product{}

	for rows.Next() {
		var product Product
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Description,
			&product.CategoryID,
			&product.UPC,
			&product.DiscountID,
			&product.Quantity,
			&product.UnitID,
			&product.Image,
			&product.BrandID,
			&product.CountryID,
			&product.Step,
		)
		if err != nil {
			return nil, Metadata{}, err // Update this to return an empty Metadata struct.
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err // Update this to return an empty Metadata struct.
	}
	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Include the metadata struct when returning.
	return products, metadata, nil
}

func (p ProductModel) Get(id int64) (*Product, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, name, price, description, category_id, upc, discount_id, quantity, unit_id, image, brand_id, country_id, step
		FROM products
		WHERE id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var product Product
	err := p.DB.QueryRow(query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Description,
		&product.CategoryID,
		&product.UPC,
		&product.DiscountID,
		&product.Quantity,
		&product.UnitID,
		&product.Image,
		&product.BrandID,
		&product.CountryID,
		&product.Step,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &product, nil
}

func (p ProductModel) Update(product *Product) error {
	query := `UPDATE products
	SET name = $1, price = $2, description = $3, category_id = $4, upc = $5, discount_id = $6, quantity = $7, unit_id = $8, image = $9, brand_id = $10, country_id = $11, step = $12
	WHERE id = $13
	RETURNING id`

	args := []any{
		product.Name,
		product.Price,
		product.Description,
		product.CategoryID,
		product.UPC,
		product.DiscountID,
		product.Quantity,
		product.UnitID,
		product.Image,
		product.BrandID,
		product.CountryID,
		product.Step,
		product.ID,
	}

	return p.DB.QueryRow(query, args...).Scan(&product.ID)
}

func (p ProductModel) Delete(id int64) error {
	query := `
		DELETE FROM products
		WHERE id = $1`
	result, err := p.DB.Exec(query, id)
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
