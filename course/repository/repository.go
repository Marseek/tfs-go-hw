package repository

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/Marseek/tfs-go-hw/course/domain"
	"github.com/Marseek/tfs-go-hw/course/pkg/telegrampb"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type Repo struct {
	pool       *pgxpool.Pool
	logger     logrus.FieldLogger
	httpClient http.Client
	secrets    map[string]string
	tgClient   telegrampb.MessageServiceClient
}

func NewRepository(pgxPool *pgxpool.Pool, logger logrus.FieldLogger) Repository {
	var publicAPIKey = flag.String("public", "", "public key from Kraken")
	var privatAPIKey = flag.String("privat", "", "Privat key from Kraken")
	flag.Parse()
	if *publicAPIKey == "" || *privatAPIKey == "" {
		logger.Fatalln("You should pass to command line args public and privat key's from Kraken")
	}
	sec := map[string]string{"public": *publicAPIKey, "privat": *privatAPIKey}
	tgclient, err := telegrampb.SetTelegramClient("localhost:5005")
	if err != nil {
		logger.Fatalln(err)
	}
	return &Repo{
		pool:   pgxPool,
		logger: logger,
		httpClient: http.Client{
			Timeout: time.Second * 5,
		},
		secrets:  sec,
		tgClient: tgclient,
	}
}

type Repository interface {
	SendOrder(symbol, side string, size int, addr string) (domain.APIResp, error)
	SetWSConnection(addr string, tick string) (chan domain.WsResponse, func(), error)
	GetTotalProfitDb(ctx context.Context) (float32, error)
	WriteOrderToDb(ctx context.Context, inst string, size int, side string, price float32, ordtype string, profit float32, stoploss float32) error
	WriteToTelegramBot(text string)
	GetUsersMap(string) map[string]string
}
