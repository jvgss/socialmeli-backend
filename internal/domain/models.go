package domain

import "time"

type User struct {
	ID       int    `json:"user_id"`
	Name     string `json:"user_name"`
	IsSeller bool   `json:"is_seller"`
}

type Product struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	Type        string `json:"type"`
	Brand       string `json:"brand"`
	Color       string `json:"color"`
	Notes       string `json:"notes"`

	ImageURL string `json:"image_url,omitempty"`
}

type Post struct {
	PostID   int       `json:"post_id"`
	UserID   int       `json:"user_id"`
	Date     time.Time `json:"-"`
	DateStr  string    `json:"date"`
	Product  Product   `json:"product"`
	Category int       `json:"category"`
	Price    float64   `json:"price"`

	HasPromo bool    `json:"has_promo"`
	Discount float64 `json:"discount"`

	FinalPrice float64 `json:"final_price"`
}
