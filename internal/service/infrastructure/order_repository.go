package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noggrj/autorepair/internal/platform/db"
	"github.com/noggrj/autorepair/internal/service/domain"
	"github.com/noggrj/autorepair/internal/sharedkernel"
)

type PostgresOrderRepository struct {
	db db.Connection
}

func NewPostgresOrderRepository(db db.Connection) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Save(order *domain.Order) error {
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			// Log rollback error but don't fail the function
			_ = err // ignore error
		}
	}()

	// Save Order
	query := `INSERT INTO orders (id, client_id, vehicle_id, status, total_service, total_parts, total, created_at, updated_at, started_at, finished_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	          ON CONFLICT (id) DO UPDATE SET
	          status = EXCLUDED.status,
	          total_service = EXCLUDED.total_service,
	          total_parts = EXCLUDED.total_parts,
	          total = EXCLUDED.total,
	          updated_at = EXCLUDED.updated_at,
	          started_at = EXCLUDED.started_at,
	          finished_at = EXCLUDED.finished_at`

	_, err = tx.Exec(ctx, query,
		order.ID, order.ClientID, order.VehicleID, order.Status,
		float64(order.TotalService), float64(order.TotalParts), float64(order.Total),
		order.CreatedAt, order.UpdatedAt, order.StartedAt, order.FinishedAt)
	if err != nil {
		return err
	}

	// Save Items (Naive approach: delete all and recreate for simplicity in MVP update)
	// For Create, insert is enough. For Update, delete+insert is easiest for MVP.
	_, err = tx.Exec(ctx, "DELETE FROM order_items WHERE order_id = $1", order.ID)
	if err != nil {
		return err
	}

	itemQuery := `INSERT INTO order_items (id, order_id, ref_id, type, name, quantity, unit_price, total)
	              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, itemQuery,
			item.ID, item.OrderID, item.RefID, item.Type, item.Name,
			item.Quantity, float64(item.UnitPrice), float64(item.Total))
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) GetByID(id uuid.UUID) (*domain.Order, error) {
	query := `SELECT id, client_id, vehicle_id, status, total_service, total_parts, total, created_at, updated_at, started_at, finished_at 
	          FROM orders WHERE id = $1`

	row := r.db.QueryRow(context.Background(), query, id)

	var o domain.Order
	var ts, tp, t float64
	var statusStr string

	err := row.Scan(&o.ID, &o.ClientID, &o.VehicleID, &statusStr, &ts, &tp, &t, &o.CreatedAt, &o.UpdatedAt, &o.StartedAt, &o.FinishedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	o.Status = domain.OrderStatus(statusStr)
	o.TotalService = sharedkernel.Money(ts)
	o.TotalParts = sharedkernel.Money(tp)
	o.Total = sharedkernel.Money(t)

	// Fetch Items
	itemsQuery := `SELECT id, order_id, ref_id, type, name, quantity, unit_price, total FROM order_items WHERE order_id = $1`
	rows, err := r.db.Query(context.Background(), itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var i domain.OrderItem
		var typeStr string
		var up, tot float64
		if err := rows.Scan(&i.ID, &i.OrderID, &i.RefID, &typeStr, &i.Name, &i.Quantity, &up, &tot); err != nil {
			return nil, err
		}
		i.Type = domain.OrderItemType(typeStr)
		i.UnitPrice = sharedkernel.Money(up)
		i.Total = sharedkernel.Money(tot)
		o.Items = append(o.Items, &i)
	}

	return &o, nil
}

func (r *PostgresOrderRepository) List() ([]*domain.Order, error) {
	// Basic listing without pagination for MVP
	query := `SELECT id, client_id, vehicle_id, status, total_service, total_parts, total, created_at, updated_at, started_at, finished_at FROM orders`
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		var ts, tp, t float64
		var statusStr string
		err := rows.Scan(&o.ID, &o.ClientID, &o.VehicleID, &statusStr, &ts, &tp, &t, &o.CreatedAt, &o.UpdatedAt, &o.StartedAt, &o.FinishedAt)
		if err != nil {
			return nil, err
		}
		o.Status = domain.OrderStatus(statusStr)
		o.TotalService = sharedkernel.Money(ts)
		o.TotalParts = sharedkernel.Money(tp)
		o.Total = sharedkernel.Money(t)
		orders = append(orders, &o)
	}
	return orders, nil
}

func (r *PostgresOrderRepository) ListActive() ([]*domain.Order, error) {
	query := `SELECT id, client_id, vehicle_id, status, total_service, total_parts, total, created_at, updated_at, started_at, finished_at
	          FROM orders
	          WHERE status NOT IN ($1, $2)
	          ORDER BY
	            CASE status
	              WHEN $3 THEN 1
	              WHEN $4 THEN 2
	              WHEN $5 THEN 3
	              WHEN $6 THEN 4
	              ELSE 5
	            END,
	            created_at ASC`

	rows, err := r.db.Query(context.Background(), query,
		string(domain.OrderStatusCompleted),
		string(domain.OrderStatusDelivered),
		string(domain.OrderStatusInExecution),
		string(domain.OrderStatusAwaitingApproval),
		string(domain.OrderStatusInDiagnosis),
		string(domain.OrderStatusReceived),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		var ts, tp, t float64
		var statusStr string
		err := rows.Scan(&o.ID, &o.ClientID, &o.VehicleID, &statusStr, &ts, &tp, &t, &o.CreatedAt, &o.UpdatedAt, &o.StartedAt, &o.FinishedAt)
		if err != nil {
			return nil, err
		}
		o.Status = domain.OrderStatus(statusStr)
		o.TotalService = sharedkernel.Money(ts)
		o.TotalParts = sharedkernel.Money(tp)
		o.Total = sharedkernel.Money(t)
		orders = append(orders, &o)
	}
	return orders, nil
}
