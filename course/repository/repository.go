package repository

import (
	"context"
	"course/domain"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type repo struct {
	pool   *pgxpool.Pool
	logger logrus.FieldLogger
}

func NewRepository(pgxPool *pgxpool.Pool, logger logrus.FieldLogger) Repository {
	return &repo{
		pool:   pgxPool,
		logger: logger,
	}
}

type Repository interface {
	SendOrder(symbol, side string, size int) (domain.APIResp, error)
	SetWSConnection(tick string) (chan domain.WsResponse, func(), error)
	GetTotalProfitDb(ctx context.Context) (float32, error)
	WriteOrderToDb(ctx context.Context, inst string, size int, side string, price float32, ordtype string, profit float32, stoploss float32) error
	WriteToTelegramBot(text string)
	GetUsersMap() map[string]string
}
