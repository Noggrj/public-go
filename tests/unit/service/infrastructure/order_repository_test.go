package infrastructure_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noggrj/autorepair/internal/service/domain"
	"github.com/noggrj/autorepair/internal/service/infrastructure"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestPostgresOrderRepository_Save(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresOrderRepository(mock)
	order, _ := domain.NewOrder(uuid.New(), uuid.New())

	// Success
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders`)).
		WithArgs(order.ID, order.ClientID, order.VehicleID, order.Status, float64(order.TotalService), float64(order.TotalParts), float64(order.Total), order.CreatedAt, order.UpdatedAt, order.StartedAt, order.FinishedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM order_items`)).
		WithArgs(order.ID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectCommit()

	err = repo.Save(order)
	assert.NoError(t, err)

	// Begin Error
	mock.ExpectBegin().WillReturnError(errors.New("tx error"))
	err = repo.Save(order)
	assert.Error(t, err)

	// Exec Error (Order)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders`)).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err = repo.Save(order)
	assert.Error(t, err)

	// Exec Error (Delete Items)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders`)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM order_items`)).
		WillReturnError(errors.New("delete error"))
	mock.ExpectRollback()

	err = repo.Save(order)
	assert.Error(t, err)
}

func TestPostgresOrderRepository_Save_WithItems(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresOrderRepository(mock)
	order, _ := domain.NewOrder(uuid.New(), uuid.New())
	_ = order.AddItem(uuid.New(), domain.ItemTypeService, "S1", 1, 100.0)

	// Success
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders`)).
		WithArgs(order.ID, order.ClientID, order.VehicleID, order.Status, float64(order.TotalService), float64(order.TotalParts), float64(order.Total), order.CreatedAt, order.UpdatedAt, order.StartedAt, order.FinishedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM order_items`)).
		WithArgs(order.ID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	item := order.Items[0]
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO order_items`)).
		WithArgs(item.ID, item.OrderID, item.RefID, item.Type, item.Name, item.Quantity, float64(item.UnitPrice), float64(item.Total)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	err = repo.Save(order)
	assert.NoError(t, err)

	// Item Error
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders`)).
		WithArgs(order.ID, order.ClientID, order.VehicleID, order.Status, float64(order.TotalService), float64(order.TotalParts), float64(order.Total), order.CreatedAt, order.UpdatedAt, order.StartedAt, order.FinishedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM order_items`)).
		WithArgs(order.ID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO order_items`)).
		WithArgs(item.ID, item.OrderID, item.RefID, item.Type, item.Name, item.Quantity, float64(item.UnitPrice), float64(item.Total)).
		WillReturnError(errors.New("item error"))
	mock.ExpectRollback()

	err = repo.Save(order)
	assert.Error(t, err)
}

func TestPostgresOrderRepository_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresOrderRepository(mock)
	id := uuid.New()
	clientID := uuid.New()
	vehicleID := uuid.New()
	now := time.Now()

	// Success
	rows := pgxmock.NewRows([]string{"id", "client_id", "vehicle_id", "status", "total_service", "total_parts", "total", "created_at", "updated_at", "started_at", "finished_at"}).
		AddRow(id, clientID, vehicleID, "Received", 100.0, 0.0, 100.0, now, now, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, client_id, vehicle_id, status, total_service, total_parts, total, created_at, updated_at, started_at, finished_at FROM orders WHERE id = $1`)).
		WithArgs(id).
		WillReturnRows(rows)

	itemRows := pgxmock.NewRows([]string{"id", "order_id", "ref_id", "type", "name", "quantity", "unit_price", "total"}).
		AddRow(uuid.New(), id, uuid.New(), "service", "S1", 1, 100.0, 100.0)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, order_id, ref_id, type, name, quantity, unit_price, total FROM order_items WHERE order_id = $1`)).
		WithArgs(id).
		WillReturnRows(itemRows)

	order, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, id, order.ID)
	assert.Len(t, order.Items, 1)

	// Not Found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(id)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrOrderNotFound, err)

	// DB Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("connection error"))

	_, err = repo.GetByID(id)
	assert.Error(t, err)

	// Items DB Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnRows(rows)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnError(errors.New("items error"))

	_, err = repo.GetByID(id)
	assert.Error(t, err)

	// Items Scan Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnRows(rows)

	itemRowsScanErr := pgxmock.NewRows([]string{"id", "order_id", "ref_id", "type", "name", "quantity", "unit_price", "total"}).
		AddRow(uuid.New(), id, uuid.New(), "service", "S1", "invalid-qty", 100.0, 100.0)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnRows(itemRowsScanErr)

	_, err = repo.GetByID(id)
	assert.Error(t, err)
}

func TestPostgresOrderRepository_List(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresOrderRepository(mock)
	now := time.Now()

	// Success
	rows := pgxmock.NewRows([]string{"id", "client_id", "vehicle_id", "status", "total_service", "total_parts", "total", "created_at", "updated_at", "started_at", "finished_at"}).
		AddRow(uuid.New(), uuid.New(), uuid.New(), "Received", 100.0, 0.0, 100.0, now, now, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, client_id, vehicle_id, status, total_service, total_parts, total, created_at, updated_at, started_at, finished_at FROM orders`)).
		WillReturnRows(rows)

	orders, err := repo.List()
	assert.NoError(t, err)
	assert.Len(t, orders, 1)

	// DB Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("db error"))

	_, err = repo.List()
	assert.Error(t, err)

	// Scan Error
	rowsScanErr := pgxmock.NewRows([]string{"id", "client_id", "vehicle_id", "status", "total_service", "total_parts", "total", "created_at", "updated_at", "started_at", "finished_at"}).
		AddRow(uuid.New(), uuid.New(), uuid.New(), "Received", "invalid-total", 0.0, 100.0, now, now, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnRows(rowsScanErr)

	_, err = repo.List()
	assert.Error(t, err)
}
