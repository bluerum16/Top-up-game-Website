package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/irham/topup-backend/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	db *pgxpool.Pool
}

func NewOrderRepo(db *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{db: db}
}

// Create inserts an order. order_number, status (pending), id, timestamps
// are filled by DB defaults / triggers.
func (r *OrderRepo) Create(ctx context.Context, o *model.Order) error {
	err := r.db.QueryRow(ctx, `
		INSERT INTO orders (
			user_id, product_id, variant_id,
			player_id, server_id, player_display_name,
			quantity, unit_price, subtotal, discount_amount, voucher_code,
			platform_fee, total_amount,
			payment_method, payment_env
		) VALUES (
			$1, $2, $3,
			NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''),
			$7, $8, $9, $10, NULLIF($11, ''),
			$12, $13,
			NULLIF($14, '')::payment_method, $15::payment_env
		)
		RETURNING id, order_number, status::text, created_at, updated_at
	`,
		o.UserID, o.ProductID, o.VariantID,
		o.PlayerID, o.ServerID, o.PlayerDisplayName,
		o.Quantity, o.UnitPrice, o.Subtotal, o.DiscountAmount, o.VoucherCode,
		o.PlatformFee, o.TotalAmount,
		o.PaymentMethod, o.PaymentEnv,
	).Scan(&o.ID, &o.OrderNumber, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("order create: %w", err)
	}
	return nil
}

func (r *OrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var o model.Order
	err := r.db.QueryRow(ctx, `
		SELECT id, order_number, user_id, product_id, variant_id,
		       COALESCE(player_id, ''), COALESCE(server_id, ''), COALESCE(player_display_name, ''),
		       quantity, unit_price, subtotal, discount_amount, COALESCE(voucher_code, ''),
		       platform_fee, total_amount,
		       status::text, COALESCE(payment_method::text, ''), payment_env::text,
		       paid_at, completed_at, expired_at, COALESCE(failure_reason, ''),
		       created_at, updated_at
		FROM orders WHERE id = $1
	`, id).Scan(
		&o.ID, &o.OrderNumber, &o.UserID, &o.ProductID, &o.VariantID,
		&o.PlayerID, &o.ServerID, &o.PlayerDisplayName,
		&o.Quantity, &o.UnitPrice, &o.Subtotal, &o.DiscountAmount, &o.VoucherCode,
		&o.PlatformFee, &o.TotalAmount,
		&o.Status, &o.PaymentMethod, &o.PaymentEnv,
		&o.PaidAt, &o.CompletedAt, &o.ExpiredAt, &o.FailureReason,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("order get by id: %w", err)
	}
	return &o, nil
}

// GetDetailByID joins game/product/variant for full display.
func (r *OrderRepo) GetDetailByID(ctx context.Context, id uuid.UUID) (*model.OrderDetail, error) {
	var d model.OrderDetail
	err := r.db.QueryRow(ctx, `
		SELECT o.id, o.order_number, o.user_id, o.product_id, o.variant_id,
		       COALESCE(o.player_id, ''), COALESCE(o.server_id, ''), COALESCE(o.player_display_name, ''),
		       o.quantity, o.unit_price, o.subtotal, o.discount_amount, COALESCE(o.voucher_code, ''),
		       o.platform_fee, o.total_amount,
		       o.status::text, COALESCE(o.payment_method::text, ''), o.payment_env::text,
		       o.paid_at, o.completed_at, o.expired_at, COALESCE(o.failure_reason, ''),
		       o.created_at, o.updated_at,
		       g.name, g.slug, COALESCE(g.thumbnail_url, ''),
		       p.name, pv.name, COALESCE(pv.currency_unit, ''),
		       COALESCE(pt.snap_token, ''), COALESCE(pt.payment_type, '')
		FROM orders o
		JOIN products p          ON p.id  = o.product_id
		JOIN games g             ON g.id  = p.game_id
		JOIN product_variants pv ON pv.id = o.variant_id
		LEFT JOIN payment_transactions pt ON pt.order_id = o.id
		WHERE o.id = $1
	`, id).Scan(
		&d.ID, &d.OrderNumber, &d.UserID, &d.ProductID, &d.VariantID,
		&d.PlayerID, &d.ServerID, &d.PlayerDisplayName,
		&d.Quantity, &d.UnitPrice, &d.Subtotal, &d.DiscountAmount, &d.VoucherCode,
		&d.PlatformFee, &d.TotalAmount,
		&d.Status, &d.PaymentMethod, &d.PaymentEnv,
		&d.PaidAt, &d.CompletedAt, &d.ExpiredAt, &d.FailureReason,
		&d.CreatedAt, &d.UpdatedAt,
		&d.GameName, &d.GameSlug, &d.GameThumb,
		&d.ProductName, &d.VariantName, &d.CurrencyUnit,
		&d.SnapToken, &d.PaymentType,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("order detail: %w", err)
	}
	return &d, nil
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.OrderListItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.Query(ctx, `
		SELECT o.id, o.order_number,
		       g.name, COALESCE(g.thumbnail_url, ''),
		       pv.name, o.total_amount, o.status::text, o.created_at
		FROM orders o
		JOIN products p          ON p.id  = o.product_id
		JOIN games g             ON g.id  = p.game_id
		JOIN product_variants pv ON pv.id = o.variant_id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("order list by user: %w", err)
	}
	defer rows.Close()

	var out []model.OrderListItem
	for rows.Next() {
		var it model.OrderListItem
		if err := rows.Scan(
			&it.ID, &it.OrderNumber,
			&it.GameName, &it.GameThumb,
			&it.VariantName, &it.TotalAmount, &it.Status, &it.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("order list scan: %w", err)
		}
		out = append(out, it)
	}
	return out, nil
}

// UpdateStatus transitions order to a new status. Sets timestamps on transition
// to paid / success / failed / expired. Trigger logs the change automatically.
func (r *OrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, failureReason string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE orders
		SET status = $2::order_status,
		    paid_at      = CASE WHEN $2 = 'paid'      AND paid_at      IS NULL THEN NOW() ELSE paid_at      END,
		    completed_at = CASE WHEN $2 = 'success'   AND completed_at IS NULL THEN NOW() ELSE completed_at END,
		    failed_at    = CASE WHEN $2 = 'failed'    AND failed_at    IS NULL THEN NOW() ELSE failed_at    END,
		    expired_at   = CASE WHEN $2 = 'expired'   AND expired_at   IS NULL THEN NOW() ELSE expired_at   END,
		    failure_reason = COALESCE(NULLIF($3, ''), failure_reason)
		WHERE id = $1
	`, id, status, failureReason)
	if err != nil {
		return fmt.Errorf("order update status: %w", err)
	}
	return nil
}
