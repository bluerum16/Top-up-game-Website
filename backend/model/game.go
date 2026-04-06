package model

import (
	"time"

	"github.com/google/uuid"
)

type Game struct {
	ID uuid.UUID `db:"id"	json:"id"`
	Slug string `db:"slug" json:"slug"`
	Name string `db:"name" json:"name"`
	PublisherName string     `db:"publisher_name" json:"publisher_name"`
    Description   string     `db:"description"    json:"description"`
    ThumbnailURL  string     `db:"thumbnail_url"  json:"thumbnail_url"`
    BannerURL     string     `db:"banner_url"     json:"banner_url"`
    CategoryID    int        `db:"category_id"    json:"category_id"`
    Platform      []string   `db:"platform"       json:"platform"`
    Genre         []string   `db:"genre"          json:"genre"`
    IsActive      bool       `db:"is_active"      json:"is_active"`
    IsFeatured    bool       `db:"is_featured"    json:"is_featured"`
    CreatedAt     time.Time  `db:"created_at"     json:"created_at"`

	Category   *Category        `db:"-" json:"category,omitempty"`
    Tags       []string         `db:"-" json:"tags,omitempty"`
    Products   []Product        `db:"-" json:"products,omitempty"`
    TopupConfig *GameTopupConfig `db:"-" json:"topup_config,omitempty"`
}

type Category struct {
    ID       int    `db:"id"        json:"id"`
    Slug     string `db:"slug"      json:"slug"`
    Name     string `db:"name"      json:"name"`
    IconURL  string `db:"icon_url"  json:"icon_url"`
}

type Product struct {
    ID           uuid.UUID `db:"id"            json:"id"`
    GameID       uuid.UUID `db:"game_id"       json:"game_id"`
    Name         string    `db:"name"          json:"name"`
    ProductType  string    `db:"product_type"  json:"product_type"`
    DeliveryType string    `db:"delivery_type" json:"delivery_type"`
    IsActive     bool      `db:"is_active"     json:"is_active"`

    Variants []ProductVariant `db:"-" json:"variants,omitempty"`
}

type ProductVariant struct {
    ID           uuid.UUID `db:"id"             json:"id"`
    ProductID    uuid.UUID `db:"product_id"     json:"product_id"`
    SKU          string    `db:"sku"            json:"sku"`
    Name         string    `db:"name"           json:"name"`
    Amount       int       `db:"amount"         json:"amount"`
    CurrencyUnit string    `db:"currency_unit"  json:"currency_unit"`
    Price        int64     `db:"price"          json:"price"`
    OriginalPrice int64    `db:"original_price" json:"original_price"`

    CostPrice    int64     `db:"cost_price"     json:"-"`
    Stock        int       `db:"stock"          json:"stock"`
    IsActive     bool      `db:"is_active"      json:"is_active"`
    SortOrder    int       `db:"sort_order"     json:"sort_order"`
}

type GameTopupConfig struct {
    GameID         uuid.UUID       `db:"game_id"         json:"game_id"`
    Fields         []TopupField    `db:"fields"          json:"fields"`
    VerifyEndpoint string          `db:"verify_endpoint" json:"-"` // internal only
}

type TopupField struct {
    Key      string        `json:"key"`
    Label    string        `json:"label"`
    Type     string        `json:"type"`     // text, select, number
    Required bool          `json:"required"`
    Options  []FieldOption `json:"options,omitempty"`
}

type FieldOption struct {
    Label string `json:"label"`
    Value string `json:"value"`
}
