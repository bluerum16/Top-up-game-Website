package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/irham/topup-backend/kafka"
	"github.com/irham/topup-backend/model"
	"github.com/irham/topup-backend/pkg"
	"github.com/irham/topup-backend/repository"
)

type OrderService struct {
	orders   *repository.OrderRepo
	payments *repository.PaymentRepo
	games    *repository.GameRepo
	users    *repository.UserRepo
	midtrans *pkg.MidtransClient
	producer *kafka.Producer
}

func NewOrderService(
	orders *repository.OrderRepo,
	payments *repository.PaymentRepo,
	games *repository.GameRepo,
	users *repository.UserRepo,
	midtrans *pkg.MidtransClient,
	producer *kafka.Producer,
) *OrderService {
	return &OrderService{
		orders: orders, payments: payments, games: games,
		users: users, midtrans: midtrans, producer: producer,
	}
}

type CreateOrderInput struct {
	UserID        uuid.UUID
	VariantID     uuid.UUID
	PlayerID      string
	ServerID      string
	PaymentMethod string
	VoucherCode   string
}

type CreateOrderResult struct {
	Order       *model.Order
	SnapToken   string
	RedirectURL string
}

var (
	ErrVariantNotFound = errors.New("variant tidak ditemukan atau tidak aktif")
	ErrInsufficientStock = errors.New("stok tidak mencukupi")
)

func (s *OrderService) Create(ctx context.Context, in CreateOrderInput) (*CreateOrderResult, error) {
	variant, err := s.games.GetVariantWithGame(ctx, in.VariantID)
	if err != nil {
		return nil, fmt.Errorf("lookup variant: %w", err)
	}
	if variant == nil {
		return nil, ErrVariantNotFound
	}
	if variant.Stock != -1 && variant.Stock < 1 {
		return nil, ErrInsufficientStock
	}

	user, err := s.users.FindByID(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	order := &model.Order{
		UserID:        in.UserID,
		ProductID:     variant.ProductID,
		VariantID:     variant.ID,
		PlayerID:      in.PlayerID,
		ServerID:      in.ServerID,
		Quantity:      1,
		UnitPrice:     variant.Price,
		Subtotal:      variant.Price,
		TotalAmount:   variant.Price,
		PaymentMethod: in.PaymentMethod,
		PaymentEnv:    s.midtrans.Env(),
		VoucherCode:   in.VoucherCode,
	}
	if err := s.orders.Create(ctx, order); err != nil {
		return nil, err
	}

	snap, err := s.midtrans.CreateSnapTransaction(pkg.SnapRequest{
		OrderID:       order.OrderNumber,
		GrossAmount:   order.TotalAmount,
		CustomerName:  fallback(user.FullName, user.Username),
		CustomerEmail: user.Email,
		CustomerPhone: user.Phone,
		ItemName:      variant.GameName + " - " + variant.Name,
	})
	if err != nil {
		_ = s.orders.UpdateStatus(ctx, order.ID, "failed", "snap creation failed")
		return nil, fmt.Errorf("midtrans snap: %w", err)
	}

	payment := &model.PaymentTransaction{
		OrderID:           order.ID,
		Env:               s.midtrans.Env(),
		MidtransOrderID:   order.OrderNumber,
		GrossAmount:       order.TotalAmount,
		TransactionStatus: "pending",
		SnapToken:         snap.Token,
		SnapRedirectURL:   snap.RedirectURL,
	}
	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("save payment: %w", err)
	}

	_ = s.orders.UpdateStatus(ctx, order.ID, "awaiting_payment", "")
	order.Status = "awaiting_payment"

	if err := s.producer.Publish(ctx, kafka.TopicOrders, order.ID.String(), map[string]any{
		"event":        "created",
		"order_id":     order.ID,
		"order_number": order.OrderNumber,
		"user_id":      order.UserID,
		"game_id":      variant.GameID,
		"variant_id":   variant.ID,
		"total_amount": order.TotalAmount,
		"created_at":   order.CreatedAt,
	}); err != nil {
		// non-fatal: log saja, jangan rollback
		fmt.Printf("kafka publish order created: %v\n", err)
	}

	return &CreateOrderResult{Order: order, SnapToken: snap.Token, RedirectURL: snap.RedirectURL}, nil
}

func fallback(s, alt string) string {
	if s == "" {
		return alt
	}
	return s
}
