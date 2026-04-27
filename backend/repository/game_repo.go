package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/irham/topup-backend/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameRepo struct {
	db *pgxpool.Pool
}

func NewGameRepo(db *pgxpool.Pool) *GameRepo {
	return &GameRepo{db: db}
}

func (r *GameRepo) ListGames(ctx context.Context) ([]model.Game, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			g.id, g.slug, g.name, g.publisher_name,
			g.thumbnail_url, g.banner_url, g.logo_url,
			g.platform, g.genre,
			g.is_active, g.is_featured, g.sort_order,
			c.id       AS category_id,
			c.slug     AS category_slug,
			c.name     AS category_name,
			c.icon_url AS category_icon
		FROM games g
		LEFT JOIN categories c ON c.id = g.category_id
		WHERE g.is_active = true
		ORDER BY g.is_featured DESC, g.sort_order ASC, g.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("ListGames: %w", err)
	}
	defer rows.Close()

	var games []model.Game
	for rows.Next() {
		var g model.Game
		var cat model.Category
		err := rows.Scan(
			&g.ID, &g.Slug, &g.Name, &g.PublisherName,
			&g.ThumbnailURL, &g.BannerURL, &g.LogoURL,
			&g.Platform, &g.Genre,
			&g.IsActive, &g.IsFeatured, &g.SortOrder,
			&cat.ID, &cat.Slug, &cat.Name, &cat.IconURL,
		)
		if err != nil {
			return nil, fmt.Errorf("ListGames scan: %w", err)
		}
		g.Category = &cat
		games = append(games, g)
	}

	// Ambil tags untuk semua game sekaligus (1 query, bukan N query)
	if len(games) > 0 {
		if err := r.loadTagsForGames(ctx, games); err != nil {
			return nil, err
		}
	}

	return games, nil
}

// ListGamesByCategory — filter berdasarkan slug kategori
func (r *GameRepo) ListGamesByCategory(ctx context.Context, categorySlug string) ([]model.Game, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			g.id, g.slug, g.name, g.publisher_name,
			g.thumbnail_url, g.banner_url, g.logo_url,
			g.platform, g.genre,
			g.is_active, g.is_featured, g.sort_order,
			c.id, c.slug, c.name, c.icon_url
		FROM games g
		LEFT JOIN categories c ON c.id = g.category_id
		WHERE g.is_active = true
		  AND c.slug = $1
		ORDER BY g.is_featured DESC, g.sort_order ASC
	`, categorySlug)
	if err != nil {
		return nil, fmt.Errorf("ListGamesByCategory: %w", err)
	}
	defer rows.Close()

	var games []model.Game
	for rows.Next() {
		var g model.Game
		var cat model.Category
		err := rows.Scan(
			&g.ID, &g.Slug, &g.Name, &g.PublisherName,
			&g.ThumbnailURL, &g.BannerURL, &g.LogoURL,
			&g.Platform, &g.Genre,
			&g.IsActive, &g.IsFeatured, &g.SortOrder,
			&cat.ID, &cat.Slug, &cat.Name, &cat.IconURL,
		)
		if err != nil {
			return nil, fmt.Errorf("ListGamesByCategory scan: %w", err)
		}
		g.Category = &cat
		games = append(games, g)
	}
	return games, nil
}

// GetGameBySlug — detail satu game beserta topup config-nya
// Dipakai saat user buka halaman game
func (r *GameRepo) GetGameBySlug(ctx context.Context, slug string) (*model.Game, error) {
	var g model.Game
	var cat model.Category

	err := r.db.QueryRow(ctx, `
		SELECT
			g.id, g.slug, g.name, g.publisher_name,
			g.description, g.thumbnail_url, g.banner_url, g.logo_url,
			g.platform, g.genre,
			g.is_active, g.is_featured, g.sort_order, g.metadata,
			c.id, c.slug, c.name, c.icon_url
		FROM games g
		LEFT JOIN categories c ON c.id = g.category_id
		WHERE g.slug = $1 AND g.is_active = true
	`, slug).Scan(
		&g.ID, &g.Slug, &g.Name, &g.PublisherName,
		&g.Description, &g.ThumbnailURL, &g.BannerURL, &g.LogoURL,
		&g.Platform, &g.Genre,
		&g.IsActive, &g.IsFeatured, &g.SortOrder, &g.Metadata,
		&cat.ID, &cat.Slug, &cat.Name, &cat.IconURL,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // game tidak ditemukan
		}
		return nil, fmt.Errorf("GetGameBySlug: %w", err)
	}
	g.Category = &cat

	// Load tags
	tags, err := r.GetTagsByGameID(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	g.Tags = tags

	// Load topup config (field Player ID, Server ID, dll)
	config, err := r.GetTopupConfig(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	g.TopupConfig = config

	return &g, nil
}

// GetGameByID — cari game berdasarkan UUID
func (r *GameRepo) GetGameByID(ctx context.Context, id uuid.UUID) (*model.Game, error) {
	var g model.Game
	err := r.db.QueryRow(ctx, `
		SELECT id, slug, name, publisher_name, thumbnail_url, is_active
		FROM games
		WHERE id = $1
	`, id).Scan(
		&g.ID, &g.Slug, &g.Name, &g.PublisherName, &g.ThumbnailURL, &g.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetGameByID: %w", err)
	}
	return &g, nil
}

// SearchGames — fuzzy search menggunakan pg_trgm
// Dipakai untuk search bar di frontend
func (r *GameRepo) SearchGames(ctx context.Context, query string) ([]model.Game, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			g.id, g.slug, g.name, g.publisher_name,
			g.thumbnail_url, g.is_featured,
			similarity(g.name, $1) AS sim_score
		FROM games g
		WHERE g.is_active = true
		  AND (
		      g.name ILIKE '%' || $1 || '%'
		      OR g.name % $1
		  )
		ORDER BY sim_score DESC, g.is_featured DESC
		LIMIT 10
	`, query)
	if err != nil {
		return nil, fmt.Errorf("SearchGames: %w", err)
	}
	defer rows.Close()

	var games []model.Game
	for rows.Next() {
		var g model.Game
		var score float64
		err := rows.Scan(
			&g.ID, &g.Slug, &g.Name, &g.PublisherName,
			&g.ThumbnailURL, &g.IsFeatured, &score,
		)
		if err != nil {
			return nil, fmt.Errorf("SearchGames scan: %w", err)
		}
		games = append(games, g)
	}
	return games, nil
}

