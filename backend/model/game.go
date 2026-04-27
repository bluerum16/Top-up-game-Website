package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Game struct {
	ID            uuid.UUID       `db:"id"             json:"id"`
	Slug          string          `db:"slug"           json:"slug"`
	Name          string          `db:"name"           json:"name"`
	PublisherName string          `db:"publisher_name" json:"publisher_name"`
	Description   string          `db:"description"    json:"description,omitempty"`
	ThumbnailURL  string          `db:"thumbnail_url"  json:"thumbnail_url"`
	BannerURL     string          `db:"banner_url"     json:"banner_url,omitempty"`
	LogoURL       string          `db:"logo_url"       json:"logo_url,omitempty"`
	CategoryID    int             `db:"category_id"    json:"category_id,omitempty"`
	Platform      []string        `db:"platform"       json:"platform,omitempty"`
	Genre         []string        `db:"genre"          json:"genre,omitempty"`
	IsActive      bool            `db:"is_active"      json:"is_active"`
	IsFeatured    bool            `db:"is_featured"    json:"is_featured"`
	SortOrder     int             `db:"sort_order"     json:"sort_order"`
	Metadata      json.RawMessage `db:"metadata"       json:"metadata,omitempty"`
	CreatedAt     time.Time       `db:"created_at"     json:"created_at,omitempty"`

	Category    *Category        `db:"-" json:"category,omitempty"`
	Tags        []string         `db:"-" json:"tags,omitempty"`
	Products    []Product        `db:"-" json:"products,omitempty"`
	TopupConfig *GameTopupConfig `db:"-" json:"topup_config,omitempty"`
}

type Category struct {
	ID        int    `db:"id"         json:"id"`
	Slug      string `db:"slug"       json:"slug"`
	Name      string `db:"name"       json:"name"`
	IconURL   string `db:"icon_url"   json:"icon_url,omitempty"`
	SortOrder int    `db:"sort_order" json:"sort_order"`
}

type Product struct {
	ID           uuid.UUID  `db:"id"            json:"id"`
	GameID       uuid.UUID  `db:"game_id"       json:"game_id"`
	SellerID     *uuid.UUID `db:"seller_id"     json:"seller_id,omitempty"`
	Name         string     `db:"name"          json:"name"`
	Slug         string     `db:"slug"          json:"slug,omitempty"`
	Description  string     `db:"description"   json:"description,omitempty"`
	ProductType  string     `db:"product_type"  json:"product_type"`
	DeliveryType string     `db:"delivery_type" json:"delivery_type"`
	ThumbnailURL string     `db:"thumbnail_url" json:"thumbnail_url,omitempty"`
	IsActive     bool       `db:"is_active"     json:"is_active"`
	IsFeatured   bool       `db:"is_featured"   json:"is_featured"`
	SortOrder    int        `db:"sort_order"    json:"sort_order"`

	Variants []ProductVariant `db:"-" json:"variants,omitempty"`
}

type ProductVariant struct {
	ID            uuid.UUID `db:"id"             json:"id"`
	ProductID     uuid.UUID `db:"product_id"     json:"product_id"`
	SKU           string    `db:"sku"            json:"sku"`
	Name          string    `db:"name"           json:"name"`
	Amount        int       `db:"amount"         json:"amount"`
	CurrencyUnit  string    `db:"currency_unit"  json:"currency_unit"`
	Price         int64     `db:"price"          json:"price"`
	OriginalPrice int64     `db:"original_price" json:"original_price"`
	CostPrice     int64     `db:"cost_price"     json:"-"`
	Stock         int       `db:"stock"          json:"stock"`
	IsActive      bool      `db:"is_active"      json:"is_active"`
	SortOrder     int       `db:"sort_order"     json:"sort_order"`
}

// VariantWithGame — variant + minimal game info, dipakai saat create order
type VariantWithGame struct {
	ID            uuid.UUID `json:"id"`
	ProductID     uuid.UUID `json:"product_id"`
	Name          string    `json:"name"`
	Amount        int       `json:"amount"`
	CurrencyUnit  string    `json:"currency_unit"`
	Price         int64     `json:"price"`
	CostPrice     int64     `json:"-"`
	Stock         int       `json:"stock"`
	IsActive      bool      `json:"is_active"`
	GameID        uuid.UUID `json:"game_id"`
	GameName      string    `json:"game_name"`
	GameSlug      string    `json:"game_slug"`
	GameThumbnail string    `json:"game_thumbnail"`
	ProductType   string    `json:"product_type"`
	DeliveryType  string    `json:"delivery_type"`
}

type GameTopupConfig struct {
	GameID         uuid.UUID       `db:"game_id"         json:"game_id"`
	Fields         json.RawMessage `db:"fields"          json:"fields"`
	VerifyEndpoint string          `db:"verify_endpoint" json:"-"`
	VerifyField    string          `db:"verify_field"    json:"-"`
}

type TopupField struct {
	Key      string        `json:"key"`
	Label    string        `json:"label"`
	Type     string        `json:"type"`
	Required bool          `json:"required"`
	Options  []FieldOption `json:"options,omitempty"`
}

type FieldOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
