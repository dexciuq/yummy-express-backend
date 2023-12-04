package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"github.com/lib/pq"
	"time"
)

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Price       int64     `json:"price"`
	Description string    `json:"description"`
	CategoryID  int64     `json:"category_id"`
	UPC         string    `json:"upc"`
	DiscountID  int64     `json:"discount_id"`
	Quantity    int64     `json:"quantity"`
	UnitID      int64     `json:"unit_id"`
	Image       string    `json:"image"`
	BrandID     int64     `json:"brand_id"`
	CountryID   int64     `json:"country_id"`
	Step        float64   `json:"step"`
	CreatedAt   time.Time `json:"created_at"`
	Version     int       `json:"-"`
}

type productDB struct {
	ID                  int64     `json:"id"`
	Name                string    `json:"name"`
	Price               int64     `json:"price"`
	Description         string    `json:"description"`
	UPC                 string    `json:"upc"`
	Quantity            int64     `json:"quantity"`
	Image               string    `json:"image"`
	Step                float64   `json:"step"`
	CategoryID          int64     `json:"category_id"`
	CategoryName        string    `json:"category_name"`
	CategoryDescription string    `json:"category_description"`
	CategoryImage       string    `json:"category_image"`
	DiscountID          int64     `json:"discount_id"`
	DiscountName        string    `json:"discount_name"`
	DiscountDescription string    `json:"discount_description"`
	DiscountPercent     int       `json:"discount_percent"`
	DiscountCreatedAt   time.Time `json:"discount_created_at"`
	DiscountStartedAt   time.Time `json:"discount_started_at"`
	DiscountEndedAt     time.Time `json:"discount_ended_at"`
	UnitID              int64     `json:"unit_id"`
	UnitName            string    `json:"unit_name"`
	UnitDescription     string    `json:"unit_description"`
	BrandID             int64     `json:"brand_id"`
	BrandName           string    `json:"brand_name"`
	BrandDescription    string    `json:"brand_description"`
	CountryID           int64     `json:"country_id"`
	CountryName         string    `json:"country_name"`
	CountryDescription  string    `json:"country_description"`
	Alpha2              string    `json:"alpha2"`
	Alpha3              string    `json:"alpha3"`
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

func (p ProductModel) GetAll(category int, brand []int, country int, name string, filters Filters) ([]*productDB, Metadata, error) {
	// Update the SQL query to include the window function which counts the total
	// (filtered) records.
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), 
			products.id, products.name, products.price, products.description, products.upc, products.quantity, products.image, products.step,
			categories.id, categories.name, categories.description, categories.image, 
			discounts.id, discounts.name, discounts.description, discounts.discount_percent, discounts.created_at, discounts.started_at, discounts.ended_at,
			units.id, units.name, units.description,
			brands.id, brands.name, brands.description,
			countries.id, countries.name, countries.description, countries.alpha2, countries.alpha3
		FROM products
	    LEFT JOIN categories ON products.category_id = categories.id
		LEFT JOIN discounts ON products.discount_id = discounts.id
		LEFT JOIN units ON products.unit_id = units.id
		LEFT JOIN brands ON products.brand_id = brands.id
		LEFT JOIN countries ON products.country_id = countries.id
		WHERE LOWER(products.name) LIKE LOWER($1)/*(to_tsvector('simple', products.name) @@ plainto_tsquery('simple', $1) OR $1 = '')*/
		AND (products.category_id = $2 OR $2 = 0)
  		AND (brand_id = ANY($3) OR COALESCE(array_length($3, 1), 0) = 0)
		AND (products.country_id = $4 OR $4 = 0)
		ORDER BY %s %s, products.id ASC
		LIMIT $5 OFFSET $6`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []any{"%" + name + "%", category, pq.Array(brand), country, filters.limit(), filters.offset()}

	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err // Update this to return an empty Metadata struct.
	}

	defer rows.Close()

	totalRecords := 0

	var products []*productDB

	for rows.Next() {
		var product productDB
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Description,
			&product.UPC,
			&product.Quantity,
			&product.Image,
			&product.Step,
			&product.CategoryID,
			&product.CategoryName,
			&product.CategoryDescription,
			&product.CategoryImage,
			&product.DiscountID,
			&product.DiscountName,
			&product.DiscountDescription,
			&product.DiscountPercent,
			&product.DiscountCreatedAt,
			&product.DiscountStartedAt,
			&product.DiscountEndedAt,
			&product.UnitID,
			&product.UnitName,
			&product.UnitDescription,
			&product.BrandID,
			&product.BrandName,
			&product.BrandDescription,
			&product.CountryID,
			&product.CountryName,
			&product.CountryDescription,
			&product.Alpha2,
			&product.Alpha3,
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

func (p ProductModel) GetDB(id int64) (*productDB, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT products.id, products.name, products.price, products.description, products.upc, products.quantity, products.image, products.step,
			categories.id, categories.name, categories.description, categories.image, 
			discounts.id, discounts.name, discounts.description, discounts.discount_percent, discounts.created_at, discounts.started_at, discounts.ended_at,
			units.id, units.name, units.description,
			brands.id, brands.name, brands.description,
			countries.id, countries.name, countries.description, countries.alpha2, countries.alpha3
		FROM products
	    LEFT JOIN categories ON products.category_id = categories.id
		LEFT JOIN discounts ON products.discount_id = discounts.id
		LEFT JOIN units ON products.unit_id = units.id
		LEFT JOIN brands ON products.brand_id = brands.id
		LEFT JOIN countries ON products.country_id = countries.id
		WHERE products.id = $1`
	// Declare a Movie struct to hold the data returned by the query.
	var product productDB
	err := p.DB.QueryRow(query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Description,
		&product.UPC,
		&product.Quantity,
		&product.Image,
		&product.Step,
		&product.CategoryID,
		&product.CategoryName,
		&product.CategoryDescription,
		&product.CategoryImage,
		&product.DiscountID,
		&product.DiscountName,
		&product.DiscountDescription,
		&product.DiscountPercent,
		&product.DiscountCreatedAt,
		&product.DiscountStartedAt,
		&product.DiscountEndedAt,
		&product.UnitID,
		&product.UnitName,
		&product.UnitDescription,
		&product.BrandID,
		&product.BrandName,
		&product.BrandDescription,
		&product.CountryID,
		&product.CountryName,
		&product.CountryDescription,
		&product.Alpha2,
		&product.Alpha3,
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

func (p ProductModel) Init() error {
	var count int
	err := p.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		products := []*Product{
			{
				Name:        "Fresh Peach",
				Price:       65000,
				Description: "Sweet and juicy peaches, perfect for a refreshing and healthy snack. Enjoy the natural goodness of ripe peaches, known for their vibrant flavor and nutritional benefits. Add them to your fruit salads, desserts, or enjoy them on their own for a delightful taste of summer.",
				CategoryID:  2,
				UPC:         "123123",
				DiscountID:  1,
				Quantity:    10,
				UnitID:      1,
				Image:       "https://pngfre.com/wp-content/uploads/peach-png-image-from-pngfre-33-1024x815.png", // url
				BrandID:     1,
				CountryID:   1,
				Step:        1.0,
			},
			{
				Name:        "Lemon",
				Price:       23000,
				Description: "Bright and zesty lemons, known for their tangy flavor and versatility. Fresh lemons are a kitchen essential, perfect for adding a burst of citrusy goodness to both sweet and savory dishes. Whether you're making lemonade, salad dressings, desserts, or savory meals, fresh lemons bring a refreshing twist to your culinary creations.",
				CategoryID:  2,
				UPC:         "1231231",
				DiscountID:  1,
				Quantity:    8,
				UnitID:      1,
				Image:       "https://pngimg.com/d/lemon_PNG25198.png",
				BrandID:     1,
				CountryID:   1,
				Step:        1.0,
			},
			{
				Name:        "Cucumber",
				Price:       23000,
				Description: "Crunchy and hydrating cucumbers, prized for their refreshing taste and versatility. Fresh cucumbers are a low-calorie, nutrient-packed addition to your meals. Enjoy them sliced in salads, pickled for a tangy snack, or add a crisp touch to your water. With their high water content, cucumbers are perfect for staying hydrated while savoring a delightful, cool crunch.",
				CategoryID:  3,
				UPC:         "12312312",
				DiscountID:  1,
				Quantity:    12,
				UnitID:      1,
				Image:       "https://pngimg.com/d/cucumber_PNG12602.png",
				BrandID:     1,
				CountryID:   1,
				Step:        1.0,
			},
		}

		for _, product := range products {
			err := p.Insert(product)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
