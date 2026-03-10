package strategies

import "online-bookstore/internal/models"

// BookQueryBuilder builds complex book queries (Builder pattern - optional).
// Enables fluent API for constructing search criteria.
type BookQueryBuilder struct {
	query models.BookQuery
}

func NewBookQueryBuilder() *BookQueryBuilder {
	return &BookQueryBuilder{
		query: models.BookQuery{},
	}
}

func (b *BookQueryBuilder) WithTitle(title string) *BookQueryBuilder {
	b.query.Title = title
	return b
}

func (b *BookQueryBuilder) WithAuthor(author string) *BookQueryBuilder {
	b.query.Author = author
	return b
}

func (b *BookQueryBuilder) WithGenre(genre string) *BookQueryBuilder {
	b.query.Genre = genre
	return b
}

func (b *BookQueryBuilder) WithPriceRange(min, max float64) *BookQueryBuilder {
	b.query.MinPrice = min
	b.query.MaxPrice = max
	return b
}

func (b *BookQueryBuilder) InStockOnly() *BookQueryBuilder {
	b.query.InStock = true
	return b
}

func (b *BookQueryBuilder) Build() models.BookQuery {
	return b.query
}