// GetTagsByGameID — ambil semua tag milik satu game
func (r *GameRepo) GetTagsByGameID(ctx context.Context, gameID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT tag FROM game_tags WHERE game_id = $1
	`, gameID)
	if err != nil {
		return nil, fmt.Errorf("GetTagsByGameID: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// GetTopupConfig — ambil config field per game (Player ID, Server ID, dll)
func (r *GameRepo) GetTopupConfig(ctx context.Context, gameID uuid.UUID) (*model.GameTopupConfig, error) {
	var cfg model.GameTopupConfig
	err := r.db.QueryRow(ctx, `
		SELECT game_id, fields, verify_endpoint, verify_field
		FROM game_topup_config
		WHERE game_id = $1
	`, gameID).Scan(
		&cfg.GameID, &cfg.Fields,
		&cfg.VerifyEndpoint, &cfg.VerifyField,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // game ini tidak punya topup config
		}
		return nil, fmt.Errorf("GetTopupConfig: %w", err)
	}
	return &cfg, nil
}

// loadTagsForGames — helper: ambil tag untuk banyak game sekaligus
// Hindari N+1 query — cukup 1 query untuk semua game
func (r *GameRepo) loadTagsForGames(ctx context.Context, games []model.Game) error {
	// Kumpulkan semua game ID
	ids := make([]uuid.UUID, len(games))
	for i, g := range games {
		ids[i] = g.ID
	}

	rows, err := r.db.Query(ctx, `
		SELECT game_id, tag FROM game_tags WHERE game_id = ANY($1)
	`, ids)
	if err != nil {
		return fmt.Errorf("loadTagsForGames: %w", err)
	}
	defer rows.Close()

	// Map game_id → []tags
	tagMap := make(map[uuid.UUID][]string)
	for rows.Next() {
		var gameID uuid.UUID
		var tag string
		if err := rows.Scan(&gameID, &tag); err != nil {
			return err
		}
		tagMap[gameID] = append(tagMap[gameID], tag)
	}

	// Pasang ke masing-masing game
	for i := range games {
		games[i].Tags = tagMap[games[i].ID]
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────
// PRODUCTS
// ─────────────────────────────────────────────────────────────────

// ListProductsByGame — semua produk aktif dalam satu game
// Beserta semua variantnya sekaligus
func (r *GameRepo) ListProductsByGame(ctx context.Context, gameID uuid.UUID) ([]model.Product, error) {
	// Step 1: ambil semua produk
	rows, err := r.db.Query(ctx, `
		SELECT
			id, game_id, seller_id,
			name, slug, description,
			product_type, delivery_type,
			thumbnail_url, is_active, is_featured, sort_order
		FROM products
		WHERE game_id = $1 AND is_active = true
		ORDER BY is_featured DESC, sort_order ASC, name ASC
	`, gameID)
	if err != nil {
		return nil, fmt.Errorf("ListProductsByGame: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	var productIDs []uuid.UUID

	for rows.Next() {
		var p model.Product
		err := rows.Scan(
			&p.ID, &p.GameID, &p.SellerID,
			&p.Name, &p.Slug, &p.Description,
			&p.ProductType, &p.DeliveryType,
			&p.ThumbnailURL, &p.IsActive, &p.IsFeatured, &p.SortOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("ListProductsByGame scan: %w", err)
		}
		products = append(products, p)
		productIDs = append(productIDs, p.ID)
	}

	if len(products) == 0 {
		return products, nil
	}

	// Step 2: ambil semua variant sekaligus (1 query, bukan N query)
	variantMap, err := r.loadVariantsByProductIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	// Step 3: pasang variant ke masing-masing product
	for i := range products {
		products[i].Variants = variantMap[products[i].ID]
	}

	return products, nil
}

// GetProductByID — detail satu produk beserta variantnya
func (r *GameRepo) GetProductByID(ctx context.Context, productID uuid.UUID) (*model.Product, error) {
	var p model.Product
	err := r.db.QueryRow(ctx, `
		SELECT
			id, game_id, seller_id,
			name, slug, description,
			product_type, delivery_type,
			thumbnail_url, is_active, is_featured, sort_order
		FROM products
		WHERE id = $1 AND is_active = true
	`, productID).Scan(
		&p.ID, &p.GameID, &p.SellerID,
		&p.Name, &p.Slug, &p.Description,
		&p.ProductType, &p.DeliveryType,
		&p.ThumbnailURL, &p.IsActive, &p.IsFeatured, &p.SortOrder,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetProductByID: %w", err)
	}

	variants, err := r.ListVariantsByProduct(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Variants = variants

	return &p, nil
}

// ─────────────────────────────────────────────────────────────────
// PRODUCT VARIANTS
// ─────────────────────────────────────────────────────────────────

// ListVariantsByProduct — semua variant aktif dalam satu produk
func (r *GameRepo) ListVariantsByProduct(ctx context.Context, productID uuid.UUID) ([]model.ProductVariant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, product_id, sku,
			name, amount, currency_unit,
			price, original_price, cost_price,
			stock, is_active, sort_order
		FROM product_variants
		WHERE product_id = $1 AND is_active = true
		ORDER BY sort_order ASC, amount ASC
	`, productID)
	if err != nil {
		return nil, fmt.Errorf("ListVariantsByProduct: %w", err)
	}
	defer rows.Close()

	var variants []model.ProductVariant
	for rows.Next() {
		var v model.ProductVariant
		err := rows.Scan(
			&v.ID, &v.ProductID, &v.SKU,
			&v.Name, &v.Amount, &v.CurrencyUnit,
			&v.Price, &v.OriginalPrice, &v.CostPrice,
			&v.Stock, &v.IsActive, &v.SortOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("ListVariantsByProduct scan: %w", err)
		}
		variants = append(variants, v)
	}
	return variants, nil
}

