package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dexciuq/yummy-express-backend/internal/validator"
)

type Category struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
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
	INSERT INTO categories (name, description, image)
	VALUES ($1, $2, $3)
	RETURNING id`

	args := []any{
		category.Name,
		category.Description,
		category.Image,
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
	query := `SELECT count(*) OVER(), id, name, description, image FROM categories`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	totalRecords := 0

	categories := []*Category{}

	for rows.Next() {
		var category Category
		err := rows.Scan(
			&totalRecords,
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Image,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func (c CategoryModel) Get(id int64) (*Category, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, description, image
		FROM categories
		WHERE id = $1`

	var category Category
	err := c.DB.QueryRow(query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Image,
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
	SET name = $1, description = $2, image = $3
	WHERE id = $4
	RETURNING id`

	args := []any{
		category.Name,
		category.Description,
		category.Image,
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

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (c CategoryModel) Init() error {
	var count int
	err := c.DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		categories := []*Category{
			{
				Name:        "Discount",
				Description: "Product list that has discounts.",
				Image:       "https://png.pngtree.com/png-vector/20230408/ourmid/pngtree-price-tag-with-the-discount-icon-vector-png-image_6686659.png",
			},
			{
				Name:        "Fruits",
				Description: "Various fresh fruits.",
				Image:       "https://www.freepnglogos.com/uploads/fruits-png/fruits-png-image-pngpix-40.png",
			},
			{
				Name:        "Vegetables",
				Description: "A variety of fresh vegetables.",
				Image:       "https://freepngimg.com/thumb/vegetable/3-2-vegetable-transparent-thumb.png",
			},
			{
				Name:        "Dairy",
				Description: "Milk, cheese, and other dairy products.",
				Image:       "https://png.monster/wp-content/uploads/2022/06/png.monster-790.png",
			},
			{
				Name:        "Meat",
				Description: "Different types of meat products.",
				Image:       "https://pngimg.com/d/pork_PNG50.png",
			},
			{
				Name:        "Seafood",
				Description: "Fresh seafood items.",
				Image:       "https://pngimg.com/uploads/fish/fish_PNG25091.png",
			},
			{
				Name:        "Bakery",
				Description: "Bread, pastries, and baked goods.",
				Image:       "https://shopepicure.ca/cdn/shop/products/image_7c1f2ad1-b5be-4f36-abcf-5b1dc8c2425f_500x500.png?v=1613673106",
			},
			{
				Name:        "Cereal",
				Description: "Various breakfast cereals.",
				Image:       "https://static.vecteezy.com/system/resources/previews/024/851/122/original/cereal-dry-breakfast-in-a-plate-transparent-background-png.png",
			},
			{
				Name:        "Snacks",
				Description: "Assorted snacks and finger foods.",
				Image:       "https://cpjmarket.com/cdn/shop/products/2018571_500x.png?v=1663091578",
			},
			{
				Name:        "Beverages",
				Description: "Non-alcoholic drinks.",
				Image:       "https://purepng.com/public/uploads/large/drinks-igr.png",
			},
			{
				Name:        "Sweets",
				Description: "Candies, chocolates, and desserts.",
				Image:       "https://freepngimg.com/thumb/sweets/4-2-sweets-transparent.png",
			},
			{
				Name:        "Condiments",
				Description: "Sauces, dressings, and condiments.",
				Image:       "https://pngimg.com/d/sauce_PNG71.png",
			},
			{
				Name:        "Frozen Foods",
				Description: "Various frozen food items.",
				Image:       "https://www.foxpak.com/wp-content/uploads/2019/02/Header-frozeon-food.png",
			},
			{
				Name:        "Spices and Herbs",
				Description: "Various spices and herbs.",
				Image:       "https://freepngimg.com/thumb/herbs/27287-5-herbs.png",
			},
		}

		for _, category := range categories {
			err := c.Insert(category)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
