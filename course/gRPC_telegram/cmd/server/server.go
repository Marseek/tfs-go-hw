package main

import (
	"context"
	"fmt"
	"gRPC/datapb"
	"google.golang.org/grpc"
	"log"
	"net"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type server struct {
	datapb.UnsafeMessageServiceServer
	bot *tgbotapi.BotAPI
}

func (s *server) SendMessage(ctx context.Context, req *datapb.Request) (*datapb.Response, error) {
	fmt.Println(req.Req)
	msg1 := tgbotapi.NewMessage(1689529148, req.Req)
	_, err := s.bot.Send(msg1)
	if err != nil {
		log.Println("Can't send message to telegram: ", err)
	}
	return &datapb.Response{Resp: 1}, nil
}


func main() {
	listen, err := net.Listen("tcp", ":5005")
	if err != nil {
		log.Fatalf("can't listen on port: %v", err)
	}
	s := grpc.NewServer()

	bot, err := tgbotapi.NewBotAPI("2146099871:AAF3T4RRFw6UhlG07i4e31O7grLwvuXXLH4")
	if err != nil {
		log.Fatalln("Can't connect to telegram: ", err)
		return
	}
	serv := server{bot: bot}

	datapb.RegisterMessageServiceServer(s, &serv)
	if err := s.Serve(listen); err != nil {
		log.Fatalf("can't register service server: %v", err)
	}
}
