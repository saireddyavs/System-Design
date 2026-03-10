package models

import "time"

// Product represents a sellable item in the catalog
type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CategoryID  string    `json:"category_id"`
	Stock       int       `json:"stock"`
	SKU         string    `json:"sku"`
	Images      []string  `json:"images"`
	Rating      float64   `json:"rating"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductBuilder implements Builder pattern for Product creation
type ProductBuilder struct {
	product *Product
}

// NewProductBuilder creates a new ProductBuilder
func NewProductBuilder() *ProductBuilder {
	return &ProductBuilder{
		product: &Product{},
	}
}

// WithID sets the product ID
func (b *ProductBuilder) WithID(id string) *ProductBuilder {
	b.product.ID = id
	return b
}

// WithName sets the product name
func (b *ProductBuilder) WithName(name string) *ProductBuilder {
	b.product.Name = name
	return b
}

// WithDescription sets the product description
func (b *ProductBuilder) WithDescription(desc string) *ProductBuilder {
	b.product.Description = desc
	return b
}

// WithPrice sets the product price
func (b *ProductBuilder) WithPrice(price float64) *ProductBuilder {
	b.product.Price = price
	return b
}

// WithCategoryID sets the category
func (b *ProductBuilder) WithCategoryID(categoryID string) *ProductBuilder {
	b.product.CategoryID = categoryID
	return b
}

// WithStock sets the initial stock
func (b *ProductBuilder) WithStock(stock int) *ProductBuilder {
	b.product.Stock = stock
	return b
}

// WithSKU sets the SKU
func (b *ProductBuilder) WithSKU(sku string) *ProductBuilder {
	b.product.SKU = sku
	return b
}

// WithImages sets the product images
func (b *ProductBuilder) WithImages(images []string) *ProductBuilder {
	b.product.Images = images
	return b
}

// WithRating sets the product rating
func (b *ProductBuilder) WithRating(rating float64) *ProductBuilder {
	b.product.Rating = rating
	return b
}

// Build returns the constructed Product
func (b *ProductBuilder) Build() *Product {
	return b.product
}
