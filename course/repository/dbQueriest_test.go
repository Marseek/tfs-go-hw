package repository

import (
	"context"
	"testing"

	"github.com/Marseek/tfs-go-hw/course/domain"
	pkgpostgres "github.com/Marseek/tfs-go-hw/course/pkg/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestWriteOrderToDb(t *testing.T) {
	t.Skip("Born to fail")
	logger := log.New()

	dsn := "postgres://jlexie:passwd@localhost:5442/fintech" +
		"?sslmode=disable"
	pool, err := pkgpostgres.NewPool(dsn, logger)
	if err != nil {
		t.Error(err)
	}
	defer pool.Close()

	r := &Repo{pool: pool}

	_ = r.WriteOrderToDb(context.Background(), "TEST_QUERY", 12, "SIDE", 1000, "open", 500, 100)
	total1, err1 := r.GetTotalProfitDb(context.Background())
	_ = r.WriteOrderToDb(context.Background(), "TEST_QUERY", 12, "SIDE", 1000, "open", 500, 100)
	total2, err2 := r.GetTotalProfitDb(context.Background())
	if err1 != nil || err2 != nil {
		t.Errorf("Error, while writing to db: %v, %v", err2, err1)
	}

	const deleteTestData = `DELETE FROM orders WHERE instrument = 'TEST_QUERY'`
	_, _ = r.pool.Exec(context.Background(), deleteTestData)

	assert.NoError(t, err1)
	assert.Equal(t, float32(500), total2-total1)
}

func TestGetTotalProfitDb(t *testing.T) {
	t.Skip("Born to fail")
	logger := log.New()

	dsn := "postgres://jlexie:passwd@localhost:5442/fintech" +
		"?sslmode=disable"
	pool, err := pkgpostgres.NewPool(dsn, logger)
	if err != nil {
		t.Error(err)
	}
	defer pool.Close()

	r := &Repo{pool: pool}
	params := domain.Options{Start: 10, Ticker: "TEST_QUERY", Size: 1, Profit: 0.1, Side: "buy"}
	err1 := r.WriteOrderToDb(context.Background(), params.Ticker, params.Size, params.Side, 1000, "open", 0, params.Profit)
	err2 := r.WriteOrderToDb(context.Background(), params.Ticker, params.Size, params.Side, 1000, "open", 0, params.Profit)
	if err1 != nil || err2 != nil {
		t.Errorf("Error, while writing to db: %v, %v", err2, err1)
	}

	const checkIfWriteSuccess = `SELECT COUNT(profit) FROM orders WHERE instrument = 'TEST_QUERY'`
	row := r.pool.QueryRow(context.Background(), checkIfWriteSuccess)

	var res int
	err = row.Scan(&res)

	const deleteTestData = `DELETE FROM orders WHERE instrument = 'TEST_QUERY'`
	_, _ = r.pool.Exec(context.Background(), deleteTestData)

	assert.NoError(t, err)
	assert.Equal(t, 2, res)
}
