package repository

import (
	"context"
	"fmt"
)

func (r *repo) WriteOrderToDb(ctx context.Context, inst string, size int, side string, price float32, ordtype string, profit float32, stoploss float32) error {
	Query := fmt.Sprintf(`INSERT INTO orders (instrument, size, side, price, ts, type, profit, stop_loss) VALUES ('%s', %d, '%s', %f, now(), '%s', %f, %f)`, inst, size, side, price, ordtype, profit, stoploss)
	_, err := r.pool.Exec(ctx, Query)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) GetTotalProfitDb(ctx context.Context) (float32, error) {
	const selectCandlesQuery = `SELECT SUM(profit) FROM orders`
	row := r.pool.QueryRow(ctx, selectCandlesQuery)

	var res float32
	err := row.Scan(&res)
	if err != nil {
		return 0, err
	}

	return res, nil
}
