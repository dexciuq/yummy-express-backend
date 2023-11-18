package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Products ProductModel
	Brands   BrandModel
	Category CategoryModel
	Units    UnitModel
	Country  CountryModel
	Discount DiscountModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Products: ProductModel{DB: db},
		Brands:   BrandModel{DB: db},
		Category: CategoryModel{DB: db},
		Units:    UnitModel{DB: db},
		Country:  CountryModel{DB: db},
		Discount: DiscountModel{DB: db},
	}
}
