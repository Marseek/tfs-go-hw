package telegrampb

import (
	"errors"

	"google.golang.org/grpc"
)

func SetTelegramClient(addr string) (MessageServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.New("can't connect to gRPC_telegram server: " + err.Error())
	}
	return NewMessageServiceClient(conn), nil
}
