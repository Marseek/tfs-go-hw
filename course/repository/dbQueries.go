package repository

import (
	"context"
)

func (r *Repo) WriteOrderToDb(ctx context.Context, inst string, size int, side string, price float32, ordtype string, profit float32, stoploss float32) error { // ($1, $2)
	_, err := r.pool.Exec(ctx, `INSERT INTO orders (instrument, size, side, price, ts, type, profit, stop_loss) VALUES ($1, $2, $3, $4, now(), $5, $6, $7)`, inst, size, side, price, ordtype, profit, stoploss)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) GetTotalProfitDb(ctx context.Context) (float32, error) {
	const selectCandlesQuery = `SELECT SUM(profit) FROM orders`
	row := r.pool.QueryRow(ctx, selectCandlesQuery)

	var res float32
	err := row.Scan(&res)
	if err != nil {
		return 0, err
	}

	return res, nil
}