// GetVariantByID — cari satu variant berdasarkan UUID
// Dipakai saat create order — validasi variant masih aktif dan ada stoknya
func (r *GameRepo) GetVariantByID(ctx context.Context, variantID uuid.UUID) (*model.ProductVariant, error) {
	var v model.ProductVariant
	err := r.db.QueryRow(ctx, `
		SELECT
			id, product_id, sku,
			name, amount, currency_unit,
			price, original_price, cost_price,
			stock, is_active, sort_order
		FROM product_variants
		WHERE id = $1
	`, variantID).Scan(
		&v.ID, &v.ProductID, &v.SKU,
		&v.Name, &v.Amount, &v.CurrencyUnit,
		&v.Price, &v.OriginalPrice, &v.CostPrice,
		&v.Stock, &v.IsActive, &v.SortOrder,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetVariantByID: %w", err)
	}
	return &v, nil
}

// GetVariantWithGame — ambil variant beserta info game-nya sekaligus
// Dipakai saat create order — perlu tahu game_id untuk publisher API
func (r *GameRepo) GetVariantWithGame(ctx context.Context, variantID uuid.UUID) (*model.VariantWithGame, error) {
	var vg model.VariantWithGame
	err := r.db.QueryRow(ctx, `
		SELECT
			pv.id, pv.product_id, pv.name,
			pv.amount, pv.currency_unit,
			pv.price, pv.cost_price,
			pv.stock, pv.is_active,
			p.game_id,
			p.product_type, p.delivery_type,
			g.name AS game_name,
			g.slug AS game_slug,
			g.thumbnail_url AS game_thumbnail
		FROM product_variants pv
		JOIN products p ON p.id = pv.product_id
		JOIN games g    ON g.id = p.game_id
		WHERE pv.id = $1
		  AND pv.is_active = true
		  AND p.is_active  = true
		  AND g.is_active  = true
	`, variantID).Scan(
		&vg.ID, &vg.ProductID, &vg.Name,
		&vg.Amount, &vg.CurrencyUnit,
		&vg.Price, &vg.CostPrice,
		&vg.Stock, &vg.IsActive,
		&vg.GameID,
		&vg.ProductType, &vg.DeliveryType,
		&vg.GameName, &vg.GameSlug, &vg.GameThumbnail,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetVariantWithGame: %w", err)
	}
	return &vg, nil
}

// DecreaseStock — kurangi stok saat order dibuat
// Hanya berlaku jika stock != -1 (stock = -1 artinya unlimited)
func (r *GameRepo) DecreaseStock(ctx context.Context, variantID uuid.UUID, qty int) error {
	result, err := r.db.Exec(ctx, `
		UPDATE product_variants
		SET stock = stock - $2
		WHERE id = $1
		  AND stock != -1
		  AND stock >= $2
	`, variantID, qty)
	if err != nil {
		return fmt.Errorf("DecreaseStock: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("stok tidak mencukupi atau produk tidak ditemukan")
	}
	return nil
}

// loadVariantsByProductIDs — helper: ambil variant untuk banyak product sekaligus
// Menghindari N+1 query di ListProductsByGame
func (r *GameRepo) loadVariantsByProductIDs(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID][]model.ProductVariant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, product_id, sku,
			name, amount, currency_unit,
			price, original_price, cost_price,
			stock, is_active, sort_order
		FROM product_variants
		WHERE product_id = ANY($1) AND is_active = true
		ORDER BY product_id, sort_order ASC, amount ASC
	`, productIDs)
	if err != nil {
		return nil, fmt.Errorf("loadVariantsByProductIDs: %w", err)
	}
	defer rows.Close()

	variantMap := make(map[uuid.UUID][]model.ProductVariant)
	for rows.Next() {
		var v model.ProductVariant
		err := rows.Scan(
			&v.ID, &v.ProductID, &v.SKU,
			&v.Name, &v.Amount, &v.CurrencyUnit,
			&v.Price, &v.OriginalPrice, &v.CostPrice,
			&v.Stock, &v.IsActive, &v.SortOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("loadVariantsByProductIDs scan: %w", err)
		}
		variantMap[v.ProductID] = append(variantMap[v.ProductID], v)
	}
	return variantMap, nil
}

// ─────────────────────────────────────────────────────────────────
// CATEGORIES
// ─────────────────────────────────────────────────────────────────

// ListCategories — semua kategori aktif
func (r *GameRepo) ListCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, name, icon_url, sort_order
		FROM categories
		WHERE is_active = true
		ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("ListCategories: %w", err)
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.IconURL, &c.SortOrder); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}